package list

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// memStore is an in-memory store for demo and testing purposes.
//
// This store is used when running in "fake" mode (no database required).
// It provides the same TodoListStore interface as postgresStore, making
// it easy to swap implementations via configuration.
type memStore struct {
	mu    sync.RWMutex
	lists map[uuid.UUID]*TodoList
}

// NewMemStore creates an in-memory TodoListStore.
func NewMemStore() TodoListStore {
	return &memStore{
		lists: make(map[uuid.UUID]*TodoList),
	}
}

func (s *memStore) Save(ctx context.Context, list *TodoList) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Create a copy to avoid external modifications
	listCopy := *list
	listsCopy := make([]TodoItem, len(list.Items))
	copy(listsCopy, list.Items)
	listCopy.Items = listsCopy

	s.lists[list.UserID] = &listCopy
	return nil
}

func (s *memStore) FindByUserID(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list, ok := s.lists[userID]
	if !ok {
		return nil, ErrNotFound
	}

	// Return a copy to avoid external modifications
	listCopy := *list
	itemsCopy := make([]TodoItem, len(list.Items))
	copy(itemsCopy, list.Items)
	listCopy.Items = itemsCopy

	return &listCopy, nil
}

func (s *memStore) Delete(ctx context.Context, listID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for userID, list := range s.lists {
		if list.ListID == listID {
			delete(s.lists, userID)
			return nil
		}
	}

	return ErrNotFound
}

var _ TodoListStore = (*memStore)(nil)
