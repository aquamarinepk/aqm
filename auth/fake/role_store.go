package fake

import (
	"context"
	"sync"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

type RoleStore struct {
	mu            sync.RWMutex
	roles         map[uuid.UUID]*auth.Role
	rolesByName   map[string]*auth.Role
}

func NewRoleStore() *RoleStore {
	return &RoleStore{
		roles:       make(map[uuid.UUID]*auth.Role),
		rolesByName: make(map[string]*auth.Role),
	}
}

func (s *RoleStore) Create(ctx context.Context, role *auth.Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[role.ID]; exists {
		return auth.ErrRoleAlreadyExists
	}

	if _, exists := s.rolesByName[role.Name]; exists {
		return auth.ErrRoleAlreadyExists
	}

	s.roles[role.ID] = role
	s.rolesByName[role.Name] = role

	return nil
}

func (s *RoleStore) Get(ctx context.Context, id uuid.UUID) (*auth.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.roles[id]
	if !exists {
		return nil, auth.ErrRoleNotFound
	}

	return role, nil
}

func (s *RoleStore) GetByName(ctx context.Context, name string) (*auth.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	role, exists := s.rolesByName[name]
	if !exists {
		return nil, auth.ErrRoleNotFound
	}

	return role, nil
}

func (s *RoleStore) Update(ctx context.Context, role *auth.Role) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.roles[role.ID]; !exists {
		return auth.ErrRoleNotFound
	}

	s.roles[role.ID] = role
	s.rolesByName[role.Name] = role

	return nil
}

func (s *RoleStore) Delete(ctx context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	role, exists := s.roles[id]
	if !exists {
		return auth.ErrRoleNotFound
	}

	role.Status = "deleted"
	return nil
}

func (s *RoleStore) List(ctx context.Context) ([]*auth.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := make([]*auth.Role, 0, len(s.roles))
	for _, role := range s.roles {
		roles = append(roles, role)
	}

	return roles, nil
}

func (s *RoleStore) ListByStatus(ctx context.Context, status auth.RoleStatus) ([]*auth.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := make([]*auth.Role, 0)
	for _, role := range s.roles {
		if role.Status == status {
			roles = append(roles, role)
		}
	}

	return roles, nil
}
