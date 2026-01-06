package internal

import (
	"context"
	"fmt"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/preflight"
	"github.com/go-chi/chi/v5"
)

// Service coordinates the tickedweb service components and manages lifecycle.
type Service struct {
	cfg          *config.Config
	log          log.Logger
	todoStore    TodoListStore
	sessionStore *SessionStore
	handler      *Handler
}

// New creates a new Service with the given configuration.
func New(cfg *config.Config, log log.Logger) (*Service, error) {
	s := &Service{
		cfg: cfg,
		log: log,
	}

	// Create HTTP-backed todo store
	tickedURL := cfg.GetStringOrDef("services.ticked.url", "http://localhost:8084")
	s.todoStore = NewHTTPTodoListStore(tickedURL, log)

	// Create session store
	sessionTTL := cfg.GetDurationOrDef("auth.session.ttl", 24*3600*1000000000) // 24 hours in nanoseconds
	s.sessionStore = NewSessionStore(sessionTTL)

	// Create handler
	s.handler = NewHandler(s.todoStore, s.sessionStore, cfg, log)

	return s, nil
}

// Start initializes the service and runs preflight checks.
func (s *Service) Start(ctx context.Context) error {
	// Preflight checks
	if s.cfg.GetBoolOrDef("preflight.enabled", true) {
		checker := preflight.New(s.log)

		// Check ticked API availability
		tickedURL := s.cfg.GetStringOrDef("services.ticked.url", "http://localhost:8084")
		checker.Add(preflight.HTTPCheck("ticked-api", tickedURL+"/health"))

		if err := checker.RunAll(ctx); err != nil {
			return fmt.Errorf("preflight checks failed: %w", err)
		}
	}

	s.log.Info("Service started successfully")
	return nil
}

// Stop gracefully shuts down the service.
func (s *Service) Stop(ctx context.Context) error {
	s.log.Info("Service stopped successfully")
	return nil
}

// RegisterRoutes registers all HTTP routes for the service.
func (s *Service) RegisterRoutes(r chi.Router) {
	s.handler.RegisterRoutes(r)
}
