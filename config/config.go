package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	log "github.com/aquamarinepk/aqm/log"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
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
	NATS     NATSConfig     `koanf:"nats"`
	Assets   AssetsConfig   `koanf:"assets"`
	Auth     AuthConfig     `koanf:"auth"`

	// Internal fields (not marshaled by koanf)
	k      *koanf.Koanf
	logger log.Logger
}

// LogConfig holds logging configuration.
type LogConfig struct {
	Level string `koanf:"level"`
}

// ServerConfig holds HTTP server configuration.
type ServerConfig struct {
	Port string `koanf:"port"`
}

// DatabaseConfig holds database connection configuration.
type DatabaseConfig struct {
	Driver   string `koanf:"driver"`
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

// NATSConfig holds NATS connection configuration.
type NATSConfig struct {
	URL          string `koanf:"url"`
	ClusterID    string `koanf:"clusterid"`
	ClientID     string `koanf:"clientid"`
	MaxReconnect int    `koanf:"maxreconnect"`
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

// Option configures Config during initialization.
type Option func(*configOptions) error

// configOptions holds option values during initialization.
type configOptions struct {
	prefix       string
	file         string
	defaults     map[string]interface{}
	envExpansion bool
}

// WithPrefix sets the environment variable prefix (e.g., "AUTHN_").
func WithPrefix(prefix string) Option {
	return func(opts *configOptions) error {
		opts.prefix = prefix
		return nil
	}
}

// WithFile loads configuration from a YAML file.
func WithFile(path string) Option {
	return func(opts *configOptions) error {
		opts.file = path
		return nil
	}
}

// WithDefaults provides default values via a map.
func WithDefaults(defaults map[string]interface{}) Option {
	return func(opts *configOptions) error {
		opts.defaults = defaults
		return nil
	}
}

// WithEnvExpansion enables ${VAR} expansion in config files.
func WithEnvExpansion() Option {
	return func(opts *configOptions) error {
		opts.envExpansion = true
		return nil
	}
}

// New creates a new Config with logger and options.
func New(logger log.Logger, opts ...Option) (*Config, error) {
	cfg := &Config{
		logger: logger,
		k:      koanf.New("."),
	}

	// Apply options
	options := &configOptions{
		prefix:       "",
		file:         "",
		defaults:     make(map[string]interface{}),
		envExpansion: false,
	}

	for _, opt := range opts {
		if err := opt(options); err != nil {
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	// Set baseline defaults
	baselineDefaults := map[string]interface{}{
		"log.level":                          "info",
		"server.port":                        ":8080",
		"database.driver":                    "fake",
		"database.host":                      "localhost",
		"database.port":                      5432,
		"database.user":                      "dev",
		"database.password":                  "dev",
		"database.database":                  "dev",
		"database.schema":                    "pulap_lite",
		"database.sslmode":                   "disable",
		"nats.url":                           "nats://localhost:4222",
		"nats.clusterid":                     "",
		"nats.clientid":                      "",
		"nats.maxreconnect":                  10,
		"assets.storage":                     "local",
		"assets.local.path":                  "./data/uploads",
		"auth.session_secret":                "change-this-in-production",
		"auth.token_ttl":                     "24h",
		"auth.encryption_key":                "12345678901234567890123456789012",
		"auth.signing_key":                   "abcdefghijklmnopqrstuvwxyz123456",
		"auth.token_private_key":             "ygvuJ/guxUMFKeIcz29Ab763Cq5DT+g2+3mRfGlNiYp0GVI1wTXGsqYlDWqYjPw4G416Z6P2hag8E+/B9GxrSA==",
		"auth.registration_token_ttl":        "72h",
		"auth.password_reset_token_ttl":      "1h",
		"auth.auto_approve_registrations":    false,
	}

	// Merge baseline defaults with user-provided defaults
	for k, v := range baselineDefaults {
		if _, exists := options.defaults[k]; !exists {
			options.defaults[k] = v
		}
	}

	// Load defaults
	if err := cfg.k.Load(confmap.Provider(options.defaults, "."), nil); err != nil {
		return nil, fmt.Errorf("failed to load defaults: %w", err)
	}

	// Load file if specified
	if options.file != "" {
		raw, err := os.ReadFile(options.file)
		if err != nil {
			logger.Debugf("Config file not found: %s (using defaults)", options.file)
		} else {
			if options.envExpansion {
				raw = []byte(os.ExpandEnv(string(raw)))
			}
			if err := cfg.k.Load(rawbytes.Provider(raw), yaml.Parser()); err != nil {
				return nil, fmt.Errorf("failed to parse config file: %w", err)
			}
			logger.Debugf("Loaded config from file: %s", options.file)
		}
	}

	// Load environment variables if prefix specified
	if options.prefix != "" {
		if err := cfg.k.Load(env.Provider(options.prefix, ".", func(s string) string {
			return strings.Replace(strings.ToLower(
				strings.TrimPrefix(s, options.prefix)), "_", ".", -1)
		}), nil); err != nil {
			return nil, fmt.Errorf("failed to load environment variables: %w", err)
		}
	}

	// Unmarshal to struct
	if err := cfg.k.Unmarshal("", cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	logger.Infof("Configuration loaded: driver=%s, port=%s, log=%s",
		cfg.Database.Driver, cfg.Server.Port, cfg.Log.Level)

	return cfg, nil
}

// GetString returns the string value for the given path.
func (c *Config) GetString(path string) string {
	return c.k.String(path)
}

// GetInt returns the int value for the given path.
func (c *Config) GetInt(path string) int {
	return c.k.Int(path)
}

// GetBool returns the bool value for the given path.
func (c *Config) GetBool(path string) bool {
	return c.k.Bool(path)
}

// GetFloat returns the float64 value for the given path.
func (c *Config) GetFloat(path string) float64 {
	return c.k.Float64(path)
}

// GetDuration parses and returns a time.Duration for the given path.
func (c *Config) GetDuration(path string) (time.Duration, error) {
	s := c.k.String(path)
	if s == "" {
		return 0, fmt.Errorf("no value found for path: %s", path)
	}
	return time.ParseDuration(s)
}

// Exists returns true if the given path exists in the configuration.
func (c *Config) Exists(path string) bool {
	return c.k.Exists(path)
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	// Validate Server
	if c.Server.Port == "" {
		return fmt.Errorf("server.port is required")
	}

	// Validate Database
	validDrivers := map[string]bool{"fake": true, "postgres": true, "mongo": true}
	if !validDrivers[c.Database.Driver] {
		return fmt.Errorf("database.driver must be 'fake', 'postgres', or 'mongo', got '%s'", c.Database.Driver)
	}

	if c.Database.Driver == "postgres" || c.Database.Driver == "mongo" {
		if c.Database.Host == "" {
			return fmt.Errorf("database.host is required for %s driver", c.Database.Driver)
		}
	}

	// Validate Log
	validLevels := map[string]bool{"debug": true, "info": true, "error": true}
	if !validLevels[c.Log.Level] {
		return fmt.Errorf("log.level must be 'debug', 'info', or 'error', got '%s'", c.Log.Level)
	}

	c.logger.Debugf("Configuration validated successfully")

	return nil
}

// LoadConfig loads configuration from a YAML file with environment variable
// and command-line flag overrides.
//
// Deprecated: Use New() with Options pattern instead for better flexibility.
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
	// Create a simple logger for backward compatibility
	logger := log.NewLogger("info")

	// Use new Options pattern
	cfg, err := New(logger,
		WithPrefix(envPrefix),
		WithFile(path),
		WithEnvExpansion(),
	)
	if err != nil {
		return nil, err
	}

	// Handle command-line flags (legacy support)
	if len(args) > 1 {
		k := cfg.k
		fs := pflag.NewFlagSet(args[0], pflag.ExitOnError)
		fs.String("log.level", cfg.Log.Level, "Log level (debug, info, error)")
		fs.String("server.port", cfg.Server.Port, "HTTP server port")
		fs.String("database.driver", cfg.Database.Driver, "Database driver (fake, postgres, mongo)")
		fs.String("database.host", cfg.Database.Host, "Database host")
		fs.Int("database.port", cfg.Database.Port, "Database port")
		fs.String("database.user", cfg.Database.User, "Database user")
		fs.String("database.password", cfg.Database.Password, "Database password")
		fs.String("database.database", cfg.Database.Database, "Database name")
		fs.String("database.schema", cfg.Database.Schema, "Database schema")
		fs.String("database.sslmode", cfg.Database.SSLMode, "Database SSL mode")
		fs.String("assets.storage", cfg.Assets.Storage, "Asset storage backend (local)")
		fs.String("assets.local.path", cfg.Assets.Local.Path, "Local storage path")
		fs.String("auth.session_secret", cfg.Auth.SessionSecret, "Session secret")
		fs.String("auth.token_ttl", cfg.Auth.TokenTTL, "Auth token TTL")
		fs.String("auth.encryption_key", cfg.Auth.EncryptionKey, "Email encryption key (32 bytes)")
		fs.String("auth.signing_key", cfg.Auth.SigningKey, "Email signing key for lookup hash (32 bytes)")
		fs.String("auth.token_private_key", cfg.Auth.TokenPrivateKey, "PASETO token private key (Ed25519 base64)")
		fs.String("auth.registration_token_ttl", cfg.Auth.RegistrationTokenTTL, "Registration token TTL")
		fs.String("auth.password_reset_token_ttl", cfg.Auth.PasswordResetTokenTTL, "Password reset token TTL")
		fs.Bool("auth.auto_approve_registrations", cfg.Auth.AutoApproveRegistrations, "Auto-approve new registrations")
		fs.Parse(args[1:])

		if err := k.Load(posflag.Provider(fs, ".", k), nil); err != nil {
			return nil, fmt.Errorf("cannot load flags: %w", err)
		}

		// Re-unmarshal with flags applied
		if err := k.Unmarshal("", cfg); err != nil {
			return nil, fmt.Errorf("cannot unmarshal config: %w", err)
		}
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
