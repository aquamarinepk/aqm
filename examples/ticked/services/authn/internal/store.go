package internal

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/fake"
	"github.com/aquamarinepk/aqm/auth/postgres"
	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/migrate"
	_ "github.com/lib/pq"
)

// NewPostgresStores creates and returns Postgres-backed store implementations.
// It opens a database connection using the provided connection string and
// returns stores for users, roles, and grants, along with the database handle.
// Runs migrations before returning stores.
// The caller is responsible for closing the database connection.
func NewPostgresStores(connStr string, migrationsFS embed.FS, logger log.Logger) (
	auth.UserStore,
	auth.RoleStore,
	auth.GrantStore,
	*sql.DB,
	error,
) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, nil, nil, nil, err
	}

	// Run migrations
	migrator := migrate.New(migrationsFS, "postgres", logger)
	migrator.SetDB(db)
	migrator.SetPath("migrations")

	if err := migrator.Run(context.Background()); err != nil {
		db.Close()
		return nil, nil, nil, nil, fmt.Errorf("migration failed: %w", err)
	}

	userStore := postgres.NewUserStore(db)
	roleStore := postgres.NewRoleStore(db)
	grantStore := postgres.NewGrantStore(db)

	return userStore, roleStore, grantStore, db, nil
}

// NewFakeStores creates and returns in-memory fake store implementations.
// These stores are useful for testing and development without requiring
// a real database. All data is stored in memory and will be lost when
// the process exits.
func NewFakeStores() (auth.UserStore, auth.RoleStore, auth.GrantStore) {
	userStore := fake.NewUserStore()
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)

	return userStore, roleStore, grantStore
}
