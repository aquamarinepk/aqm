package fake

import (
	"context"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

func TestNewUserStore(t *testing.T) {
	store := NewUserStore()
	if store == nil {
		t.Fatal("NewUserStore() returned nil")
	}
	if store.users == nil {
		t.Error("users map not initialized")
	}
	if store.usersByUsername == nil {
		t.Error("usersByUsername map not initialized")
	}
	if store.usersByEmailLookup == nil {
		t.Error("usersByEmailLookup map not initialized")
	}
	if store.usersByPINLookup == nil {
		t.Error("usersByPINLookup map not initialized")
	}
}

func TestUserStore_Create(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*UserStore) *auth.User
		user    *auth.User
		wantErr bool
		errType error
	}{
		{
			name: "create new user",
			setup: func(s *UserStore) *auth.User {
				return nil
			},
			user: &auth.User{
				ID:          uuid.New(),
				Username:    "testuser",
				EmailLookup: []byte("email-lookup"),
				PINLookup:   []byte("pin-lookup"),
			},
			wantErr: false,
		},
		{
			name: "duplicate user ID",
			setup: func(s *UserStore) *auth.User {
				existing := &auth.User{
					ID:       uuid.New(),
					Username: "existing",
				}
				_ = s.Create(context.Background(), existing)
				return existing
			},
			user:    nil,
			wantErr: true,
			errType: auth.ErrUserAlreadyExists,
		},
		{
			name: "duplicate username",
			setup: func(s *UserStore) *auth.User {
				existing := &auth.User{
					ID:       uuid.New(),
					Username: "duplicate",
				}
				_ = s.Create(context.Background(), existing)
				return nil
			},
			user: &auth.User{
				ID:       uuid.New(),
				Username: "duplicate",
			},
			wantErr: true,
			errType: auth.ErrUsernameExists,
		},
		{
			name: "user without email lookup",
			setup: func(s *UserStore) *auth.User {
				return nil
			},
			user: &auth.User{
				ID:          uuid.New(),
				Username:    "nomail",
				EmailLookup: []byte{},
			},
			wantErr: false,
		},
		{
			name: "user without PIN lookup",
			setup: func(s *UserStore) *auth.User {
				return nil
			},
			user: &auth.User{
				ID:        uuid.New(),
				Username:  "nopin",
				PINLookup: []byte{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewUserStore()
			setupUser := tt.setup(store)
			if tt.user == nil && setupUser != nil {
				tt.user = &auth.User{
					ID:       setupUser.ID,
					Username: "different",
				}
			}

			err := store.Create(context.Background(), tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != tt.errType {
				t.Errorf("Create() error = %v, want %v", err, tt.errType)
			}
		})
	}
}

func TestUserStore_Get(t *testing.T) {
	store := NewUserStore()
	user := &auth.User{
		ID:       uuid.New(),
		Username: "testuser",
	}
	_ = store.Create(context.Background(), user)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr bool
	}{
		{
			name:    "existing user",
			id:      user.ID,
			wantErr: false,
		},
		{
			name:    "nonexistent user",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Get(context.Background(), tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("Get() returned wrong user")
			}
		})
	}
}

func TestUserStore_GetByEmailLookup(t *testing.T) {
	store := NewUserStore()
	user := &auth.User{
		ID:          uuid.New(),
		Username:    "testuser",
		EmailLookup: []byte("email-lookup"),
	}
	_ = store.Create(context.Background(), user)

	tests := []struct {
		name    string
		lookup  []byte
		wantErr bool
	}{
		{
			name:    "existing email lookup",
			lookup:  []byte("email-lookup"),
			wantErr: false,
		},
		{
			name:    "nonexistent email lookup",
			lookup:  []byte("nonexistent"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByEmailLookup(context.Background(), tt.lookup)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByEmailLookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got.EmailLookup) != string(tt.lookup) {
				t.Errorf("GetByEmailLookup() returned wrong user")
			}
		})
	}
}

func TestUserStore_GetByUsername(t *testing.T) {
	store := NewUserStore()
	user := &auth.User{
		ID:       uuid.New(),
		Username: "testuser",
	}
	_ = store.Create(context.Background(), user)

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "existing username",
			username: "testuser",
			wantErr:  false,
		},
		{
			name:     "nonexistent username",
			username: "nonexistent",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByUsername(context.Background(), tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.Username != tt.username {
				t.Errorf("GetByUsername() returned wrong user")
			}
		})
	}
}

func TestUserStore_GetByPINLookup(t *testing.T) {
	store := NewUserStore()
	user := &auth.User{
		ID:        uuid.New(),
		Username:  "testuser",
		PINLookup: []byte("pin-lookup"),
	}
	_ = store.Create(context.Background(), user)

	tests := []struct {
		name    string
		lookup  []byte
		wantErr bool
	}{
		{
			name:    "existing PIN lookup",
			lookup:  []byte("pin-lookup"),
			wantErr: false,
		},
		{
			name:    "nonexistent PIN lookup",
			lookup:  []byte("nonexistent"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByPINLookup(context.Background(), tt.lookup)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByPINLookup() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got.PINLookup) != string(tt.lookup) {
				t.Errorf("GetByPINLookup() returned wrong user")
			}
		})
	}
}

func TestUserStore_Update(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*UserStore) *auth.User
		update  func(*auth.User)
		wantErr bool
	}{
		{
			name: "update existing user",
			setup: func(s *UserStore) *auth.User {
				user := &auth.User{
					ID:       uuid.New(),
					Username: "testuser",
				}
				_ = s.Create(context.Background(), user)
				return user
			},
			update: func(u *auth.User) {
				u.Username = "updated"
			},
			wantErr: false,
		},
		{
			name: "update user with new email lookup",
			setup: func(s *UserStore) *auth.User {
				user := &auth.User{
					ID:       uuid.New(),
					Username: "testuser",
				}
				_ = s.Create(context.Background(), user)
				return user
			},
			update: func(u *auth.User) {
				u.EmailLookup = []byte("new-email-lookup")
			},
			wantErr: false,
		},
		{
			name: "update user with new PIN lookup",
			setup: func(s *UserStore) *auth.User {
				user := &auth.User{
					ID:       uuid.New(),
					Username: "testuser",
				}
				_ = s.Create(context.Background(), user)
				return user
			},
			update: func(u *auth.User) {
				u.PINLookup = []byte("new-pin-lookup")
			},
			wantErr: false,
		},
		{
			name: "update nonexistent user",
			setup: func(s *UserStore) *auth.User {
				return &auth.User{
					ID:       uuid.New(),
					Username: "nonexistent",
				}
			},
			update:  func(u *auth.User) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewUserStore()
			user := tt.setup(store)
			tt.update(user)

			err := store.Update(context.Background(), user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestUserStore_Delete(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*UserStore) uuid.UUID
		wantErr bool
	}{
		{
			name: "delete existing user",
			setup: func(s *UserStore) uuid.UUID {
				user := &auth.User{
					ID:       uuid.New(),
					Username: "testuser",
					Status:   auth.UserStatusActive,
				}
				_ = s.Create(context.Background(), user)
				return user.ID
			},
			wantErr: false,
		},
		{
			name: "delete nonexistent user",
			setup: func(s *UserStore) uuid.UUID {
				return uuid.New()
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewUserStore()
			id := tt.setup(store)

			err := store.Delete(context.Background(), id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				user, _ := store.Get(context.Background(), id)
				if user.Status != "deleted" {
					t.Error("Delete() did not set status to deleted")
				}
			}
		})
	}
}

func TestUserStore_List(t *testing.T) {
	store := NewUserStore()

	user1 := &auth.User{ID: uuid.New(), Username: "user1"}
	user2 := &auth.User{ID: uuid.New(), Username: "user2"}
	_ = store.Create(context.Background(), user1)
	_ = store.Create(context.Background(), user2)

	users, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}
	if len(users) != 2 {
		t.Errorf("List() returned %d users, want 2", len(users))
	}
}

func TestUserStore_ListByStatus(t *testing.T) {
	store := NewUserStore()

	activeUser := &auth.User{ID: uuid.New(), Username: "active", Status: auth.UserStatusActive}
	suspendedUser := &auth.User{ID: uuid.New(), Username: "suspended", Status: auth.UserStatusSuspended}
	_ = store.Create(context.Background(), activeUser)
	_ = store.Create(context.Background(), suspendedUser)

	tests := []struct {
		name       string
		status     auth.UserStatus
		wantCount  int
	}{
		{
			name:       "list active users",
			status:     auth.UserStatusActive,
			wantCount:  1,
		},
		{
			name:       "list suspended users",
			status:     auth.UserStatusSuspended,
			wantCount:  1,
		},
		{
			name:       "list pending users",
			status:     auth.UserStatusPending,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users, err := store.ListByStatus(context.Background(), tt.status)
			if err != nil {
				t.Fatalf("ListByStatus() error = %v", err)
			}
			if len(users) != tt.wantCount {
				t.Errorf("ListByStatus() returned %d users, want %d", len(users), tt.wantCount)
			}
		})
	}
}
