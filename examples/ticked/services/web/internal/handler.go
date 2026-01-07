package internal

import (
	"context"
	"net/http"
	"time"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/web"
	"github.com/aquamarinepk/aqm/web/htmx"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler wires HTTP routes for the web interface.
type Handler struct {
	todoStore    TodoListStore
	sessionStore *SessionStore
	tmplMgr      *web.TemplateManager
	cfg          *config.Config
	log          log.Logger
}

// NewHandler creates a new handler instance.
func NewHandler(todoStore TodoListStore, sessionStore *SessionStore, tmplMgr *web.TemplateManager, cfg *config.Config, log log.Logger) *Handler {
	return &Handler{
		todoStore:    todoStore,
		sessionStore: sessionStore,
		tmplMgr:      tmplMgr,
		cfg:          cfg,
		log:          log,
	}
}

// RegisterRoutes registers all HTTP routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	fileServer := http.FileServer(http.Dir("assets/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/signin", h.ShowSignIn)
	r.Post("/signin", h.HandleSignIn)

	r.Group(func(r chi.Router) {
		r.Use(h.SessionMiddleware)

		r.Get("/", h.HandleList)
		r.Get("/list", h.HandleList)
		r.Post("/signout", h.HandleSignOut)

		r.Post("/list/items", h.HandleAddItem)
		r.Post("/list/items/{itemID}/toggle", h.HandleToggleItem)
		r.Delete("/list/items/{itemID}", h.HandleRemoveItem)
	})
}

// SessionMiddleware validates cookie and injects user into context.
func (h *Handler) SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionName := h.cfg.GetStringOrDef("auth.session.name", "ticked_session")
		cookie, err := r.Cookie(sessionName)
		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		session, err := h.sessionStore.Get(cookie.Value)
		if err != nil {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		ctx := context.WithValue(r.Context(), "session", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetSession extracts session from context.
func GetSession(r *http.Request) *Session {
	if session := r.Context().Value("session"); session != nil {
		return session.(*Session)
	}
	return nil
}

// ShowSignIn displays the signin form.
func (h *Handler) ShowSignIn(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Sign In - Ticked",
	}
	h.tmplMgr.Render(w, "auth", "signin", data)
}

// Demo user credentials (until seeding and authn integration)
const (
	demoEmail    = "john.doe@localhost"
	demoPassword = "johndoe"
)

var demoUserID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// HandleSignIn processes signin requests.
// NOTE: WIP - Temporary signin until proper authn integration
// Currently accepts john.doe@localhost / johndoe
// TODO: Integrate with authn service /auth/signin endpoint
func (h *Handler) HandleSignIn(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.renderError(w, "auth", "signin", "Failed to parse form. Please try again.")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		h.renderError(w, "auth", "signin", "Email and password are required.")
		return
	}

	// Demo validation (until authn integration)
	if email != demoEmail || password != demoPassword {
		h.renderError(w, "auth", "signin", "Invalid credentials. Try: john.doe@localhost / johndoe")
		return
	}

	userID := demoUserID

	session := &Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(h.sessionStore.ttl),
	}

	if err := h.sessionStore.Save(session); err != nil {
		h.log.Errorf("Failed to save session: %v", err)
		h.renderError(w, "auth", "signin", "Session error. Please try again.")
		return
	}

	sessionName := h.cfg.GetStringOrDef("auth.session.name", "ticked_session")
	http.SetCookie(w, &http.Cookie{
		Name:     sessionName,
		Value:    session.ID,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(h.sessionStore.ttl.Seconds()),
	})

	w.Header().Set("HX-Redirect", "/list")
	w.WriteHeader(http.StatusOK)
}

// HandleSignOut processes sign-out requests.
func (h *Handler) HandleSignOut(w http.ResponseWriter, r *http.Request) {
	sessionName := h.cfg.GetStringOrDef("auth.session.name", "ticked_session")
	cookie, err := r.Cookie(sessionName)
	if err == nil && cookie.Value != "" {
		h.sessionStore.Delete(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	w.Header().Set("HX-Redirect", "/signin")
	w.WriteHeader(http.StatusOK)
}

// HandleList displays the todo list.
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)

	list, err := h.todoStore.Get(r.Context(), session.UserID)
	if err != nil {
		list = &TodoList{Items: []TodoItem{}}
	}

	data := map[string]interface{}{
		"Title":     "My Todo List",
		"Items":     list.Items,
		"UserID":    session.UserID,
		"UserEmail": session.Email,
	}

	h.tmplMgr.Render(w, "todos", "list", data)
}

// HandleAddItem adds a new item (htmx endpoint).
func (h *Handler) HandleAddItem(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)
	text := r.FormValue("text")

	list, err := h.todoStore.AddItem(r.Context(), session.UserID, text)
	if err != nil {
		h.log.Errorf("Failed to add item: %v", err)
		http.Error(w, "Failed to add item", http.StatusInternalServerError)
		return
	}

	if len(list.Items) > 0 {
		newItem := list.Items[len(list.Items)-1]
		h.tmplMgr.RenderPartial(w, "todos", "item", newItem)
	}
}

// HandleToggleItem toggles completion status (htmx endpoint).
func (h *Handler) HandleToggleItem(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)
	itemID, err := web.ParseIDParam(r, "itemID")
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	list, err := h.todoStore.Get(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	var currentCompleted bool
	found := false
	for _, item := range list.Items {
		if item.ID == itemID {
			currentCompleted = item.Completed
			found = true
			break
		}
	}

	if !found {
		http.Error(w, "Item not found", http.StatusNotFound)
		return
	}

	newCompleted := !currentCompleted
	list, err = h.todoStore.UpdateItem(r.Context(), session.UserID, itemID, nil, &newCompleted)
	if err != nil {
		h.log.Errorf("Failed to update item: %v", err)
		http.Error(w, "Failed to update item", http.StatusInternalServerError)
		return
	}

	for _, item := range list.Items {
		if item.ID == itemID {
			h.tmplMgr.RenderPartial(w, "todos", "item", item)
			return
		}
	}
}

// HandleRemoveItem removes an item (htmx endpoint).
func (h *Handler) HandleRemoveItem(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)
	itemID, err := web.ParseIDParam(r, "itemID")
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	_, err = h.todoStore.RemoveItem(r.Context(), session.UserID, itemID)
	htmx.RespondDelete(w, err, h.log)
}

// renderError renders an error message on a template.
func (h *Handler) renderError(w http.ResponseWriter, namespace, template, message string) {
	data := map[string]interface{}{
		"Title": "Sign In - Ticked",
		"Error": message,
	}
	h.tmplMgr.Render(w, namespace, template, data)
}
