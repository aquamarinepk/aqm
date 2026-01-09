package internal

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/migrate"
	"github.com/aquamarinepk/aqm/pubsub"
	"github.com/aquamarinepk/aqm/pubsub/nats"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

const (
	// Topic is the pubsub topic for audit events.
	Topic = "audit.todo"
	// SubscriberID is the persistent subscriber ID for audit.
	SubscriberID = "audit-persistence"
)

// Service coordinates the audit service components and manages lifecycle.
type Service struct {
	cfg    *config.Config
	log    log.Logger
	db     *sql.DB
	broker *nats.Broker
	store  Store
}

// New creates a new Service with the given configuration.
func New(cfg *config.Config, migrationsFS embed.FS, logger log.Logger) (*Service, error) {
	s := &Service{
		cfg: cfg,
		log: logger,
	}

	// Initialize database
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
		s.store = NewPostgresStore(db)
	} else {
		return nil, fmt.Errorf("audit service requires postgres database")
	}

	// Initialize NATS broker using static config
	natsCfg := nats.Config{
		URL:            cfg.NATS.URL,
		MaxReconnect:   cfg.NATS.MaxReconnect,
		ReconnectWait:  nats.DefaultConfig().ReconnectWait,
		ConnectTimeout: nats.DefaultConfig().ConnectTimeout,
	}
	if natsCfg.URL == "" {
		natsCfg.URL = nats.DefaultConfig().URL
	}
	if natsCfg.MaxReconnect == 0 {
		natsCfg.MaxReconnect = nats.DefaultConfig().MaxReconnect
	}
	s.broker = nats.NewBroker(natsCfg, logger)

	return s, nil
}

// Start initializes the service and subscribes to audit events.
func (s *Service) Start(ctx context.Context) error {
	// Start NATS broker
	if err := s.broker.Start(ctx); err != nil {
		return fmt.Errorf("cannot start NATS broker: %w", err)
	}

	// Subscribe to audit topic
	if err := s.broker.Subscribe(ctx, Topic, s.handleEvent, pubsub.SubscribeOptions{
		SubscriberID: SubscriberID,
	}); err != nil {
		return fmt.Errorf("cannot subscribe to %s: %w", Topic, err)
	}

	s.log.Infof("Subscribed to %s", Topic)
	s.log.Info("Service started successfully")
	return nil
}

// RegisterRoutes registers HTTP routes (none for audit service currently).
func (s *Service) RegisterRoutes(r chi.Router) {
	// No HTTP routes for audit service - it only consumes events
}

// Stop gracefully shuts down the service.
func (s *Service) Stop(ctx context.Context) error {
	if s.broker != nil {
		if err := s.broker.Stop(ctx); err != nil {
			s.log.Errorf("Error stopping broker: %v", err)
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

// handleEvent processes incoming audit events.
func (s *Service) handleEvent(ctx context.Context, env pubsub.Envelope) error {
	// JSON unmarshal produces map[string]interface{}, convert to map[string]string
	rawPayload, ok := env.Payload.(map[string]interface{})
	if !ok {
		s.log.Errorf("invalid payload type: %T", env.Payload)
		return fmt.Errorf("invalid payload type: %T", env.Payload)
	}

	payload := make(map[string]string)
	for k, v := range rawPayload {
		if str, ok := v.(string); ok {
			payload[k] = str
		}
	}

	record := &Record{
		ID:        env.ID,
		EventType: payload["event_type"],
		ItemID:    payload["item_id"],
		UserID:    env.Metadata["user_id"],
		Payload:   payload,
		Source:    env.Metadata["source"],
		CreatedAt: env.Timestamp,
	}

	if err := s.store.Save(ctx, record); err != nil {
		s.log.Errorf("cannot save audit record: %v", err)
		return err
	}

	s.log.Debugf("Persisted audit event %s: %s", env.ID, record.EventType)
	return nil
}
