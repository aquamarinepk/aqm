package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aquamarinepk/aqm/examples/ticked/services/audit/internal/audit/sqlcgen"
	_ "github.com/lib/pq"
	"github.com/sqlc-dev/pqtype"
)

// Record represents an audit log entry.
type Record struct {
	ID        string            `json:"id"`
	EventType string            `json:"event_type"`
	ItemID    string            `json:"item_id"`
	UserID    string            `json:"user_id"`
	Payload   map[string]string `json:"payload"`
	Source    string            `json:"source"`
	CreatedAt time.Time         `json:"created_at"`
}

// Store defines persistence operations for audit records.
type Store interface {
	Save(ctx context.Context, record *Record) error
	List(ctx context.Context, limit int) ([]Record, error)
}

// PostgresStore implements Store using PostgreSQL with sqlc.
type PostgresStore struct {
	queries *sqlcgen.Queries
}

// NewPostgresStore creates a new PostgreSQL-backed store.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{queries: sqlcgen.New(db)}
}

// Save persists an audit record to the database.
func (s *PostgresStore) Save(ctx context.Context, record *Record) error {
	payloadJSON, err := json.Marshal(record.Payload)
	if err != nil {
		return fmt.Errorf("cannot marshal payload: %w", err)
	}

	params := sqlcgen.InsertAuditRecordParams{
		ID:        record.ID,
		EventType: record.EventType,
		ItemID:    sql.NullString{String: record.ItemID, Valid: record.ItemID != ""},
		UserID:    record.UserID,
		Payload:   pqtype.NullRawMessage{RawMessage: payloadJSON, Valid: true},
		Source:    record.Source,
		CreatedAt: record.CreatedAt,
	}

	if err := s.queries.InsertAuditRecord(ctx, params); err != nil {
		return fmt.Errorf("cannot insert audit record: %w", err)
	}

	return nil
}

// List retrieves the most recent audit records.
func (s *PostgresStore) List(ctx context.Context, limit int) ([]Record, error) {
	rows, err := s.queries.ListAuditRecords(ctx, int32(limit))
	if err != nil {
		return nil, fmt.Errorf("cannot query audit records: %w", err)
	}

	records := make([]Record, 0, len(rows))
	for _, row := range rows {
		var payload map[string]string
		if row.Payload.Valid && len(row.Payload.RawMessage) > 0 {
			if err := json.Unmarshal(row.Payload.RawMessage, &payload); err != nil {
				payload = make(map[string]string)
			}
		}

		records = append(records, Record{
			ID:        row.ID,
			EventType: row.EventType,
			ItemID:    row.ItemID,
			UserID:    row.UserID,
			Payload:   payload,
			Source:    row.Source,
			CreatedAt: row.CreatedAt,
		})
	}

	return records, nil
}

// FakeStore is an in-memory implementation of Store for testing.
type FakeStore struct {
	mu      sync.RWMutex
	records []Record

	SaveFunc func(ctx context.Context, record *Record) error
	ListFunc func(ctx context.Context, limit int) ([]Record, error)
}

// NewFakeStore creates a new in-memory fake store.
func NewFakeStore() *FakeStore {
	return &FakeStore{
		records: make([]Record, 0),
	}
}

// Save stores a record in memory.
func (s *FakeStore) Save(ctx context.Context, record *Record) error {
	if s.SaveFunc != nil {
		return s.SaveFunc(ctx, record)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, *record)
	return nil
}

// List returns the most recent records.
func (s *FakeStore) List(ctx context.Context, limit int) ([]Record, error) {
	if s.ListFunc != nil {
		return s.ListFunc(ctx, limit)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.records) {
		limit = len(s.records)
	}

	result := make([]Record, limit)
	for i := 0; i < limit; i++ {
		result[i] = s.records[len(s.records)-1-i]
	}

	return result, nil
}

// Records returns all stored records for test assertions.
func (s *FakeStore) Records() []Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Record, len(s.records))
	copy(result, s.records)
	return result
}

// Reset clears all stored records.
func (s *FakeStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = make([]Record, 0)
}
