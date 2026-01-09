package audit

import (
	"context"
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/migrate"
	"github.com/aquamarinepk/aqm/pubsub"
	"github.com/aquamarinepk/aqm/pubsub/nats"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

const (
	topic        = "audit.todo"
	subscriberID = "audit-persistence"
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
func New(migrationsFS embed.FS, cfg *config.Config, logger log.Logger) (*Service, error) {
	s := &Service{
		cfg: cfg,
		log: logger,
	}

	if cfg.Database.Driver != "postgres" {
		return nil, fmt.Errorf("audit service requires postgres database")
	}

	connStr := cfg.Database.ConnectionString()
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	migrator := migrate.New(migrationsFS, "postgres", logger)
	migrator.SetDB(db)
	migrator.SetPath("assets/migration/postgres")

	if err := migrator.Run(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	s.db = db
	s.store = NewPostgresStore(db)

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

	return s, nil
}

// Start initializes the service and subscribes to audit events.
func (s *Service) Start(ctx context.Context) error {
	if s.broker != nil {
		if err := s.broker.Start(ctx); err != nil {
			return fmt.Errorf("cannot start NATS broker: %w", err)
		}
		s.log.Info("NATS broker started")

		if err := s.broker.Subscribe(ctx, topic, s.handleEvent, pubsub.SubscribeOptions{
			SubscriberID: subscriberID,
		}); err != nil {
			return fmt.Errorf("cannot subscribe to %s: %w", topic, err)
		}
		s.log.Infof("Subscribed to %s", topic)
	}

	s.log.Info("Service started successfully")
	return nil
}

// RegisterRoutes registers HTTP routes for the audit service.
func (s *Service) RegisterRoutes(r chi.Router) {
	r.Get("/events", s.handleListEvents)
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

func (s *Service) handleEvent(ctx context.Context, env pubsub.Envelope) error {
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

func (s *Service) handleListEvents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limitStr := r.URL.Query().Get("limit")
	limit := 100
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	events, err := s.store.List(ctx, limit)
	if err != nil {
		s.log.Errorf("Failed to list events: %v", err)
		http.Error(w, "Failed to retrieve events", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"data": events,
	})
}
