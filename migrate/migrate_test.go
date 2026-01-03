package migrate

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/testhelper"
)

//go:embed testdata
var testAssetsFS embed.FS

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()

	db, schema, cleanup := testhelper.SetupTestDB(t)

	// Set search_path for this test
	ctx := context.Background()
	_, err := db.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", schema))
	if err != nil {
		cleanup()
		t.Fatalf("cannot set search_path: %v", err)
	}

	return db, cleanup
}

func TestNew(t *testing.T) {
	log := logger.NewLogger("error")

	tests := []struct {
		name   string
		engine string
	}{
		{
			name:   "postgres engine",
			engine: "postgres",
		},
		{
			name:   "sqlite engine",
			engine: "sqlite",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator := New(testAssetsFS, tt.engine, log)

			if migrator == nil {
				t.Error("expected migrator to be created")
			}

			if migrator.engine != tt.engine {
				t.Errorf("expected engine %q, got %q", tt.engine, migrator.engine)
			}
		})
	}
}

func TestMigratorRunCreatesTable(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := logger.NewLogger("error")
	migrator := New(testAssetsFS, "postgres", log)
	migrator.SetDB(db)
	migrator.SetPath("testdata/migration/postgres")

	err := migrator.Run(context.Background())
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'migrations'
		)
	`).Scan(&exists)

	if err != nil {
		t.Fatalf("failed to check table existence: %v", err)
	}

	if !exists {
		t.Error("migrations table was not created")
	}
}

func TestMigratorRunWithMigrations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := logger.NewLogger("error")
	migrator := New(testAssetsFS, "postgres", log)
	migrator.SetDB(db)
	migrator.SetPath("testdata/migration/postgres")

	err := migrator.Run(context.Background())
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	var exists bool
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'test_table'
		)
	`).Scan(&exists)

	if err != nil {
		t.Fatalf("failed to check table existence: %v", err)
	}

	if !exists {
		t.Error("test_table was not created by migration")
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to count migrations: %v", err)
	}

	if count < 1 {
		t.Errorf("expected at least 1 migration, got %d", count)
	}
}

func TestMigratorRunIdempotent(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := logger.NewLogger("error")
	migrator := New(testAssetsFS, "postgres", log)
	migrator.SetDB(db)
	migrator.SetPath("testdata/migration/postgres")

	err := migrator.Run(context.Background())
	if err != nil {
		t.Fatalf("first Run failed: %v", err)
	}

	var firstCount int
	err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&firstCount)
	if err != nil {
		t.Fatalf("failed to count migrations: %v", err)
	}

	err = migrator.Run(context.Background())
	if err != nil {
		t.Fatalf("second Run failed: %v", err)
	}

	var secondCount int
	err = db.QueryRow("SELECT COUNT(*) FROM migrations").Scan(&secondCount)
	if err != nil {
		t.Fatalf("failed to count migrations: %v", err)
	}

	if firstCount != secondCount {
		t.Errorf("migrations reran: first=%d, second=%d", firstCount, secondCount)
	}
}

func TestSetPath(t *testing.T) {
	log := logger.NewLogger("error")
	migrator := New(testAssetsFS, "postgres", log)

	tests := []struct {
		name string
		path string
	}{
		{
			name: "custom path",
			path: "custom/migrations",
		},
		{
			name: "empty path",
			path: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			migrator.SetPath(tt.path)

			if migrator.path != tt.path {
				t.Errorf("expected path %q, got %q", tt.path, migrator.path)
			}
		})
	}
}

func TestSetDB(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	log := logger.NewLogger("error")
	migrator := New(testAssetsFS, "postgres", log)

	if migrator.db != nil {
		t.Error("expected db to be nil before SetDB")
	}

	migrator.SetDB(db)

	if migrator.db == nil {
		t.Error("expected db to be set after SetDB")
	}
}
