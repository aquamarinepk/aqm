package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/posflag"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
	"github.com/spf13/pflag"
)

// Config holds the application configuration.
type Config struct {
	Log      LogConfig      `koanf:"log"`
	Server   ServerConfig   `koanf:"server"`
	Database DatabaseConfig `koanf:"database"`
	Assets   AssetsConfig   `koanf:"assets"`
	Auth     AuthConfig     `koanf:"auth"`
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `koanf:"level"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port string `koanf:"port"`
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	Database string `koanf:"database"`
	Schema   string `koanf:"schema"`
	SSLMode  string `koanf:"sslmode"`
}

// AssetsConfig holds asset storage configuration.
type AssetsConfig struct {
	Storage string      `koanf:"storage"`
	Local   LocalConfig `koanf:"local"`
}

// LocalConfig holds local filesystem storage configuration.
type LocalConfig struct {
	Path string `koanf:"path"`
}

// AuthConfig holds authentication and session configuration.
type AuthConfig struct {
	SessionSecret            string `koanf:"session_secret"`
	TokenTTL                 string `koanf:"token_ttl"`
	EncryptionKey            string `koanf:"encryption_key"`
	SigningKey               string `koanf:"signing_key"`
	TokenPrivateKey          string `koanf:"token_private_key"`
	RegistrationTokenTTL     string `koanf:"registration_token_ttl"`
	PasswordResetTokenTTL    string `koanf:"password_reset_token_ttl"`
	AutoApproveRegistrations bool   `koanf:"auto_approve_registrations"`
}

// New creates a new Config with sensible defaults.
func New() *Config {
	return &Config{
		Log: LogConfig{
			Level: "info",
		},
		Server: ServerConfig{
			Port: ":8080",
		},
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "dev",
			Password: "dev",
			Database: "dev",
			Schema:   "pulap_lite",
			SSLMode:  "disable",
		},
		Assets: AssetsConfig{
			Storage: "local",
			Local: LocalConfig{
				Path: "./data/uploads",
			},
		},
		Auth: AuthConfig{
			SessionSecret:            "change-this-in-production",
			TokenTTL:                 "24h",
			EncryptionKey:            "12345678901234567890123456789012",
			SigningKey:               "abcdefghijklmnopqrstuvwxyz123456",
			TokenPrivateKey:          "ygvuJ/guxUMFKeIcz29Ab763Cq5DT+g2+3mRfGlNiYp0GVI1wTXGsqYlDWqYjPw4G416Z6P2hag8E+/B9GxrSA==",
			RegistrationTokenTTL:     "72h",
			PasswordResetTokenTTL:    "1h",
			AutoApproveRegistrations: false,
		},
	}
}

// LoadConfig loads configuration from a YAML file with environment variable
// and command-line flag overrides.
//
// Configuration precedence (highest to lowest):
//  1. Command-line flags
//  2. Environment variables (with envPrefix)
//  3. YAML file (with env var expansion)
//  4. Default values
//
// Parameters:
//   - path: Path to YAML config file
//   - envPrefix: Prefix for environment variables (e.g., "PULAP_")
//   - args: Command-line arguments (typically os.Args)
//
// Returns the loaded configuration or an error.
func LoadConfig(path, envPrefix string, args []string) (*Config, error) {
	k := koanf.New(".")
	cfg := New()

	fs := pflag.NewFlagSet(args[0], pflag.ExitOnError)
	fs.String("log.level", "info", "Log level (debug, info, error)")
	fs.String("server.port", ":8080", "HTTP server port")
	fs.String("database.host", "localhost", "Database host")
	fs.Int("database.port", 5432, "Database port")
	fs.String("database.user", "dev", "Database user")
	fs.String("database.password", "dev", "Database password")
	fs.String("database.database", "dev", "Database name")
	fs.String("database.schema", "pulap_lite", "Database schema")
	fs.String("database.sslmode", "disable", "Database SSL mode")
	fs.String("assets.storage", "local", "Asset storage backend (local)")
	fs.String("assets.local.path", "./data/uploads", "Local storage path")
	fs.String("auth.session_secret", "change-this-in-production", "Session secret")
	fs.String("auth.token_ttl", "24h", "Auth token TTL")
	fs.String("auth.encryption_key", "12345678901234567890123456789012", "Email encryption key (32 bytes)")
	fs.String("auth.signing_key", "abcdefghijklmnopqrstuvwxyz123456", "Email signing key for lookup hash (32 bytes)")
	fs.String("auth.token_private_key", "ygvuJ/guxUMFKeIcz29Ab763Cq5DT+g2+3mRfGlNiYp0GVI1wTXGsqYlDWqYjPw4G416Z6P2hag8E+/B9GxrSA==", "PASETO token private key (Ed25519 base64)")
	fs.String("auth.registration_token_ttl", "72h", "Registration token TTL")
	fs.String("auth.password_reset_token_ttl", "1h", "Password reset token TTL")
	fs.Bool("auth.auto_approve_registrations", false, "Auto-approve new registrations")
	fs.Parse(args[1:])

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read config file: %w", err)
	}
	expanded := []byte(os.ExpandEnv(string(raw)))

	if err := k.Load(rawbytes.Provider(expanded), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("cannot parse yaml: %w", err)
	}

	if err := k.Load(env.Provider(envPrefix, ".", func(s string) string {
		return strings.Replace(strings.ToLower(
			strings.TrimPrefix(s, envPrefix)), "_", ".", -1)
	}), nil); err != nil {
		return nil, fmt.Errorf("cannot load env vars: %w", err)
	}

	if err := k.Load(posflag.Provider(fs, ".", k), nil); err != nil {
		return nil, fmt.Errorf("cannot load flags: %w", err)
	}

	if err := k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("cannot unmarshal config: %w", err)
	}

	return cfg, nil
}

// ConnectionString builds a PostgreSQL connection string with schema support.
func (d DatabaseConfig) ConnectionString() string {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Database, d.SSLMode)

	if d.Schema != "" {
		connStr += fmt.Sprintf(" search_path=%s", d.Schema)
	}

	return connStr
}
