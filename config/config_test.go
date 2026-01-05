package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	log "github.com/aquamarinepk/aqm/log"
)

func TestNewWithDefaults(t *testing.T) {
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
		{"log level", cfg.Log.Level, "info"},
		{"server port", cfg.Server.Port, ":8080"},
		{"database driver", cfg.Database.Driver, "fake"},
		{"database host", cfg.Database.Host, "localhost"},
		{"database port", cfg.Database.Port, 5432},
		{"database user", cfg.Database.User, "dev"},
		{"database password", cfg.Database.Password, "dev"},
		{"database name", cfg.Database.Database, "dev"},
		{"database sslmode", cfg.Database.SSLMode, "disable"},
		{"nats url", cfg.NATS.URL, "nats://localhost:4222"},
		{"nats maxreconnect", cfg.NATS.MaxReconnect, 10},
		{"assets storage", cfg.Assets.Storage, "local"},
		{"assets local path", cfg.Assets.Local.Path, "./data/uploads"},
		{"auth session secret", cfg.Auth.SessionSecret, "change-this-in-production"},
		{"auth token ttl", cfg.Auth.TokenTTL, "24h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestNewWithCustomDefaults(t *testing.T) {
	logger := log.NewLogger("info")

	customDefaults := map[string]interface{}{
		"server.port":      ":3000",
		"database.driver":  "postgres",
		"custom.field":     "custom-value",
	}

	cfg, err := New(logger, WithDefaults(customDefaults))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"custom server port", cfg.Server.Port, ":3000"},
		{"custom database driver", cfg.Database.Driver, "postgres"},
		{"baseline log level", cfg.Log.Level, "info"},
		{"baseline database host", cfg.Database.Host, "localhost"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}

	// Test custom field via dynamic access
	if cfg.GetString("custom.field") != "custom-value" {
		t.Errorf("GetString(custom.field) = %q, want %q", cfg.GetString("custom.field"), "custom-value")
	}
}

func TestNewWithFile(t *testing.T) {
	logger := log.NewLogger("info")

	yaml := `
log:
  level: debug
server:
  port: ":9090"
database:
  driver: postgres
  host: db.example.com
  port: 5433
  user: testuser
  password: testpass
  database: testdb
  sslmode: require
nats:
  url: nats://nats.example.com:4222
  clusterid: test-cluster
  clientid: test-client
  maxreconnect: 20
assets:
  storage: local
  local:
    path: /tmp/uploads
auth:
  session_secret: test-secret
  token_ttl: 48h
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	cfg, err := New(logger, WithFile(configPath))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"log level", cfg.Log.Level, "debug"},
		{"server port", cfg.Server.Port, ":9090"},
		{"database driver", cfg.Database.Driver, "postgres"},
		{"database host", cfg.Database.Host, "db.example.com"},
		{"database port", cfg.Database.Port, 5433},
		{"database user", cfg.Database.User, "testuser"},
		{"database password", cfg.Database.Password, "testpass"},
		{"database name", cfg.Database.Database, "testdb"},
		{"database sslmode", cfg.Database.SSLMode, "require"},
		{"nats url", cfg.NATS.URL, "nats://nats.example.com:4222"},
		{"nats clusterid", cfg.NATS.ClusterID, "test-cluster"},
		{"nats clientid", cfg.NATS.ClientID, "test-client"},
		{"nats maxreconnect", cfg.NATS.MaxReconnect, 20},
		{"assets storage", cfg.Assets.Storage, "local"},
		{"assets local path", cfg.Assets.Local.Path, "/tmp/uploads"},
		{"auth session secret", cfg.Auth.SessionSecret, "test-secret"},
		{"auth token ttl", cfg.Auth.TokenTTL, "48h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestNewWithEnvExpansion(t *testing.T) {
	logger := log.NewLogger("info")

	os.Setenv("DB_HOST", "env.db.com")
	os.Setenv("DB_PORT", "6543")
	defer os.Unsetenv("DB_HOST")
	defer os.Unsetenv("DB_PORT")

	yaml := `
database:
  host: ${DB_HOST}
  port: ${DB_PORT}
  user: envuser
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	cfg, err := New(logger, WithFile(configPath), WithEnvExpansion())
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	if cfg.Database.Host != "env.db.com" {
		t.Errorf("host = %q, want %q", cfg.Database.Host, "env.db.com")
	}
	if cfg.Database.Port != 6543 {
		t.Errorf("port = %d, want %d", cfg.Database.Port, 6543)
	}
}

func TestNewWithPrefix(t *testing.T) {
	logger := log.NewLogger("info")

	os.Setenv("TEST_LOG_LEVEL", "error")
	os.Setenv("TEST_SERVER_PORT", ":3000")
	os.Setenv("TEST_DATABASE_HOST", "override.db.com")
	os.Setenv("TEST_DATABASE_DRIVER", "mongo")
	defer os.Unsetenv("TEST_LOG_LEVEL")
	defer os.Unsetenv("TEST_SERVER_PORT")
	defer os.Unsetenv("TEST_DATABASE_HOST")
	defer os.Unsetenv("TEST_DATABASE_DRIVER")

	cfg, err := New(logger, WithPrefix("TEST_"))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"log level from env", cfg.Log.Level, "error"},
		{"server port from env", cfg.Server.Port, ":3000"},
		{"database host from env", cfg.Database.Host, "override.db.com"},
		{"database driver from env", cfg.Database.Driver, "mongo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestNewWithMultipleOptions(t *testing.T) {
	logger := log.NewLogger("info")

	// Set env vars
	os.Setenv("TEST_LOG_LEVEL", "error")
	defer os.Unsetenv("TEST_LOG_LEVEL")

	// Create config file
	yaml := `
server:
  port: ":9090"
database:
  host: file.db.com
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	// Custom defaults
	defaults := map[string]interface{}{
		"custom.value": 42,
	}

	cfg, err := New(logger,
		WithDefaults(defaults),
		WithFile(configPath),
		WithPrefix("TEST_"),
	)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"env overrides file", cfg.Log.Level, "error"},
		{"file loads correctly", cfg.Server.Port, ":9090"},
		{"file loads db host", cfg.Database.Host, "file.db.com"},
		{"defaults baseline", cfg.Database.Port, 5432},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}

	// Test custom field
	if cfg.GetInt("custom.value") != 42 {
		t.Errorf("GetInt(custom.value) = %d, want 42", cfg.GetInt("custom.value"))
	}
}

func TestNewMissingFile(t *testing.T) {
	logger := log.NewLogger("info")

	// Missing file should not fail, just log and continue with defaults
	cfg, err := New(logger, WithFile("/nonexistent/config.yaml"))
	if err != nil {
		t.Fatalf("New() should not fail on missing file: %v", err)
	}

	// Should have defaults
	if cfg.Server.Port != ":8080" {
		t.Errorf("expected default port :8080, got %s", cfg.Server.Port)
	}
}

func TestGetString(t *testing.T) {
	logger := log.NewLogger("info")

	defaults := map[string]interface{}{
		"custom.string": "test-value",
		"nested.deep.value": "deep-value",
	}

	cfg, err := New(logger, WithDefaults(defaults))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{"existing custom", "custom.string", "test-value"},
		{"nested value", "nested.deep.value", "deep-value"},
		{"baseline value", "log.level", "info"},
		{"nonexistent", "does.not.exist", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetString(tt.path)
			if got != tt.want {
				t.Errorf("GetString(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	logger := log.NewLogger("info")

	defaults := map[string]interface{}{
		"custom.int": 42,
		"custom.zero": 0,
	}

	cfg, err := New(logger, WithDefaults(defaults))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name string
		path string
		want int
	}{
		{"custom int", "custom.int", 42},
		{"custom zero", "custom.zero", 0},
		{"baseline int", "database.port", 5432},
		{"nonexistent", "does.not.exist", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetInt(tt.path)
			if got != tt.want {
				t.Errorf("GetInt(%q) = %d, want %d", tt.path, got, tt.want)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	logger := log.NewLogger("info")

	defaults := map[string]interface{}{
		"custom.bool.true": true,
		"custom.bool.false": false,
	}

	cfg, err := New(logger, WithDefaults(defaults))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"custom true", "custom.bool.true", true},
		{"custom false", "custom.bool.false", false},
		{"baseline bool", "auth.auto_approve_registrations", false},
		{"nonexistent", "does.not.exist", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetBool(tt.path)
			if got != tt.want {
				t.Errorf("GetBool(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestGetFloat(t *testing.T) {
	logger := log.NewLogger("info")

	defaults := map[string]interface{}{
		"custom.float": 3.14,
		"custom.zero": 0.0,
	}

	cfg, err := New(logger, WithDefaults(defaults))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name string
		path string
		want float64
	}{
		{"custom float", "custom.float", 3.14},
		{"custom zero", "custom.zero", 0.0},
		{"nonexistent", "does.not.exist", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.GetFloat(tt.path)
			if got != tt.want {
				t.Errorf("GetFloat(%q) = %f, want %f", tt.path, got, tt.want)
			}
		})
	}
}

func TestGetDuration(t *testing.T) {
	logger := log.NewLogger("info")

	defaults := map[string]interface{}{
		"custom.duration": "5m",
		"custom.hours": "2h30m",
	}

	cfg, err := New(logger, WithDefaults(defaults))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name    string
		path    string
		want    time.Duration
		wantErr bool
	}{
		{"custom duration", "custom.duration", 5 * time.Minute, false},
		{"custom hours", "custom.hours", 2*time.Hour + 30*time.Minute, false},
		{"baseline duration", "auth.token_ttl", 24 * time.Hour, false},
		{"nonexistent", "does.not.exist", 0, true},
		{"invalid format", "log.level", 0, true}, // "info" is not a duration
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.GetDuration(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDuration(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("GetDuration(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestExists(t *testing.T) {
	logger := log.NewLogger("info")

	defaults := map[string]interface{}{
		"custom.exists": "value",
	}

	cfg, err := New(logger, WithDefaults(defaults))
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}

	tests := []struct {
		name string
		path string
		want bool
	}{
		{"custom exists", "custom.exists", true},
		{"baseline exists", "log.level", true},
		{"nested exists", "database.host", true},
		{"does not exist", "does.not.exist", false},
		{"partial path", "custom", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cfg.Exists(tt.path)
			if got != tt.want {
				t.Errorf("Exists(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		modify  func(*Config)
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid config",
			modify:  func(c *Config) {},
			wantErr: false,
		},
		{
			name: "missing server port",
			modify: func(c *Config) {
				c.Server.Port = ""
			},
			wantErr: true,
			errMsg:  "server.port is required",
		},
		{
			name: "invalid database driver",
			modify: func(c *Config) {
				c.Database.Driver = "invalid"
			},
			wantErr: true,
			errMsg:  "database.driver must be",
		},
		{
			name: "postgres missing host",
			modify: func(c *Config) {
				c.Database.Driver = "postgres"
				c.Database.Host = ""
			},
			wantErr: true,
			errMsg:  "database.host is required",
		},
		{
			name: "mongo missing host",
			modify: func(c *Config) {
				c.Database.Driver = "mongo"
				c.Database.Host = ""
			},
			wantErr: true,
			errMsg:  "database.host is required",
		},
		{
			name: "fake driver no host required",
			modify: func(c *Config) {
				c.Database.Driver = "fake"
				c.Database.Host = ""
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			modify: func(c *Config) {
				c.Log.Level = "invalid"
			},
			wantErr: true,
			errMsg:  "log.level must be",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewLogger("info")
			cfg, err := New(logger)
			if err != nil {
				t.Fatalf("New() failed: %v", err)
			}

			tt.modify(cfg)

			err = cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("Validate() error = %q, want to contain %q", err.Error(), tt.errMsg)
				}
			}
		})
	}
}

func TestLoadConfigBackwardCompatibility(t *testing.T) {
	yaml := `
log:
  level: debug
server:
  port: ":9090"
database:
  driver: postgres
  host: db.example.com
  port: 5433
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath, "TEST_", []string{"test"})
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"log level", cfg.Log.Level, "debug"},
		{"server port", cfg.Server.Port, ":9090"},
		{"database driver", cfg.Database.Driver, "postgres"},
		{"database host", cfg.Database.Host, "db.example.com"},
		{"database port", cfg.Database.Port, 5433},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoadConfigWithFlags(t *testing.T) {
	yaml := `
log:
  level: debug
server:
  port: ":9090"
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	args := []string{
		"test",
		"--log.level=error",
		"--database.driver=postgres",
		"--database.host=flag.db.com",
	}

	cfg, err := LoadConfig(configPath, "TEST_", args)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"flag overrides file", cfg.Log.Level, "error"},
		{"flag sets driver", cfg.Database.Driver, "postgres"},
		{"flag sets host", cfg.Database.Host, "flag.db.com"},
		{"file value used", cfg.Server.Port, ":9090"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoadConfigWithEnvVars(t *testing.T) {
	os.Setenv("TEST_DATABASE_DRIVER", "mongo")
	os.Setenv("TEST_DATABASE_HOST", "env.db.com")
	os.Setenv("TEST_SERVER_PORT", ":4000")
	defer os.Unsetenv("TEST_DATABASE_DRIVER")
	defer os.Unsetenv("TEST_DATABASE_HOST")
	defer os.Unsetenv("TEST_SERVER_PORT")

	yaml := `
log:
  level: debug
server:
  port: ":9090"
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath, "TEST_", []string{"test"})
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"env overrides file port", cfg.Server.Port, ":4000"},
		{"env sets driver", cfg.Database.Driver, "mongo"},
		{"env sets host", cfg.Database.Host, "env.db.com"},
		{"file value used", cfg.Log.Level, "debug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestNewErrorOnInvalidYAML(t *testing.T) {
	logger := log.NewLogger("info")

	invalidYaml := `
log:
  level: debug
server:
  - invalid
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(invalidYaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	_, err := New(logger, WithFile(configPath))
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestNewErrorOnValidationFailure(t *testing.T) {
	logger := log.NewLogger("info")

	yaml := `
log:
  level: invalid-level
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	_, err := New(logger, WithFile(configPath))
	if err == nil {
		t.Error("expected validation error for invalid log level")
	}
}

func TestConnectionString(t *testing.T) {
	tests := []struct {
		name   string
		config DatabaseConfig
		want   string
	}{
		{
			name: "with schema",
			config: DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "testuser",
				Password: "testpass",
				Database: "testdb",
				SSLMode:  "disable",
				Schema:   "public",
			},
			want: "host=localhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable search_path=public",
		},
		{
			name: "without schema",
			config: DatabaseConfig{
				Host:     "db.example.com",
				Port:     5433,
				User:     "admin",
				Password: "secret",
				Database: "proddb",
				SSLMode:  "require",
				Schema:   "",
			},
			want: "host=db.example.com port=5433 user=admin password=secret dbname=proddb sslmode=require",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.config.ConnectionString()
			if got != tt.want {
				t.Errorf("ConnectionString() = %q, want %q", got, tt.want)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
