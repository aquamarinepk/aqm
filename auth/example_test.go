package auth_test

import (
	"context"
	"fmt"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/fake"
	"github.com/aquamarinepk/aqm/auth/seed"
	"github.com/aquamarinepk/aqm/auth/service"
	"github.com/aquamarinepk/aqm/log"
	"github.com/google/uuid"
)

// noopLogger is a no-op implementation of logger.Logger for examples
type noopLogger struct{}

func (l *noopLogger) Debug(v ...interface{})                      {}
func (l *noopLogger) Debugf(format string, args ...interface{})   {}
func (l *noopLogger) Info(v ...interface{})                       {}
func (l *noopLogger) Infof(format string, args ...interface{})    {}
func (l *noopLogger) Error(v ...interface{})                      {}
func (l *noopLogger) Errorf(format string, args ...interface{})   {}
func (l *noopLogger) With(args ...interface{}) logger.Logger { return l }

// Example of manual wiring with fake implementations
func Example_manualWiring() {
	ctx := context.Background()

	// 1. Create crypto services (fake for testing/examples)
	crypto := fake.NewCryptoService()
	tokenGen := fake.NewTokenGenerator()

	// 2. Create stores (fake for testing/examples)
	userStore := fake.NewUserStore()

	// 3. Sign up a new user using the service function
	user, err := service.SignUp(ctx, userStore, crypto, "test@example.com", "Password123!", "testuser", "Test User")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Created user: %s\n", user.Username)

	// 4. Sign in the user
	authUser, token, err := service.SignIn(ctx, userStore, crypto, tokenGen, "test@example.com", "Password123!")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Authenticated: %t\n", authUser != nil)
	fmt.Printf("Token generated: %t\n", token != "")

	// Output:
	// Created user: testuser
	// Authenticated: true
	// Token generated: true
}

// Example of using the seed package
func Example_seedPackage() {
	ctx := context.Background()

	// Setup stores
	userStore := fake.NewUserStore()
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)

	// Create seeder with config for encryption keys
	cfg := &seed.Config{
		EncryptionKey: []byte("12345678901234567890123456789012"),
		SigningKey:    []byte("12345678901234567890123456789012"),
	}
	// For examples, we can use a no-op logger
	seeder := seed.New(userStore, roleStore, grantStore, cfg, &noopLogger{})

	// Seed an admin role
	role, _ := seeder.SeedRole(ctx, seed.RoleInput{
		Name:        "admin",
		Description: "Administrator role",
		Permissions: []string{"users:read", "users:write", "roles:read"},
		CreatedBy:   "system",
	})
	fmt.Printf("Seeded role: %s with %d permissions\n", role.Name, len(role.Permissions))

	// Seed an admin user
	user, err := seeder.SeedUser(ctx, seed.UserInput{
		Username:  "admin",
		Name:      "Administrator",
		Email:     "admin@example.com",
		Password:  "Admin123!",
		PIN:       "000000",
		CreatedBy: "system",
	})
	if err != nil {
		fmt.Printf("Error seeding user: %v\n", err)
		return
	}
	fmt.Printf("Seeded user: %s\n", user.Username)

	// Grant admin role to user
	grant, err := seeder.SeedGrant(ctx, seed.GrantInput{
		UserID:     user.ID,
		RoleID:     role.ID,
		AssignedBy: "system",
	})
	if err != nil {
		fmt.Printf("Error seeding grant: %v\n", err)
		return
	}
	fmt.Printf("Granted role to user: %t\n", grant != nil)

	// Output:
	// Seeded role: admin with 3 permissions
	// Seeded user: admin
	// Granted role to user: true
}

// Example of role-based access control
func Example_roleBasedAccessControl() {
	ctx := context.Background()

	// Setup
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)

	// Create roles with different permissions
	adminRole := &auth.Role{
		ID:          uuid.New(),
		Name:        "admin",
		Permissions: []string{"users:read", "users:write", "users:delete"},
		Status:      auth.RoleStatusActive,
	}
	_ = roleStore.Create(ctx, adminRole)

	editorRole := &auth.Role{
		ID:          uuid.New(),
		Name:        "editor",
		Permissions: []string{"users:read", "users:write"},
		Status:      auth.RoleStatusActive,
	}
	_ = roleStore.Create(ctx, editorRole)

	// Create user and grant editor role
	userID := uuid.New()
	grant := &auth.Grant{
		UserID: userID,
		RoleID: editorRole.ID,
	}
	_ = grantStore.Create(ctx, grant)

	// Check permissions
	hasRead, _ := grantStore.HasRole(ctx, userID, "editor")
	fmt.Printf("User has editor role: %t\n", hasRead)

	// Get all user roles
	roles, _ := grantStore.GetUserRoles(ctx, userID)
	for _, role := range roles {
		fmt.Printf("User role: %s with permissions: %v\n", role.Name, role.Permissions)
	}

	// Output:
	// User has editor role: true
	// User role: editor with permissions: [users:read users:write]
}

// Example of permission checking
func Example_permissionChecking() {
	permissions := []string{"users:read", "users:write", "posts:read"}

	// Check single permission
	hasRead := auth.HasPermission(permissions, "users:read")
	hasDelete := auth.HasPermission(permissions, "users:delete")
	fmt.Printf("Has users:read: %t\n", hasRead)
	fmt.Printf("Has users:delete: %t\n", hasDelete)

	// Check any permission
	hasAny := auth.HasAnyPermission(permissions, []string{"users:delete", "posts:read"})
	fmt.Printf("Has any [users:delete, posts:read]: %t\n", hasAny)

	// Check all permissions
	hasAll := auth.HasAllPermissions(permissions, []string{"users:read", "users:write"})
	hasAllIncludingMissing := auth.HasAllPermissions(permissions, []string{"users:read", "users:delete"})
	fmt.Printf("Has all [users:read, users:write]: %t\n", hasAll)
	fmt.Printf("Has all [users:read, users:delete]: %t\n", hasAllIncludingMissing)

	// Output:
	// Has users:read: true
	// Has users:delete: false
	// Has any [users:delete, posts:read]: true
	// Has all [users:read, users:write]: true
	// Has all [users:read, users:delete]: false
}

// Example of role normalization
func Example_roleNormalization() {
	// Role names are normalized to lowercase
	normalized1 := auth.NormalizeRoleName("Admin")
	normalized2 := auth.NormalizeRoleName("EDITOR")
	normalized3 := auth.NormalizeRoleName("super-admin")

	fmt.Printf("'Admin' normalizes to: '%s'\n", normalized1)
	fmt.Printf("'EDITOR' normalizes to: '%s'\n", normalized2)
	fmt.Printf("'super-admin' normalizes to: '%s'\n", normalized3)

	// Output:
	// 'Admin' normalizes to: 'admin'
	// 'EDITOR' normalizes to: 'editor'
	// 'super-admin' normalizes to: 'super-admin'
}

// Example of user status management
func Example_userStatusManagement() {
	ctx := context.Background()
	userStore := fake.NewUserStore()

	// Create users with different statuses
	activeUser := &auth.User{
		ID:       uuid.New(),
		Username: "active_user",
		Status:   auth.UserStatusActive,
	}
	_ = userStore.Create(ctx, activeUser)

	suspendedUser := &auth.User{
		ID:       uuid.New(),
		Username: "suspended_user",
		Status:   auth.UserStatusSuspended,
	}
	_ = userStore.Create(ctx, suspendedUser)

	// List users by status
	activeUsers, _ := userStore.ListByStatus(ctx, auth.UserStatusActive)
	suspendedUsers, _ := userStore.ListByStatus(ctx, auth.UserStatusSuspended)

	fmt.Printf("Active users: %d\n", len(activeUsers))
	fmt.Printf("Suspended users: %d\n", len(suspendedUsers))

	// Delete user (soft delete - sets status to deleted)
	_ = userStore.Delete(ctx, activeUser.ID)
	user, _ := userStore.Get(ctx, activeUser.ID)
	fmt.Printf("Deleted user status: %s\n", user.Status)

	// Output:
	// Active users: 1
	// Suspended users: 1
	// Deleted user status: deleted
}

// Example of complete authentication flow
func Example_authenticationFlow() {
	ctx := context.Background()

	// Setup complete auth system
	crypto := fake.NewCryptoService()
	tokenGen := fake.NewTokenGenerator()
	userStore := fake.NewUserStore()
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)

	// 1. Register a new user
	user, _ := service.SignUp(ctx, userStore, crypto, "john@example.com", "Secure123!", "john", "John Doe")
	fmt.Printf("Step 1 - Registered user: %s\n", user.Username)

	// 2. Create a role
	role := &auth.Role{
		ID:          uuid.New(),
		Name:        "member",
		Description: "Member role",
		Permissions: []string{"posts:read"},
		CreatedBy:   "system",
	}
	role.BeforeCreate()
	_ = roleStore.Create(ctx, role)
	fmt.Printf("Step 2 - Created role: %s\n", role.Name)

	// 3. Grant role to user
	grant := auth.NewGrant(user.ID, role.ID, "system")
	_ = grantStore.Create(ctx, grant)
	fmt.Printf("Step 3 - Granted role to user\n")

	// 4. Sign in user
	authUser, token, _ := service.SignIn(ctx, userStore, crypto, tokenGen, "john@example.com", "Secure123!")
	fmt.Printf("Step 4 - Authenticated: %t\n", authUser != nil)
	fmt.Printf("Step 5 - Generated token: %s\n", token[:6])

	// 6. Check user's roles
	roles, _ := grantStore.GetUserRoles(ctx, user.ID)
	fmt.Printf("Step 6 - User has %d role(s)\n", len(roles))

	// 7. Check permissions
	hasRead, _ := grantStore.HasRole(ctx, user.ID, "member")
	fmt.Printf("Step 7 - User has member role: %t\n", hasRead)

	// Output:
	// Step 1 - Registered user: john
	// Step 2 - Created role: member
	// Step 3 - Granted role to user
	// Step 4 - Authenticated: true
	// Step 5 - Generated token: token-
	// Step 6 - User has 1 role(s)
	// Step 7 - User has member role: true
}
