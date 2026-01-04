package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

type roleStore struct {
	db *sql.DB
}

func NewRoleStore(db *sql.DB) auth.RoleStore {
	return &roleStore{db: db}
}

func (s *roleStore) Create(ctx context.Context, role *auth.Role) error {
	permsJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO roles (
			id, name, description, permissions, status,
			created_at, created_by, updated_at, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9
		)
	`
	_, err = s.db.ExecContext(ctx, query,
		role.ID, role.Name, role.Description, permsJSON, role.Status,
		role.CreatedAt, role.CreatedBy, role.UpdatedAt, role.UpdatedBy,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *roleStore) Get(ctx context.Context, id uuid.UUID) (*auth.Role, error) {
	query := `
		SELECT id, name, description, permissions, status,
			created_at, created_by, updated_at, updated_by
		FROM roles
		WHERE id = $1
	`
	role := &auth.Role{}
	var permsJSON []byte
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&role.ID, &role.Name, &role.Description, &permsJSON, &role.Status,
		&role.CreatedAt, &role.CreatedBy, &role.UpdatedAt, &role.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, auth.ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(permsJSON, &role.Permissions); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleStore) GetByName(ctx context.Context, name string) (*auth.Role, error) {
	query := `
		SELECT id, name, description, permissions, status,
			created_at, created_by, updated_at, updated_by
		FROM roles
		WHERE name = $1
	`
	role := &auth.Role{}
	var permsJSON []byte
	err := s.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description, &permsJSON, &role.Status,
		&role.CreatedAt, &role.CreatedBy, &role.UpdatedAt, &role.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, auth.ErrRoleNotFound
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(permsJSON, &role.Permissions); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *roleStore) Update(ctx context.Context, role *auth.Role) error {
	permsJSON, err := json.Marshal(role.Permissions)
	if err != nil {
		return err
	}

	query := `
		UPDATE roles SET
			name = $2, description = $3, permissions = $4, status = $5,
			updated_at = $6, updated_by = $7
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query,
		role.ID, role.Name, role.Description, permsJSON, role.Status,
		role.UpdatedAt, role.UpdatedBy,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return auth.ErrRoleNotFound
	}
	return nil
}

func (s *roleStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE roles SET status = 'inactive', updated_at = NOW() WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return auth.ErrRoleNotFound
	}
	return nil
}

func (s *roleStore) List(ctx context.Context) ([]*auth.Role, error) {
	query := `
		SELECT id, name, description, permissions, status,
			created_at, created_by, updated_at, updated_by
		FROM roles
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*auth.Role
	for rows.Next() {
		role := &auth.Role{}
		var permsJSON []byte
		err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &permsJSON, &role.Status,
			&role.CreatedAt, &role.CreatedBy, &role.UpdatedAt, &role.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(permsJSON, &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

func (s *roleStore) ListByStatus(ctx context.Context, status auth.RoleStatus) ([]*auth.Role, error) {
	query := `
		SELECT id, name, description, permissions, status,
			created_at, created_by, updated_at, updated_by
		FROM roles
		WHERE status = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []*auth.Role
	for rows.Next() {
		role := &auth.Role{}
		var permsJSON []byte
		err := rows.Scan(
			&role.ID, &role.Name, &role.Description, &permsJSON, &role.Status,
			&role.CreatedAt, &role.CreatedBy, &role.UpdatedAt, &role.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		if err := json.Unmarshal(permsJSON, &role.Permissions); err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}
	return roles, rows.Err()
}

var _ auth.RoleStore = (*roleStore)(nil)
