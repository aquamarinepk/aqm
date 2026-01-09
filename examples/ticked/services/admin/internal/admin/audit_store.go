package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
)

// AuditEvent represents an audit log entry from the audit service.
type AuditEvent struct {
	ID        string            `json:"id"`
	EventType string            `json:"event_type"`
	ItemID    string            `json:"item_id"`
	UserID    string            `json:"user_id"`
	Payload   map[string]string `json:"payload"`
	Source    string            `json:"source"`
	CreatedAt time.Time         `json:"created_at"`
}

// AuditEventStore abstracts access to audit events.
// This interface allows for swappable implementations (HTTP, Fake for tests).
type AuditEventStore interface {
	List(ctx context.Context, limit int) ([]*AuditEvent, error)
}

// HTTPAuditEventStore implements AuditEventStore using HTTP calls to audit API.
// This wraps httpclient.Client which provides automatic retry with exponential backoff.
type HTTPAuditEventStore struct {
	client *httpclient.Client
	log    log.Logger
}

// NewHTTPAuditEventStore creates a new HTTP-backed audit event store.
func NewHTTPAuditEventStore(auditURL string, logger log.Logger) AuditEventStore {
	client := httpclient.New(
		auditURL,
		logger,
		httpclient.WithRetryMax(3),
		httpclient.WithRetryDelay(100*time.Millisecond),
		httpclient.WithTimeout(30*time.Second),
	)

	return &HTTPAuditEventStore{
		client: client,
		log:    logger,
	}
}

// List retrieves the most recent audit events.
func (s *HTTPAuditEventStore) List(ctx context.Context, limit int) ([]*AuditEvent, error) {
	path := fmt.Sprintf("/events?limit=%d", limit)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("audit API error: status %d", resp.StatusCode)
	}

	var result struct {
		Data []*AuditEvent `json:"data"`
	}

	if err := resp.JSON(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Data, nil
}

// FakeAuditEventStore is a test double for AuditEventStore.
type FakeAuditEventStore struct {
	events []*AuditEvent
}

// NewFakeAuditEventStore creates a new fake store for testing.
func NewFakeAuditEventStore() *FakeAuditEventStore {
	return &FakeAuditEventStore{
		events: make([]*AuditEvent, 0),
	}
}

// List returns stored events.
func (s *FakeAuditEventStore) List(ctx context.Context, limit int) ([]*AuditEvent, error) {
	if limit <= 0 || limit > len(s.events) {
		return s.events, nil
	}
	return s.events[:limit], nil
}

// Add adds an event to the fake store (for test setup).
func (s *FakeAuditEventStore) Add(event *AuditEvent) {
	s.events = append([]*AuditEvent{event}, s.events...)
}
