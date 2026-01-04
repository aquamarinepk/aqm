package seed

import (
	"context"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/fake"
	"github.com/aquamarinepk/aqm/log"
	"github.com/google/uuid"
)

func TestSeeder_SeedRole(t *testing.T) {
	tests := []struct {
		name      string
		input     RoleInput
		wantErr   bool
		checkFunc func(t *testing.T, role *auth.Role)
	}{
		{
			name: "valid role with permissions",
			input: RoleInput{
				Name:        "admin",
				Description: "Administrator role",
				Permissions: []string{"users:read", "users:write"},
				CreatedBy:   "system",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, role *auth.Role) {
				if role.Name != "admin" {
					t.Errorf("expected name=admin, got=%s", role.Name)
				}
				if role.Description != "Administrator role" {
					t.Errorf("expected description='Administrator role', got=%s", role.Description)
				}
				if len(role.Permissions) != 2 {
					t.Errorf("expected 2 permissions, got=%d", len(role.Permissions))
				}
				if role.CreatedBy != "system" {
					t.Errorf("expected created_by=system, got=%s", role.CreatedBy)
				}
				if role.ID == uuid.Nil {
					t.Error("expected ID to be set")
				}
				if role.CreatedAt.IsZero() {
					t.Error("expected CreatedAt to be set")
				}
			},
		},
		{
			name: "valid role without permissions",
			input: RoleInput{
				Name:        "viewer",
				Description: "Read-only viewer",
				Permissions: nil,
				CreatedBy:   "system",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, role *auth.Role) {
				if len(role.Permissions) != 0 {
					t.Errorf("expected empty permissions, got=%d", len(role.Permissions))
				}
			},
		},
		{
			name: "invalid role name",
			input: RoleInput{
				Name:        "",
				Description: "Invalid",
				Permissions: nil,
				CreatedBy:   "system",
			},
			wantErr: true,
		},
		{
			name: "duplicate role name",
			input: RoleInput{
				Name:        "admin",
				Description: "Duplicate admin",
				Permissions: nil,
				CreatedBy:   "system",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleStore := fake.NewRoleStore()
			userStore := fake.NewUserStore()
			grantStore := fake.NewGrantStore(roleStore)

			if tt.name == "duplicate role name" {
				existingRole := auth.NewRole()
				existingRole.Name = "admin"
				existingRole.Description = "Existing admin"
				existingRole.CreatedBy = "system"
				existingRole.UpdatedBy = "system"
				existingRole.BeforeCreate()
				_ = roleStore.Create(context.Background(), existingRole)
			}

			cfg := &Config{
				EncryptionKey: []byte("test-encryption-key-32-bytes!"),
				SigningKey:    []byte("test-signing-key"),
			}
			seeder := New(userStore, roleStore, grantStore, cfg, logger.NewNoopLogger())

			role, err := seeder.SeedRole(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("SeedRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, role)
			}
		})
	}
}

func TestSeeder_SeedUser(t *testing.T) {
	encKey := make([]byte, 32)
	sigKey := make([]byte, 32)
	for i := range encKey {
		encKey[i] = byte(i)
		sigKey[i] = byte(i + 32)
	}

	tests := []struct {
		name      string
		input     UserInput
		wantErr   bool
		checkFunc func(t *testing.T, user *auth.User, cfg *Config)
	}{
		{
			name: "valid user with PIN",
			input: UserInput{
				Username:  "johndoe",
				Name:      "John Doe",
				Email:     "john@example.com",
				Password:  "SecurePass123!",
				PIN:       "1234",
				CreatedBy: "system",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *auth.User, cfg *Config) {
				if user.Username != "johndoe" {
					t.Errorf("expected username=johndoe, got=%s", user.Username)
				}
				if user.Name != "John Doe" {
					t.Errorf("expected name='John Doe', got=%s", user.Name)
				}
				if len(user.EmailCT) == 0 {
					t.Error("expected email to be encrypted")
				}
				if len(user.PasswordHash) == 0 {
					t.Error("expected password to be hashed")
				}
				if len(user.PINCT) == 0 {
					t.Error("expected PIN to be encrypted")
				}

				email, err := user.GetEmail(cfg.EncryptionKey)
				if err != nil {
					t.Fatalf("failed to decrypt email: %v", err)
				}
				if email != "john@example.com" {
					t.Errorf("expected email=john@example.com, got=%s", email)
				}

				if !user.VerifyPassword("SecurePass123!") {
					t.Error("password verification failed")
				}

				if !user.VerifyPIN("1234", cfg.SigningKey) {
					t.Error("PIN verification failed")
				}
			},
		},
		{
			name: "valid user without PIN",
			input: UserInput{
				Username:  "janedoe",
				Name:      "Jane Doe",
				Email:     "jane@example.com",
				Password:  "AnotherPass456!",
				PIN:       "",
				CreatedBy: "system",
			},
			wantErr: false,
			checkFunc: func(t *testing.T, user *auth.User, cfg *Config) {
				if len(user.PINCT) != 0 {
					t.Error("expected PIN to be empty")
				}
			},
		},
		{
			name: "invalid email",
			input: UserInput{
				Username:  "baduser",
				Name:      "Bad User",
				Email:     "not-an-email",
				Password:  "ValidPass123!",
				PIN:       "",
				CreatedBy: "system",
			},
			wantErr: true,
		},
		{
			name: "invalid password",
			input: UserInput{
				Username:  "weakuser",
				Name:      "Weak User",
				Email:     "weak@example.com",
				Password:  "123",
				PIN:       "",
				CreatedBy: "system",
			},
			wantErr: true,
		},
		{
			name: "invalid username",
			input: UserInput{
				Username:  "",
				Name:      "No Username",
				Email:     "nouser@example.com",
				Password:  "ValidPass123!",
				PIN:       "",
				CreatedBy: "system",
			},
			wantErr: true,
		},
		{
			name: "duplicate username",
			input: UserInput{
				Username:  "johndoe",
				Name:      "Another John",
				Email:     "john2@example.com",
				Password:  "ValidPass123!",
				PIN:       "",
				CreatedBy: "system",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleStore := fake.NewRoleStore()
			userStore := fake.NewUserStore()
			grantStore := fake.NewGrantStore(roleStore)

			if tt.name == "duplicate username" {
				existingUser := auth.NewUser()
				existingUser.Username = "johndoe"
				existingUser.Name = "John Doe"
				existingUser.CreatedBy = "system"
				existingUser.UpdatedBy = "system"
				_ = existingUser.SetEmail("existing@example.com", encKey, sigKey)
				_ = existingUser.SetPassword("ExistingPass123!")
				existingUser.BeforeCreate()
				_ = userStore.Create(context.Background(), existingUser)
			}

			cfg := &Config{
				EncryptionKey: encKey,
				SigningKey:    sigKey,
			}
			seeder := New(userStore, roleStore, grantStore, cfg, logger.NewNoopLogger())

			user, err := seeder.SeedUser(context.Background(), tt.input)

			if (err != nil) != tt.wantErr {
				t.Errorf("SeedUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, user, cfg)
			}
		})
	}
}

func TestSeeder_SeedGrant(t *testing.T) {
	encKey := make([]byte, 32)
	sigKey := make([]byte, 32)
	for i := range encKey {
		encKey[i] = byte(i)
		sigKey[i] = byte(i + 32)
	}

	tests := []struct {
		name       string
		setupFunc  func(ctx context.Context, seeder *Seeder) (userID, roleID uuid.UUID)
		assignedBy string
		wantErr    bool
		checkFunc  func(t *testing.T, grant *auth.Grant, userID, roleID uuid.UUID)
	}{
		{
			name: "valid grant",
			setupFunc: func(ctx context.Context, seeder *Seeder) (uuid.UUID, uuid.UUID) {
				user, _ := seeder.SeedUser(ctx, UserInput{
					Username:  "testuser",
					Name:      "Test User",
					Email:     "test@example.com",
					Password:  "TestPass123!",
					CreatedBy: "system",
				})
				role, _ := seeder.SeedRole(ctx, RoleInput{
					Name:        "admin",
					Description: "Admin role",
					CreatedBy:   "system",
				})
				return user.ID, role.ID
			},
			assignedBy: "system",
			wantErr:    false,
			checkFunc: func(t *testing.T, grant *auth.Grant, userID, roleID uuid.UUID) {
				if grant.UserID != userID {
					t.Errorf("expected user_id=%s, got=%s", userID, grant.UserID)
				}
				if grant.RoleID != roleID {
					t.Errorf("expected role_id=%s, got=%s", roleID, grant.RoleID)
				}
				if grant.AssignedBy != "system" {
					t.Errorf("expected assigned_by=system, got=%s", grant.AssignedBy)
				}
				if grant.ID == uuid.Nil {
					t.Error("expected ID to be set")
				}
				if grant.AssignedAt.IsZero() {
					t.Error("expected AssignedAt to be set")
				}
			},
		},
		{
			name: "invalid user ID",
			setupFunc: func(ctx context.Context, seeder *Seeder) (uuid.UUID, uuid.UUID) {
				role, _ := seeder.SeedRole(ctx, RoleInput{
					Name:        "admin",
					Description: "Admin role",
					CreatedBy:   "system",
				})
				return uuid.Nil, role.ID
			},
			assignedBy: "system",
			wantErr:    true,
		},
		{
			name: "invalid role ID",
			setupFunc: func(ctx context.Context, seeder *Seeder) (uuid.UUID, uuid.UUID) {
				user, _ := seeder.SeedUser(ctx, UserInput{
					Username:  "testuser",
					Name:      "Test User",
					Email:     "test@example.com",
					Password:  "TestPass123!",
					CreatedBy: "system",
				})
				return user.ID, uuid.Nil
			},
			assignedBy: "system",
			wantErr:    true,
		},
		{
			name: "empty assigned by",
			setupFunc: func(ctx context.Context, seeder *Seeder) (uuid.UUID, uuid.UUID) {
				user, _ := seeder.SeedUser(ctx, UserInput{
					Username:  "testuser",
					Name:      "Test User",
					Email:     "test@example.com",
					Password:  "TestPass123!",
					CreatedBy: "system",
				})
				role, _ := seeder.SeedRole(ctx, RoleInput{
					Name:        "admin",
					Description: "Admin role",
					CreatedBy:   "system",
				})
				return user.ID, role.ID
			},
			assignedBy: "",
			wantErr:    true,
		},
		{
			name: "duplicate grant",
			setupFunc: func(ctx context.Context, seeder *Seeder) (uuid.UUID, uuid.UUID) {
				user, _ := seeder.SeedUser(ctx, UserInput{
					Username:  "testuser",
					Name:      "Test User",
					Email:     "test@example.com",
					Password:  "TestPass123!",
					CreatedBy: "system",
				})
				role, _ := seeder.SeedRole(ctx, RoleInput{
					Name:        "admin",
					Description: "Admin role",
					CreatedBy:   "system",
				})

				_, _ = seeder.SeedGrant(ctx, GrantInput{
					UserID:     user.ID,
					RoleID:     role.ID,
					AssignedBy: "system",
				})

				return user.ID, role.ID
			},
			assignedBy: "system",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roleStore := fake.NewRoleStore()
			userStore := fake.NewUserStore()
			grantStore := fake.NewGrantStore(roleStore)

			cfg := &Config{
				EncryptionKey: encKey,
				SigningKey:    sigKey,
			}
			seeder := New(userStore, roleStore, grantStore, cfg, logger.NewNoopLogger())

			userID, roleID := tt.setupFunc(context.Background(), seeder)

			grant, err := seeder.SeedGrant(context.Background(), GrantInput{
				UserID:     userID,
				RoleID:     roleID,
				AssignedBy: tt.assignedBy,
			})

			if (err != nil) != tt.wantErr {
				t.Errorf("SeedGrant() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, grant, userID, roleID)
			}
		})
	}
}

func TestSeeder_New(t *testing.T) {
	roleStore := fake.NewRoleStore()
	userStore := fake.NewUserStore()
	grantStore := fake.NewGrantStore(roleStore)
	cfg := &Config{
		EncryptionKey: []byte("test-key"),
		SigningKey:    []byte("test-sig"),
	}
	log := logger.NewNoopLogger()

	seeder := New(userStore, roleStore, grantStore, cfg, log)

	if seeder == nil {
		t.Fatal("expected seeder to be created")
	}
	if seeder.users != userStore {
		t.Error("users store not set correctly")
	}
	if seeder.roles != roleStore {
		t.Error("roles store not set correctly")
	}
	if seeder.grants != grantStore {
		t.Error("grants store not set correctly")
	}
	if seeder.cfg != cfg {
		t.Error("config not set correctly")
	}
	if seeder.log != log {
		t.Error("logger not set correctly")
	}
}
