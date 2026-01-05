package database

import (
	"context"
	"embed"
	"testing"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/testhelper"
)

//go:embed testdata
var testAssetsFS embed.FS

func setupTestPostgres(t *testing.T) (*config.Config, func()) {
	t.Helper()
	return testhelper.SetupTestDBWithConfig(t)
}

func TestNew(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "dev",
			Password: "dev",
			Database: "dev",
			Schema:   "test",
			SSLMode:  "disable",
		},
	}

	db := New(testAssetsFS, "postgres", cfg, logger)

	if db == nil {
		t.Error("expected database to be created")
	}

	if db.DB != nil {
		t.Error("expected DB to be nil before Start")
	}
}

func TestStartAndStop(t *testing.T) {
	cfg, cleanup := setupTestPostgres(t)
	defer cleanup()

	logger := log.NewLogger("error")
	db := New(testAssetsFS, "postgres", cfg, logger)
	db.SetMigrationPath("testdata/migration/postgres")

	ctx := context.Background()

	if err := db.Start(ctx); err != nil {
		t.Fatalf("unexpected error starting database: %v", err)
	}

	if db.DB == nil {
		t.Error("expected DB to be set after Start")
	}

	if err := db.DB.PingContext(ctx); err != nil {
		t.Errorf("expected database to be reachable: %v", err)
	}

	var schemaExists bool
	err := db.DB.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = $1)",
		cfg.Database.Schema).Scan(&schemaExists)
	if err != nil {
		t.Fatalf("cannot check schema existence: %v", err)
	}

	if !schemaExists {
		t.Errorf("expected schema %s to exist", cfg.Database.Schema)
	}

	if err := db.Stop(ctx); err != nil {
		t.Errorf("unexpected error stopping database: %v", err)
	}
}

func TestStartWithMigrations(t *testing.T) {
	cfg, cleanup := setupTestPostgres(t)
	defer cleanup()

	logger := log.NewLogger("error")
	db := New(testAssetsFS, "postgres", cfg, logger)
	db.SetMigrationPath("testdata/migration/postgres")

	ctx := context.Background()

	if err := db.Start(ctx); err != nil {
		t.Fatalf("unexpected error starting database: %v", err)
	}
	defer db.Stop(ctx)

	var tableExists bool
	err := db.DB.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = $1 AND table_name = 'test_table')",
		cfg.Database.Schema).Scan(&tableExists)
	if err != nil {
		t.Fatalf("cannot check table existence: %v", err)
	}

	if !tableExists {
		t.Error("expected test_table to be created by migration from testdata")
	}
}

func TestStartWithInvalidConfig(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     "invalid-host",
			Port:     9999,
			User:     "invalid",
			Password: "invalid",
			Database: "invalid",
			SSLMode:  "disable",
		},
	}

	db := New(testAssetsFS, "postgres", cfg, logger)
	ctx := context.Background()

	if err := db.Start(ctx); err == nil {
		t.Error("expected error with invalid config")
	}
}

func TestStopWithoutStart(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := &config.Config{}

	db := New(testAssetsFS, "postgres", cfg, logger)
	ctx := context.Background()

	if err := db.Stop(ctx); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestConnectionString(t *testing.T) {
	tests := []struct {
		name     string
		cfg      config.DatabaseConfig
		expected string
	}{
		{
			name: "with schema",
			cfg: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "dev",
				Password: "dev",
				Database: "dev",
				Schema:   "pulap_lite",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=dev password=dev dbname=dev sslmode=disable search_path=pulap_lite",
		},
		{
			name: "without schema",
			cfg: config.DatabaseConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "dev",
				Password: "dev",
				Database: "dev",
				Schema:   "",
				SSLMode:  "disable",
			},
			expected: "host=localhost port=5432 user=dev password=dev dbname=dev sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cfg.ConnectionString()
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
