package internal

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/aquamarinepk/aqm/config"
	aqmlog "github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/examples/ticked/services/ticked/internal/list"
	"github.com/aquamarinepk/aqm/examples/ticked/services/ticked/internal/list/fake"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

// Service coordinates the ticked service components and manages lifecycle.
type Service struct {
	cfg *config.Config
	log aqmlog.Logger
	db  *sql.DB

	listHandler *list.Handler
}

// New creates a new Service with the given configuration.
func New(cfg *config.Config, log aqmlog.Logger) (*Service, error) {
	s := &Service{
		cfg: cfg,
		log: log,
	}

	var repo list.Repo

	// Initialize repository based on driver
	if cfg.Database.Driver == "postgres" {
		connStr := cfg.Database.ConnectionString()
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}

		s.db = db
		repo = list.NewPostgresRepo(db)
	} else {
		repo = &fake.Repo{}
	}

	// Initialize service and handler
	listService := list.NewService(repo, cfg, log)
	s.listHandler = list.NewHandler(listService, cfg, log)

	return s, nil
}

// Start initializes the service.
func (s *Service) Start(ctx context.Context) error {
	s.log.Info("Service started successfully")
	return nil
}

// RegisterRoutes registers all HTTP routes for the service.
func (s *Service) RegisterRoutes(r chi.Router) {
	s.listHandler.RegisterRoutes(r)
}

// Stop gracefully shuts down the service and closes database connections.
func (s *Service) Stop(ctx context.Context) error {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("database close error: %w", err)
		}
	}

	s.log.Info("Service stopped successfully")
	return nil
}
