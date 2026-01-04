package service

import (
	"crypto/ed25519"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewDefaultCryptoService(t *testing.T) {
	encKey := make([]byte, 32)
	sigKey := make([]byte, 32)

	service := NewDefaultCryptoService(encKey, sigKey)
	if service == nil {
		t.Fatal("NewDefaultCryptoService() returned nil")
	}

	if len(service.EncryptionKey()) != 32 {
		t.Errorf("EncryptionKey() length = %v, want 32", len(service.EncryptionKey()))
	}

	if len(service.SigningKey()) != 32 {
		t.Errorf("SigningKey() length = %v, want 32", len(service.SigningKey()))
	}
}

func TestDefaultCryptoServiceComputeLookupHash(t *testing.T) {
	encKey := make([]byte, 32)
	sigKey := make([]byte, 32)
	service := NewDefaultCryptoService(encKey, sigKey)

	hash := service.ComputeLookupHash("test@example.com")
	if len(hash) == 0 {
		t.Error("ComputeLookupHash() returned empty hash")
	}

	hash2 := service.ComputeLookupHash("test@example.com")
	if string(hash) != string(hash2) {
		t.Error("ComputeLookupHash() is not deterministic")
	}
}

func TestDefaultCryptoServiceComputePINLookupHash(t *testing.T) {
	encKey := make([]byte, 32)
	sigKey := make([]byte, 32)
	service := NewDefaultCryptoService(encKey, sigKey)

	hash := service.ComputePINLookupHash("123456")
	if len(hash) == 0 {
		t.Error("ComputePINLookupHash() returned empty hash")
	}

	hash2 := service.ComputePINLookupHash("123456")
	if string(hash) != string(hash2) {
		t.Error("ComputePINLookupHash() is not deterministic")
	}
}

func TestNewDefaultTokenGenerator(t *testing.T) {
	_, privKey, _ := ed25519.GenerateKey(nil)
	ttl := time.Hour

	generator := NewDefaultTokenGenerator(privKey, ttl)
	if generator == nil {
		t.Fatal("NewDefaultTokenGenerator() returned nil")
	}
}

func TestDefaultTokenGeneratorGenerateToken(t *testing.T) {
	_, privKey, _ := ed25519.GenerateKey(nil)
	ttl := time.Hour
	generator := NewDefaultTokenGenerator(privKey, ttl)

	token, err := generator.GenerateToken(uuid.New())
	if err != nil {
		t.Fatalf("GenerateToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateToken() returned empty token")
	}

	token2, _ := generator.GenerateToken(uuid.New())
	if token == token2 {
		t.Error("GenerateToken() should generate unique tokens")
	}
}

func TestNewDefaultPasswordGenerator(t *testing.T) {
	generator := NewDefaultPasswordGenerator(32)
	if generator == nil {
		t.Fatal("NewDefaultPasswordGenerator() returned nil")
	}

	// Test with short length (should default to 32)
	generator2 := NewDefaultPasswordGenerator(8)
	password := generator2.GeneratePassword()
	if len(password) < 16 {
		t.Errorf("NewDefaultPasswordGenerator(8) password length = %v, want at least 16", len(password))
	}
}

func TestDefaultPasswordGeneratorGeneratePassword(t *testing.T) {
	generator := NewDefaultPasswordGenerator(32)

	password := generator.GeneratePassword()
	if len(password) < 16 {
		t.Errorf("GeneratePassword() length = %v, want at least 16", len(password))
	}

	password2 := generator.GeneratePassword()
	if password == password2 {
		t.Error("GeneratePassword() should generate unique passwords")
	}
}

func TestNewDefaultPINGenerator(t *testing.T) {
	generator := NewDefaultPINGenerator()
	if generator == nil {
		t.Fatal("NewDefaultPINGenerator() returned nil")
	}
}

func TestDefaultPINGeneratorGeneratePIN(t *testing.T) {
	generator := NewDefaultPINGenerator()

	pin := generator.GeneratePIN()
	if len(pin) != 6 {
		t.Errorf("GeneratePIN() length = %v, want 6", len(pin))
	}

	for _, char := range pin {
		if char < '0' || char > '9' {
			t.Errorf("GeneratePIN() contains non-digit character: %c", char)
		}
	}

	pin2 := generator.GeneratePIN()
	if pin == pin2 {
		t.Error("GeneratePIN() should generate different PINs (note: small chance of collision)")
	}
}
