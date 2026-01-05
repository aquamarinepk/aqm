package postgres

import (
	"context"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

func setupGrantTestDB(t *testing.T) (*grantStore, *roleStore, func()) {
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
		);
		CREATE TABLE IF NOT EXISTS grants (
			id UUID PRIMARY KEY,
			user_id UUID NOT NULL,
			role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
			assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			assigned_by TEXT NOT NULL,
			UNIQUE(user_id, role_id)
		)
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	gstore := NewGrantStore(db).(*grantStore)
	rstore := NewRoleStore(db).(*roleStore)

	return gstore, rstore, func() {
		db.Exec("DROP TABLE IF EXISTS grants")
		db.Exec("DROP TABLE IF EXISTS roles")
		cleanup()
	}
}

func TestGrantStoreCreate(t *testing.T) {
	gstore, rstore, cleanup := setupGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	role := auth.NewRole()
	role.Name = "admin"
	role.CreatedBy = "system"
	role.UpdatedBy = "system"
	role.BeforeCreate()
	rstore.Create(ctx, role)

	userID := uuid.New()
	grant := auth.NewGrant(userID, role.ID, "system")

	err := gstore.Create(ctx, grant)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	grants, _ := gstore.GetUserGrants(ctx, userID)
	if len(grants) != 1 {
		t.Errorf("GetUserGrants() count = %v, want 1", len(grants))
	}
}

func TestGrantStoreHasRole(t *testing.T) {
	gstore, rstore, cleanup := setupGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	role := auth.NewRole()
	role.Name = "viewer"
	role.CreatedBy = "system"
	role.UpdatedBy = "system"
	role.BeforeCreate()
	rstore.Create(ctx, role)

	userID := uuid.New()
	grant := auth.NewGrant(userID, role.ID, "system")
	gstore.Create(ctx, grant)

	has, err := gstore.HasRole(ctx, userID, "viewer")
	if err != nil {
		t.Fatalf("HasRole() error = %v", err)
	}

	if !has {
		t.Errorf("HasRole() = false, want true")
	}
}

func TestGrantStoreDelete(t *testing.T) {
	gstore, rstore, cleanup := setupGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	role := auth.NewRole()
	role.Name = "editor"
	role.CreatedBy = "system"
	role.UpdatedBy = "system"
	role.BeforeCreate()
	rstore.Create(ctx, role)

	userID := uuid.New()
	grant := auth.NewGrant(userID, role.ID, "system")
	gstore.Create(ctx, grant)

	err := gstore.Delete(ctx, userID, role.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	has, err := gstore.HasRole(ctx, userID, "editor")
	if err != nil {
		t.Fatalf("HasRole() error = %v", err)
	}

	if has {
		t.Errorf("HasRole() after Delete() = true, want false")
	}
}

func TestGrantStoreGetRoleGrants(t *testing.T) {
	gstore, rstore, cleanup := setupGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	role := auth.NewRole()
	role.Name = "moderator"
	role.CreatedBy = "system"
	role.UpdatedBy = "system"
	role.BeforeCreate()
	rstore.Create(ctx, role)

	user1 := uuid.New()
	user2 := uuid.New()

	grant1 := auth.NewGrant(user1, role.ID, "system")
	gstore.Create(ctx, grant1)

	grant2 := auth.NewGrant(user2, role.ID, "system")
	gstore.Create(ctx, grant2)

	grants, err := gstore.GetRoleGrants(ctx, role.ID)
	if err != nil {
		t.Fatalf("GetRoleGrants() error = %v", err)
	}

	if len(grants) != 2 {
		t.Errorf("GetRoleGrants() count = %v, want 2", len(grants))
	}
}

func TestGrantStoreGetUserRoles(t *testing.T) {
	gstore, rstore, cleanup := setupGrantTestDB(t)
	defer cleanup()

	ctx := context.Background()

	role1 := &auth.Role{Name: "admin", Status: auth.RoleStatusActive, CreatedBy: "system", UpdatedBy: "system"}
	role1.BeforeCreate()
	rstore.Create(ctx, role1)

	role2 := &auth.Role{Name: "editor", Status: auth.RoleStatusActive, CreatedBy: "system", UpdatedBy: "system"}
	role2.BeforeCreate()
	rstore.Create(ctx, role2)

	userID := uuid.New()

	grant1 := auth.NewGrant(userID, role1.ID, "system")
	gstore.Create(ctx, grant1)

	grant2 := auth.NewGrant(userID, role2.ID, "system")
	gstore.Create(ctx, grant2)

	roles, err := gstore.GetUserRoles(ctx, userID)
	if err != nil {
		t.Fatalf("GetUserRoles() error = %v", err)
	}

	if len(roles) != 2 {
		t.Errorf("GetUserRoles() count = %v, want 2", len(roles))
	}

	roleNames := make(map[string]bool)
	for _, r := range roles {
		roleNames[r.Name] = true
	}

	if !roleNames["admin"] || !roleNames["editor"] {
		t.Errorf("GetUserRoles() missing expected roles")
	}
}
