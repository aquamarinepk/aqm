package internal

import (
	"net/http"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
	"github.com/go-chi/chi/v5"
)

type Handler struct {
	authnClient *httpclient.Client
	authzClient *httpclient.Client
	cfg         *config.Config
	log         log.Logger
}

func NewHandler(authnClient, authzClient *httpclient.Client, cfg *config.Config, logger log.Logger) *Handler {
	return &Handler{
		authnClient: authnClient,
		authzClient: authzClient,
		cfg:         cfg,
		log:         logger,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/", h.handleIndex)

	r.Route("/admin", func(r chi.Router) {
		r.Get("/", h.handleDashboard)
	})
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Admin</title></head>
<body>
<h1>Admin Service</h1>
<p><a href="/admin">Dashboard</a></p>
</body>
</html>`))
}

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<!DOCTYPE html>
<html>
<head><title>Admin Dashboard</title></head>
<body>
<h1>Dashboard</h1>
<p>Admin interface - coming soon</p>
</body>
</html>`))
}
