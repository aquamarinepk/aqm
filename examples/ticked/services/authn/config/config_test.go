package config

import (
	"encoding/base64"
	"encoding/hex"
	"os"
	"testing"

	log "github.com/aquamarinepk/aqm/log"
)

func TestNew(t *testing.T) {
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	// Set required env vars
	os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
	os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
	os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
	defer os.Unsetenv("AUTHN_CRYPTO_ENCRYPTIONKEY")
	defer os.Unsetenv("AUTHN_CRYPTO_SIGNINGKEY")
	defer os.Unsetenv("AUTHN_CRYPTO_TOKENPRIVATEKEY")

	logger := log.NewLogger("info")
	cfg, err := New(logger)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"server port", cfg.Server.Port, ":8080"},
		{"database driver", cfg.Database.Driver, "fake"},
		{"log level", cfg.Log.Level, "info"},
		{"log format", cfg.GetLogFormat(), "text"},
		{"password length", cfg.GetPasswordLength(), 32},
		{"bootstrap enabled", cfg.IsBootstrapEnabled(), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestNewWithEnvOverrides(t *testing.T) {
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	tests := []struct {
		name     string
		envVars  map[string]string
		wantErr  bool
		validate func(*testing.T, *Config)
	}{
		{
			name: "valid with all required keys",
			envVars: map[string]string{
				"AUTHN_CRYPTO_ENCRYPTIONKEY":   validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":      validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Server.Port != ":8080" {
					t.Errorf("expected port :8080, got %s", cfg.Server.Port)
				}
			},
		},
		{
			name: "missing encryption key",
			envVars: map[string]string{
				"AUTHN_CRYPTO_SIGNINGKEY":      validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: true,
		},
		{
			name: "missing signing key",
			envVars: map[string]string{
				"AUTHN_CRYPTO_ENCRYPTIONKEY":   validEncKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: true,
		},
		{
			name: "missing token private key",
			envVars: map[string]string{
				"AUTHN_CRYPTO_ENCRYPTIONKEY": validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":    validSignKey,
			},
			wantErr: true,
		},
		{
			name: "override port",
			envVars: map[string]string{
				"AUTHN_SERVER_PORT":            ":9090",
				"AUTHN_CRYPTO_ENCRYPTIONKEY":   validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":      validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Server.Port != ":9090" {
					t.Errorf("expected port :9090, got %s", cfg.Server.Port)
				}
			},
		},
		{
			name: "override database driver",
			envVars: map[string]string{
				"AUTHN_DATABASE_DRIVER":        "postgres",
				"AUTHN_DATABASE_HOST":          "testhost",
				"AUTHN_DATABASE_DATABASE":      "testdb",
				"AUTHN_CRYPTO_ENCRYPTIONKEY":   validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":      validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: false,
			validate: func(t *testing.T, cfg *Config) {
				if cfg.Database.Driver != "postgres" {
					t.Errorf("expected driver postgres, got %s", cfg.Database.Driver)
				}
				if cfg.Database.Host != "testhost" {
					t.Errorf("expected host testhost, got %s", cfg.Database.Host)
				}
			},
		},
		{
			name: "invalid encryption key length",
			envVars: map[string]string{
				"AUTHN_CRYPTO_ENCRYPTIONKEY":   "tooshort",
				"AUTHN_CRYPTO_SIGNINGKEY":      validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: true,
		},
		{
			name: "invalid signing key length",
			envVars: map[string]string{
				"AUTHN_CRYPTO_ENCRYPTIONKEY":   validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":      "tooshort",
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: true,
		},
		{
			name: "invalid token key length",
			envVars: map[string]string{
				"AUTHN_CRYPTO_ENCRYPTIONKEY":   validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":      validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": base64.StdEncoding.EncodeToString(make([]byte, 32)), // wrong size
			},
			wantErr: true,
		},
		{
			name: "invalid token ttl",
			envVars: map[string]string{
				"AUTHN_AUTH_TOKENTTL":          "invalid",
				"AUTHN_CRYPTO_ENCRYPTIONKEY":   validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":      validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: true,
		},
		{
			name: "password length too short",
			envVars: map[string]string{
				"AUTHN_AUTH_PASSWORDLENGTH":    "8",
				"AUTHN_CRYPTO_ENCRYPTIONKEY":   validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":      validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: true,
		},
		{
			name: "invalid log format",
			envVars: map[string]string{
				"AUTHN_LOG_FORMAT":             "invalid",
				"AUTHN_CRYPTO_ENCRYPTIONKEY":   validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":      validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear and set env vars
			clearAuthnEnvVars()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}
			defer clearAuthnEnvVars()

			logger := log.NewLogger("info")
			cfg, err := New(logger)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, cfg)
			}
		})
	}
}

func TestDecodeKeys(t *testing.T) {
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
	os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
	os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
	defer clearAuthnEnvVars()

	logger := log.NewLogger("info")
	cfg, err := New(logger)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	// Test DecodeEncryptionKey
	encKey, err := cfg.DecodeEncryptionKey()
	if err != nil {
		t.Errorf("DecodeEncryptionKey() error = %v", err)
	}
	if len(encKey) != 32 {
		t.Errorf("expected encryption key length 32, got %d", len(encKey))
	}

	// Test DecodeSigningKey
	signKey, err := cfg.DecodeSigningKey()
	if err != nil {
		t.Errorf("DecodeSigningKey() error = %v", err)
	}
	if len(signKey) != 32 {
		t.Errorf("expected signing key length 32, got %d", len(signKey))
	}

	// Test DecodeTokenPrivateKey
	tokenKey, err := cfg.DecodeTokenPrivateKey()
	if err != nil {
		t.Errorf("DecodeTokenPrivateKey() error = %v", err)
	}
	if len(tokenKey) != 64 {
		t.Errorf("expected token key length 64, got %d", len(tokenKey))
	}
}

func TestParseTokenTTL(t *testing.T) {
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	tests := []struct {
		name    string
		ttl     string
		wantErr bool
	}{
		{"valid 24h", "24h", false},
		{"valid 1h30m", "1h30m", false},
		{"valid 5m", "5m", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clearAuthnEnvVars()
			os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
			os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
			os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
			os.Setenv("AUTHN_AUTH_TOKENTTL", tt.ttl)
			defer clearAuthnEnvVars()

			logger := log.NewLogger("info")
			cfg, err := New(logger)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}

			duration, err := cfg.ParseTokenTTL()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTokenTTL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && duration == 0 {
				t.Error("expected non-zero duration")
			}
		})
	}
}

func TestConnectionString(t *testing.T) {
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	clearAuthnEnvVars()
	os.Setenv("AUTHN_DATABASE_DRIVER", "postgres")
	os.Setenv("AUTHN_DATABASE_HOST", "testhost")
	os.Setenv("AUTHN_DATABASE_PORT", "5433")
	os.Setenv("AUTHN_DATABASE_USER", "testuser")
	os.Setenv("AUTHN_DATABASE_PASSWORD", "testpass")
	os.Setenv("AUTHN_DATABASE_DATABASE", "testdb")
	os.Setenv("AUTHN_DATABASE_SSLMODE", "require")
	os.Setenv("AUTHN_DATABASE_SCHEMA", "public")
	os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
	os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
	os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
	defer clearAuthnEnvVars()

	logger := log.NewLogger("info")
	cfg, err := New(logger)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	connStr := cfg.Database.ConnectionString()
	if connStr == "" {
		t.Error("expected non-empty connection string")
	}

	expected := "host=testhost port=5433 user=testuser password=testpass dbname=testdb sslmode=require search_path=public"
	if connStr != expected {
		t.Errorf("ConnectionString() = %q, want %q", connStr, expected)
	}
}

// Helper function to clear all AUTHN_ env vars
func clearAuthnEnvVars() {
	envVars := []string{
		"AUTHN_SERVER_PORT",
		"AUTHN_DATABASE_DRIVER",
		"AUTHN_DATABASE_HOST",
		"AUTHN_DATABASE_PORT",
		"AUTHN_DATABASE_USER",
		"AUTHN_DATABASE_PASSWORD",
		"AUTHN_DATABASE_DATABASE",
		"AUTHN_DATABASE_SSLMODE",
		"AUTHN_DATABASE_SCHEMA",
		"AUTHN_CRYPTO_ENCRYPTIONKEY",
		"AUTHN_CRYPTO_SIGNINGKEY",
		"AUTHN_CRYPTO_TOKENPRIVATEKEY",
		"AUTHN_AUTH_TOKENTTL",
		"AUTHN_AUTH_PASSWORDLENGTH",
		"AUTHN_AUTH_ENABLEBOOTSTRAP",
		"AUTHN_LOG_LEVEL",
		"AUTHN_LOG_FORMAT",
	}

	for _, v := range envVars {
		os.Unsetenv(v)
	}
}
