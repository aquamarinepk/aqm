package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/auth"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	ctx := context.Background()

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:15-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	connStr, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("failed to get connection string: %v", err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatalf("failed to connect to postgres: %v", err)
	}

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("failed to ping postgres: %v", err)
	}

	// Create tables
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			username TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			email_ct BYTEA NOT NULL,
			email_iv BYTEA NOT NULL,
			email_tag BYTEA NOT NULL,
			email_lookup BYTEA UNIQUE NOT NULL,
			password_hash BYTEA NOT NULL,
			password_salt BYTEA NOT NULL,
			mfa_secret_ct BYTEA,
			pin_ct BYTEA,
			pin_iv BYTEA,
			pin_tag BYTEA,
			pin_lookup BYTEA UNIQUE,
			status TEXT NOT NULL DEFAULT 'active',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			created_by TEXT NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			updated_by TEXT NOT NULL
		)`,
	}

	for _, migration := range migrations {
		if _, err := db.ExecContext(ctx, migration); err != nil {
			t.Fatalf("failed to run migration: %v", err)
		}
	}

	cleanup := func() {
		db.Close()
		if err := postgresContainer.Terminate(context.Background()); err != nil {
			t.Logf("failed to terminate container: %v", err)
		}
	}

	return db, cleanup
}

func TestUserStoreCreate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)
	ctx := context.Background()

	user := auth.NewUser()
	user.Username = "testuser"
	user.Name = "Test User"
	user.EmailCT = []byte("encrypted")
	user.EmailIV = []byte("iv")
	user.EmailTag = []byte("tag")
	user.EmailLookup = []byte("lookup123")
	user.PasswordHash = []byte("hash")
	user.PasswordSalt = []byte("salt")
	user.Status = auth.UserStatusActive
	user.BeforeCreate()

	err := store.Create(ctx, user)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// Verify user was created
	retrieved, err := store.Get(ctx, user.ID)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Username != user.Username {
		t.Errorf("Username = %v, want %v", retrieved.Username, user.Username)
	}
}

func TestUserStoreGet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() *auth.User
		wantErr error
	}{
		{
			name: "existing user",
			setup: func() *auth.User {
				user := auth.NewUser()
				user.Username = "existing"
				user.Name = "Existing User"
				user.EmailCT = []byte("encrypted")
				user.EmailIV = []byte("iv")
				user.EmailTag = []byte("tag")
				user.EmailLookup = []byte("lookup_existing")
				user.PasswordHash = []byte("hash")
				user.PasswordSalt = []byte("salt")
				user.BeforeCreate()
				store.Create(ctx, user)
				return user
			},
			wantErr: nil,
		},
		{
			name: "non-existing user",
			setup: func() *auth.User {
				user := auth.NewUser()
				return user
			},
			wantErr: auth.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.setup()
			got, err := store.Get(ctx, user.ID)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Get() unexpected error = %v", err)
			}

			if got.ID != user.ID {
				t.Errorf("Get() ID = %v, want %v", got.ID, user.ID)
			}
		})
	}
}

func TestUserStoreGetByUsername(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)
	ctx := context.Background()

	user := auth.NewUser()
	user.Username = "findme"
	user.Name = "Find Me"
	user.EmailCT = []byte("encrypted")
	user.EmailIV = []byte("iv")
	user.EmailTag = []byte("tag")
	user.EmailLookup = []byte("lookup_findme")
	user.PasswordHash = []byte("hash")
	user.PasswordSalt = []byte("salt")
	user.BeforeCreate()
	store.Create(ctx, user)

	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "existing username",
			username: "findme",
			wantErr:  nil,
		},
		{
			name:     "non-existing username",
			username: "notfound",
			wantErr:  auth.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByUsername(ctx, tt.username)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByUsername() unexpected error = %v", err)
			}

			if got.Username != tt.username {
				t.Errorf("GetByUsername() username = %v, want %v", got.Username, tt.username)
			}
		})
	}
}

func TestUserStoreUpdate(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() *auth.User
		wantErr error
	}{
		{
			name: "successful update",
			setup: func() *auth.User {
				user := auth.NewUser()
				user.Username = "updateme"
				user.Name = "Original Name"
				user.EmailCT = []byte("encrypted")
				user.EmailIV = []byte("iv")
				user.EmailTag = []byte("tag")
				user.EmailLookup = []byte("lookup_updateme")
				user.PasswordHash = []byte("hash")
				user.PasswordSalt = []byte("salt")
				user.CreatedBy = "test"
				user.UpdatedBy = "test"
				user.BeforeCreate()
				store.Create(ctx, user)
				return user
			},
			wantErr: nil,
		},
		{
			name: "update non-existent user",
			setup: func() *auth.User {
				user := auth.NewUser()
				user.Username = "nonexistent"
				user.Name = "Nonexistent"
				user.EmailCT = []byte("encrypted")
				user.EmailIV = []byte("iv")
				user.EmailTag = []byte("tag")
				user.EmailLookup = []byte("lookup_nonexistent")
				user.PasswordHash = []byte("hash")
				user.PasswordSalt = []byte("salt")
				user.CreatedBy = "test"
				user.UpdatedBy = "test"
				user.BeforeCreate()
				return user
			},
			wantErr: auth.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.setup()
			user.Name = "Updated Name"
			user.BeforeUpdate()

			err := store.Update(ctx, user)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Update() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Update() error = %v", err)
			}

			retrieved, _ := store.Get(ctx, user.ID)
			if retrieved.Name != "Updated Name" {
				t.Errorf("Update() name = %v, want %v", retrieved.Name, "Updated Name")
			}
		})
	}
}

func TestUserStoreDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func() *auth.User
		wantErr error
	}{
		{
			name: "successful delete",
			setup: func() *auth.User {
				user := auth.NewUser()
				user.Username = "deleteme"
				user.Name = "Delete Me"
				user.EmailCT = []byte("encrypted")
				user.EmailIV = []byte("iv")
				user.EmailTag = []byte("tag")
				user.EmailLookup = []byte("lookup_deleteme")
				user.PasswordHash = []byte("hash")
				user.PasswordSalt = []byte("salt")
				user.PINLookup = []byte("pin_deleteme")
				user.CreatedBy = "test"
				user.UpdatedBy = "test"
				user.BeforeCreate()
				store.Create(ctx, user)
				return user
			},
			wantErr: nil,
		},
		{
			name: "delete non-existent user",
			setup: func() *auth.User {
				user := auth.NewUser()
				user.BeforeCreate()
				return user
			},
			wantErr: auth.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := tt.setup()

			err := store.Delete(ctx, user.ID)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("Delete() error = %v, want %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("Delete() error = %v", err)
			}

			retrieved, _ := store.Get(ctx, user.ID)
			if retrieved.Status != "deleted" {
				t.Errorf("Delete() status = %v, want deleted", retrieved.Status)
			}
		})
	}
}

func TestUserStoreList(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)
	ctx := context.Background()

	// Create 3 users
	created := 0
	for i := 0; i < 3; i++ {
		user := auth.NewUser()
		user.Username = "listuser" + string(rune('a'+i))
		user.Name = "User " + string(rune('A'+i))
		user.EmailCT = []byte("encrypted")
		user.EmailIV = []byte("iv")
		user.EmailTag = []byte("tag")
		user.EmailLookup = []byte{byte('l'), byte('i'), byte('s'), byte('t'), byte('a' + i)}
		user.PasswordHash = []byte("hash")
		user.PasswordSalt = []byte("salt")
		user.PINLookup = []byte{byte('p'), byte('i'), byte('n'), byte('a' + i)}
		user.CreatedBy = "test"
		user.UpdatedBy = "test"
		user.BeforeCreate()
		if err := store.Create(ctx, user); err != nil {
			t.Fatalf("Create user %d failed: %v", i, err)
		}
		created++
	}

	users, err := store.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(users) < created {
		t.Errorf("List() count = %v, want at least %v", len(users), created)
	}
}

func TestUserStoreGetByEmailLookup(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)
	ctx := context.Background()

	user := auth.NewUser()
	user.Username = "emailtest"
	user.Name = "Email Test"
	user.EmailCT = []byte("encrypted")
	user.EmailIV = []byte("iv")
	user.EmailTag = []byte("tag")
	user.EmailLookup = []byte("unique_email_lookup_123")
	user.PasswordHash = []byte("hash")
	user.PasswordSalt = []byte("salt")
	user.BeforeCreate()
	store.Create(ctx, user)

	tests := []struct {
		name        string
		emailLookup []byte
		wantErr     error
	}{
		{
			name:        "existing email",
			emailLookup: []byte("unique_email_lookup_123"),
			wantErr:     nil,
		},
		{
			name:        "non-existing email",
			emailLookup: []byte("notfound"),
			wantErr:     auth.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByEmailLookup(ctx, tt.emailLookup)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("GetByEmailLookup() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByEmailLookup() unexpected error = %v", err)
			}

			if got.Username != user.Username {
				t.Errorf("GetByEmailLookup() username = %v, want %v", got.Username, user.Username)
			}
		})
	}
}

func TestUserStoreGetByPINLookup(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)
	ctx := context.Background()

	user := auth.NewUser()
	user.Username = "pintest"
	user.Name = "PIN Test"
	user.EmailCT = []byte("encrypted")
	user.EmailIV = []byte("iv")
	user.EmailTag = []byte("tag")
	user.EmailLookup = []byte("email_lookup_pin")
	user.PasswordHash = []byte("hash")
	user.PasswordSalt = []byte("salt")
	user.PINCT = []byte("encrypted_pin")
	user.PINIV = []byte("pin_iv")
	user.PINTag = []byte("pin_tag")
	user.PINLookup = []byte("unique_pin_lookup_456")
	user.BeforeCreate()
	store.Create(ctx, user)

	tests := []struct {
		name      string
		pinLookup []byte
		wantErr   error
	}{
		{
			name:      "existing pin",
			pinLookup: []byte("unique_pin_lookup_456"),
			wantErr:   nil,
		},
		{
			name:      "non-existing pin",
			pinLookup: []byte("notfound"),
			wantErr:   auth.ErrUserNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByPINLookup(ctx, tt.pinLookup)

			if tt.wantErr != nil {
				if err != tt.wantErr {
					t.Errorf("GetByPINLookup() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}

			if err != nil {
				t.Fatalf("GetByPINLookup() unexpected error = %v", err)
			}

			if got.Username != user.Username {
				t.Errorf("GetByPINLookup() username = %v, want %v", got.Username, user.Username)
			}
		})
	}
}

func TestUserStoreListByStatus(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)
	ctx := context.Background()

	activeUser := auth.NewUser()
	activeUser.Username = "active"
	activeUser.Name = "Active User"
	activeUser.EmailCT = []byte("encrypted")
	activeUser.EmailIV = []byte("iv")
	activeUser.EmailTag = []byte("tag")
	activeUser.EmailLookup = []byte("active_lookup")
	activeUser.PasswordHash = []byte("hash")
	activeUser.PasswordSalt = []byte("salt")
	activeUser.PINLookup = []byte("active_pin_lookup")
	activeUser.Status = auth.UserStatusActive
	activeUser.CreatedBy = "test"
	activeUser.UpdatedBy = "test"
	activeUser.BeforeCreate()
	if err := store.Create(ctx, activeUser); err != nil {
		t.Fatalf("Failed to create active user: %v", err)
	}

	suspendedUser := auth.NewUser()
	suspendedUser.Username = "suspended"
	suspendedUser.Name = "Suspended User"
	suspendedUser.EmailCT = []byte("encrypted2")
	suspendedUser.EmailIV = []byte("iv2")
	suspendedUser.EmailTag = []byte("tag2")
	suspendedUser.EmailLookup = []byte("suspended_lookup")
	suspendedUser.PasswordHash = []byte("hash2")
	suspendedUser.PasswordSalt = []byte("salt2")
	suspendedUser.PINLookup = []byte("suspended_pin_lookup")
	suspendedUser.Status = auth.UserStatusSuspended
	suspendedUser.CreatedBy = "test"
	suspendedUser.UpdatedBy = "test"
	suspendedUser.BeforeCreate()

	if err := store.Create(ctx, suspendedUser); err != nil {
		t.Fatalf("Failed to create suspended user: %v", err)
	}

	tests := []struct {
		name      string
		status    auth.UserStatus
		wantCount int
	}{
		{
			name:      "active users",
			status:    auth.UserStatusActive,
			wantCount: 1,
		},
		{
			name:      "suspended users",
			status:    auth.UserStatusSuspended,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := store.ListByStatus(ctx, tt.status)
			if err != nil {
				t.Fatalf("ListByStatus() error = %v", err)
			}

			if len(users) != tt.wantCount {
				t.Errorf("ListByStatus() returned %d users, want %d", len(users), tt.wantCount)
			}
		})
	}
}
