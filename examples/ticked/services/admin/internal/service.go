package internal

import (
	"context"
	"fmt"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/preflight"
	"github.com/go-chi/chi/v5"
)

type Service struct {
	cfg    *config.Config
	log    log.Logger

	authnClient *httpclient.Client
	authzClient *httpclient.Client

	handler *Handler
}

func New(cfg *config.Config, logger log.Logger) (*Service, error) {
	s := &Service{
		cfg: cfg,
		log: logger,
	}

	authnURL := cfg.GetStringOrDef("services.authn.url", "http://localhost:8082")
	authzURL := cfg.GetStringOrDef("services.authz.url", "http://localhost:8083")

	s.authnClient = httpclient.New(authnURL, logger)
	s.authzClient = httpclient.New(authzURL, logger)

	s.handler = NewHandler(s.authnClient, s.authzClient, cfg, logger)

	return s, nil
}

func (s *Service) Start(ctx context.Context) error {
	if s.cfg.GetBoolOrDef("preflight.enabled", true) {
		authnURL := s.cfg.GetStringOrDef("services.authn.url", "http://localhost:8082")
		authzURL := s.cfg.GetStringOrDef("services.authz.url", "http://localhost:8083")

		checker := preflight.New(s.log)
		checker.Add(preflight.HTTPCheck("authn", authnURL+"/health"))
		checker.Add(preflight.HTTPCheck("authz", authzURL+"/health"))

		if err := checker.RunAll(ctx); err != nil {
			return fmt.Errorf("preflight checks failed: %w", err)
		}
	}

	s.log.Infof("Admin service started successfully")
	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	s.log.Infof("Admin service stopped")
	return nil
}

func (s *Service) RegisterRoutes(r chi.Router) {
	s.handler.RegisterRoutes(r)
}
