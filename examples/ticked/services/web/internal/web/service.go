package web

import (
	"context"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/web"
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
func New(tmplMgr *web.TemplateManager, cfg *config.Config, log log.Logger) (*Service, error) {
	s := &Service{
		cfg: cfg,
		log: log,
	}

	tickedURL := cfg.GetStringOrDef("services.ticked.url", "http://localhost:8084")
	s.todoStore = NewHTTPTodoListStore(tickedURL, log)

	authnURL := cfg.GetStringOrDef("services.authn.url", "http://localhost:8080")
	authnClient := NewAuthNClient(httpclient.New(authnURL, log))

	sessionTTL := cfg.GetDurationOrDef("auth.session.ttl", 24*3600*1000000000)
	s.sessionStore = NewSessionStore(sessionTTL)

	s.handler = NewHandler(s.todoStore, authnClient, s.sessionStore, tmplMgr, cfg, log)

	return s, nil
}

// Start initializes the service.
func (s *Service) Start(ctx context.Context) error {
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
