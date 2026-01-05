package internal

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Handler struct {
	authnClient *AuthNClient
	authzClient *AuthZClient
	cfg         *config.Config
	log         log.Logger
	templates   *template.Template
}

func NewHandler(authnClient *AuthNClient, authzClient *AuthZClient, cfg *config.Config, logger log.Logger) *Handler {
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		tmpl, err = template.ParseGlob("internal/templates/*.html")
		if err != nil {
			logger.Errorf("Failed to load templates: %v", err)
			tmpl = template.New("empty")
		}
	}

	return &Handler{
		authnClient: authnClient,
		authzClient: authzClient,
		cfg:         cfg,
		log:         logger,
		templates:   tmpl,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	fileServer := http.FileServer(http.Dir("internal/static"))
	r.Handle("/static/*", http.StripPrefix("/static/", fileServer))

	r.Get("/", h.handleIndex)

	r.Route("/admin", func(r chi.Router) {
		r.Get("/", h.handleDashboard)
		r.Get("/list-users", h.handleListUsers)
		r.Get("/get-user", h.handleGetUser)
	})
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{
		"Title": "Dashboard",
	}

	if err := h.templates.ExecuteTemplate(w, "dashboard.html", data); err != nil {
		h.log.Errorf("Failed to render dashboard: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	users, err := h.authnClient.ListUsers(ctx)
	if err != nil {
		h.log.Errorf("Failed to list users: %v", err)
		http.Error(w, "Failed to retrieve users", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Title": "Users",
		"Users": users,
	}

	if err := h.templates.ExecuteTemplate(w, "users-list.html", data); err != nil {
		h.log.Errorf("Failed to render users list: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}

func (h *Handler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userIDStr := r.URL.Query().Get("id")
	if userIDStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	user, err := h.authnClient.GetUser(ctx, userID)
	if err != nil {
		h.log.Errorf("Failed to get user: %v", err)
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	roles, err := h.authzClient.GetUserRoles(ctx, userID)
	if err != nil {
		h.log.Errorf("Failed to get user roles: %v", err)
		roles = []*Role{}
	}

	data := map[string]interface{}{
		"Title": fmt.Sprintf("User: %s", user.Username),
		"User":  user,
		"Roles": roles,
	}

	if err := h.templates.ExecuteTemplate(w, "user-detail.html", data); err != nil {
		h.log.Errorf("Failed to render user detail: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
