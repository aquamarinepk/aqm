package crypto

import (
	"bytes"
	"crypto/rand"
	"io"
	"testing"
)

func TestEncryptEmail(t *testing.T) {
	validKey := make([]byte, 32)
	io.ReadFull(rand.Reader, validKey)

	tests := []struct {
		name      string
		plaintext string
		key       []byte
		wantErr   error
	}{
		{
			name:      "valid encryption",
			plaintext: "test@example.com",
			key:       validKey,
			wantErr:   nil,
		},
		{
			name:      "long plaintext",
			plaintext: "very.long.email.address@subdomain.example.com",
			key:       validKey,
			wantErr:   nil,
		},
		{
			name:      "invalid key length short",
			plaintext: "test@example.com",
			key:       make([]byte, 16),
			wantErr:   ErrInvalidKey,
		},
		{
			name:      "invalid key length long",
			plaintext: "test@example.com",
			key:       make([]byte, 64),
			wantErr:   ErrInvalidKey,
		},
		{
			name:      "nil key",
			plaintext: "test@example.com",
			key:       nil,
			wantErr:   ErrInvalidKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, iv, tag, err := EncryptEmail(tt.plaintext, tt.key)

			if err != tt.wantErr {
				t.Errorf("EncryptEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				if ct == "" || iv == "" || tag == "" {
					t.Error("EncryptEmail() returned empty ciphertext, iv, or tag")
				}
			}
		})
	}
}

func TestDecryptEmail(t *testing.T) {
	validKey := make([]byte, 32)
	io.ReadFull(rand.Reader, validKey)

	plaintext := "test@example.com"
	ct, iv, tag, _ := EncryptEmail(plaintext, validKey)

	tests := []struct {
		name       string
		ciphertext string
		iv         string
		tag        string
		key        []byte
		want       string
		wantErr    error
	}{
		{
			name:       "valid decryption",
			ciphertext: ct,
			iv:         iv,
			tag:        tag,
			key:        validKey,
			want:       plaintext,
			wantErr:    nil,
		},
		{
			name:       "invalid key length",
			ciphertext: ct,
			iv:         iv,
			tag:        tag,
			key:        make([]byte, 16),
			want:       "",
			wantErr:    ErrInvalidKey,
		},
		{
			name:       "invalid ciphertext base64",
			ciphertext: "!!!invalid!!!",
			iv:         iv,
			tag:        tag,
			key:        validKey,
			want:       "",
			wantErr:    ErrInvalidCiphertext,
		},
		{
			name:       "invalid iv base64",
			ciphertext: ct,
			iv:         "!!!invalid!!!",
			tag:        tag,
			key:        validKey,
			want:       "",
			wantErr:    ErrInvalidIV,
		},
		{
			name:       "invalid tag base64",
			ciphertext: ct,
			iv:         iv,
			tag:        "!!!invalid!!!",
			key:        validKey,
			want:       "",
			wantErr:    ErrInvalidTag,
		},
		{
			name:       "wrong key",
			ciphertext: ct,
			iv:         iv,
			tag:        tag,
			key:        make([]byte, 32),
			want:       "",
			wantErr:    ErrDecryptionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DecryptEmail(tt.ciphertext, tt.iv, tt.tag, tt.key)

			if err != tt.wantErr {
				t.Errorf("DecryptEmail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("DecryptEmail() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEncryptDecryptRoundtrip(t *testing.T) {
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key)

	tests := []struct {
		name      string
		plaintext string
	}{
		{
			name:      "standard email",
			plaintext: "user@example.com",
		},
		{
			name:      "email with plus",
			plaintext: "user+tag@example.com",
		},
		{
			name:      "long email",
			plaintext: "very.long.email.address.with.many.dots@subdomain.example.com",
		},
		{
			name:      "empty string",
			plaintext: "",
		},
		{
			name:      "unicode email",
			plaintext: "用户@例え.jp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, iv, tag, err := EncryptEmail(tt.plaintext, key)
			if err != nil {
				t.Fatalf("EncryptEmail() error = %v", err)
			}

			got, err := DecryptEmail(ct, iv, tag, key)
			if err != nil {
				t.Fatalf("DecryptEmail() error = %v", err)
			}

			if got != tt.plaintext {
				t.Errorf("Roundtrip failed: got %v, want %v", got, tt.plaintext)
			}
		})
	}
}

func TestComputeLookupHash(t *testing.T) {
	validKey := make([]byte, 32)
	io.ReadFull(rand.Reader, validKey)

	tests := []struct {
		name       string
		email      string
		signingKey []byte
		wantEmpty  bool
	}{
		{
			name:       "valid email and key",
			email:      "test@example.com",
			signingKey: validKey,
			wantEmpty:  false,
		},
		{
			name:       "empty email",
			email:      "",
			signingKey: validKey,
			wantEmpty:  false,
		},
		{
			name:       "invalid key length",
			email:      "test@example.com",
			signingKey: make([]byte, 16),
			wantEmpty:  true,
		},
		{
			name:       "nil key",
			email:      "test@example.com",
			signingKey: nil,
			wantEmpty:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeLookupHash(tt.email, tt.signingKey)

			if tt.wantEmpty && got != "" {
				t.Errorf("ComputeLookupHash() = %v, want empty", got)
			}

			if !tt.wantEmpty && got == "" {
				t.Error("ComputeLookupHash() returned empty, want non-empty")
			}
		})
	}
}

func TestComputeLookupHashConsistency(t *testing.T) {
	key := make([]byte, 32)
	io.ReadFull(rand.Reader, key)

	email := "test@example.com"
	hash1 := ComputeLookupHash(email, key)
	hash2 := ComputeLookupHash(email, key)

	if hash1 != hash2 {
		t.Error("ComputeLookupHash() not deterministic")
	}

	differentEmail := "other@example.com"
	hash3 := ComputeLookupHash(differentEmail, key)

	if hash1 == hash3 {
		t.Error("ComputeLookupHash() collision for different emails")
	}
}

func TestHashPassword(t *testing.T) {
	validSalt := make([]byte, 32)
	io.ReadFull(rand.Reader, validSalt)

	tests := []struct {
		name     string
		password string
		salt     []byte
		wantNil  bool
	}{
		{
			name:     "valid password and salt",
			password: "SecurePassword123!",
			salt:     validSalt,
			wantNil:  false,
		},
		{
			name:     "empty password",
			password: "",
			salt:     validSalt,
			wantNil:  false,
		},
		{
			name:     "invalid salt length",
			password: "password",
			salt:     make([]byte, 16),
			wantNil:  true,
		},
		{
			name:     "nil salt",
			password: "password",
			salt:     nil,
			wantNil:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HashPassword(tt.password, tt.salt)

			if tt.wantNil && got != nil {
				t.Errorf("HashPassword() = %v, want nil", got)
			}

			if !tt.wantNil && got == nil {
				t.Error("HashPassword() returned nil, want non-nil")
			}

			if !tt.wantNil && len(got) != 32 {
				t.Errorf("HashPassword() length = %d, want 32", len(got))
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	salt := make([]byte, 32)
	io.ReadFull(rand.Reader, salt)

	password := "CorrectPassword123!"
	hash := HashPassword(password, salt)

	tests := []struct {
		name     string
		password string
		hash     []byte
		salt     []byte
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			salt:     salt,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "WrongPassword123!",
			hash:     hash,
			salt:     salt,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			salt:     salt,
			want:     false,
		},
		{
			name:     "invalid salt length",
			password: password,
			hash:     hash,
			salt:     make([]byte, 16),
			want:     false,
		},
		{
			name:     "nil salt",
			password: password,
			hash:     hash,
			salt:     nil,
			want:     false,
		},
		{
			name:     "nil hash",
			password: password,
			hash:     nil,
			salt:     salt,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := VerifyPassword(tt.password, tt.hash, tt.salt)

			if got != tt.want {
				t.Errorf("VerifyPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGenerateSalt(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "generate salt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateSalt()

			if err != nil {
				t.Errorf("GenerateSalt() error = %v", err)
			}

			if len(got) != 32 {
				t.Errorf("GenerateSalt() length = %d, want 32", len(got))
			}

			got2, _ := GenerateSalt()
			if bytes.Equal(got, got2) {
				t.Error("GenerateSalt() not random, got same salt twice")
			}
		})
	}
}

func TestGenerateSecureToken(t *testing.T) {
	tests := []struct {
		name       string
		length     int
		wantMinLen int
	}{
		{
			name:       "default length 32",
			length:     32,
			wantMinLen: 32,
		},
		{
			name:       "custom length 64",
			length:     64,
			wantMinLen: 64,
		},
		{
			name:       "length below minimum enforces 32",
			length:     16,
			wantMinLen: 32,
		},
		{
			name:       "zero length enforces 32",
			length:     0,
			wantMinLen: 32,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GenerateSecureToken(tt.length)

			if err != nil {
				t.Errorf("GenerateSecureToken() error = %v", err)
			}

			if len(got) < tt.wantMinLen {
				t.Errorf("GenerateSecureToken() length = %d, want >= %d", len(got), tt.wantMinLen)
			}

			got2, _ := GenerateSecureToken(tt.length)
			if got == got2 {
				t.Error("GenerateSecureToken() not random, got same token twice")
			}
		})
	}
}
