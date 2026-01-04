package fake

import (
	"context"
	"sync"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

type UserStore struct {
	mu                sync.RWMutex
	users             map[uuid.UUID]*auth.User
	usersByUsername   map[string]*auth.User
	usersByEmailLookup map[string]*auth.User
	usersByPINLookup  map[string]*auth.User
}

func NewUserStore() *UserStore {
	return &UserStore{
		users:              make(map[uuid.UUID]*auth.User),
		usersByUsername:    make(map[string]*auth.User),
		usersByEmailLookup: make(map[string]*auth.User),
		usersByPINLookup:   make(map[string]*auth.User),
	}
}

func (s *UserStore) Create(ctx context.Context, user *auth.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; exists {
		return auth.ErrUserAlreadyExists
	}

	if _, exists := s.usersByUsername[user.Username]; exists {
		return auth.ErrUsernameExists
	}

	s.users[user.ID] = user
	s.usersByUsername[user.Username] = user
	if len(user.EmailLookup) > 0 {
		s.usersByEmailLookup[string(user.EmailLookup)] = user
	}
	if len(user.PINLookup) > 0 {
		s.usersByPINLookup[string(user.PINLookup)] = user
	}

	return nil
}

func (s *UserStore) Get(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[id]
	if !exists {
		return nil, auth.ErrUserNotFound
	}

	return user, nil
}

func (s *UserStore) GetByEmailLookup(ctx context.Context, lookup []byte) (*auth.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByEmailLookup[string(lookup)]
	if !exists {
		return nil, auth.ErrUserNotFound
	}

	return user, nil
}

func (s *UserStore) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByUsername[username]
	if !exists {
		return nil, auth.ErrUserNotFound
	}

	return user, nil
}

func (s *UserStore) GetByPINLookup(ctx context.Context, lookup []byte) (*auth.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.usersByPINLookup[string(lookup)]
	if !exists {
		return nil, auth.ErrUserNotFound
	}

	return user, nil
}

func (s *UserStore) Update(ctx context.Context, user *auth.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.ID]; !exists {
		return auth.ErrUserNotFound
	}

	s.users[user.ID] = user
	s.usersByUsername[user.Username] = user
	if len(user.EmailLookup) > 0 {
		s.usersByEmailLookup[string(user.EmailLookup)] = user
	}
	if len(user.PINLookup) > 0 {
		s.usersByPINLookup[string(user.PINLookup)] = user
	}

	return nil
}

func (s *UserStore) Delete(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	user, exists := s.users[id]
	if !exists {
		return auth.ErrUserNotFound
	}

	user.Status = "deleted"
	return nil
}

func (s *UserStore) List(ctx context.Context) ([]*auth.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*auth.User, 0, len(s.users))
	for _, user := range s.users {
		users = append(users, user)
	}

	return users, nil
}

func (s *UserStore) ListByStatus(ctx context.Context, status auth.UserStatus) ([]*auth.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]*auth.User, 0)
	for _, user := range s.users {
		if user.Status == status {
			users = append(users, user)
		}
	}

	return users, nil
}
