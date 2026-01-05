package config

import (
	"encoding/base64"
	"encoding/hex"
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := New()

	if cfg.Server.Port != ":8080" {
		t.Errorf("expected port :8080, got %s", cfg.Server.Port)
	}

	if cfg.Database.Driver != "fake" {
		t.Errorf("expected driver fake, got %s", cfg.Database.Driver)
	}

	if cfg.Auth.TokenTTL != "24h" {
		t.Errorf("expected token_ttl 24h, got %s", cfg.Auth.TokenTTL)
	}

	if cfg.Log.Level != "info" {
		t.Errorf("expected log level info, got %s", cfg.Log.Level)
	}
}

func TestLoad(t *testing.T) {
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	baseEnvVars := map[string]string{
		"AUTHN_CRYPTO_ENCRYPTIONKEY":    validEncKey,
		"AUTHN_CRYPTO_SIGNINGKEY":       validSignKey,
		"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
	}

	tests := []struct {
		name     string
		envVars  map[string]string
		wantErr  bool
		wantPort string
	}{
		{
			name:     "valid with all required keys",
			envVars:  baseEnvVars,
			wantErr:  false,
			wantPort: ":8080",
		},
		{
			name: "missing encryption key",
			envVars: map[string]string{
				"AUTHN_CRYPTO_SIGNINGKEY":       validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: true,
		},
		{
			name: "missing signing key",
			envVars: map[string]string{
				"AUTHN_CRYPTO_ENCRYPTIONKEY":    validEncKey,
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
				"AUTHN_SERVER_PORT":              ":9090",
				"AUTHN_CRYPTO_ENCRYPTIONKEY":    validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":       validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr:  false,
			wantPort: ":9090",
		},
		{
			name: "override database driver",
			envVars: map[string]string{
				"AUTHN_DATABASE_DRIVER":          "postgres",
				"AUTHN_DATABASE_HOST":            "testhost",
				"AUTHN_DATABASE_DATABASE":        "testdb",
				"AUTHN_CRYPTO_ENCRYPTIONKEY":    validEncKey,
				"AUTHN_CRYPTO_SIGNINGKEY":       validSignKey,
				"AUTHN_CRYPTO_TOKENPRIVATEKEY": validTokenKey,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean env first
			os.Clearenv()

			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg, err := Load("AUTHN_")

			if tt.wantErr {
				if err == nil {
					t.Errorf("Load() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Load() unexpected error: %v", err)
				return
			}

			if tt.wantPort != "" {
				if cfg.Server.Port != tt.wantPort {
					t.Errorf("expected port %s, got %s", tt.wantPort, cfg.Server.Port)
				}
			}

			if tt.envVars["AUTHN_DATABASE_DRIVER"] != "" {
				if cfg.Database.Driver != tt.envVars["AUTHN_DATABASE_DRIVER"] {
					t.Errorf("expected driver %s, got %s", tt.envVars["AUTHN_DATABASE_DRIVER"], cfg.Database.Driver)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: &Config{
				Crypto: CryptoConfig{
					EncryptionKey:   validEncKey,
					SigningKey:      validSignKey,
					TokenPrivateKey: validTokenKey,
				},
				Auth: AuthConfig{
					TokenTTL:       "24h",
					PasswordLength: 32,
				},
				Database: DatabaseConfig{
					Driver: "fake",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: false,
		},
		{
			name: "empty encryption key",
			cfg: &Config{
				Crypto: CryptoConfig{
					EncryptionKey:   "",
					SigningKey:      validSignKey,
					TokenPrivateKey: validTokenKey,
				},
				Auth: AuthConfig{
					TokenTTL:       "24h",
					PasswordLength: 32,
				},
				Database: DatabaseConfig{
					Driver: "fake",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid encryption key length",
			cfg: &Config{
				Crypto: CryptoConfig{
					EncryptionKey:   hex.EncodeToString([]byte("short")),
					SigningKey:      validSignKey,
					TokenPrivateKey: validTokenKey,
				},
				Auth: AuthConfig{
					TokenTTL:       "24h",
					PasswordLength: 32,
				},
				Database: DatabaseConfig{
					Driver: "fake",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid token ttl",
			cfg: &Config{
				Crypto: CryptoConfig{
					EncryptionKey:   validEncKey,
					SigningKey:      validSignKey,
					TokenPrivateKey: validTokenKey,
				},
				Auth: AuthConfig{
					TokenTTL:       "invalid",
					PasswordLength: 32,
				},
				Database: DatabaseConfig{
					Driver: "fake",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "password length too short",
			cfg: &Config{
				Crypto: CryptoConfig{
					EncryptionKey:   validEncKey,
					SigningKey:      validSignKey,
					TokenPrivateKey: validTokenKey,
				},
				Auth: AuthConfig{
					TokenTTL:       "24h",
					PasswordLength: 8,
				},
				Database: DatabaseConfig{
					Driver: "fake",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid database driver",
			cfg: &Config{
				Crypto: CryptoConfig{
					EncryptionKey:   validEncKey,
					SigningKey:      validSignKey,
					TokenPrivateKey: validTokenKey,
				},
				Auth: AuthConfig{
					TokenTTL:       "24h",
					PasswordLength: 32,
				},
				Database: DatabaseConfig{
					Driver: "invalid",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "postgres missing host",
			cfg: &Config{
				Crypto: CryptoConfig{
					EncryptionKey:   validEncKey,
					SigningKey:      validSignKey,
					TokenPrivateKey: validTokenKey,
				},
				Auth: AuthConfig{
					TokenTTL:       "24h",
					PasswordLength: 32,
				},
				Database: DatabaseConfig{
					Driver:   "postgres",
					Host:     "",
					Database: "testdb",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "postgres missing database",
			cfg: &Config{
				Crypto: CryptoConfig{
					EncryptionKey:   validEncKey,
					SigningKey:      validSignKey,
					TokenPrivateKey: validTokenKey,
				},
				Auth: AuthConfig{
					TokenTTL:       "24h",
					PasswordLength: 32,
				},
				Database: DatabaseConfig{
					Driver:   "postgres",
					Host:     "localhost",
					Database: "",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: &Config{
				Crypto: CryptoConfig{
					EncryptionKey:   validEncKey,
					SigningKey:      validSignKey,
					TokenPrivateKey: validTokenKey,
				},
				Auth: AuthConfig{
					TokenTTL:       "24h",
					PasswordLength: 32,
				},
				Database: DatabaseConfig{
					Driver: "fake",
				},
				Log: LogConfig{
					Level:  "invalid",
					Format: "text",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log format",
			cfg: &Config{
				Crypto: CryptoConfig{
					EncryptionKey:   validEncKey,
					SigningKey:      validSignKey,
					TokenPrivateKey: validTokenKey,
				},
				Auth: AuthConfig{
					TokenTTL:       "24h",
					PasswordLength: 32,
				},
				Database: DatabaseConfig{
					Driver: "fake",
				},
				Log: LogConfig{
					Level:  "info",
					Format: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Validate() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Validate() unexpected error: %v", err)
			}
		})
	}
}

func TestDatabaseConfig_ConnectionString(t *testing.T) {
	tests := []struct {
		name string
		cfg  DatabaseConfig
		want string
	}{
		{
			name: "postgres connection string",
			cfg: DatabaseConfig{
				Driver:   "postgres",
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "disable",
			},
			want: "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable",
		},
		{
			name: "fake driver returns empty",
			cfg: DatabaseConfig{
				Driver: "fake",
			},
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.ConnectionString()
			if got != tt.want {
				t.Errorf("ConnectionString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCryptoConfig_DecodeKeys(t *testing.T) {
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	tests := []struct {
		name    string
		cfg     CryptoConfig
		wantErr bool
	}{
		{
			name: "decode encryption key",
			cfg: CryptoConfig{
				EncryptionKey: validEncKey,
			},
			wantErr: false,
		},
		{
			name: "decode signing key",
			cfg: CryptoConfig{
				SigningKey: validSignKey,
			},
			wantErr: false,
		},
		{
			name: "decode token private key",
			cfg: CryptoConfig{
				TokenPrivateKey: validTokenKey,
			},
			wantErr: false,
		},
		{
			name: "invalid hex encryption key",
			cfg: CryptoConfig{
				EncryptionKey: "not-hex",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.cfg.EncryptionKey != "" {
				_, err := tt.cfg.DecodeEncryptionKey()
				if (err != nil) != tt.wantErr {
					t.Errorf("DecodeEncryptionKey() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			if tt.cfg.SigningKey != "" {
				_, err := tt.cfg.DecodeSigningKey()
				if (err != nil) != tt.wantErr {
					t.Errorf("DecodeSigningKey() error = %v, wantErr %v", err, tt.wantErr)
				}
			}

			if tt.cfg.TokenPrivateKey != "" {
				_, err := tt.cfg.DecodeTokenPrivateKey()
				if (err != nil) != tt.wantErr {
					t.Errorf("DecodeTokenPrivateKey() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

func TestAuthConfig_ParseTokenTTL(t *testing.T) {
	tests := []struct {
		name    string
		cfg     AuthConfig
		wantErr bool
	}{
		{
			name: "valid duration 24h",
			cfg: AuthConfig{
				TokenTTL: "24h",
			},
			wantErr: false,
		},
		{
			name: "valid duration 1h30m",
			cfg: AuthConfig{
				TokenTTL: "1h30m",
			},
			wantErr: false,
		},
		{
			name: "invalid duration",
			cfg: AuthConfig{
				TokenTTL: "invalid",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.cfg.ParseTokenTTL()
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTokenTTL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	// Create a temporary config file
	configYAML := `
server:
  port: ":3000"

database:
  driver: "postgres"
  host: "testhost"
  database: "testdb"

crypto:
  encryptionkey: "` + validEncKey + `"
  signingkey: "` + validSignKey + `"
  tokenprivatekey: "` + validTokenKey + `"

auth:
  tokenttl: "48h"
  passwordlength: 64

log:
  level: "debug"
  format: "json"
`

	tmpfile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(configYAML)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	// Set the config file env var
	os.Setenv("AUTHN_CONFIG_FILE", tmpfile.Name())
	defer os.Unsetenv("AUTHN_CONFIG_FILE")

	cfg, err := Load("AUTHN_")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Port != ":3000" {
		t.Errorf("expected port :3000, got %s", cfg.Server.Port)
	}

	if cfg.Database.Driver != "postgres" {
		t.Errorf("expected driver postgres, got %s", cfg.Database.Driver)
	}

	if cfg.Auth.TokenTTL != "48h" {
		t.Errorf("expected tokenttl 48h, got %s", cfg.Auth.TokenTTL)
	}

	if cfg.Log.Level != "debug" {
		t.Errorf("expected log level debug, got %s", cfg.Log.Level)
	}

	if cfg.Log.Format != "json" {
		t.Errorf("expected log format json, got %s", cfg.Log.Format)
	}
}
