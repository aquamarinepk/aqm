package postgres

import (
	"context"
	"database/sql"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	db, err := sql.Open("postgres", "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable")
	if err != nil {
		t.Skip("PostgreSQL not available, skipping integration tests")
	}

	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Skip("PostgreSQL not available, skipping integration tests")
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
		db.ExecContext(context.Background(), "DROP TABLE IF EXISTS users")
		db.Close()
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

	user := auth.NewUser()
	user.Username = "updateme"
	user.Name = "Original Name"
	user.EmailCT = []byte("encrypted")
	user.EmailIV = []byte("iv")
	user.EmailTag = []byte("tag")
	user.EmailLookup = []byte("lookup_updateme")
	user.PasswordHash = []byte("hash")
	user.PasswordSalt = []byte("salt")
	user.BeforeCreate()
	store.Create(ctx, user)

	user.Name = "Updated Name"
	user.BeforeUpdate()

	err := store.Update(ctx, user)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	retrieved, _ := store.Get(ctx, user.ID)
	if retrieved.Name != "Updated Name" {
		t.Errorf("Update() name = %v, want %v", retrieved.Name, "Updated Name")
	}
}

func TestUserStoreDelete(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	store := NewUserStore(db)
	ctx := context.Background()

	user := auth.NewUser()
	user.Username = "deleteme"
	user.Name = "Delete Me"
	user.EmailCT = []byte("encrypted")
	user.EmailIV = []byte("iv")
	user.EmailTag = []byte("tag")
	user.EmailLookup = []byte("lookup_deleteme")
	user.PasswordHash = []byte("hash")
	user.PasswordSalt = []byte("salt")
	user.BeforeCreate()
	store.Create(ctx, user)

	err := store.Delete(ctx, user.ID)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	retrieved, _ := store.Get(ctx, user.ID)
	if retrieved.Status != "deleted" {
		t.Errorf("Delete() status = %v, want deleted", retrieved.Status)
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
		user.BeforeCreate()
		if err := store.Create(ctx, user); err != nil {
			t.Logf("Create user %d failed: %v", i, err)
			continue
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
