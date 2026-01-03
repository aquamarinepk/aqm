package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	cfg := New()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"log level", cfg.Log.Level, "info"},
		{"server port", cfg.Server.Port, ":8080"},
		{"database host", cfg.Database.Host, "localhost"},
		{"database port", cfg.Database.Port, 5432},
		{"database user", cfg.Database.User, "dev"},
		{"database password", cfg.Database.Password, "dev"},
		{"database name", cfg.Database.Database, "dev"},
		{"database sslmode", cfg.Database.SSLMode, "disable"},
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

func TestLoadConfigFromYAML(t *testing.T) {
	yaml := `
log:
  level: debug
server:
  port: ":9090"
database:
  host: db.example.com
  port: 5433
  user: testuser
  password: testpass
  database: testdb
  sslmode: require
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
		{"database host", cfg.Database.Host, "db.example.com"},
		{"database port", cfg.Database.Port, 5433},
		{"database user", cfg.Database.User, "testuser"},
		{"database password", cfg.Database.Password, "testpass"},
		{"database name", cfg.Database.Database, "testdb"},
		{"database sslmode", cfg.Database.SSLMode, "require"},
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

func TestLoadConfigWithEnvVarExpansion(t *testing.T) {
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

	cfg, err := LoadConfig(configPath, "TEST_", []string{"test"})
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Database.Host != "env.db.com" {
		t.Errorf("host = %q, want %q", cfg.Database.Host, "env.db.com")
	}
	if cfg.Database.Port != 6543 {
		t.Errorf("port = %d, want %d", cfg.Database.Port, 6543)
	}
}

func TestLoadConfigWithEnvOverrides(t *testing.T) {
	os.Setenv("TEST_LOG_LEVEL", "error")
	os.Setenv("TEST_SERVER_PORT", ":3000")
	os.Setenv("TEST_DATABASE_HOST", "override.db.com")
	defer os.Unsetenv("TEST_LOG_LEVEL")
	defer os.Unsetenv("TEST_SERVER_PORT")
	defer os.Unsetenv("TEST_DATABASE_HOST")

	yaml := `
log:
  level: debug
server:
  port: ":9090"
database:
  host: original.db.com
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
		{"log level overridden", cfg.Log.Level, "error"},
		{"server port overridden", cfg.Server.Port, ":3000"},
		{"database host overridden", cfg.Database.Host, "override.db.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoadConfigWithFlagOverrides(t *testing.T) {
	yaml := `
log:
  level: debug
server:
  port: ":9090"
database:
  host: yaml.db.com
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	args := []string{
		"test",
		"--log.level=warn",
		"--server.port=:4000",
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
		{"log level from flag", cfg.Log.Level, "warn"},
		{"server port from flag", cfg.Server.Port, ":4000"},
		{"database host from flag", cfg.Database.Host, "flag.db.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}

func TestLoadConfigPrecedence(t *testing.T) {
	os.Setenv("TEST_LOG_LEVEL", "error")
	defer os.Unsetenv("TEST_LOG_LEVEL")

	yaml := `
log:
  level: debug
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(yaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	args := []string{"test", "--log.level=warn"}

	cfg, err := LoadConfig(configPath, "TEST_", args)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	if cfg.Log.Level != "warn" {
		t.Errorf("expected flag to override env and yaml, got %q", cfg.Log.Level)
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/config.yaml", "TEST_", []string{"test"})
	if err == nil {
		t.Error("expected error for missing config file")
	}
}

func TestLoadConfigInvalidYAML(t *testing.T) {
	invalidYaml := `
log:
  level: debug
server
  port: invalid
`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(invalidYaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	_, err := LoadConfig(configPath, "TEST_", []string{"test"})
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	emptyYaml := `{}`
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(emptyYaml), 0644); err != nil {
		t.Fatalf("cannot write test config: %v", err)
	}

	cfg, err := LoadConfig(configPath, "TEST_", []string{"test"})
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	defaults := New()

	tests := []struct {
		name string
		got  interface{}
		want interface{}
	}{
		{"log level defaults", cfg.Log.Level, defaults.Log.Level},
		{"server port defaults", cfg.Server.Port, defaults.Server.Port},
		{"database host defaults", cfg.Database.Host, defaults.Database.Host},
		{"assets storage defaults", cfg.Assets.Storage, defaults.Assets.Storage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %v, want %v", tt.got, tt.want)
			}
		})
	}
}
