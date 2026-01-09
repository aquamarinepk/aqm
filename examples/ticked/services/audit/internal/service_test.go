package internal

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/pubsub"
)

func TestHandleEvent(t *testing.T) {
	tests := []struct {
		name     string
		envelope pubsub.Envelope
		store    *testStore
		wantErr  bool
		validate func(*testing.T, *Record)
	}{
		{
			name: "valid event",
			envelope: pubsub.Envelope{
				ID:        "event-123",
				Topic:     Topic,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"event_type": "todo.item.added",
					"item_id":    "item-456",
					"title":      "Test task",
				},
				Metadata: map[string]string{
					"user_id": "user-789",
					"source":  "ticked",
				},
			},
			store: &testStore{
				saveFunc: func(ctx context.Context, record *Record) error {
					return nil
				},
			},
			wantErr: false,
			validate: func(t *testing.T, record *Record) {
				if record == nil {
					t.Fatal("record is nil")
				}
				if record.EventType != "todo.item.added" {
					t.Errorf("EventType = %q, want %q", record.EventType, "todo.item.added")
				}
				if record.ItemID != "item-456" {
					t.Errorf("ItemID = %q, want %q", record.ItemID, "item-456")
				}
				if record.UserID != "user-789" {
					t.Errorf("UserID = %q, want %q", record.UserID, "user-789")
				}
				if record.Source != "ticked" {
					t.Errorf("Source = %q, want %q", record.Source, "ticked")
				}
			},
		},
		{
			name: "invalid payload type",
			envelope: pubsub.Envelope{
				ID:        "event-123",
				Topic:     Topic,
				Timestamp: time.Now(),
				Payload:   "invalid payload",
				Metadata: map[string]string{
					"user_id": "user-789",
					"source":  "ticked",
				},
			},
			store:   &testStore{},
			wantErr: true,
			validate: func(t *testing.T, record *Record) {
				// No validation needed for error case
			},
		},
		{
			name: "store error",
			envelope: pubsub.Envelope{
				ID:        "event-123",
				Topic:     Topic,
				Timestamp: time.Now(),
				Payload: map[string]interface{}{
					"event_type": "todo.item.added",
					"item_id":    "item-456",
				},
				Metadata: map[string]string{
					"user_id": "user-789",
					"source":  "ticked",
				},
			},
			store: &testStore{
				saveFunc: func(ctx context.Context, record *Record) error {
					return errors.New("database error")
				},
			},
			wantErr: true,
			validate: func(t *testing.T, record *Record) {
				// No validation needed for error case
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var savedRecord *Record
			if tt.store.saveFunc == nil {
				tt.store.saveFunc = func(ctx context.Context, record *Record) error {
					savedRecord = record
					return nil
				}
			} else {
				originalSave := tt.store.saveFunc
				tt.store.saveFunc = func(ctx context.Context, record *Record) error {
					savedRecord = record
					return originalSave(ctx, record)
				}
			}

			svc := &Service{
				store: tt.store,
				log:   log.NewNoopLogger(),
			}

			err := svc.handleEvent(context.Background(), tt.envelope)

			if (err != nil) != tt.wantErr {
				t.Errorf("handleEvent() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				tt.validate(t, savedRecord)
			}
		})
	}
}

func TestHandleEventPayloadConversion(t *testing.T) {
	var savedRecord *Record
	store := &testStore{
		saveFunc: func(ctx context.Context, record *Record) error {
			savedRecord = record
			return nil
		},
	}

	svc := &Service{
		store: store,
		log:   log.NewNoopLogger(),
	}

	// Test with mixed types in payload (only strings should be kept)
	envelope := pubsub.Envelope{
		ID:        "event-123",
		Topic:     Topic,
		Timestamp: time.Now(),
		Payload: map[string]interface{}{
			"event_type": "todo.item.added",
			"item_id":    "item-456",
			"count":      42, // non-string, should be skipped
		},
		Metadata: map[string]string{
			"user_id": "user-789",
			"source":  "ticked",
		},
	}

	err := svc.handleEvent(context.Background(), envelope)
	if err != nil {
		t.Fatalf("handleEvent() error = %v", err)
	}

	if savedRecord == nil {
		t.Fatal("record was not saved")
	}

	// Verify string values are present
	if savedRecord.Payload["event_type"] != "todo.item.added" {
		t.Errorf("event_type = %q, want %q", savedRecord.Payload["event_type"], "todo.item.added")
	}

	// Verify non-string values are not in payload
	if _, ok := savedRecord.Payload["count"]; ok {
		t.Error("count should not be in payload (non-string value)")
	}
}
