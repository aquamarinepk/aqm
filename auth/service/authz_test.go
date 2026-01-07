package service

import (
	"context"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/fake"
	"github.com/google/uuid"
)

func TestCreateRole(t *testing.T) {
	store := fake.NewRoleStore()
	ctx := context.Background()

	tests := []struct {
		name        string
		roleName    string
		description string
		permissions []string
		createdBy   string
		wantErr     bool
	}{
		{
			name:        "valid role",
			roleName:    "admin",
			description: "Administrator role",
			permissions: []string{"users:read", "users:write"},
			createdBy:   "system",
			wantErr:     false,
		},
		{
			name:        "duplicate role",
			roleName:    "admin",
			description: "Another admin",
			permissions: []string{"users:read"},
			createdBy:   "system",
			wantErr:     true,
		},
		{
			name:        "invalid role name",
			roleName:    "",
			description: "Invalid",
			permissions: []string{"users:read"},
			createdBy:   "system",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role, err := CreateRole(ctx, store, tt.roleName, tt.description, tt.permissions, tt.createdBy)

			if (err != nil) != tt.wantErr {
				t.Errorf("CreateRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if role == nil {
					t.Error("CreateRole() returned nil role")
				}
				if role.Name != tt.roleName {
					t.Errorf("CreateRole() name = %v, want %v", role.Name, tt.roleName)
				}
				if len(role.Permissions) != len(tt.permissions) {
					t.Errorf("CreateRole() permissions count = %v, want %v", len(role.Permissions), len(tt.permissions))
				}
			}
		})
	}
}

func TestGetRoleByID(t *testing.T) {
	store := fake.NewRoleStore()
	ctx := context.Background()

	role, _ := CreateRole(ctx, store, "viewer", "Viewer role", []string{"users:read"}, "system")

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
			name:    "non-existing role",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRoleByID(ctx, store, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRoleByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("GetRoleByID() id = %v, want %v", got.ID, tt.id)
			}
		})
	}
}

func TestGetRoleByName(t *testing.T) {
	store := fake.NewRoleStore()
	ctx := context.Background()

	CreateRole(ctx, store, "editor", "Editor role", []string{"content:write"}, "system")

	tests := []struct {
		name     string
		roleName string
		wantErr  bool
	}{
		{
			name:     "existing role",
			roleName: "editor",
			wantErr:  false,
		},
		{
			name:     "non-existing role",
			roleName: "notfound",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetRoleByName(ctx, store, tt.roleName)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetRoleByName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.Name != tt.roleName {
				t.Errorf("GetRoleByName() name = %v, want %v", got.Name, tt.roleName)
			}
		})
	}
}

func TestListRoles(t *testing.T) {
	store := fake.NewRoleStore()
	ctx := context.Background()

	CreateRole(ctx, store, "role1", "Role 1", []string{"perm1"}, "system")
	CreateRole(ctx, store, "role2", "Role 2", []string{"perm2"}, "system")

	roles, err := ListRoles(ctx, store)
	if err != nil {
		t.Fatalf("ListRoles() error = %v", err)
	}

	if len(roles) < 2 {
		t.Errorf("ListRoles() count = %v, want at least 2", len(roles))
	}
}

func TestListRolesByStatus(t *testing.T) {
	store := fake.NewRoleStore()
	ctx := context.Background()

	CreateRole(ctx, store, "active1", "Active 1", []string{"perm1"}, "system")
	role2, _ := CreateRole(ctx, store, "active2", "Active 2", []string{"perm2"}, "system")

	DeleteRole(ctx, store, role2.ID)

	activeRoles, err := ListRolesByStatus(ctx, store, auth.RoleStatusActive)
	if err != nil {
		t.Fatalf("ListRolesByStatus() error = %v", err)
	}

	if len(activeRoles) < 1 {
		t.Errorf("ListRolesByStatus(active) count = %v, want at least 1", len(activeRoles))
	}
}

func TestUpdateRole(t *testing.T) {
	store := fake.NewRoleStore()
	ctx := context.Background()

	role, _ := CreateRole(ctx, store, "updatable", "Original", []string{"perm1"}, "system")

	role.Description = "Updated"
	err := UpdateRole(ctx, store, role, "admin")
	if err != nil {
		t.Fatalf("UpdateRole() error = %v", err)
	}

	retrieved, _ := GetRoleByID(ctx, store, role.ID)
	if retrieved.Description != "Updated" {
		t.Errorf("UpdateRole() description = %v, want Updated", retrieved.Description)
	}
	if retrieved.UpdatedBy != "admin" {
		t.Errorf("UpdateRole() updatedBy = %v, want admin", retrieved.UpdatedBy)
	}
}

func TestDeleteRole(t *testing.T) {
	store := fake.NewRoleStore()
	ctx := context.Background()

	role, _ := CreateRole(ctx, store, "deletable", "To delete", []string{"perm1"}, "system")

	err := DeleteRole(ctx, store, role.ID)
	if err != nil {
		t.Fatalf("DeleteRole() error = %v", err)
	}

	retrieved, _ := GetRoleByID(ctx, store, role.ID)
	if retrieved.Status != "deleted" {
		t.Errorf("DeleteRole() status = %v, want deleted", retrieved.Status)
	}
}

func TestAssignRole(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	ctx := context.Background()

	role, _ := CreateRole(ctx, roleStore, "assignable", "Assignable", []string{"perm1"}, "system")
	username := "testuser"

	tests := []struct {
		name       string
		username   string
		roleID     uuid.UUID
		assignedBy string
		wantErr    bool
	}{
		{
			name:       "valid assignment",
			username:   username,
			roleID:     role.ID,
			assignedBy: "admin",
			wantErr:    false,
		},
		{
			name:       "duplicate assignment",
			username:   username,
			roleID:     role.ID,
			assignedBy: "admin",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			grant, err := AssignRole(ctx, grantStore, tt.username, tt.roleID, tt.assignedBy)

			if (err != nil) != tt.wantErr {
				t.Errorf("AssignRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if grant == nil {
					t.Error("AssignRole() returned nil grant")
				}
				if grant.Username != tt.username {
					t.Errorf("AssignRole() userID = %v, want %v", grant.Username, tt.username)
				}
				if grant.RoleID != tt.roleID {
					t.Errorf("AssignRole() roleID = %v, want %v", grant.RoleID, tt.roleID)
				}
			}
		})
	}
}

func TestRevokeRole(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	ctx := context.Background()

	role, _ := CreateRole(ctx, roleStore, "revokable", "Revokable", []string{"perm1"}, "system")
	username := "testuser"
	AssignRole(ctx, grantStore, username, role.ID, "admin")

	err := RevokeRole(ctx, grantStore, username, role.ID)
	if err != nil {
		t.Fatalf("RevokeRole() error = %v", err)
	}

	has, _ := HasRole(ctx, grantStore, username, "revokable")
	if has {
		t.Error("RevokeRole() role still assigned after revocation")
	}
}

func TestGetUserRoles(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	ctx := context.Background()

	role1, _ := CreateRole(ctx, roleStore, "role1", "Role 1", []string{"perm1"}, "system")
	role2, _ := CreateRole(ctx, roleStore, "role2", "Role 2", []string{"perm2"}, "system")
	username := "testuser"

	AssignRole(ctx, grantStore, username, role1.ID, "admin")
	AssignRole(ctx, grantStore, username, role2.ID, "admin")

	roles, err := GetUserRoles(ctx, grantStore, username)
	if err != nil {
		t.Fatalf("GetUserRoles() error = %v", err)
	}

	if len(roles) != 2 {
		t.Errorf("GetUserRoles() count = %v, want 2", len(roles))
	}
}

func TestGetUserGrants(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	ctx := context.Background()

	role, _ := CreateRole(ctx, roleStore, "test", "Test", []string{"perm1"}, "system")
	username := "testuser"
	AssignRole(ctx, grantStore, username, role.ID, "admin")

	grants, err := GetUserGrants(ctx, grantStore, username)
	if err != nil {
		t.Fatalf("GetUserGrants() error = %v", err)
	}

	if len(grants) != 1 {
		t.Errorf("GetUserGrants() count = %v, want 1", len(grants))
	}
}

func TestGetRoleGrants(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	ctx := context.Background()

	role, _ := CreateRole(ctx, roleStore, "popular", "Popular", []string{"perm1"}, "system")
	user1 := "testuser1"
	user2 := "testuser2"

	AssignRole(ctx, grantStore, user1, role.ID, "admin")
	AssignRole(ctx, grantStore, user2, role.ID, "admin")

	grants, err := GetRoleGrants(ctx, grantStore, role.ID)
	if err != nil {
		t.Fatalf("GetRoleGrants() error = %v", err)
	}

	if len(grants) != 2 {
		t.Errorf("GetRoleGrants() count = %v, want 2", len(grants))
	}
}

func TestCheckPermission(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	ctx := context.Background()

	role, _ := CreateRole(ctx, roleStore, "perms", "Permissions", []string{"users:read", "users:write"}, "system")
	username := "testuser"
	AssignRole(ctx, grantStore, username, role.ID, "admin")

	tests := []struct {
		name       string
		permission string
		want       bool
	}{
		{
			name:       "has permission",
			permission: "users:read",
			want:       true,
		},
		{
			name:       "does not have permission",
			permission: "admin:delete",
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckPermission(ctx, grantStore, username, tt.permission)
			if err != nil {
				t.Fatalf("CheckPermission() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("CheckPermission() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test with deleted role
	DeleteRole(ctx, roleStore, role.ID)
	got, err := CheckPermission(ctx, grantStore, username, "users:read")
	if err != nil {
		t.Fatalf("CheckPermission() with deleted role error = %v", err)
	}
	if got {
		t.Error("CheckPermission() with deleted role should return false")
	}
}

func TestCheckAnyPermission(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	ctx := context.Background()

	role, _ := CreateRole(ctx, roleStore, "multi", "Multi", []string{"users:read", "content:write"}, "system")
	username := "testuser"
	AssignRole(ctx, grantStore, username, role.ID, "admin")

	tests := []struct {
		name        string
		permissions []string
		want        bool
	}{
		{
			name:        "has one permission",
			permissions: []string{"users:read", "admin:delete"},
			want:        true,
		},
		{
			name:        "has none",
			permissions: []string{"admin:delete", "system:restart"},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckAnyPermission(ctx, grantStore, username, tt.permissions)
			if err != nil {
				t.Fatalf("CheckAnyPermission() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("CheckAnyPermission() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test with deleted role
	DeleteRole(ctx, roleStore, role.ID)
	got, err := CheckAnyPermission(ctx, grantStore, username, []string{"users:read"})
	if err != nil {
		t.Fatalf("CheckAnyPermission() with deleted role error = %v", err)
	}
	if got {
		t.Error("CheckAnyPermission() with deleted role should return false")
	}
}

func TestCheckAllPermissions(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	ctx := context.Background()

	role, _ := CreateRole(ctx, roleStore, "complete", "Complete", []string{"users:read", "users:write", "content:read"}, "system")
	username := "testuser"
	AssignRole(ctx, grantStore, username, role.ID, "admin")

	tests := []struct {
		name        string
		permissions []string
		want        bool
	}{
		{
			name:        "has all",
			permissions: []string{"users:read", "users:write"},
			want:        true,
		},
		{
			name:        "missing one",
			permissions: []string{"users:read", "admin:delete"},
			want:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckAllPermissions(ctx, grantStore, username, tt.permissions)
			if err != nil {
				t.Fatalf("CheckAllPermissions() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("CheckAllPermissions() = %v, want %v", got, tt.want)
			}
		})
	}

	// Test with deleted role
	DeleteRole(ctx, roleStore, role.ID)
	got, err := CheckAllPermissions(ctx, grantStore, username, []string{"users:read"})
	if err != nil {
		t.Fatalf("CheckAllPermissions() with deleted role error = %v", err)
	}
	if got {
		t.Error("CheckAllPermissions() with deleted role should return false")
	}
}

func TestHasRole(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	ctx := context.Background()

	role, _ := CreateRole(ctx, roleStore, "checker", "Checker", []string{"check:perm"}, "system")
	username := "testuser"
	AssignRole(ctx, grantStore, username, role.ID, "admin")

	tests := []struct {
		name     string
		roleName string
		want     bool
	}{
		{
			name:     "has role",
			roleName: "checker",
			want:     true,
		},
		{
			name:     "does not have role",
			roleName: "nonexistent",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := HasRole(ctx, grantStore, username, tt.roleName)
			if err != nil {
				t.Fatalf("HasRole() error = %v", err)
			}

			if got != tt.want {
				t.Errorf("HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}
