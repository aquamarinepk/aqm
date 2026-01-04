package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewRole(t *testing.T) {
	role := NewRole()
	if role == nil {
		t.Fatal("NewRole() returned nil")
	}
	if role.Status != RoleStatusActive {
		t.Errorf("NewRole() status = %v, want %v", role.Status, RoleStatusActive)
	}
	if role.Permissions == nil {
		t.Error("NewRole() permissions is nil")
	}
	if len(role.Permissions) != 0 {
		t.Errorf("NewRole() permissions length = %d, want 0", len(role.Permissions))
	}
}

func TestRoleEnsureID(t *testing.T) {
	tests := []struct {
		name    string
		initial uuid.UUID
		wantNil bool
	}{
		{"generates ID when nil", uuid.Nil, false},
		{"keeps existing ID", uuid.New(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &Role{ID: tt.initial}
			role.EnsureID()
			if (role.ID == uuid.Nil) == !tt.wantNil {
				t.Errorf("EnsureID() ID nil = %v, wantNil %v", role.ID == uuid.Nil, tt.wantNil)
			}
			if tt.initial != uuid.Nil && role.ID != tt.initial {
				t.Error("EnsureID() changed existing ID")
			}
		})
	}
}

func TestRoleBeforeCreate(t *testing.T) {
	role := &Role{
		Name:        "  SuperAdmin  ",
		Description: "  Super Administrator  ",
		Permissions: nil,
	}

	role.BeforeCreate()

	if role.ID == uuid.Nil {
		t.Error("BeforeCreate() did not generate ID")
	}
	if role.Name != "superadmin" {
		t.Errorf("BeforeCreate() name = %q, want %q", role.Name, "superadmin")
	}
	if role.Description != "Super Administrator" {
		t.Errorf("BeforeCreate() description = %q, want %q", role.Description, "Super Administrator")
	}
	if role.Permissions == nil {
		t.Error("BeforeCreate() did not initialize permissions")
	}
	if len(role.Permissions) != 0 {
		t.Error("BeforeCreate() permissions should be empty")
	}
	if role.CreatedAt.IsZero() {
		t.Error("BeforeCreate() did not set CreatedAt")
	}
	if role.UpdatedAt.IsZero() {
		t.Error("BeforeCreate() did not set UpdatedAt")
	}
}

func TestRoleBeforeUpdate(t *testing.T) {
	role := &Role{
		Name:        "  Admin  ",
		Description: "  Administrator  ",
		Permissions: nil,
	}

	role.BeforeUpdate()

	if role.Name != "admin" {
		t.Errorf("BeforeUpdate() name = %q, want %q", role.Name, "admin")
	}
	if role.Description != "Administrator" {
		t.Errorf("BeforeUpdate() description = %q, want %q", role.Description, "Administrator")
	}
	if role.Permissions == nil {
		t.Error("BeforeUpdate() did not initialize permissions")
	}
	if role.UpdatedAt.IsZero() {
		t.Error("BeforeUpdate() did not set UpdatedAt")
	}
}

func TestRoleHasPermission(t *testing.T) {
	role := &Role{
		Permissions: []string{"users:read", "users:write"},
	}

	tests := []struct {
		name       string
		permission string
		want       bool
	}{
		{"has permission", "users:read", true},
		{"does not have permission", "users:delete", false},
		{"wildcard", "*", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := role.HasPermission(tt.permission); got != tt.want {
				t.Errorf("HasPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRoleAddPermission(t *testing.T) {
	tests := []struct {
		name       string
		initial    []string
		add        string
		wantCount  int
		wantHas    bool
	}{
		{"add new permission", []string{"users:read"}, "users:write", 2, true},
		{"add duplicate permission", []string{"users:read"}, "users:read", 1, true},
		{"add to empty", []string{}, "users:read", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &Role{Permissions: tt.initial}
			role.AddPermission(tt.add)
			if len(role.Permissions) != tt.wantCount {
				t.Errorf("AddPermission() count = %d, want %d", len(role.Permissions), tt.wantCount)
			}
			if got := role.HasPermission(tt.add); got != tt.wantHas {
				t.Errorf("HasPermission() after add = %v, want %v", got, tt.wantHas)
			}
		})
	}
}

func TestRoleRemovePermission(t *testing.T) {
	tests := []struct {
		name      string
		initial   []string
		remove    string
		wantCount int
		wantHas   bool
	}{
		{"remove existing", []string{"users:read", "users:write"}, "users:read", 1, false},
		{"remove non-existing", []string{"users:read"}, "users:write", 1, false},
		{"remove from empty", []string{}, "users:read", 0, false},
		{"remove all", []string{"users:read"}, "users:read", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			role := &Role{Permissions: tt.initial}
			role.RemovePermission(tt.remove)
			if len(role.Permissions) != tt.wantCount {
				t.Errorf("RemovePermission() count = %d, want %d", len(role.Permissions), tt.wantCount)
			}
			if got := role.HasPermission(tt.remove); got != tt.wantHas {
				t.Errorf("HasPermission() after remove = %v, want %v", got, tt.wantHas)
			}
		})
	}
}

func TestRoleValidate(t *testing.T) {
	tests := []struct {
		name    string
		role    *Role
		wantErr bool
	}{
		{
			"valid role",
			&Role{Name: "admin", Status: RoleStatusActive},
			false,
		},
		{
			"invalid name",
			&Role{Name: "a", Status: RoleStatusActive},
			true,
		},
		{
			"invalid status",
			&Role{Name: "admin", Status: RoleStatus("invalid")},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.role.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
