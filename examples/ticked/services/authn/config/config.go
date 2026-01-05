package config

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

type Config struct {
	Server   ServerConfig   `koanf:"server"`
	Database DatabaseConfig `koanf:"database"`
	Crypto   CryptoConfig   `koanf:"crypto"`
	Auth     AuthConfig     `koanf:"auth"`
	Log      LogConfig      `koanf:"log"`
}

type ServerConfig struct {
	Port string `koanf:"port"`
}

type DatabaseConfig struct {
	Driver   string `koanf:"driver"`
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	Database string `koanf:"database"`
	SSLMode  string `koanf:"sslmode"`
}

type CryptoConfig struct {
	EncryptionKey   string `koanf:"encryptionkey"`
	SigningKey      string `koanf:"signingkey"`
	TokenPrivateKey string `koanf:"tokenprivatekey"`
}

type AuthConfig struct {
	TokenTTL         string `koanf:"tokenttl"`
	PasswordLength   int    `koanf:"passwordlength"`
	EnableBootstrap  bool   `koanf:"enablebootstrap"`
}

type LogConfig struct {
	Level  string `koanf:"level"`
	Format string `koanf:"format"`
}

func New() *Config {
	return &Config{
		Server: ServerConfig{
			Port: ":8080",
		},
		Database: DatabaseConfig{
			Driver:   "fake",
			Host:     "localhost",
			Port:     5432,
			User:     "dev",
			Password: "dev",
			Database: "authn_dev",
			SSLMode:  "disable",
		},
		Crypto: CryptoConfig{
			EncryptionKey:   "",
			SigningKey:      "",
			TokenPrivateKey: "",
		},
		Auth: AuthConfig{
			TokenTTL:        "24h",
			PasswordLength:  32,
			EnableBootstrap: true,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func Load(envPrefix string) (*Config, error) {
	k := koanf.New(".")
	cfg := New()

	if err := k.Load(rawbytes.Provider(defaultConfigYAML), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("load default config: %w", err)
	}

	if configPath := os.Getenv(envPrefix + "CONFIG_FILE"); configPath != "" {
		if data, err := os.ReadFile(configPath); err == nil {
			if err := k.Load(rawbytes.Provider(data), yaml.Parser()); err != nil {
				return nil, fmt.Errorf("load config file %s: %w", configPath, err)
			}
		}
	}

	if err := k.Load(env.Provider(envPrefix, ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, envPrefix)), "_", ".", -1)
	}), nil); err != nil {
		return nil, fmt.Errorf("load environment variables: %w", err)
	}

	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Crypto.EncryptionKey == "" {
		return fmt.Errorf("crypto.encryptionkey is required")
	}

	encKey, err := hex.DecodeString(c.Crypto.EncryptionKey)
	if err != nil {
		return fmt.Errorf("crypto.encryptionkey must be hex-encoded: %w", err)
	}
	if len(encKey) != 32 {
		return fmt.Errorf("crypto.encryptionkey must be 32 bytes (64 hex chars), got %d bytes", len(encKey))
	}

	if c.Crypto.SigningKey == "" {
		return fmt.Errorf("crypto.signingkey is required")
	}

	signKey, err := hex.DecodeString(c.Crypto.SigningKey)
	if err != nil {
		return fmt.Errorf("crypto.signingkey must be hex-encoded: %w", err)
	}
	if len(signKey) != 32 {
		return fmt.Errorf("crypto.signingkey must be 32 bytes (64 hex chars), got %d bytes", len(signKey))
	}

	if c.Crypto.TokenPrivateKey == "" {
		return fmt.Errorf("crypto.tokenprivatekey is required")
	}

	tokenKey, err := base64.StdEncoding.DecodeString(c.Crypto.TokenPrivateKey)
	if err != nil {
		return fmt.Errorf("crypto.tokenprivatekey must be base64-encoded: %w", err)
	}
	if len(tokenKey) != 64 {
		return fmt.Errorf("crypto.tokenprivatekey must be 64 bytes (Ed25519), got %d bytes", len(tokenKey))
	}

	if c.Auth.TokenTTL != "" {
		if _, err := time.ParseDuration(c.Auth.TokenTTL); err != nil {
			return fmt.Errorf("auth.tokenttl invalid duration: %w", err)
		}
	}

	if c.Auth.PasswordLength < 16 {
		return fmt.Errorf("auth.passwordlength must be at least 16, got %d", c.Auth.PasswordLength)
	}

	validDrivers := map[string]bool{"fake": true, "postgres": true}
	if !validDrivers[c.Database.Driver] {
		return fmt.Errorf("database.driver must be 'fake' or 'postgres', got '%s'", c.Database.Driver)
	}

	if c.Database.Driver == "postgres" {
		if c.Database.Host == "" {
			return fmt.Errorf("database.host is required for postgres driver")
		}
		if c.Database.Database == "" {
			return fmt.Errorf("database.database is required for postgres driver")
		}
	}

	validLevels := map[string]bool{"debug": true, "info": true, "error": true}
	if !validLevels[c.Log.Level] {
		return fmt.Errorf("log.level must be 'debug', 'info', or 'error', got '%s'", c.Log.Level)
	}

	validFormats := map[string]bool{"text": true, "json": true}
	if !validFormats[c.Log.Format] {
		return fmt.Errorf("log.format must be 'text' or 'json', got '%s'", c.Log.Format)
	}

	return nil
}

func (c *DatabaseConfig) ConnectionString() string {
	if c.Driver != "postgres" {
		return ""
	}

	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Database, c.SSLMode,
	)
}

func (c *CryptoConfig) DecodeEncryptionKey() ([]byte, error) {
	return hex.DecodeString(c.EncryptionKey)
}

func (c *CryptoConfig) DecodeSigningKey() ([]byte, error) {
	return hex.DecodeString(c.SigningKey)
}

func (c *CryptoConfig) DecodeTokenPrivateKey() ([]byte, error) {
	return base64.StdEncoding.DecodeString(c.TokenPrivateKey)
}

func (c *AuthConfig) ParseTokenTTL() (time.Duration, error) {
	return time.ParseDuration(c.TokenTTL)
}

var defaultConfigYAML = []byte(`
server:
  port: ":8080"

database:
  driver: "fake"
  host: "localhost"
  port: 5432
  user: "dev"
  password: "dev"
  database: "authn_dev"
  sslmode: "disable"

crypto:
  encryptionkey: ""
  signingkey: ""
  tokenprivatekey: ""

auth:
  tokenttl: "24h"
  passwordlength: 32
  enablebootstrap: true

log:
  level: "info"
  format: "text"
`)
