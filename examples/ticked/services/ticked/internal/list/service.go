package list

import (
	"context"
	"errors"
	"time"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/pubsub"
	"github.com/google/uuid"
)

const (
	// AuditTopic is the pubsub topic for audit events.
	AuditTopic = "audit.todo"
)

// Service defines the business logic operations for todo lists.
type Service interface {
	GetOrCreateList(ctx context.Context, userID uuid.UUID) (*TodoList, error)
	GetList(ctx context.Context, userID uuid.UUID) (*TodoList, error)
	AddItem(ctx context.Context, userID uuid.UUID, text string) (*TodoList, error)
	UpdateItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, text *string, completed *bool) (*TodoList, error)
	RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) (*TodoList, error)
}

// service contains business logic for todo lists.
type service struct {
	store     TodoListStore
	publisher pubsub.Publisher
	cfg       *config.Config
	log       log.Logger
}

// NewService creates a new service instance.
func NewService(store TodoListStore, publisher pubsub.Publisher, cfg *config.Config, logger log.Logger) Service {
	if logger == nil {
		logger = log.NewNoopLogger()
	}
	return &service{
		store:     store,
		publisher: publisher,
		cfg:       cfg,
		log:       logger,
	}
}

// GetOrCreateList retrieves a user's list or creates it if it doesn't exist.
func (s *service) GetOrCreateList(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	list, err := s.store.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			list = NewTodoList(userID)
			if err := s.store.Save(ctx, list); err != nil {
				return nil, err
			}
			list.SortByCreatedAt()
			return list, nil
		}
		return nil, err
	}
	list.SortByCreatedAt()
	return list, nil
}

// GetList retrieves a user's list.
func (s *service) GetList(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	list, err := s.store.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	list.SortByCreatedAt()
	return list, nil
}

// AddItem adds an item to a user's list.
func (s *service) AddItem(ctx context.Context, userID uuid.UUID, text string) (*TodoList, error) {
	list, err := s.GetOrCreateList(ctx, userID)
	if err != nil {
		return nil, err
	}

	item, err := list.AddItem(text)
	if err != nil {
		return nil, err
	}

	if err := s.store.Save(ctx, list); err != nil {
		return nil, err
	}

	// Publish audit event
	s.publishEvent(ctx, "todo.item.added", userID.String(), item.ItemID.String(), map[string]string{
		"title": text,
	})

	list.SortByCreatedAt()
	return list, nil
}

// UpdateItem updates an item in a user's list.
func (s *service) UpdateItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, text *string, completed *bool) (*TodoList, error) {
	list, err := s.store.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := list.UpdateItem(itemID, text, completed); err != nil {
		return nil, err
	}

	if err := s.store.Save(ctx, list); err != nil {
		return nil, err
	}

	// Publish audit event
	if completed != nil && *completed {
		s.publishEvent(ctx, "todo.item.completed", userID.String(), itemID.String(), nil)
	}

	list.SortByCreatedAt()
	return list, nil
}

// RemoveItem removes an item from a user's list.
func (s *service) RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) (*TodoList, error) {
	list, err := s.store.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := list.RemoveItem(itemID); err != nil {
		return nil, err
	}

	if err := s.store.Save(ctx, list); err != nil {
		return nil, err
	}

	// Publish audit event
	s.publishEvent(ctx, "todo.item.removed", userID.String(), itemID.String(), nil)

	list.SortByCreatedAt()
	return list, nil
}

// publishEvent publishes an audit event via the configured publisher.
func (s *service) publishEvent(ctx context.Context, eventType, userID, itemID string, data map[string]string) {
	if s.publisher == nil {
		return
	}

	payload := map[string]string{
		"event_type": eventType,
		"item_id":    itemID,
	}
	for k, v := range data {
		payload[k] = v
	}

	env := pubsub.Envelope{
		ID:        uuid.New().String(),
		Topic:     AuditTopic,
		Timestamp: time.Now(),
		Payload:   payload,
		Metadata: map[string]string{
			"user_id": userID,
			"source":  "ticked",
		},
	}

	if err := s.publisher.Publish(ctx, AuditTopic, env); err != nil {
		s.log.Errorf("failed to publish audit event: %v", err)
	}
}

