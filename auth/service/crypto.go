package service

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/aquamarinepk/aqm/crypto"
	"github.com/google/uuid"
)

// DefaultCryptoService implements CryptoService using aqm crypto primitives
type DefaultCryptoService struct {
	encryptionKey []byte
	signingKey    []byte
}

func NewDefaultCryptoService(encryptionKey, signingKey []byte) *DefaultCryptoService {
	return &DefaultCryptoService{
		encryptionKey: encryptionKey,
		signingKey:    signingKey,
	}
}

func (s *DefaultCryptoService) EncryptionKey() []byte {
	return s.encryptionKey
}

func (s *DefaultCryptoService) SigningKey() []byte {
	return s.signingKey
}

func (s *DefaultCryptoService) ComputeLookupHash(value string) []byte {
	hash := crypto.ComputeLookupHash(value, s.signingKey)
	return []byte(hash)
}

func (s *DefaultCryptoService) ComputePINLookupHash(pin string) []byte {
	hash := crypto.ComputeLookupHash(pin, s.signingKey)
	return []byte(hash)
}

// DefaultTokenGenerator implements TokenGenerator using PASETO v4
type DefaultTokenGenerator struct {
	privateKey ed25519.PrivateKey
	ttl        time.Duration
}

func NewDefaultTokenGenerator(privateKey ed25519.PrivateKey, ttl time.Duration) *DefaultTokenGenerator {
	return &DefaultTokenGenerator{
		privateKey: privateKey,
		ttl:        ttl,
	}
}

func (g *DefaultTokenGenerator) GenerateToken(userID uuid.UUID) (string, error) {
	sessionID := crypto.GenerateSessionID()
	claims := crypto.TokenClaims{
		Subject:   userID.String(),
		SessionID: sessionID,
		ExpiresAt: time.Now().Add(g.ttl).Unix(),
	}
	token, err := crypto.GenerateToken(claims, g.privateKey)
	if err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return token, nil
}

// DefaultPasswordGenerator implements PasswordGenerator
type DefaultPasswordGenerator struct {
	length int
}

func NewDefaultPasswordGenerator(length int) *DefaultPasswordGenerator {
	if length < 16 {
		length = 32
	}
	return &DefaultPasswordGenerator{length: length}
}

func (g *DefaultPasswordGenerator) GeneratePassword() string {
	token, err := crypto.GenerateSecureToken(g.length)
	if err != nil {
		return uuid.New().String()
	}
	return token
}

// DevPasswordGenerator returns a fixed password for development mode.
// Use this when AQM_DEV_MODE=true to avoid having to look up random passwords.
type DevPasswordGenerator struct {
	password string
}

func NewDevPasswordGenerator(password string) *DevPasswordGenerator {
	if password == "" {
		password = "Superadmin123!"
	}
	return &DevPasswordGenerator{password: password}
}

func (g *DevPasswordGenerator) GeneratePassword() string {
	return g.password
}

// DefaultPINGenerator implements PINGenerator
type DefaultPINGenerator struct{}

func NewDefaultPINGenerator() *DefaultPINGenerator {
	return &DefaultPINGenerator{}
}

func (g *DefaultPINGenerator) GeneratePIN() string {
	// Generate 6-digit PIN
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "000000"
	}
	return fmt.Sprintf("%06d", n.Int64())
}
