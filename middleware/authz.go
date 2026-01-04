package middleware

import (
	"context"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

// AuthzChecker implements RoleChecker using auth.GrantStore.
type AuthzChecker struct {
	grantStore auth.GrantStore
}

// NewAuthzChecker creates a new authorization checker.
func NewAuthzChecker(grantStore auth.GrantStore) *AuthzChecker {
	return &AuthzChecker{
		grantStore: grantStore,
	}
}

// HasRole checks if a user has a specific role.
func (a *AuthzChecker) HasRole(ctx context.Context, userID string, roleName string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}

	roleName = auth.NormalizeRoleName(roleName)
	return a.grantStore.HasRole(ctx, uid, roleName)
}

// CheckPermission checks if a user has a specific permission.
func (a *AuthzChecker) CheckPermission(ctx context.Context, userID string, permission string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}

	roles, err := a.grantStore.GetUserRoles(ctx, uid)
	if err != nil {
		return false, err
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

// CheckAnyPermission checks if a user has any of the specified permissions.
func (a *AuthzChecker) CheckAnyPermission(ctx context.Context, userID string, permissions []string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}

	roles, err := a.grantStore.GetUserRoles(ctx, uid)
	if err != nil {
		return false, err
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

// CheckAllPermissions checks if a user has all of the specified permissions.
func (a *AuthzChecker) CheckAllPermissions(ctx context.Context, userID string, permissions []string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}

	roles, err := a.grantStore.GetUserRoles(ctx, uid)
	if err != nil {
		return false, err
	}

	allPerms := make([]string, 0)
	for _, role := range roles {
		if role.Status != auth.RoleStatusActive {
			continue
		}
		allPerms = append(allPerms, role.Permissions...)
	}

	return auth.HasAllPermissions(allPerms, permissions), nil
}
