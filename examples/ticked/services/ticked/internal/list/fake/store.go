package fake

import (
	"context"

	"github.com/aquamarinepk/aqm/examples/ticked/services/ticked/internal/list"
	"github.com/google/uuid"
)

// Store is a fake repository for testing.
type Store struct {
	SaveFunc         func(ctx context.Context, l *list.TodoList) error
	FindByUserIDFunc func(ctx context.Context, userID uuid.UUID) (*list.TodoList, error)
	DeleteFunc       func(ctx context.Context, listID uuid.UUID) error
}

func (r *Store) Save(ctx context.Context, l *list.TodoList) error {
	if r.SaveFunc != nil {
		return r.SaveFunc(ctx, l)
	}
	return nil
}

func (r *Store) FindByUserID(ctx context.Context, userID uuid.UUID) (*list.TodoList, error) {
	if r.FindByUserIDFunc != nil {
		return r.FindByUserIDFunc(ctx, userID)
	}
	return nil, list.ErrNotFound
}

func (r *Store) Delete(ctx context.Context, listID uuid.UUID) error {
	if r.DeleteFunc != nil {
		return r.DeleteFunc(ctx, listID)
	}
	return nil
}
