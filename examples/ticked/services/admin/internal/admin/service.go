package admin

import (
	"context"
	"embed"
	"fmt"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/preflight"
	"github.com/go-chi/chi/v5"
)

type Service struct {
	assetsFS embed.FS
	cfg      *config.Config
	log      log.Logger

	authnClient *AuthNClient
	authzClient *AuthZClient
	auditStore  AuditEventStore

	handler *Handler
}

func New(assetsFS embed.FS, cfg *config.Config, logger log.Logger) (*Service, error) {
	s := &Service{
		assetsFS: assetsFS,
		cfg:      cfg,
		log:      logger,
	}

	authnURL := cfg.GetStringOrDef("services.authn.url", "http://localhost:8082")
	authzURL := cfg.GetStringOrDef("services.authz.url", "http://localhost:8083")
	auditURL := cfg.GetStringOrDef("services.audit.url", "http://localhost:8085")

	httpAuthnClient := httpclient.New(authnURL, logger)
	httpAuthzClient := httpclient.New(authzURL, logger)

	s.authnClient = NewAuthNClient(httpAuthnClient)
	s.authzClient = NewAuthZClient(httpAuthzClient)
	s.auditStore = NewHTTPAuditEventStore(auditURL, logger)

	s.handler = NewHandler(assetsFS, s.authnClient, s.authzClient, s.auditStore, cfg, logger)

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
