package fake

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/examples/ticked/services/audit/internal"
)

func TestNewStore(t *testing.T) {
	store := NewStore()

	if store == nil {
		t.Fatal("NewStore() returned nil")
	}

	if store.records == nil {
		t.Error("records slice not initialized")
	}
}

func TestStoreSave(t *testing.T) {
	tests := []struct {
		name     string
		saveFunc func(ctx context.Context, record *internal.Record) error
		wantErr  bool
	}{
		{
			name:     "success",
			saveFunc: nil,
			wantErr:  false,
		},
		{
			name: "custom save func success",
			saveFunc: func(ctx context.Context, record *internal.Record) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "custom save func error",
			saveFunc: func(ctx context.Context, record *internal.Record) error {
				return errors.New("save error")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()
			store.SaveFunc = tt.saveFunc

			record := &internal.Record{
				ID:        "test-id",
				EventType: "todo.item.added",
				ItemID:    "item-123",
				UserID:    "user-456",
				Payload:   map[string]string{"title": "Test"},
				Source:    "ticked",
				CreatedAt: time.Now(),
			}

			err := store.Save(context.Background(), record)

			if (err != nil) != tt.wantErr {
				t.Errorf("Save() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStoreList(t *testing.T) {
	tests := []struct {
		name      string
		setup     func(*Store)
		listFunc  func(ctx context.Context, limit int) ([]internal.Record, error)
		limit     int
		wantCount int
		wantErr   bool
	}{
		{
			name: "empty store",
			setup: func(s *Store) {
				// no records
			},
			limit:     10,
			wantCount: 0,
			wantErr:   false,
		},
		{
			name: "with records",
			setup: func(s *Store) {
				for i := 0; i < 5; i++ {
					s.Save(context.Background(), &internal.Record{ID: string(rune('a' + i))})
				}
			},
			limit:     10,
			wantCount: 5,
			wantErr:   false,
		},
		{
			name: "limit less than records",
			setup: func(s *Store) {
				for i := 0; i < 5; i++ {
					s.Save(context.Background(), &internal.Record{ID: string(rune('a' + i))})
				}
			},
			limit:     3,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "limit zero returns all",
			setup: func(s *Store) {
				for i := 0; i < 3; i++ {
					s.Save(context.Background(), &internal.Record{ID: string(rune('a' + i))})
				}
			},
			limit:     0,
			wantCount: 3,
			wantErr:   false,
		},
		{
			name: "custom list func",
			setup: func(s *Store) {
				s.ListFunc = func(ctx context.Context, limit int) ([]internal.Record, error) {
					return []internal.Record{{ID: "custom"}}, nil
				}
			},
			limit:     10,
			wantCount: 1,
			wantErr:   false,
		},
		{
			name: "custom list func error",
			setup: func(s *Store) {
				s.ListFunc = func(ctx context.Context, limit int) ([]internal.Record, error) {
					return nil, errors.New("list error")
				}
			},
			limit:     10,
			wantCount: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewStore()
			tt.setup(store)

			records, err := store.List(context.Background(), tt.limit)

			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr && len(records) != tt.wantCount {
				t.Errorf("List() returned %d records, want %d", len(records), tt.wantCount)
			}
		})
	}
}

func TestStoreListOrder(t *testing.T) {
	store := NewStore()

	// Add records in order a, b, c
	store.Save(context.Background(), &internal.Record{ID: "a"})
	store.Save(context.Background(), &internal.Record{ID: "b"})
	store.Save(context.Background(), &internal.Record{ID: "c"})

	records, err := store.List(context.Background(), 10)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Should return in reverse order (most recent first): c, b, a
	if len(records) != 3 {
		t.Fatalf("expected 3 records, got %d", len(records))
	}

	if records[0].ID != "c" {
		t.Errorf("first record ID = %q, want %q", records[0].ID, "c")
	}

	if records[1].ID != "b" {
		t.Errorf("second record ID = %q, want %q", records[1].ID, "b")
	}

	if records[2].ID != "a" {
		t.Errorf("third record ID = %q, want %q", records[2].ID, "a")
	}
}

func TestStoreRecords(t *testing.T) {
	store := NewStore()

	store.Save(context.Background(), &internal.Record{ID: "a"})
	store.Save(context.Background(), &internal.Record{ID: "b"})

	records := store.Records()

	if len(records) != 2 {
		t.Errorf("Records() returned %d, want 2", len(records))
	}

	// Should be in insertion order
	if records[0].ID != "a" {
		t.Errorf("first record ID = %q, want %q", records[0].ID, "a")
	}

	if records[1].ID != "b" {
		t.Errorf("second record ID = %q, want %q", records[1].ID, "b")
	}
}

func TestStoreReset(t *testing.T) {
	store := NewStore()

	store.Save(context.Background(), &internal.Record{ID: "a"})
	store.Save(context.Background(), &internal.Record{ID: "b"})

	if len(store.Records()) != 2 {
		t.Fatal("expected 2 records before reset")
	}

	store.Reset()

	if len(store.Records()) != 0 {
		t.Errorf("expected 0 records after reset, got %d", len(store.Records()))
	}
}
