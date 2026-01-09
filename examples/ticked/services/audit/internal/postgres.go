package internal

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// PostgresStore implements Store using PostgreSQL.
type PostgresStore struct {
	db *sql.DB
}

// NewPostgresStore creates a new PostgreSQL-backed store.
func NewPostgresStore(db *sql.DB) *PostgresStore {
	return &PostgresStore{db: db}
}

// Save persists an audit record to the database.
func (s *PostgresStore) Save(ctx context.Context, record *Record) error {
	payloadJSON, err := json.Marshal(record.Payload)
	if err != nil {
		return fmt.Errorf("cannot marshal payload: %w", err)
	}

	query := `
		INSERT INTO audit_log (id, event_type, item_id, user_id, payload, source, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err = s.db.ExecContext(ctx, query,
		record.ID,
		record.EventType,
		sql.NullString{String: record.ItemID, Valid: record.ItemID != ""},
		record.UserID,
		payloadJSON,
		record.Source,
		record.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("cannot insert audit record: %w", err)
	}

	return nil
}

// List retrieves the most recent audit records.
func (s *PostgresStore) List(ctx context.Context, limit int) ([]Record, error) {
	query := `
		SELECT id, event_type, COALESCE(item_id, ''), user_id, payload, source, created_at
		FROM audit_log
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := s.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("cannot query audit records: %w", err)
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var r Record
		var payloadJSON []byte
		if err := rows.Scan(&r.ID, &r.EventType, &r.ItemID, &r.UserID, &payloadJSON, &r.Source, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("cannot scan audit record: %w", err)
		}

		if len(payloadJSON) > 0 {
			if err := json.Unmarshal(payloadJSON, &r.Payload); err != nil {
				r.Payload = make(map[string]string)
			}
		}

		records = append(records, r)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return records, nil
}
