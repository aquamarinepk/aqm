package internal

import (
	"context"
	"errors"
	"testing"
	"time"
)

// testStore is a simple fake store for testing.
type testStore struct {
	saveFunc func(ctx context.Context, record *Record) error
	listFunc func(ctx context.Context, limit int) ([]Record, error)
}

func (s *testStore) Save(ctx context.Context, record *Record) error {
	if s.saveFunc != nil {
		return s.saveFunc(ctx, record)
	}
	return nil
}

func (s *testStore) List(ctx context.Context, limit int) ([]Record, error) {
	if s.listFunc != nil {
		return s.listFunc(ctx, limit)
	}
	return nil, nil
}

func TestNewRecord(t *testing.T) {
	record := &Record{
		ID:        "test-id",
		EventType: "todo.item.added",
		ItemID:    "item-123",
		UserID:    "user-456",
		Payload:   map[string]string{"title": "Test task"},
		Source:    "ticked",
		CreatedAt: time.Now(),
	}

	if record.ID != "test-id" {
		t.Errorf("ID = %q, want %q", record.ID, "test-id")
	}
	if record.EventType != "todo.item.added" {
		t.Errorf("EventType = %q, want %q", record.EventType, "todo.item.added")
	}
	if record.ItemID != "item-123" {
		t.Errorf("ItemID = %q, want %q", record.ItemID, "item-123")
	}
	if record.UserID != "user-456" {
		t.Errorf("UserID = %q, want %q", record.UserID, "user-456")
	}
	if record.Source != "ticked" {
		t.Errorf("Source = %q, want %q", record.Source, "ticked")
	}
}

func TestStoreInterface(t *testing.T) {
	var saved *Record
	store := &testStore{
		saveFunc: func(ctx context.Context, record *Record) error {
			saved = record
			return nil
		},
		listFunc: func(ctx context.Context, limit int) ([]Record, error) {
			return []Record{*saved}, nil
		},
	}

	record := &Record{
		ID:        "test-id",
		EventType: "todo.item.added",
		ItemID:    "item-123",
		UserID:    "user-456",
		Payload:   map[string]string{"title": "Test task"},
		Source:    "ticked",
		CreatedAt: time.Now(),
	}

	// Test Save
	err := store.Save(context.Background(), record)
	if err != nil {
		t.Errorf("Save() error = %v", err)
	}

	if saved == nil {
		t.Fatal("record was not saved")
	}

	if saved.ID != record.ID {
		t.Errorf("saved ID = %q, want %q", saved.ID, record.ID)
	}

	// Test List
	records, err := store.List(context.Background(), 10)
	if err != nil {
		t.Errorf("List() error = %v", err)
	}

	if len(records) != 1 {
		t.Errorf("List() returned %d records, want 1", len(records))
	}
}

func TestStoreSaveError(t *testing.T) {
	expectedErr := errors.New("database error")
	store := &testStore{
		saveFunc: func(ctx context.Context, record *Record) error {
			return expectedErr
		},
	}

	record := &Record{
		ID:        "test-id",
		EventType: "todo.item.added",
	}

	err := store.Save(context.Background(), record)
	if !errors.Is(err, expectedErr) {
		t.Errorf("Save() error = %v, want %v", err, expectedErr)
	}
}

func TestStoreListError(t *testing.T) {
	expectedErr := errors.New("database error")
	store := &testStore{
		listFunc: func(ctx context.Context, limit int) ([]Record, error) {
			return nil, expectedErr
		},
	}

	_, err := store.List(context.Background(), 10)
	if !errors.Is(err, expectedErr) {
		t.Errorf("List() error = %v, want %v", err, expectedErr)
	}
}
