package fake

import (
	"context"

	"github.com/aquamarinepk/aqm/examples/ticked/services/ticked/internal/list"
	"github.com/google/uuid"
)

// Service is a fake service for testing.
type Service struct {
	GetOrCreateListFunc func(ctx context.Context, userID uuid.UUID) (*list.TodoList, error)
	GetListFunc         func(ctx context.Context, userID uuid.UUID) (*list.TodoList, error)
	AddItemFunc         func(ctx context.Context, userID uuid.UUID, text string) (*list.TodoList, error)
	UpdateItemFunc func(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, text *string, completed *bool) (*list.TodoList, error)
	RemoveItemFunc      func(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) (*list.TodoList, error)
}

func (s *Service) GetOrCreateList(ctx context.Context, userID uuid.UUID) (*list.TodoList, error) {
	if s.GetOrCreateListFunc != nil {
		return s.GetOrCreateListFunc(ctx, userID)
	}
	return list.NewTodoList(userID), nil
}

func (s *Service) GetList(ctx context.Context, userID uuid.UUID) (*list.TodoList, error) {
	if s.GetListFunc != nil {
		return s.GetListFunc(ctx, userID)
	}
	return nil, list.ErrNotFound
}

func (s *Service) AddItem(ctx context.Context, userID uuid.UUID, text string) (*list.TodoList, error) {
	if s.AddItemFunc != nil {
		return s.AddItemFunc(ctx, userID, text)
	}
	return nil, nil
}

func (s *Service) UpdateItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, text *string, completed *bool) (*list.TodoList, error) {
	if s.UpdateItemFunc != nil {
		return s.UpdateItemFunc(ctx, userID, itemID, text, completed)
	}
	return nil, nil
}

func (s *Service) RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) (*list.TodoList, error) {
	if s.RemoveItemFunc != nil {
		return s.RemoveItemFunc(ctx, userID, itemID)
	}
	return nil, nil
}
