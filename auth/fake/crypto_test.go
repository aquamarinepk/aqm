package fake

import (
	"testing"

	"github.com/google/uuid"
)

func TestNewCryptoService(t *testing.T) {
	service := NewCryptoService()
	if service == nil {
		t.Fatal("NewCryptoService() returned nil")
	}
	if service.encKey == nil {
		t.Error("encryption key not initialized")
	}
	if service.sigKey == nil {
		t.Error("signing key not initialized")
	}
}

func TestCryptoService_EncryptionKey(t *testing.T) {
	service := NewCryptoService()
	key := service.EncryptionKey()

	if key == nil {
		t.Fatal("EncryptionKey() returned nil")
	}
	if len(key) != 32 {
		t.Errorf("EncryptionKey() returned key of length %d, want 32", len(key))
	}
}

func TestCryptoService_SigningKey(t *testing.T) {
	service := NewCryptoService()
	key := service.SigningKey()

	if key == nil {
		t.Fatal("SigningKey() returned nil")
	}
	if len(key) != 32 {
		t.Errorf("SigningKey() returned key of length %d, want 32", len(key))
	}
}

func TestCryptoService_ComputeLookupHash(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "hash email",
			value: "user@example.com",
		},
		{
			name:  "hash empty string",
			value: "",
		},
		{
			name:  "hash special characters",
			value: "test!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCryptoService()
			hash1 := service.ComputeLookupHash(tt.value)
			hash2 := service.ComputeLookupHash(tt.value)

			if hash1 == nil {
				t.Fatal("ComputeLookupHash() returned nil")
			}
			if string(hash1) != string(hash2) {
				t.Error("ComputeLookupHash() not consistent for same input")
			}
		})
	}
}

func TestCryptoService_ComputePINLookupHash(t *testing.T) {
	tests := []struct {
		name string
		pin  string
	}{
		{
			name: "hash numeric PIN",
			pin:  "123456",
		},
		{
			name: "hash empty PIN",
			pin:  "",
		},
		{
			name: "hash alphanumeric PIN",
			pin:  "ABC123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := NewCryptoService()
			hash1 := service.ComputePINLookupHash(tt.pin)
			hash2 := service.ComputePINLookupHash(tt.pin)

			if hash1 == nil {
				t.Fatal("ComputePINLookupHash() returned nil")
			}
			if string(hash1) != string(hash2) {
				t.Error("ComputePINLookupHash() not consistent for same input")
			}
		})
	}
}

func TestNewTokenGenerator(t *testing.T) {
	generator := NewTokenGenerator()
	if generator == nil {
		t.Fatal("NewTokenGenerator() returned nil")
	}
}

func TestTokenGenerator_GenerateToken(t *testing.T) {
	tests := []struct {
		name   string
		userID uuid.UUID
	}{
		{
			name:   "generate token for user",
			userID: uuid.New(),
		},
		{
			name:   "generate token for another user",
			userID: uuid.New(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			generator := NewTokenGenerator()
			token, err := generator.GenerateToken(tt.userID)

			if err != nil {
				t.Fatalf("GenerateToken() error = %v", err)
			}
			if token == "" {
				t.Error("GenerateToken() returned empty token")
			}
			if len(token) == 0 {
				t.Error("GenerateToken() returned zero-length token")
			}
		})
	}
}

func TestNewPasswordGenerator(t *testing.T) {
	generator := NewPasswordGenerator()
	if generator == nil {
		t.Fatal("NewPasswordGenerator() returned nil")
	}
}

func TestPasswordGenerator_GeneratePassword(t *testing.T) {
	generator := NewPasswordGenerator()
	password := generator.GeneratePassword()

	if password == "" {
		t.Error("GeneratePassword() returned empty password")
	}
	if len(password) == 0 {
		t.Error("GeneratePassword() returned zero-length password")
	}
}

func TestNewPINGenerator(t *testing.T) {
	generator := NewPINGenerator()
	if generator == nil {
		t.Fatal("NewPINGenerator() returned nil")
	}
}

func TestPINGenerator_GeneratePIN(t *testing.T) {
	generator := NewPINGenerator()
	pin := generator.GeneratePIN()

	if pin == "" {
		t.Error("GeneratePIN() returned empty PIN")
	}
	if len(pin) == 0 {
		t.Error("GeneratePIN() returned zero-length PIN")
	}
}
