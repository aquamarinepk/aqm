package postgres

import (
	"context"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
)

func setupRoleTestDB(t *testing.T) (*roleStore, func()) {
	t.Helper()

	db, cleanup := setupTestDB(t)

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS roles (
			id UUID PRIMARY KEY,
			name TEXT UNIQUE NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			permissions JSONB NOT NULL DEFAULT '[]'::jsonb,
			status TEXT NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by TEXT NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_by TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("failed to create roles table: %v", err)
	}

	store := NewRoleStore(db).(*roleStore)

	return store, func() {
		db.Exec("DROP TABLE IF EXISTS roles")
		cleanup()
	}
}

func TestRoleStoreCreate(t *testing.T) {
	store, cleanup := setupRoleTestDB(t)
	defer cleanup()

	ctx := context.Background()
	role := auth.NewRole()
	role.Name = "admin"
	role.Description = "Administrator"
	role.Permissions = []string{"*"}
	role.Status = auth.RoleStatusActive
	role.CreatedBy = "system"
	role.UpdatedBy = "system"
	role.BeforeCreate()

	err := store.Create(ctx, role)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	retrieved, err := store.Get(ctx, role.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Name != role.Name {
		t.Errorf("Name = %v, want %v", retrieved.Name, role.Name)
	}
}

func TestRoleStoreGetByName(t *testing.T) {
	store, cleanup := setupRoleTestDB(t)
	defer cleanup()

	ctx := context.Background()
	role := auth.NewRole()
	role.Name = "editor"
	role.Permissions = []string{"content:*"}
	role.CreatedBy = "system"
	role.UpdatedBy = "system"
	role.BeforeCreate()
	store.Create(ctx, role)

	tests := []struct {
		name    string
		rname   string
		wantErr error
	}{
		{
			name:    "existing role",
			rname:   "editor",
			wantErr: nil,
		},
		{
			name:    "non-existing role",
			rname:   "notfound",
			wantErr: auth.ErrRoleNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByName(ctx, tt.rname)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("GetByName() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByName() unexpected error = %v", err)
			}

			if got.Name != tt.rname {
				t.Errorf("GetByName() name = %v, want %v", got.Name, tt.rname)
			}
		})
	}
}

func TestRoleStoreUpdate(t *testing.T) {
	store, cleanup := setupRoleTestDB(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() *auth.Role
		wantErr error
	}{
		{
			name: "successful update",
			setup: func() *auth.Role {
				role := auth.NewRole()
				role.Name = "moderator"
				role.Description = "Moderator"
				role.Permissions = []string{"content:read"}
				role.CreatedBy = "system"
				role.UpdatedBy = "system"
				role.BeforeCreate()
				store.Create(ctx, role)
				return role
			},
			wantErr: nil,
		},
		{
			name: "update non-existent role",
			setup: func() *auth.Role {
				role := auth.NewRole()
				role.Name = "nonexistent"
				role.Description = "Does not exist"
				role.Permissions = []string{"none"}
				role.CreatedBy = "system"
				role.UpdatedBy = "system"
				role.BeforeCreate()
				return role
			},
			wantErr: auth.ErrRoleNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := tt.setup()
			role.Description = "Updated Moderator"
			role.Permissions = []string{"content:read", "content:update"}
			role.UpdatedBy = "admin"
			role.BeforeUpdate()

			err := store.Update(ctx, role)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Update() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}

			retrieved, err := store.Get(ctx, role.ID)
			if err != nil {
				t.Fatalf("Get() error = %v", err)
			}

			if retrieved.Description != "Updated Moderator" {
				t.Errorf("Description = %v, want %v", retrieved.Description, "Updated Moderator")
			}

			if len(retrieved.Permissions) != 2 {
				t.Errorf("len(Permissions) = %v, want %v", len(retrieved.Permissions), 2)
			}

			if retrieved.UpdatedBy != "admin" {
				t.Errorf("UpdatedBy = %v, want %v", retrieved.UpdatedBy, "admin")
			}
		})
	}
}

func TestRoleStoreDelete(t *testing.T) {
	store, cleanup := setupRoleTestDB(t)
	defer cleanup()

	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() *auth.Role
		wantErr error
	}{
		{
			name: "successful delete",
			setup: func() *auth.Role {
				role := auth.NewRole()
				role.Name = "temp"
				role.CreatedBy = "system"
				role.UpdatedBy = "system"
				role.BeforeCreate()
				store.Create(ctx, role)
				return role
			},
			wantErr: nil,
		},
		{
			name: "delete non-existent role",
			setup: func() *auth.Role {
				role := auth.NewRole()
				role.BeforeCreate()
				return role
			},
			wantErr: auth.ErrRoleNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := tt.setup()

			err := store.Delete(ctx, role.ID)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Delete() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Delete() error = %v", err)
			}

			// Delete is a soft delete, so role should still exist but with status='inactive'
			retrieved, err := store.Get(ctx, role.ID)
			if err != nil {
				t.Fatalf("Get() after Delete() error = %v, want nil", err)
			}
			if retrieved.Status != auth.RoleStatusInactive {
				t.Errorf("Get() after Delete() status = %v, want %v", retrieved.Status, auth.RoleStatusInactive)
			}
		})
	}
}

func TestRoleStoreList(t *testing.T) {
	store, cleanup := setupRoleTestDB(t)
	defer cleanup()

	ctx := context.Background()

	roles := []*auth.Role{
		{Name: "admin", Description: "Admin", Permissions: []string{"*"}, Status: auth.RoleStatusActive, CreatedBy: "system", UpdatedBy: "system"},
		{Name: "editor", Description: "Editor", Permissions: []string{"content:*"}, Status: auth.RoleStatusActive, CreatedBy: "system", UpdatedBy: "system"},
		{Name: "viewer", Description: "Viewer", Permissions: []string{"content:read"}, Status: auth.RoleStatusActive, CreatedBy: "system", UpdatedBy: "system"},
	}

	for _, role := range roles {
		role.BeforeCreate()
		if err := store.Create(ctx, role); err != nil {
			t.Fatalf("Create() error = %v", err)
		}
	}

	retrieved, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(retrieved) != 3 {
		t.Errorf("List() returned %d roles, want 3", len(retrieved))
	}
}

func TestRoleStoreListByStatus(t *testing.T) {
	store, cleanup := setupRoleTestDB(t)
	defer cleanup()

	ctx := context.Background()

	activeRole := &auth.Role{Name: "active", Status: auth.RoleStatusActive, CreatedBy: "system", UpdatedBy: "system"}
	activeRole.BeforeCreate()
	store.Create(ctx, activeRole)

	inactiveRole := &auth.Role{Name: "inactive", Status: auth.RoleStatusInactive, CreatedBy: "system", UpdatedBy: "system"}
	inactiveRole.BeforeCreate()
	store.Create(ctx, inactiveRole)

	tests := []struct {
		name      string
		status    auth.RoleStatus
		wantCount int
	}{
		{
			name:      "active roles",
			status:    auth.RoleStatusActive,
			wantCount: 1,
		},
		{
			name:      "inactive roles",
			status:    auth.RoleStatusInactive,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles, err := store.ListByStatus(ctx, tt.status)
			if err != nil {
				t.Fatalf("ListByStatus() error = %v", err)
			}

			if len(roles) != tt.wantCount {
				t.Errorf("ListByStatus() returned %d roles, want %d", len(roles), tt.wantCount)
			}
		})
	}
}
