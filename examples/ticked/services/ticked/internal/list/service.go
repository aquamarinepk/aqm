package list

import (
	"context"
	"errors"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/google/uuid"
)

// Service contains business logic for todo lists.
type Service struct {
	repo Repo
	log  log.Logger
	cfg  *config.Config
}

// NewService creates a new service instance.
func NewService(repo Repo, cfg *config.Config, log log.Logger) *Service {
	if log == nil {
		log = &noopLogger{}
	}
	return &Service{
		repo: repo,
		log:  log,
		cfg:  cfg,
	}
}

// GetOrCreateList retrieves a user's list or creates it if it doesn't exist.
func (s *Service) GetOrCreateList(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	list, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			list = NewTodoList(userID)
			if err := s.repo.Save(ctx, list); err != nil {
				return nil, err
			}
			return list, nil
		}
		return nil, err
	}
	return list, nil
}

// GetList retrieves a user's list.
func (s *Service) GetList(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	return s.repo.FindByUserID(ctx, userID)
}

// AddItem adds an item to a user's list.
func (s *Service) AddItem(ctx context.Context, userID uuid.UUID, text string) (*TodoList, error) {
	list, err := s.GetOrCreateList(ctx, userID)
	if err != nil {
		return nil, err
	}

	if _, err := list.AddItem(text); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

// UpdateItem updates an item in a user's list.
func (s *Service) UpdateItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID, text *string, completed *bool) (*TodoList, error) {
	list, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := list.UpdateItem(itemID, text, completed); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

// RemoveItem removes an item from a user's list.
func (s *Service) RemoveItem(ctx context.Context, userID uuid.UUID, itemID uuid.UUID) (*TodoList, error) {
	list, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := list.RemoveItem(itemID); err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

// noopLogger is a no-op logger implementation.
type noopLogger struct{}

func (l *noopLogger) Debug(v ...any)                  {}
func (l *noopLogger) Debugf(format string, a ...any)  {}
func (l *noopLogger) Info(v ...any)                   {}
func (l *noopLogger) Infof(format string, a ...any)   {}
func (l *noopLogger) Error(v ...any)                  {}
func (l *noopLogger) Errorf(format string, a ...any)  {}
func (l *noopLogger) With(args ...any) log.Logger     { return l }
