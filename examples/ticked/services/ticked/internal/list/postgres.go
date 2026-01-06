package list

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
)

type postgresRepo struct {
	db *sql.DB
}

// NewPostgresRepo creates a PostgreSQL-backed repository.
func NewPostgresRepo(db *sql.DB) Repo {
	return &postgresRepo{db: db}
}

func (r *postgresRepo) Save(ctx context.Context, list *TodoList) error {
	itemsJSON, err := json.Marshal(list.Items)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO todo_lists (id, user_id, items, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			items = EXCLUDED.items,
			updated_at = EXCLUDED.updated_at
	`

	_, err = r.db.ExecContext(ctx, query,
		list.ListID,
		list.UserID,
		itemsJSON,
		list.CreatedAt,
		list.UpdatedAt,
	)

	return err
}

func (r *postgresRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	query := `
		SELECT id, user_id, items, created_at, updated_at
		FROM todo_lists
		WHERE user_id = $1
	`

	var list TodoList
	var itemsJSON []byte

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&list.ListID,
		&list.UserID,
		&itemsJSON,
		&list.CreatedAt,
		&list.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	if err := json.Unmarshal(itemsJSON, &list.Items); err != nil {
		return nil, err
	}

	return &list, nil
}

func (r *postgresRepo) Delete(ctx context.Context, listID uuid.UUID) error {
	query := `DELETE FROM todo_lists WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, listID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
