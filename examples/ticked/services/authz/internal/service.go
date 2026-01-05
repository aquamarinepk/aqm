package internal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/handler"
	"github.com/aquamarinepk/aqm/config"
	"github.com/go-chi/chi/v5"
)

// Service coordinates the authz service components and manages lifecycle.
type Service struct {
	cfg *config.Config
	db  *sql.DB

	// Stores
	roleStore  auth.RoleStore
	grantStore auth.GrantStore

	// Handlers
	authzHandler *handler.AuthZHandler
}

// New creates a new Service with the given configuration.
// It initializes stores (postgres or fake based on config) and HTTP handlers.
// Returns an error if initialization fails.
func New(cfg *config.Config) (*Service, error) {
	s := &Service{
		cfg: cfg,
	}

	// Initialize stores based on driver
	if cfg.Database.Driver == "postgres" {
		connStr := cfg.Database.ConnectionString()
		roleStore, grantStore, db, err := NewPostgresStores(connStr)
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

	// Initialize handler
	s.authzHandler = handler.NewAuthZHandler(s.roleStore, s.grantStore)

	return s, nil
}

// Start initializes the service.
func (s *Service) Start(ctx context.Context) error {
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
