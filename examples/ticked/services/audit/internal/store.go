package internal

import (
	"context"
	"time"
)

// Record represents an audit log entry.
type Record struct {
	ID        string
	EventType string
	ItemID    string
	UserID    string
	Payload   map[string]string
	Source    string
	CreatedAt time.Time
}

// Store defines persistence operations for audit records.
type Store interface {
	Save(ctx context.Context, record *Record) error
	List(ctx context.Context, limit int) ([]Record, error)
}
