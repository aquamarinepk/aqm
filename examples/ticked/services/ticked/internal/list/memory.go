package list

import (
	"context"
	"sync"

	"github.com/google/uuid"
)

// memoryRepo is an in-memory repository for demo purposes.
type memoryRepo struct {
	mu    sync.RWMutex
	lists map[uuid.UUID]*TodoList
}

func NewMemoryRepo() Repo {
	return &memoryRepo{
		lists: make(map[uuid.UUID]*TodoList),
	}
}

func (r *memoryRepo) Save(ctx context.Context, list *TodoList) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Create a copy to avoid external modifications
	listCopy := *list
	listsCopy := make([]TodoItem, len(list.Items))
	copy(listsCopy, list.Items)
	listCopy.Items = listsCopy

	r.lists[list.UserID] = &listCopy
	return nil
}

func (r *memoryRepo) FindByUserID(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list, ok := r.lists[userID]
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

func (r *memoryRepo) Delete(ctx context.Context, listID uuid.UUID) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for userID, list := range r.lists {
		if list.ListID == listID {
			delete(r.lists, userID)
			return nil
		}
	}

	return ErrNotFound
}
