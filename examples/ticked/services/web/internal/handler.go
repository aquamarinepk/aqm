package internal

import (
	"context"
	"html/template"
	"net/http"
	"time"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// Handler wires HTTP routes for the web interface.
type Handler struct {
	todoStore    TodoListStore
	sessionStore *SessionStore
	cfg          *config.Config
	log          log.Logger
	templates    *template.Template
}

// NewHandler creates a new handler instance.
func NewHandler(todoStore TodoListStore, sessionStore *SessionStore, cfg *config.Config, log log.Logger) *Handler {
	// Load templates with fallback paths
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		tmpl, err = template.ParseGlob("internal/templates/*.html")
		if err != nil {
			log.Errorf("Failed to load templates: %v", err)
			tmpl = template.New("empty")
		}
	}

	// Load partials
	if partials, err := template.ParseGlob("templates/partials/*.html"); err == nil {
		for _, t := range partials.Templates() {
			tmpl.AddParseTree(t.Name(), t.Tree)
		}
	} else if partials, err := template.ParseGlob("internal/templates/partials/*.html"); err == nil {
		for _, t := range partials.Templates() {
			tmpl.AddParseTree(t.Name(), t.Tree)
		}
	}

	return &Handler{
		todoStore:    todoStore,
		sessionStore: sessionStore,
		cfg:          cfg,
		log:          log,
		templates:    tmpl,
	}
}

// RegisterRoutes registers all HTTP routes.
func (h *Handler) RegisterRoutes(r chi.Router) {
	// Static files
	fileServer := http.FileServer(http.Dir("internal/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	// Public routes
	r.Get("/signin", h.ShowSignIn)
	r.Post("/signin", h.HandleSignIn)

	// Protected routes (with session middleware)
	r.Group(func(r chi.Router) {
		r.Use(h.SessionMiddleware)

		r.Get("/", h.HandleList)
		r.Get("/list", h.HandleList)
		r.Post("/signout", h.HandleSignOut)

		// htmx endpoints (return HTML fragments)
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

		// Inject session into context
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
	h.renderTemplate(w, "signin.html", data)
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
		h.renderError(w, "signin.html", "Failed to parse form. Please try again.")
		return
	}

	email := r.FormValue("email")
	password := r.FormValue("password")

	if email == "" || password == "" {
		h.renderError(w, "signin.html", "Email and password are required.")
		return
	}

	// Demo validation (until authn integration)
	if email != demoEmail || password != demoPassword {
		h.renderError(w, "signin.html", "Invalid credentials. Try: john.doe@localhost / johndoe")
		return
	}

	userID := demoUserID

	// Create session
	session := &Session{
		ID:        uuid.New().String(),
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(h.sessionStore.ttl),
	}

	if err := h.sessionStore.Save(session); err != nil {
		h.log.Errorf("Failed to save session: %v", err)
		h.renderError(w, "signin.html", "Session error. Please try again.")
		return
	}

	// Set cookie
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

	// Redirect with HX-Redirect for htmx
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

	// Clear cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Redirect to signin
	w.Header().Set("HX-Redirect", "/signin")
	w.WriteHeader(http.StatusOK)
}

// HandleList displays the todo list.
func (h *Handler) HandleList(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)

	list, err := h.todoStore.Get(r.Context(), session.UserID)
	if err != nil {
		// List not found - show empty state
		list = &TodoList{Items: []TodoItem{}}
	}

	data := map[string]interface{}{
		"Title":     "My Todo List",
		"Items":     list.Items,
		"UserID":    session.UserID,
		"UserEmail": session.Email,
	}

	h.renderTemplate(w, "list.html", data)
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

	// Return just the new item HTML
	if len(list.Items) > 0 {
		newItem := list.Items[len(list.Items)-1]
		if err := h.templates.ExecuteTemplate(w, "item.html", newItem); err != nil {
			h.log.Errorf("Failed to render item template: %v", err)
		}
	}
}

// HandleToggleItem toggles completion status (htmx endpoint).
func (h *Handler) HandleToggleItem(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)
	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	// Get current state
	list, err := h.todoStore.Get(r.Context(), session.UserID)
	if err != nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}

	// Find item and toggle
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

	// Find updated item and return HTML
	for _, item := range list.Items {
		if item.ID == itemID {
			if err := h.templates.ExecuteTemplate(w, "item.html", item); err != nil {
				h.log.Errorf("Failed to render item template: %v", err)
			}
			return
		}
	}
}

// HandleRemoveItem removes an item (htmx endpoint).
func (h *Handler) HandleRemoveItem(w http.ResponseWriter, r *http.Request) {
	session := GetSession(r)
	itemIDStr := chi.URLParam(r, "itemID")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		http.Error(w, "Invalid item ID", http.StatusBadRequest)
		return
	}

	_, err = h.todoStore.RemoveItem(r.Context(), session.UserID, itemID)
	if err != nil {
		h.log.Errorf("Failed to remove item: %v", err)
		http.Error(w, "Failed to remove item", http.StatusInternalServerError)
		return
	}

	// Return empty content - htmx will replace element with nothing (fade out with swap:1s)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(""))
}

// renderTemplate renders a template with the given data.
func (h *Handler) renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		h.log.Errorf("Failed to render template %s: %v", name, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

// renderError renders an error message on the signin template.
func (h *Handler) renderError(w http.ResponseWriter, template string, message string) {
	data := map[string]interface{}{
		"Title": "Sign In - Ticked",
		"Error": message,
	}
	h.renderTemplate(w, template, data)
}
