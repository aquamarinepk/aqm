package fake

import (
	"context"
	"sync"

	"github.com/aquamarinepk/aqm/examples/ticked/services/audit/internal"
)

// Store is an in-memory implementation of internal.Store for testing.
type Store struct {
	mu      sync.RWMutex
	records []internal.Record

	// SaveFunc allows overriding Save behavior in tests.
	SaveFunc func(ctx context.Context, record *internal.Record) error
	// ListFunc allows overriding List behavior in tests.
	ListFunc func(ctx context.Context, limit int) ([]internal.Record, error)
}

// NewStore creates a new fake store.
func NewStore() *Store {
	return &Store{
		records: make([]internal.Record, 0),
	}
}

// Save stores a record in memory.
func (s *Store) Save(ctx context.Context, record *internal.Record) error {
	if s.SaveFunc != nil {
		return s.SaveFunc(ctx, record)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = append(s.records, *record)
	return nil
}

// List returns the most recent records.
func (s *Store) List(ctx context.Context, limit int) ([]internal.Record, error) {
	if s.ListFunc != nil {
		return s.ListFunc(ctx, limit)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.records) {
		limit = len(s.records)
	}

	// Return most recent first (reverse order)
	result := make([]internal.Record, limit)
	for i := 0; i < limit; i++ {
		result[i] = s.records[len(s.records)-1-i]
	}

	return result, nil
}

// Records returns all stored records for test assertions.
func (s *Store) Records() []internal.Record {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]internal.Record, len(s.records))
	copy(result, s.records)
	return result
}

// Reset clears all stored records.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.records = make([]internal.Record, 0)
}
