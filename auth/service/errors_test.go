package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

// Test nil parameter cases for authn functions
func TestAuthnNilParameters(t *testing.T) {
	ctx := context.Background()

	t.Run("SignUp nil store", func(t *testing.T) {
		_, err := SignUp(ctx, nil, nil, "test@example.com", "Password123!", "test", "Test")
		if err == nil {
			t.Error("SignUp() with nil store should error")
		}
	})

	t.Run("SignUp nil crypto", func(t *testing.T) {
		_, err := SignUp(ctx, nil, nil, "test@example.com", "Password123!", "test", "Test")
		if err == nil {
			t.Error("SignUp() with nil crypto should error")
		}
	})

	t.Run("SignIn nil store", func(t *testing.T) {
		_, _, err := SignIn(ctx, nil, nil, nil, "test@example.com", "password")
		if err == nil {
			t.Error("SignIn() with nil store should error")
		}
	})

	t.Run("SignIn nil crypto", func(t *testing.T) {
		_, _, err := SignIn(ctx, nil, nil, nil, "test@example.com", "password")
		if err == nil {
			t.Error("SignIn() with nil crypto should error")
		}
	})

	t.Run("SignIn nil tokenGen", func(t *testing.T) {
		_, _, err := SignIn(ctx, nil, nil, nil, "test@example.com", "password")
		if err == nil {
			t.Error("SignIn() with nil tokenGen should error")
		}
	})

	t.Run("SignInByPIN nil store", func(t *testing.T) {
		_, err := SignInByPIN(ctx, nil, nil, "1234")
		if err == nil {
			t.Error("SignInByPIN() with nil store should error")
		}
	})

	t.Run("SignInByPIN nil crypto", func(t *testing.T) {
		_, err := SignInByPIN(ctx, nil, nil, "1234")
		if err == nil {
			t.Error("SignInByPIN() with nil crypto should error")
		}
	})

	t.Run("Bootstrap nil store", func(t *testing.T) {
		_, _, err := Bootstrap(ctx, nil, nil, nil)
		if err == nil {
			t.Error("Bootstrap() with nil store should error")
		}
	})

	t.Run("Bootstrap nil crypto", func(t *testing.T) {
		_, _, err := Bootstrap(ctx, nil, nil, nil)
		if err == nil {
			t.Error("Bootstrap() with nil crypto should error")
		}
	})

	t.Run("Bootstrap nil pwdGen", func(t *testing.T) {
		_, _, err := Bootstrap(ctx, nil, nil, nil)
		if err == nil {
			t.Error("Bootstrap() with nil pwdGen should error")
		}
	})

	t.Run("GeneratePIN nil store", func(t *testing.T) {
		_, err := GeneratePIN(ctx, nil, nil, nil, nil)
		if err == nil {
			t.Error("GeneratePIN() with nil store should error")
		}
	})

	t.Run("GeneratePIN nil crypto", func(t *testing.T) {
		_, err := GeneratePIN(ctx, nil, nil, nil, nil)
		if err == nil {
			t.Error("GeneratePIN() with nil crypto should error")
		}
	})

	t.Run("GeneratePIN nil pinGen", func(t *testing.T) {
		_, err := GeneratePIN(ctx, nil, nil, nil, nil)
		if err == nil {
			t.Error("GeneratePIN() with nil pinGen should error")
		}
	})

	t.Run("GeneratePIN nil user", func(t *testing.T) {
		_, err := GeneratePIN(ctx, nil, nil, nil, nil)
		if err == nil {
			t.Error("GeneratePIN() with nil user should error")
		}
	})

	t.Run("GetUserByID nil store", func(t *testing.T) {
		_, err := GetUserByID(ctx, nil, uuid.New())
		if err == nil {
			t.Error("GetUserByID() with nil store should error")
		}
	})

	t.Run("GetUserByUsername nil store", func(t *testing.T) {
		_, err := GetUserByUsername(ctx, nil, "test")
		if err == nil {
			t.Error("GetUserByUsername() with nil store should error")
		}
	})

	t.Run("ListUsers nil store", func(t *testing.T) {
		_, err := ListUsers(ctx, nil)
		if err == nil {
			t.Error("ListUsers() with nil store should error")
		}
	})

	t.Run("ListUsersByStatus nil store", func(t *testing.T) {
		_, err := ListUsersByStatus(ctx, nil, "active")
		if err == nil {
			t.Error("ListUsersByStatus() with nil store should error")
		}
	})

	t.Run("UpdateUser nil store", func(t *testing.T) {
		err := UpdateUser(ctx, nil, nil)
		if err == nil {
			t.Error("UpdateUser() with nil store should error")
		}
	})

	t.Run("UpdateUser nil user", func(t *testing.T) {
		err := UpdateUser(ctx, nil, nil)
		if err == nil {
			t.Error("UpdateUser() with nil user should error")
		}
	})

	t.Run("DeleteUser nil store", func(t *testing.T) {
		err := DeleteUser(ctx, nil, uuid.New())
		if err == nil {
			t.Error("DeleteUser() with nil store should error")
		}
	})
}

// Test nil parameter cases for authz functions
func TestAuthzNilParameters(t *testing.T) {
	ctx := context.Background()

	t.Run("CreateRole nil store", func(t *testing.T) {
		_, err := CreateRole(ctx, nil, "test", "desc", []string{}, "user")
		if err == nil {
			t.Error("CreateRole() with nil store should error")
		}
	})

	t.Run("GetRoleByID nil store", func(t *testing.T) {
		_, err := GetRoleByID(ctx, nil, uuid.New())
		if err == nil {
			t.Error("GetRoleByID() with nil store should error")
		}
	})

	t.Run("GetRoleByName nil store", func(t *testing.T) {
		_, err := GetRoleByName(ctx, nil, "test")
		if err == nil {
			t.Error("GetRoleByName() with nil store should error")
		}
	})

	t.Run("ListRoles nil store", func(t *testing.T) {
		_, err := ListRoles(ctx, nil)
		if err == nil {
			t.Error("ListRoles() with nil store should error")
		}
	})

	t.Run("ListRolesByStatus nil store", func(t *testing.T) {
		_, err := ListRolesByStatus(ctx, nil, "active")
		if err == nil {
			t.Error("ListRolesByStatus() with nil store should error")
		}
	})

	t.Run("UpdateRole nil store", func(t *testing.T) {
		err := UpdateRole(ctx, nil, nil, "user")
		if err == nil {
			t.Error("UpdateRole() with nil store should error")
		}
	})

	t.Run("UpdateRole nil role", func(t *testing.T) {
		err := UpdateRole(ctx, nil, nil, "user")
		if err == nil {
			t.Error("UpdateRole() with nil role should error")
		}
	})

	t.Run("DeleteRole nil store", func(t *testing.T) {
		err := DeleteRole(ctx, nil, uuid.New())
		if err == nil {
			t.Error("DeleteRole() with nil store should error")
		}
	})

	t.Run("AssignRole nil store", func(t *testing.T) {
		_, err := AssignRole(ctx, nil, "testuser", uuid.New(), "user")
		if err == nil {
			t.Error("AssignRole() with nil store should error")
		}
	})

	t.Run("RevokeRole nil store", func(t *testing.T) {
		err := RevokeRole(ctx, nil, "testuser", uuid.New())
		if err == nil {
			t.Error("RevokeRole() with nil store should error")
		}
	})

	t.Run("GetUserRoles nil store", func(t *testing.T) {
		_, err := GetUserRoles(ctx, nil, "testuser")
		if err == nil {
			t.Error("GetUserRoles() with nil store should error")
		}
	})

	t.Run("GetUserGrants nil store", func(t *testing.T) {
		_, err := GetUserGrants(ctx, nil, "testuser")
		if err == nil {
			t.Error("GetUserGrants() with nil store should error")
		}
	})

	t.Run("GetRoleGrants nil store", func(t *testing.T) {
		_, err := GetRoleGrants(ctx, nil, uuid.New())
		if err == nil {
			t.Error("GetRoleGrants() with nil store should error")
		}
	})

	t.Run("CheckPermission nil store", func(t *testing.T) {
		_, err := CheckPermission(ctx, nil, "testuser", "perm")
		if err == nil {
			t.Error("CheckPermission() with nil store should error")
		}
	})

	t.Run("CheckAnyPermission nil store", func(t *testing.T) {
		_, err := CheckAnyPermission(ctx, nil, "testuser", []string{"perm"})
		if err == nil {
			t.Error("CheckAnyPermission() with nil store should error")
		}
	})

	t.Run("CheckAllPermissions nil store", func(t *testing.T) {
		_, err := CheckAllPermissions(ctx, nil, "testuser", []string{"perm"})
		if err == nil {
			t.Error("CheckAllPermissions() with nil store should error")
		}
	})

	t.Run("HasRole nil store", func(t *testing.T) {
		_, err := HasRole(ctx, nil, "testuser", "role")
		if err == nil {
			t.Error("HasRole() with nil store should error")
		}
	})
}
