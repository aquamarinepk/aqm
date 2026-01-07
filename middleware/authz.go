package middleware

import (
	"context"

	"github.com/aquamarinepk/aqm/auth"
)

// AuthzChecker implements RoleChecker using auth.GrantStore.
// NOTE: With username-based grants, this checker uses usernames directly
// instead of parsing UUIDs. The username is the natural key for authorization.
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
func (a *AuthzChecker) HasRole(ctx context.Context, username string, roleName string) (bool, error) {
	if username == "" {
		return false, nil
	}

	roleName = auth.NormalizeRoleName(roleName)
	return a.grantStore.HasRole(ctx, username, roleName)
}

// CheckPermission checks if a user has a specific permission.
func (a *AuthzChecker) CheckPermission(ctx context.Context, username string, permission string) (bool, error) {
	if username == "" {
		return false, nil
	}

	roles, err := a.grantStore.GetUserRoles(ctx, username)
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
func (a *AuthzChecker) CheckAnyPermission(ctx context.Context, username string, permissions []string) (bool, error) {
	if username == "" {
		return false, nil
	}

	roles, err := a.grantStore.GetUserRoles(ctx, username)
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
func (a *AuthzChecker) CheckAllPermissions(ctx context.Context, username string, permissions []string) (bool, error) {
	if username == "" {
		return false, nil
	}

	roles, err := a.grantStore.GetUserRoles(ctx, username)
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
