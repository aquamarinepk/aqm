package testhelper

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDB creates an isolated test database environment.
// In CI (when DB_HOST is set), it uses the existing PostgreSQL instance with a unique schema.
// Locally, it spins up a testcontainer for complete isolation.
func SetupTestDB(t *testing.T) (*sql.DB, string, func()) {
	t.Helper()
	ctx := context.Background()

	// Check if we're in CI environment
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		return setupCIDatabase(t, ctx)
	}

	// Local development: use testcontainers
	return setupTestContainer(t, ctx)
}

// setupCIDatabase creates a unique schema in the shared CI database
func setupCIDatabase(t *testing.T, ctx context.Context) (*sql.DB, string, func()) {
	t.Helper()

	dbHost := os.Getenv("DB_HOST")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "postgres")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "postgres")
	dbName := getEnvOrDefault("DB_NAME", "postgres")

	// Generate unique schema name for this test
	schema := fmt.Sprintf("test_%d_%s", time.Now().UnixNano(), randomString(8))

	// Connect to database
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("cannot open database: %v", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("cannot ping database: %v", err)
	}

	// Create unique schema
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA %s", schema))
	if err != nil {
		db.Close()
		t.Fatalf("cannot create schema %s: %v", schema, err)
	}

	// Set search_path to use this schema by default
	_, err = db.ExecContext(ctx, fmt.Sprintf("SET search_path TO %s", schema))
	if err != nil {
		db.Close()
		t.Fatalf("cannot set search_path: %v", err)
	}

	cleanup := func() {
		// Drop schema and all its objects
		_, _ = db.ExecContext(context.Background(),
			fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))
		db.Close()
	}

	return db, schema, cleanup
}

// setupTestContainer creates a new PostgreSQL container for local testing
func setupTestContainer(t *testing.T, ctx context.Context) (*sql.DB, string, func()) {
	t.Helper()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2)),
	)
	if err != nil {
		t.Fatalf("cannot start postgres container: %v", err)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("cannot get connection string: %v", err)
	}

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("cannot open database: %v", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		t.Fatalf("cannot ping database: %v", err)
	}

	// For testcontainers, we use the public schema
	schema := "public"

	cleanup := func() {
		db.Close()
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("cannot terminate container: %v", err)
		}
	}

	return db, schema, cleanup
}

// SetupTestDBWithConfig returns a config.Config for testing
func SetupTestDBWithConfig(t *testing.T) (*config.Config, func()) {
	t.Helper()
	ctx := context.Background()

	// Check if we're in CI environment
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		return setupCIDatabaseWithConfig(t, ctx)
	}

	// Local development: use testcontainers
	return setupTestContainerWithConfig(t, ctx)
}

func setupCIDatabaseWithConfig(t *testing.T, ctx context.Context) (*config.Config, func()) {
	t.Helper()

	dbHost := os.Getenv("DB_HOST")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "postgres")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "postgres")
	dbName := getEnvOrDefault("DB_NAME", "postgres")

	// Generate unique schema name for this test
	schema := fmt.Sprintf("test_%d_%s", time.Now().UnixNano(), randomString(8))

	// Connect to database to create schema
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("cannot open database: %v", err)
	}
	defer db.Close()

	// Create unique schema
	_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE SCHEMA %s", schema))
	if err != nil {
		t.Fatalf("cannot create schema %s: %v", schema, err)
	}

	// Parse port as int
	port := 5432
	fmt.Sscanf(dbPort, "%d", &port)

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     dbHost,
			Port:     port,
			User:     dbUser,
			Password: dbPassword,
			Database: dbName,
			Schema:   schema,
			SSLMode:  "disable",
		},
	}

	cleanup := func() {
		// Reconnect to drop schema
		db, err := sql.Open("pgx", connStr)
		if err != nil {
			t.Logf("cannot open database for cleanup: %v", err)
			return
		}
		defer db.Close()

		_, _ = db.ExecContext(context.Background(),
			fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))
	}

	return cfg, cleanup
}

func setupTestContainerWithConfig(t *testing.T, ctx context.Context) (*config.Config, func()) {
	t.Helper()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2)),
	)
	if err != nil {
		t.Fatalf("cannot start postgres container: %v", err)
	}

	host, err := pgContainer.Host(ctx)
	if err != nil {
		t.Fatalf("cannot get container host: %v", err)
	}

	port, err := pgContainer.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("cannot get container port: %v", err)
	}

	cfg := &config.Config{
		Database: config.DatabaseConfig{
			Host:     host,
			Port:     port.Int(),
			User:     "postgres",
			Password: "postgres",
			Database: "testdb",
			Schema:   "public",
			SSLMode:  "disable",
		},
	}

	cleanup := func() {
		if err := pgContainer.Terminate(context.Background()); err != nil {
			t.Logf("cannot terminate container: %v", err)
		}
	}

	return cfg, cleanup
}

// TestLogger returns a logger suitable for testing
func TestLogger() log.Logger {
	return log.NewLogger("error")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range b {
		b[i] = charset[rng.Intn(len(charset))]
	}
	return string(b)
}
