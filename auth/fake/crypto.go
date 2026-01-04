package fake

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"
)

type CryptoService struct {
	encKey []byte
	sigKey []byte
}

func NewCryptoService() *CryptoService {
	return &CryptoService{
		encKey: []byte("12345678901234567890123456789012"),
		sigKey: []byte("12345678901234567890123456789012"),
	}
}

func (c *CryptoService) EncryptionKey() []byte {
	return c.encKey
}

func (c *CryptoService) SigningKey() []byte {
	return c.sigKey
}

func (c *CryptoService) ComputeLookupHash(value string) []byte {
	h := hmac.New(sha256.New, c.sigKey)
	h.Write([]byte(value))
	hash := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return []byte(hash)
}

func (c *CryptoService) ComputePINLookupHash(pin string) []byte {
	h := hmac.New(sha256.New, c.sigKey)
	h.Write([]byte(pin))
	hash := base64.StdEncoding.EncodeToString(h.Sum(nil))
	return []byte(hash)
}

type TokenGenerator struct{}

func NewTokenGenerator() *TokenGenerator {
	return &TokenGenerator{}
}

func (t *TokenGenerator) GenerateToken(userID uuid.UUID) (string, error) {
	return fmt.Sprintf("token-%s", userID.String()), nil
}

type PasswordGenerator struct{}

func NewPasswordGenerator() *PasswordGenerator {
	return &PasswordGenerator{}
}

func (p *PasswordGenerator) GeneratePassword() string {
	return "GeneratedPassword123!"
}

type PINGenerator struct{}

func NewPINGenerator() *PINGenerator {
	return &PINGenerator{}
}

func (p *PINGenerator) GeneratePIN() string {
	return "123456"
}
