package service

import (
	"context"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/fake"
	"github.com/google/uuid"
)

func TestSignUp(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()

	tests := []struct {
		name        string
		email       string
		password    string
		username    string
		displayName string
		wantErr     bool
	}{
		{
			name:        "valid signup",
			email:       "test@example.com",
			password:    "Password123!",
			username:    "testuser",
			displayName: "Test User",
			wantErr:     false,
		},
		{
			name:        "invalid email",
			email:       "invalid",
			password:    "Password123!",
			username:    "testuser2",
			displayName: "Test User",
			wantErr:     true,
		},
		{
			name:        "weak password",
			email:       "test2@example.com",
			password:    "weak",
			username:    "testuser3",
			displayName: "Test User",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := SignUp(context.Background(), store, crypto, tt.email, tt.password, tt.username, tt.displayName)

			if (err != nil) != tt.wantErr {
				t.Errorf("SignUp() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && user == nil {
				t.Error("SignUp() returned nil user for successful signup")
			}

			if !tt.wantErr && user.Username != tt.username {
				t.Errorf("SignUp() username = %v, want %v", user.Username, tt.username)
			}
		})
	}
}

func TestSignIn(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	tokenGen := fake.NewTokenGenerator()
	ctx := context.Background()

	user, err := SignUp(ctx, store, crypto, "signin@example.com", "Password123!", "signinuser", "Sign In User")
	if err != nil {
		t.Fatalf("SignUp failed: %v", err)
	}

	if !user.VerifyPassword("Password123!") {
		t.Fatal("Password verification failed immediately after signup")
	}

	lookup := crypto.ComputeLookupHash("signin@example.com")
	retrievedByLookup, err := store.GetByEmailLookup(ctx, lookup)
	if err != nil {
		t.Fatalf("GetByEmailLookup failed: %v", err)
	}
	if retrievedByLookup.ID != user.ID {
		t.Fatalf("Retrieved user ID mismatch: got %v, want %v", retrievedByLookup.ID, user.ID)
	}

	tests := []struct {
		name     string
		email    string
		password string
		wantErr  bool
	}{
		{
			name:     "correct credentials",
			email:    "signin@example.com",
			password: "Password123!",
			wantErr:  false,
		},
		{
			name:     "wrong password",
			email:    "signin@example.com",
			password: "WrongPassword123!",
			wantErr:  true,
		},
		{
			name:     "non-existing user",
			email:    "notfound@example.com",
			password: "Password123!",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, token, err := SignIn(ctx, store, crypto, tokenGen, tt.email, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("SignIn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if gotUser == nil {
					t.Error("SignIn() returned nil user")
				}
				if token == "" {
					t.Error("SignIn() returned empty token")
				}
				if gotUser.ID != user.ID {
					t.Errorf("SignIn() user ID = %v, want %v", gotUser.ID, user.ID)
				}
			}
		})
	}
}

func TestBootstrap(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	pwdGen := fake.NewPasswordGenerator()
	ctx := context.Background()

	user, password, err := Bootstrap(ctx, store, crypto, pwdGen)
	if err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	if user == nil {
		t.Fatal("Bootstrap() returned nil user")
	}

	if password == "" {
		t.Error("Bootstrap() returned empty password")
	}

	if user.Username != "superadmin" {
		t.Errorf("Bootstrap() username = %v, want superadmin", user.Username)
	}

	_, password2, err := Bootstrap(ctx, store, crypto, pwdGen)
	if err != nil {
		t.Fatalf("Bootstrap() second call error = %v", err)
	}

	if password2 != "" {
		t.Error("Bootstrap() second call should return empty password when superadmin exists")
	}
}

func TestGeneratePIN(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	pinGen := fake.NewPINGenerator()
	ctx := context.Background()

	user, err := SignUp(ctx, store, crypto, "pinuser@example.com", "Password123!", "pinuser", "PIN User")
	if err != nil {
		t.Fatalf("SignUp failed: %v", err)
	}

	pin, err := GeneratePIN(ctx, store, crypto, pinGen, user)
	if err != nil {
		t.Fatalf("GeneratePIN() error = %v", err)
	}

	if pin == "" {
		t.Error("GeneratePIN() returned empty PIN")
	}

	if len(user.PINLookup) == 0 {
		t.Error("GeneratePIN() did not set PINLookup on user")
	}
}

func TestSignInByPIN(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	pinGen := fake.NewPINGenerator()
	ctx := context.Background()

	user, _ := SignUp(ctx, store, crypto, "pinlogin@example.com", "Password123!", "pinloginuser", "PIN Login User")
	pin, _ := GeneratePIN(ctx, store, crypto, pinGen, user)

	tests := []struct {
		name    string
		pin     string
		wantErr bool
	}{
		{
			name:    "correct PIN",
			pin:     pin,
			wantErr: false,
		},
		{
			name:    "wrong PIN",
			pin:     "000000",
			wantErr: true,
		},
		{
			name:    "invalid PIN length",
			pin:     "12",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotUser, err := SignInByPIN(ctx, store, crypto, tt.pin)

			if (err != nil) != tt.wantErr {
				t.Errorf("SignInByPIN() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && gotUser == nil {
				t.Error("SignInByPIN() returned nil user")
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	ctx := context.Background()

	user, _ := SignUp(ctx, store, crypto, "getbyid@example.com", "Password123!", "getbyiduser", "Get By ID User")

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
			name:    "non-existing user",
			id:      uuid.New(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserByID(ctx, store, tt.id)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("GetUserByID() id = %v, want %v", got.ID, tt.id)
			}
		})
	}
}

func TestGetUserByUsername(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	ctx := context.Background()

	SignUp(ctx, store, crypto, "getbyname@example.com", "Password123!", "getbynameuser", "Get By Name User")

	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{
			name:     "existing username",
			username: "getbynameuser",
			wantErr:  false,
		},
		{
			name:     "non-existing username",
			username: "notfound",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetUserByUsername(ctx, store, tt.username)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetUserByUsername() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got.Username != tt.username {
				t.Errorf("GetUserByUsername() username = %v, want %v", got.Username, tt.username)
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	ctx := context.Background()

	SignUp(ctx, store, crypto, "list1@example.com", "Password123!", "listuser1", "List User 1")
	SignUp(ctx, store, crypto, "list2@example.com", "Password123!", "listuser2", "List User 2")

	users, err := ListUsers(ctx, store)
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}

	if len(users) < 2 {
		t.Errorf("ListUsers() count = %v, want at least 2", len(users))
	}
}

func TestListUsersByStatus(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	ctx := context.Background()

	SignUp(ctx, store, crypto, "status1@example.com", "Password123!", "statususer1", "Status User 1")
	user2, _ := SignUp(ctx, store, crypto, "status2@example.com", "Password123!", "statususer2", "Status User 2")

	DeleteUser(ctx, store, user2.ID)

	activeUsers, err := ListUsersByStatus(ctx, store, auth.UserStatusActive)
	if err != nil {
		t.Fatalf("ListUsersByStatus() error = %v", err)
	}

	if len(activeUsers) < 1 {
		t.Errorf("ListUsersByStatus(active) count = %v, want at least 1", len(activeUsers))
	}
}

func TestUpdateUser(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	ctx := context.Background()

	user, _ := SignUp(ctx, store, crypto, "update@example.com", "Password123!", "updateuser", "Original Name")

	user.Name = "Updated Name"
	err := UpdateUser(ctx, store, user)
	if err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}

	retrieved, _ := GetUserByID(ctx, store, user.ID)
	if retrieved.Name != "Updated Name" {
		t.Errorf("UpdateUser() name = %v, want Updated Name", retrieved.Name)
	}
}

func TestDeleteUser(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	ctx := context.Background()

	user, _ := SignUp(ctx, store, crypto, "delete@example.com", "Password123!", "deleteuser", "Delete User")

	err := DeleteUser(ctx, store, user.ID)
	if err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}

	retrieved, _ := GetUserByID(ctx, store, user.ID)
	if retrieved.Status != "deleted" {
		t.Errorf("DeleteUser() status = %v, want deleted", retrieved.Status)
	}
}

func TestSignUpValidationErrors(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	ctx := context.Background()

	tests := []struct {
		name        string
		email       string
		password    string
		username    string
		displayName string
		wantErr     bool
	}{
		{
			name:        "empty username",
			email:       "test@example.com",
			password:    "Password123!",
			username:    "",
			displayName: "Test",
			wantErr:     true,
		},
		{
			name:        "empty display name",
			email:       "test@example.com",
			password:    "Password123!",
			username:    "testuser",
			displayName: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := SignUp(ctx, store, crypto, tt.email, tt.password, tt.username, tt.displayName)
			if (err != nil) != tt.wantErr {
				t.Errorf("SignUp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAssignRoleErrors(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	ctx := context.Background()

	role, _ := CreateRole(ctx, roleStore, "tester", "Tester", []string{"test"}, "system")
	username := "testuser"

	// First assignment succeeds
	_, err := AssignRole(ctx, grantStore, username, role.ID, "admin")
	if err != nil {
		t.Fatalf("First AssignRole() error = %v", err)
	}

	// Second assignment should fail
	_, err = AssignRole(ctx, grantStore, username, role.ID, "admin")
	if err == nil {
		t.Error("Second AssignRole() should fail with duplicate error")
	}
}

func TestSignInInactiveUser(t *testing.T) {
	store := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	tokenGen := fake.NewTokenGenerator()
	ctx := context.Background()

	user, _ := SignUp(ctx, store, crypto, "inactive@example.com", "Password123!", "inactiveuser", "Inactive User")

	// Soft delete the user
	DeleteUser(ctx, store, user.ID)

	// Try to sign in
	_, _, err := SignIn(ctx, store, crypto, tokenGen, "inactive@example.com", "Password123!")
	if err == nil {
		t.Error("SignIn() with inactive user should fail")
	}
}
