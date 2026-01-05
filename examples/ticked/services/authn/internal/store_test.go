package internal

import (
	"context"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

func TestNewFakeStores(t *testing.T) {
	userStore, roleStore, grantStore := NewFakeStores()

	if userStore == nil {
		t.Fatal("userStore is nil")
	}

	if roleStore == nil {
		t.Fatal("roleStore is nil")
	}

	if grantStore == nil {
		t.Fatal("grantStore is nil")
	}

	// Test that stores are functional
	ctx := context.Background()

	// Create a test user
	user := &auth.User{
		ID:       uuid.New(),
		Username: "testuser",
		Status:   auth.UserStatusActive,
	}

	err := userStore.Create(ctx, user)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Retrieve the user
	retrieved, err := userStore.Get(ctx, user.ID)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	if retrieved.ID != user.ID {
		t.Errorf("expected user ID %v, got %v", user.ID, retrieved.ID)
	}

	// Create a test role
	role := &auth.Role{
		ID:     uuid.New(),
		Name:   "testrole",
		Status: auth.RoleStatusActive,
	}

	err = roleStore.Create(ctx, role)
	if err != nil {
		t.Fatalf("failed to create role: %v", err)
	}

	// Retrieve the role
	retrievedRole, err := roleStore.Get(ctx, role.ID)
	if err != nil {
		t.Fatalf("failed to get role: %v", err)
	}

	if retrievedRole.ID != role.ID {
		t.Errorf("expected role ID %v, got %v", role.ID, retrievedRole.ID)
	}

	// Create a grant
	grant := &auth.Grant{
		UserID: user.ID,
		RoleID: role.ID,
	}

	err = grantStore.Create(ctx, grant)
	if err != nil {
		t.Fatalf("failed to create grant: %v", err)
	}

	// Check if user has the role
	hasRole, err := grantStore.HasRole(ctx, user.ID, role.Name)
	if err != nil {
		t.Fatalf("failed to check role: %v", err)
	}

	if !hasRole {
		t.Error("expected user to have role")
	}
}

func TestNewPostgresStores_InvalidConnectionString(t *testing.T) {
	_, _, _, _, err := NewPostgresStores("invalid connection string")
	if err == nil {
		t.Error("expected error for invalid connection string, got nil")
	}
}

func TestNewPostgresStores_UnreachableHost(t *testing.T) {
	// Use a connection string that points to an unreachable host
	connStr := "host=unreachable.invalid port=5432 user=test password=test dbname=test sslmode=disable connect_timeout=1"
	_, _, _, db, err := NewPostgresStores(connStr)

	if err == nil {
		if db != nil {
			db.Close()
		}
		t.Error("expected error for unreachable host, got nil")
	}
}

func TestNewPostgresStores_Success(t *testing.T) {
	// Skip if no database is available
	// Set DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME env vars for integration tests
	host := "localhost"
	if h := context.Background().Value("DB_HOST"); h != nil {
		host = h.(string)
	}

	port := "5432"
	user := "postgres"
	password := "postgres"
	dbname := "postgres"

	connStr := "host=" + host + " port=" + port + " user=" + user + " password=" + password + " dbname=" + dbname + " sslmode=disable connect_timeout=1"

	userStore, roleStore, grantStore, db, err := NewPostgresStores(connStr)

	// If connection fails, skip the test (no database available)
	if err != nil {
		t.Skip("skipping postgres integration test: no database available")
		return
	}

	defer db.Close()

	if userStore == nil {
		t.Error("userStore is nil")
	}

	if roleStore == nil {
		t.Error("roleStore is nil")
	}

	if grantStore == nil {
		t.Error("grantStore is nil")
	}

	if db == nil {
		t.Error("db is nil")
	}
}
