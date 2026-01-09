package internal

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/examples/ticked/services/ticked/internal/list"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/migrate"
	"github.com/aquamarinepk/aqm/pubsub/nats"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

// Service coordinates the ticked service components and manages lifecycle.
type Service struct {
	cfg    *config.Config
	log    log.Logger
	db     *sql.DB
	broker *nats.Broker

	listHandler *list.Handler
}

// New creates a new Service with the given configuration.
func New(cfg *config.Config, migrationsFS embed.FS, logger log.Logger) (*Service, error) {
	s := &Service{
		cfg: cfg,
		log: logger,
	}

	var store list.TodoListStore

	// Initialize store based on driver
	if cfg.Database.Driver == "postgres" {
		connStr := cfg.Database.ConnectionString()
		db, err := sql.Open("postgres", connStr)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}

		if err := db.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}

		// Run migrations
		migrator := migrate.New(migrationsFS, "postgres", logger)
		migrator.SetDB(db)
		migrator.SetPath("db/migrations")

		if err := migrator.Run(context.Background()); err != nil {
			db.Close()
			return nil, fmt.Errorf("migration failed: %w", err)
		}

		s.db = db
		store = list.NewPostgresStore(db)
	} else {
		store = list.NewMemStore()
	}

	// Initialize NATS broker if configured using static config
	if cfg.NATS.URL != "" {
		natsCfg := nats.Config{
			URL:            cfg.NATS.URL,
			MaxReconnect:   cfg.NATS.MaxReconnect,
			ReconnectWait:  nats.DefaultConfig().ReconnectWait,
			ConnectTimeout: nats.DefaultConfig().ConnectTimeout,
		}
		if natsCfg.MaxReconnect == 0 {
			natsCfg.MaxReconnect = nats.DefaultConfig().MaxReconnect
		}
		s.broker = nats.NewBroker(natsCfg, logger)
	}

	// Initialize service and handler
	listService := list.NewService(store, s.broker, cfg, logger)
	s.listHandler = list.NewHandler(listService, cfg, logger)

	return s, nil
}

// Start initializes the service.
func (s *Service) Start(ctx context.Context) error {
	// Start NATS broker if configured
	if s.broker != nil {
		if err := s.broker.Start(ctx); err != nil {
			return fmt.Errorf("cannot start NATS broker: %w", err)
		}
		s.log.Info("NATS broker started")
	}

	s.log.Info("Service started successfully")
	return nil
}

// RegisterRoutes registers all HTTP routes for the service.
func (s *Service) RegisterRoutes(r chi.Router) {
	s.listHandler.RegisterRoutes(r)
}

// Stop gracefully shuts down the service and closes database connections.
func (s *Service) Stop(ctx context.Context) error {
	if s.broker != nil {
		if err := s.broker.Stop(ctx); err != nil {
			s.log.Errorf("Error stopping NATS broker: %v", err)
		}
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("database close error: %w", err)
		}
	}

	s.log.Info("Service stopped successfully")
	return nil
}
