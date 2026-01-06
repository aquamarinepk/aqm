package list

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// Repo abstracts persistence for todo lists.
type Repo interface {
	Save(ctx context.Context, list *TodoList) error
	FindByUserID(ctx context.Context, userID uuid.UUID) (*TodoList, error)
	Delete(ctx context.Context, listID uuid.UUID) error
}

var ErrNotFound = errors.New("list not found")
