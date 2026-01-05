package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/middleware"
)

// RouterOption configures optional features for the main router.
type RouterOption func(chi.Router) error

// NewRouter creates a new chi router with the provided options applied.
func NewRouter(logger log.Logger, opts ...RouterOption) chi.Router {
	r := chi.NewRouter()

	for _, opt := range opts {
		if err := opt(r); err != nil {
			logger.Error("Cannot apply router option", "error", err)
		}
	}

	return r
}

// WithDebugRoutes enables GET /debug/routes endpoint that lists all registered routes.
func WithDebugRoutes() RouterOption {
	return func(r chi.Router) error {
		r.Get("/debug/routes", handleDebugRoutes)
		return nil
	}
}

// WithPing enables GET /ping health check endpoint.
func WithPing() RouterOption {
	return func(r chi.Router) error {
		r.Get("/ping", handlePing)
		return nil
	}
}

// WithHealthChecks enables GET /health endpoint with service information.
func WithHealthChecks(name, version string) RouterOption {
	return func(r chi.Router) error {
		r.Get("/health", handleHealthCheck(name, version))
		return nil
	}
}

// WithDefaultMiddlewares applies the default middleware stack (RequestID, RealIP, Logger, Recoverer).
func WithDefaultMiddlewares() RouterOption {
	return func(r chi.Router) error {
		r.Use(middleware.DefaultStack()...)
		return nil
	}
}

// WithDefaultInternalMiddlewares applies the default middleware stack plus InternalOnly restriction.
func WithDefaultInternalMiddlewares() RouterOption {
	return func(r chi.Router) error {
		r.Use(middleware.DefaultInternal()...)
		return nil
	}
}

// ApplyRouterOptions applies all router options.
func ApplyRouterOptions(r chi.Router, opts ...RouterOption) error {
	for _, opt := range opts {
		if err := opt(r); err != nil {
			return err
		}
	}
	return nil
}

func handlePing(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func handleHealthCheck(name, version string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health := map[string]string{
			"status":  "ok",
			"service": name,
			"version": version,
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(health)
	}
}

func handleDebugRoutes(w http.ResponseWriter, r *http.Request) {
	router := chi.RouteContext(r.Context()).Routes

	var routes []string
	walkFunc := func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error {
		// Format: METHOD /path
		routes = append(routes, fmt.Sprintf("%-6s %s", method, route))
		return nil
	}

	if err := chi.Walk(router, walkFunc); err != nil {
		http.Error(w, "Cannot walk routes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Registered Routes:\n\n"))
	w.Write([]byte(strings.Join(routes, "\n")))
}
