package service

import "github.com/google/uuid"

// CryptoService handles encryption and hashing operations
type CryptoService interface {
	EncryptionKey() []byte
	SigningKey() []byte
	ComputeLookupHash(value string) []byte
	ComputePINLookupHash(pin string) []byte
}

// TokenGenerator generates session tokens
type TokenGenerator interface {
	GenerateToken(userID uuid.UUID) (string, error)
}

// PasswordGenerator generates secure passwords
type PasswordGenerator interface {
	GeneratePassword() string
}

// PINGenerator generates PINs for lightweight authentication
type PINGenerator interface {
	GeneratePIN() string
}
