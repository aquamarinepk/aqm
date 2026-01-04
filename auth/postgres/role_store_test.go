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
