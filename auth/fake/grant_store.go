package fake

import (
	"context"
	"sync"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

type grantKey struct {
	Username string
	RoleID   uuid.UUID
}

type GrantStore struct {
	mu        sync.RWMutex
	grants    map[grantKey]*auth.Grant
	roleStore *RoleStore
}

func NewGrantStore(roleStore *RoleStore) *GrantStore {
	return &GrantStore{
		grants:    make(map[grantKey]*auth.Grant),
		roleStore: roleStore,
	}
}

func (s *GrantStore) Create(ctx context.Context, grant *auth.Grant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := grantKey{Username: grant.Username, RoleID: grant.RoleID}
	if _, exists := s.grants[key]; exists {
		return auth.ErrGrantAlreadyExists
	}

	s.grants[key] = grant
	return nil
}

func (s *GrantStore) Delete(ctx context.Context, username string, roleID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	key := grantKey{Username: username, RoleID: roleID}
	if _, exists := s.grants[key]; !exists {
		return auth.ErrGrantNotFound
	}

	delete(s.grants, key)
	return nil
}

func (s *GrantStore) GetUserGrants(ctx context.Context, username string) ([]*auth.Grant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	grants := make([]*auth.Grant, 0)
	for key, grant := range s.grants {
		if key.Username == username {
			grants = append(grants, grant)
		}
	}

	return grants, nil
}

func (s *GrantStore) GetRoleGrants(ctx context.Context, roleID uuid.UUID) ([]*auth.Grant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	grants := make([]*auth.Grant, 0)
	for key, grant := range s.grants {
		if key.RoleID == roleID {
			grants = append(grants, grant)
		}
	}

	return grants, nil
}

func (s *GrantStore) GetUserRoles(ctx context.Context, username string) ([]*auth.Role, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roles := make([]*auth.Role, 0)
	for key := range s.grants {
		if key.Username == username {
			role, err := s.roleStore.Get(ctx, key.RoleID)
			if err != nil {
				continue
			}
			roles = append(roles, role)
		}
	}

	return roles, nil
}

func (s *GrantStore) HasRole(ctx context.Context, username string, roleName string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for key := range s.grants {
		if key.Username == username {
			role, err := s.roleStore.Get(ctx, key.RoleID)
			if err != nil {
				continue
			}
			if role.Name == roleName {
				return true, nil
			}
		}
	}

	return false, nil
}
