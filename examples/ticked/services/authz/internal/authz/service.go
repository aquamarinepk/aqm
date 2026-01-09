package authz

import (
	"context"
	"database/sql"
	"embed"
	"encoding/hex"
	"fmt"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/handler"
	"github.com/aquamarinepk/aqm/auth/seed"
	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/go-chi/chi/v5"
)

// Service coordinates the authz service components and manages lifecycle.
type Service struct {
	cfg *config.Config
	db  *sql.DB
	log log.Logger

	roleStore  auth.RoleStore
	grantStore auth.GrantStore

	bootstrapService *BootstrapService

	authzHandler *handler.AuthZHandler
}

// New creates a new Service with the given configuration and logger.
// It initializes stores (postgres or fake based on config), bootstrap service, and HTTP handlers.
// Returns an error if initialization fails.
func New(migrationsFS embed.FS, cfg *config.Config, logger log.Logger) (*Service, error) {
	s := &Service{
		cfg: cfg,
		log: logger,
	}

	if cfg.Database.Driver == "postgres" {
		connStr := cfg.Database.ConnectionString()
		roleStore, grantStore, db, err := NewPostgresStores(connStr, migrationsFS, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres stores: %w", err)
		}
		s.db = db
		s.roleStore = roleStore
		s.grantStore = grantStore
	} else {
		roleStore, grantStore := NewFakeStores()
		s.roleStore = roleStore
		s.grantStore = grantStore
	}

	encKeyStr := cfg.GetString("crypto.encryptionkey")
	if encKeyStr == "" {
		return nil, fmt.Errorf("crypto.encryptionkey is required")
	}
	encKey, err := hex.DecodeString(encKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	signKeyStr := cfg.GetString("crypto.signingkey")
	if signKeyStr == "" {
		return nil, fmt.Errorf("crypto.signingkey is required")
	}
	signKey, err := hex.DecodeString(signKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signing key: %w", err)
	}

	seeder := seed.New(
		nil, // userStore not needed for authz
		s.roleStore,
		s.grantStore,
		&seed.Config{
			EncryptionKey: encKey,
			SigningKey:    signKey,
		},
		logger,
	)

	s.bootstrapService = NewBootstrapService(s.roleStore, s.grantStore, seeder, cfg, logger)

	s.authzHandler = handler.NewAuthZHandler(s.roleStore, s.grantStore)

	return s, nil
}

// Start initializes the service and runs bootstrap if enabled.
func (s *Service) Start(ctx context.Context) error {
	if s.cfg.GetBoolOrDef("auth.enablebootstrap", true) {
		if err := s.bootstrapService.Bootstrap(ctx); err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}
	}

	s.log.Infof("AuthZ service started successfully")
	return nil
}

// RegisterRoutes registers all HTTP routes for the service.
func (s *Service) RegisterRoutes(r chi.Router) {
	s.authzHandler.RegisterRoutes(r)
}

// Stop gracefully shuts down the service and closes database connections.
func (s *Service) Stop(ctx context.Context) error {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("database close error: %w", err)
		}
	}

	return nil
}
