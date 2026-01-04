package auth

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewUser(t *testing.T) {
	user := NewUser()
	if user == nil {
		t.Fatal("NewUser() returned nil")
	}
	if user.Status != UserStatusActive {
		t.Errorf("NewUser() status = %v, want %v", user.Status, UserStatusActive)
	}
}

func TestUserEnsureID(t *testing.T) {
	tests := []struct {
		name    string
		initial uuid.UUID
		wantNil bool
	}{
		{"generates ID when nil", uuid.Nil, false},
		{"keeps existing ID", uuid.New(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := &User{ID: tt.initial}
			user.EnsureID()
			if (user.ID == uuid.Nil) == !tt.wantNil {
				t.Errorf("EnsureID() ID nil = %v, wantNil %v", user.ID == uuid.Nil, tt.wantNil)
			}
			if tt.initial != uuid.Nil && user.ID != tt.initial {
				t.Error("EnsureID() changed existing ID")
			}
		})
	}
}

func TestUserBeforeCreate(t *testing.T) {
	user := &User{
		Username: "  .JohnDoe_.  ",
		Name:     "  John Doe  ",
	}

	user.BeforeCreate()

	if user.ID == uuid.Nil {
		t.Error("BeforeCreate() did not generate ID")
	}
	if user.Username != "johndoe" {
		t.Errorf("BeforeCreate() username = %q, want %q", user.Username, "johndoe")
	}
	if user.Name != "John Doe" {
		t.Errorf("BeforeCreate() name = %q, want %q", user.Name, "John Doe")
	}
	if user.CreatedAt.IsZero() {
		t.Error("BeforeCreate() did not set CreatedAt")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("BeforeCreate() did not set UpdatedAt")
	}
}

func TestUserBeforeUpdate(t *testing.T) {
	user := &User{
		Username: "  .JaneDoe_.  ",
		Name:     "  Jane Doe  ",
	}

	user.BeforeUpdate()

	if user.Username != "janedoe" {
		t.Errorf("BeforeUpdate() username = %q, want %q", user.Username, "janedoe")
	}
	if user.Name != "Jane Doe" {
		t.Errorf("BeforeUpdate() name = %q, want %q", user.Name, "Jane Doe")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("BeforeUpdate() did not set UpdatedAt")
	}
}

func TestUserSetEmail(t *testing.T) {
	encKey := make([]byte, 32)
	signKey := make([]byte, 32)

	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid email", "test@example.com", false},
		{"invalid email", "invalid", true},
		{"empty email", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser()
			err := user.SetEmail(tt.email, encKey, signKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(user.EmailCT) == 0 {
					t.Error("SetEmail() did not set EmailCT")
				}
				if len(user.EmailIV) == 0 {
					t.Error("SetEmail() did not set EmailIV")
				}
				if len(user.EmailTag) == 0 {
					t.Error("SetEmail() did not set EmailTag")
				}
				if len(user.EmailLookup) == 0 {
					t.Error("SetEmail() did not set EmailLookup")
				}
			}
		})
	}
}

func TestUserGetEmail(t *testing.T) {
	encKey := make([]byte, 32)
	signKey := make([]byte, 32)

	user := NewUser()
	email := "test@example.com"

	if err := user.SetEmail(email, encKey, signKey); err != nil {
		t.Fatalf("SetEmail() failed: %v", err)
	}

	got, err := user.GetEmail(encKey)
	if err != nil {
		t.Fatalf("GetEmail() error = %v", err)
	}

	if got != email {
		t.Errorf("GetEmail() = %q, want %q", got, email)
	}
}

func TestUserGetEmailDecryptionError(t *testing.T) {
	user := &User{
		EmailCT:  []byte("invalid"),
		EmailIV:  []byte("invalid"),
		EmailTag: []byte("invalid"),
	}

	encKey := make([]byte, 32)
	_, err := user.GetEmail(encKey)
	if err == nil {
		t.Error("GetEmail() expected error for invalid encrypted data")
	}
	if err != ErrDecryptionFailed {
		t.Errorf("GetEmail() error = %v, want %v", err, ErrDecryptionFailed)
	}
}

func TestUserSetPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid password", "Test1234!", false},
		{"invalid password too short", "Test1!", true},
		{"invalid password no uppercase", "test1234!", true},
		{"empty password", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser()
			err := user.SetPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(user.PasswordHash) == 0 {
					t.Error("SetPassword() did not set PasswordHash")
				}
				if len(user.PasswordSalt) == 0 {
					t.Error("SetPassword() did not set PasswordSalt")
				}
			}
		})
	}
}

func TestUserVerifyPassword(t *testing.T) {
	user := NewUser()
	password := "Test1234!"

	if err := user.SetPassword(password); err != nil {
		t.Fatalf("SetPassword() failed: %v", err)
	}

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"correct password", password, true},
		{"incorrect password", "Wrong1234!", false},
		{"empty password", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := user.VerifyPassword(tt.password); got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserSetPIN(t *testing.T) {
	encKey := make([]byte, 32)
	signKey := make([]byte, 32)

	tests := []struct {
		name    string
		pin     string
		wantErr bool
	}{
		{"valid 4 digit PIN", "1234", false},
		{"valid 6 digit PIN", "123456", false},
		{"valid 8 digit PIN", "12345678", false},
		{"too short", "123", true},
		{"too long", "123456789", true},
		{"empty", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUser()
			err := user.SetPIN(tt.pin, encKey, signKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("SetPIN() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if len(user.PINCT) == 0 {
					t.Error("SetPIN() did not set PINCT")
				}
				if len(user.PINLookup) == 0 {
					t.Error("SetPIN() did not set PINLookup")
				}
			}
		})
	}
}

func TestUserVerifyPIN(t *testing.T) {
	encKey := make([]byte, 32)
	signKey := make([]byte, 32)

	user := NewUser()
	pin := "1234"

	if err := user.SetPIN(pin, encKey, signKey); err != nil {
		t.Fatalf("SetPIN() failed: %v", err)
	}

	tests := []struct {
		name string
		pin  string
		want bool
	}{
		{"correct PIN", pin, true},
		{"incorrect PIN", "5678", false},
		{"empty PIN", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := user.VerifyPIN(tt.pin, signKey); got != tt.want {
				t.Errorf("VerifyPIN() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUserVerifyPINNoPINSet(t *testing.T) {
	user := NewUser()
	signKey := make([]byte, 32)

	if got := user.VerifyPIN("1234", signKey); got != false {
		t.Error("VerifyPIN() should return false when no PIN is set")
	}
}

func TestUserValidate(t *testing.T) {
	tests := []struct {
		name    string
		user    *User
		wantErr bool
	}{
		{
			"valid user",
			&User{Username: "johndoe", Name: "John Doe", Status: UserStatusActive},
			false,
		},
		{
			"invalid username",
			&User{Username: "ab", Name: "John Doe", Status: UserStatusActive},
			true,
		},
		{
			"invalid name",
			&User{Username: "johndoe", Name: "", Status: UserStatusActive},
			true,
		},
		{
			"invalid status",
			&User{Username: "johndoe", Name: "John Doe", Status: UserStatus("invalid")},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
