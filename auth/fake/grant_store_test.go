package fake

import (
	"context"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

func TestNewGrantStore(t *testing.T) {
	roleStore := NewRoleStore()
	store := NewGrantStore(roleStore)
	if store == nil {
		t.Fatal("NewGrantStore() returned nil")
	}
	if store.grants == nil {
		t.Error("grants map not initialized")
	}
	if store.roleStore != roleStore {
		t.Error("roleStore not set correctly")
	}
}

func TestGrantStore_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*GrantStore) *auth.Grant
		grant   *auth.Grant
		wantErr bool
		errType error
	}{
		{
			name: "create new grant",
			setup: func(s *GrantStore) *auth.Grant {
				return nil
			},
			grant: &auth.Grant{
				UserID: uuid.New(),
				RoleID: uuid.New(),
			},
			wantErr: false,
		},
		{
			name: "duplicate grant",
			setup: func(s *GrantStore) *auth.Grant {
				grant := &auth.Grant{
					UserID: uuid.New(),
					RoleID: uuid.New(),
				}
				_ = s.Create(context.Background(), grant)
				return grant
			},
			grant:   nil,
			wantErr: true,
			errType: auth.ErrGrantAlreadyExists,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleStore := NewRoleStore()
			store := NewGrantStore(roleStore)
			setupGrant := tt.setup(store)
			if tt.grant == nil && setupGrant != nil {
				tt.grant = setupGrant
			}

			err := store.Create(context.Background(), tt.grant)
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

func TestGrantStore_Delete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*GrantStore) (uuid.UUID, uuid.UUID)
		wantErr bool
	}{
		{
			name: "delete existing grant",
			setup: func(s *GrantStore) (uuid.UUID, uuid.UUID) {
				userID := uuid.New()
				roleID := uuid.New()
				grant := &auth.Grant{
					UserID: userID,
					RoleID: roleID,
				}
				_ = s.Create(context.Background(), grant)
				return userID, roleID
			},
			wantErr: false,
		},
		{
			name: "delete nonexistent grant",
			setup: func(s *GrantStore) (uuid.UUID, uuid.UUID) {
				return uuid.New(), uuid.New()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleStore := NewRoleStore()
			store := NewGrantStore(roleStore)
			userID, roleID := tt.setup(store)

			err := store.Delete(context.Background(), userID, roleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGrantStore_GetUserGrants(t *testing.T) {
	roleStore := NewRoleStore()
	store := NewGrantStore(roleStore)

	userID := uuid.New()
	grant1 := &auth.Grant{UserID: userID, RoleID: uuid.New()}
	grant2 := &auth.Grant{UserID: userID, RoleID: uuid.New()}
	otherGrant := &auth.Grant{UserID: uuid.New(), RoleID: uuid.New()}

	_ = store.Create(context.Background(), grant1)
	_ = store.Create(context.Background(), grant2)
	_ = store.Create(context.Background(), otherGrant)

	grants, err := store.GetUserGrants(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserGrants() error = %v", err)
	}
	if len(grants) != 2 {
		t.Errorf("GetUserGrants() returned %d grants, want 2", len(grants))
	}
}

func TestGrantStore_GetRoleGrants(t *testing.T) {
	roleStore := NewRoleStore()
	store := NewGrantStore(roleStore)

	roleID := uuid.New()
	grant1 := &auth.Grant{UserID: uuid.New(), RoleID: roleID}
	grant2 := &auth.Grant{UserID: uuid.New(), RoleID: roleID}
	otherGrant := &auth.Grant{UserID: uuid.New(), RoleID: uuid.New()}

	_ = store.Create(context.Background(), grant1)
	_ = store.Create(context.Background(), grant2)
	_ = store.Create(context.Background(), otherGrant)

	grants, err := store.GetRoleGrants(context.Background(), roleID)
	if err != nil {
		t.Fatalf("GetRoleGrants() error = %v", err)
	}
	if len(grants) != 2 {
		t.Errorf("GetRoleGrants() returned %d grants, want 2", len(grants))
	}
}

func TestGrantStore_GetUserRoles(t *testing.T) {
	roleStore := NewRoleStore()
	store := NewGrantStore(roleStore)

	role1 := &auth.Role{ID: uuid.New(), Name: "admin"}
	role2 := &auth.Role{ID: uuid.New(), Name: "editor"}
	_ = roleStore.Create(context.Background(), role1)
	_ = roleStore.Create(context.Background(), role2)

	userID := uuid.New()
	grant1 := &auth.Grant{UserID: userID, RoleID: role1.ID}
	grant2 := &auth.Grant{UserID: userID, RoleID: role2.ID}
	_ = store.Create(context.Background(), grant1)
	_ = store.Create(context.Background(), grant2)

	roles, err := store.GetUserRoles(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserRoles() error = %v", err)
	}
	if len(roles) != 2 {
		t.Errorf("GetUserRoles() returned %d roles, want 2", len(roles))
	}
}

func TestGrantStore_GetUserRoles_SkipsInvalidRoles(t *testing.T) {
	roleStore := NewRoleStore()
	store := NewGrantStore(roleStore)

	role1 := &auth.Role{ID: uuid.New(), Name: "admin"}
	_ = roleStore.Create(context.Background(), role1)

	userID := uuid.New()
	grant1 := &auth.Grant{UserID: userID, RoleID: role1.ID}
	grant2 := &auth.Grant{UserID: userID, RoleID: uuid.New()}
	_ = store.Create(context.Background(), grant1)
	_ = store.Create(context.Background(), grant2)

	roles, err := store.GetUserRoles(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserRoles() error = %v", err)
	}
	if len(roles) != 1 {
		t.Errorf("GetUserRoles() returned %d roles, want 1 (should skip invalid role)", len(roles))
	}
}

func TestGrantStore_HasRole(t *testing.T) {
	roleStore := NewRoleStore()
	store := NewGrantStore(roleStore)

	role := &auth.Role{ID: uuid.New(), Name: "admin"}
	_ = roleStore.Create(context.Background(), role)

	userID := uuid.New()
	grant := &auth.Grant{UserID: userID, RoleID: role.ID}
	_ = store.Create(context.Background(), grant)

	tests := []struct {
		name     string
		userID   uuid.UUID
		roleName string
		want     bool
	}{
		{
			name:     "user has role",
			userID:   userID,
			roleName: "admin",
			want:     true,
		},
		{
			name:     "user does not have role",
			userID:   userID,
			roleName: "editor",
			want:     false,
		},
		{
			name:     "nonexistent user",
			userID:   uuid.New(),
			roleName: "admin",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.HasRole(context.Background(), tt.userID, tt.roleName)
			if err != nil {
				t.Fatalf("HasRole() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrantStore_HasRole_SkipsInvalidRoles(t *testing.T) {
	roleStore := NewRoleStore()
	store := NewGrantStore(roleStore)

	userID := uuid.New()
	grant := &auth.Grant{UserID: userID, RoleID: uuid.New()}
	_ = store.Create(context.Background(), grant)

	hasRole, err := store.HasRole(context.Background(), userID, "admin")
	if err != nil {
		t.Fatalf("HasRole() error = %v", err)
	}
	if hasRole {
		t.Error("HasRole() = true, want false (should skip invalid role)")
	}
}
