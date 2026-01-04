package fake

import (
	"context"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

func TestNewRoleStore(t *testing.T) {
	store := NewRoleStore()
	if store == nil {
		t.Fatal("NewRoleStore() returned nil")
	}
	if store.roles == nil {
		t.Error("roles map not initialized")
	}
	if store.rolesByName == nil {
		t.Error("rolesByName map not initialized")
	}
}

func TestRoleStore_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*RoleStore) *auth.Role
		role    *auth.Role
		wantErr bool
		errType error
	}{
		{
			name: "create new role",
			setup: func(s *RoleStore) *auth.Role {
				return nil
			},
			role: &auth.Role{
				ID:   uuid.New(),
				Name: "admin",
			},
			wantErr: false,
		},
		{
			name: "duplicate role ID",
			setup: func(s *RoleStore) *auth.Role {
				existing := &auth.Role{
					ID:   uuid.New(),
					Name: "existing",
				}
				_ = s.Create(context.Background(), existing)
				return existing
			},
			role:    nil,
			wantErr: true,
			errType: auth.ErrRoleAlreadyExists,
		},
		{
			name: "duplicate role name",
			setup: func(s *RoleStore) *auth.Role {
				existing := &auth.Role{
					ID:   uuid.New(),
					Name: "duplicate",
				}
				_ = s.Create(context.Background(), existing)
				return nil
			},
			role: &auth.Role{
				ID:   uuid.New(),
				Name: "duplicate",
			},
			wantErr: true,
			errType: auth.ErrRoleAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewRoleStore()
			setupRole := tt.setup(store)
			if tt.role == nil && setupRole != nil {
				tt.role = &auth.Role{
					ID:   setupRole.ID,
					Name: "different",
				}
			}

			err := store.Create(context.Background(), tt.role)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != tt.errType {
				t.Errorf("Create() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestRoleStore_Get(t *testing.T) {
	store := NewRoleStore()
	role := &auth.Role{
		ID:   uuid.New(),
		Name: "admin",
	}
	_ = store.Create(context.Background(), role)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "existing role",
			id:      role.ID,
			wantErr: false,
		},
		{
			name:    "nonexistent role",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Get(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("Get() returned wrong role")
			}
		})
	}
}

func TestRoleStore_GetByName(t *testing.T) {
	store := NewRoleStore()
	role := &auth.Role{
		ID:   uuid.New(),
		Name: "admin",
	}
	_ = store.Create(context.Background(), role)

	tests := []struct {
		name     string
		roleName string
		wantErr  bool
	}{
		{
			name:     "existing role",
			roleName: "admin",
			wantErr:  false,
		},
		{
			name:     "nonexistent role",
			roleName: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByName(context.Background(), tt.roleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Name != tt.roleName {
				t.Errorf("GetByName() returned wrong role")
			}
		})
	}
}

func TestRoleStore_Update(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*RoleStore) *auth.Role
		update  func(*auth.Role)
		wantErr bool
	}{
		{
			name: "update existing role",
			setup: func(s *RoleStore) *auth.Role {
				role := &auth.Role{
					ID:   uuid.New(),
					Name: "admin",
				}
				_ = s.Create(context.Background(), role)
				return role
			},
			update: func(r *auth.Role) {
				r.Name = "updated"
			},
			wantErr: false,
		},
		{
			name: "update nonexistent role",
			setup: func(s *RoleStore) *auth.Role {
				return &auth.Role{
					ID:   uuid.New(),
					Name: "nonexistent",
				}
			},
			update:  func(r *auth.Role) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewRoleStore()
			role := tt.setup(store)
			tt.update(role)

			err := store.Update(context.Background(), role)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRoleStore_Delete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*RoleStore) uuid.UUID
		wantErr bool
	}{
		{
			name: "delete existing role",
			setup: func(s *RoleStore) uuid.UUID {
				role := &auth.Role{
					ID:     uuid.New(),
					Name:   "admin",
					Status: auth.RoleStatusActive,
				}
				_ = s.Create(context.Background(), role)
				return role.ID
			},
			wantErr: false,
		},
		{
			name: "delete nonexistent role",
			setup: func(s *RoleStore) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewRoleStore()
			id := tt.setup(store)

			err := store.Delete(context.Background(), id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				role, _ := store.Get(context.Background(), id)
				if role.Status != "deleted" {
					t.Error("Delete() did not set status to deleted")
				}
			}
		})
	}
}

func TestRoleStore_List(t *testing.T) {
	store := NewRoleStore()

	role1 := &auth.Role{ID: uuid.New(), Name: "admin"}
	role2 := &auth.Role{ID: uuid.New(), Name: "editor"}
	_ = store.Create(context.Background(), role1)
	_ = store.Create(context.Background(), role2)

	roles, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("List() returned %d roles, want 2", len(roles))
	}
}

func TestRoleStore_ListByStatus(t *testing.T) {
	store := NewRoleStore()

	activeRole := &auth.Role{ID: uuid.New(), Name: "active", Status: auth.RoleStatusActive}
	inactiveRole := &auth.Role{ID: uuid.New(), Name: "inactive", Status: auth.RoleStatusInactive}
	_ = store.Create(context.Background(), activeRole)
	_ = store.Create(context.Background(), inactiveRole)

	tests := []struct {
		name      string
		status    auth.RoleStatus
		wantCount int
	}{
		{
			name:      "list active roles",
			status:    auth.RoleStatusActive,
			wantCount: 1,
		},
		{
			name:      "list inactive roles",
			status:    auth.RoleStatusInactive,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles, err := store.ListByStatus(context.Background(), tt.status)
			if err != nil {
				t.Fatalf("ListByStatus() error = %v", err)
			}
			if len(roles) != tt.wantCount {
				t.Errorf("ListByStatus() returned %d roles, want %d", len(roles), tt.wantCount)
			}
		})
	}
}
