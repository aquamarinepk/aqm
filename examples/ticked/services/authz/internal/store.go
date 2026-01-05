package internal

import (
	"database/sql"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/fake"
	"github.com/aquamarinepk/aqm/auth/postgres"
	_ "github.com/lib/pq"
)

// NewPostgresStores creates and returns Postgres-backed store implementations.
// It opens a database connection using the provided connection string and
// returns stores for roles and grants, along with the database handle.
// The caller is responsible for closing the database connection.
func NewPostgresStores(connStr string) (
	auth.RoleStore,
	auth.GrantStore,
	*sql.DB,
	error,
) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, nil, nil, err
	}

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, nil, nil, err
	}

	roleStore := postgres.NewRoleStore(db)
	grantStore := postgres.NewGrantStore(db)

	return roleStore, grantStore, db, nil
}

// NewFakeStores creates and returns in-memory fake store implementations.
// These stores are useful for testing and development without requiring
// a real database. All data is stored in memory and will be lost when
// the process exits.
func NewFakeStores() (auth.RoleStore, auth.GrantStore) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)

	return roleStore, grantStore
}
