package service

import (
	"context"
	"fmt"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

// CreateRole creates a new role with permissions
func CreateRole(ctx context.Context, store auth.RoleStore, name, description string, permissions []string, createdBy string) (*auth.Role, error) {
	if store == nil {
		return nil, fmt.Errorf("role store is required")
	}

	name = auth.NormalizeRoleName(name)
	if err := auth.ValidateRoleName(name); err != nil {
		return nil, err
	}

	// Check if role already exists
	existing, err := store.GetByName(ctx, name)
	if err != nil && err != auth.ErrRoleNotFound {
		return nil, fmt.Errorf("check existing role: %w", err)
	}
	if existing != nil {
		return nil, auth.ErrRoleAlreadyExists
	}

	role := auth.NewRole()
	role.Name = name
	role.Description = description
	role.Permissions = permissions
	role.Status = auth.RoleStatusActive
	role.CreatedBy = createdBy
	role.UpdatedBy = createdBy

	role.BeforeCreate()

	if err := store.Create(ctx, role); err != nil {
		return nil, fmt.Errorf("create role: %w", err)
	}

	return role, nil
}

// GetRoleByID retrieves a role by ID
func GetRoleByID(ctx context.Context, store auth.RoleStore, id uuid.UUID) (*auth.Role, error) {
	if store == nil {
		return nil, fmt.Errorf("role store is required")
	}
	return store.Get(ctx, id)
}

// GetRoleByName retrieves a role by name
func GetRoleByName(ctx context.Context, store auth.RoleStore, name string) (*auth.Role, error) {
	if store == nil {
		return nil, fmt.Errorf("role store is required")
	}
	name = auth.NormalizeRoleName(name)
	return store.GetByName(ctx, name)
}

// ListRoles retrieves all roles
func ListRoles(ctx context.Context, store auth.RoleStore) ([]*auth.Role, error) {
	if store == nil {
		return nil, fmt.Errorf("role store is required")
	}
	return store.List(ctx)
}

// ListRolesByStatus retrieves roles by status
func ListRolesByStatus(ctx context.Context, store auth.RoleStore, status auth.RoleStatus) ([]*auth.Role, error) {
	if store == nil {
		return nil, fmt.Errorf("role store is required")
	}
	return store.ListByStatus(ctx, status)
}

// UpdateRole updates a role's information
func UpdateRole(ctx context.Context, store auth.RoleStore, role *auth.Role, updatedBy string) error {
	if store == nil {
		return fmt.Errorf("role store is required")
	}
	if role == nil {
		return fmt.Errorf("role is required")
	}
	role.UpdatedBy = updatedBy
	role.BeforeUpdate()
	return store.Update(ctx, role)
}

// DeleteRole soft-deletes a role
func DeleteRole(ctx context.Context, store auth.RoleStore, id uuid.UUID) error {
	if store == nil {
		return fmt.Errorf("role store is required")
	}
	return store.Delete(ctx, id)
}

// AssignRole assigns a role to a user
func AssignRole(ctx context.Context, store auth.GrantStore, userID, roleID uuid.UUID, assignedBy string) (*auth.Grant, error) {
	if store == nil {
		return nil, fmt.Errorf("grant store is required")
	}

	// Check if grant already exists
	grants, err := store.GetUserGrants(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("check existing grants: %w", err)
	}
	for _, g := range grants {
		if g.RoleID == roleID {
			return nil, auth.ErrGrantAlreadyExists
		}
	}

	grant := auth.NewGrant(userID, roleID, assignedBy)

	if err := store.Create(ctx, grant); err != nil {
		return nil, fmt.Errorf("create grant: %w", err)
	}

	return grant, nil
}

// RevokeRole removes a role from a user
func RevokeRole(ctx context.Context, store auth.GrantStore, userID, roleID uuid.UUID) error {
	if store == nil {
		return fmt.Errorf("grant store is required")
	}
	return store.Delete(ctx, userID, roleID)
}

// GetUserRoles retrieves all roles for a user
func GetUserRoles(ctx context.Context, store auth.GrantStore, userID uuid.UUID) ([]*auth.Role, error) {
	if store == nil {
		return nil, fmt.Errorf("grant store is required")
	}
	return store.GetUserRoles(ctx, userID)
}

// GetUserGrants retrieves all grants for a user
func GetUserGrants(ctx context.Context, store auth.GrantStore, userID uuid.UUID) ([]*auth.Grant, error) {
	if store == nil {
		return nil, fmt.Errorf("grant store is required")
	}
	return store.GetUserGrants(ctx, userID)
}

// GetRoleGrants retrieves all grants for a role
func GetRoleGrants(ctx context.Context, store auth.GrantStore, roleID uuid.UUID) ([]*auth.Grant, error) {
	if store == nil {
		return nil, fmt.Errorf("grant store is required")
	}
	return store.GetRoleGrants(ctx, roleID)
}

// CheckPermission checks if a user has a specific permission
func CheckPermission(ctx context.Context, store auth.GrantStore, userID uuid.UUID, permission string) (bool, error) {
	if store == nil {
		return false, fmt.Errorf("grant store is required")
	}

	roles, err := store.GetUserRoles(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("get user roles: %w", err)
	}

	for _, role := range roles {
		if role.Status != auth.RoleStatusActive {
			continue
		}
		if auth.HasPermission(role.Permissions, permission) {
			return true, nil
		}
	}

	return false, nil
}

// CheckAnyPermission checks if a user has any of the specified permissions
func CheckAnyPermission(ctx context.Context, store auth.GrantStore, userID uuid.UUID, permissions []string) (bool, error) {
	if store == nil {
		return false, fmt.Errorf("grant store is required")
	}

	roles, err := store.GetUserRoles(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("get user roles: %w", err)
	}

	for _, role := range roles {
		if role.Status != auth.RoleStatusActive {
			continue
		}
		if auth.HasAnyPermission(role.Permissions, permissions) {
			return true, nil
		}
	}

	return false, nil
}

// CheckAllPermissions checks if a user has all of the specified permissions
func CheckAllPermissions(ctx context.Context, store auth.GrantStore, userID uuid.UUID, permissions []string) (bool, error) {
	if store == nil {
		return false, fmt.Errorf("grant store is required")
	}

	roles, err := store.GetUserRoles(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("get user roles: %w", err)
	}

	// Collect all user permissions
	allPerms := make([]string, 0)
	for _, role := range roles {
		if role.Status != auth.RoleStatusActive {
			continue
		}
		allPerms = append(allPerms, role.Permissions...)
	}

	return auth.HasAllPermissions(allPerms, permissions), nil
}

// HasRole checks if a user has a specific role by name
func HasRole(ctx context.Context, store auth.GrantStore, userID uuid.UUID, roleName string) (bool, error) {
	if store == nil {
		return false, fmt.Errorf("grant store is required")
	}
	roleName = auth.NormalizeRoleName(roleName)
	return store.HasRole(ctx, userID, roleName)
}
