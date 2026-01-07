package postgres

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

type grantStore struct {
	db *sql.DB
}

func NewGrantStore(db *sql.DB) auth.GrantStore {
	return &grantStore{db: db}
}

func (s *grantStore) Create(ctx context.Context, grant *auth.Grant) error {
	query := `
		INSERT INTO grants (id, username, role_id, assigned_at, assigned_by)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query,
		grant.ID, grant.Username, grant.RoleID,
		grant.AssignedAt, grant.AssignedBy,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *grantStore) Delete(ctx context.Context, username string, roleID uuid.UUID) error {
	query := `DELETE FROM grants WHERE username = $1 AND role_id = $2`
	result, err := s.db.ExecContext(ctx, query, username, roleID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return auth.ErrGrantNotFound
	}
	return nil
}

func (s *grantStore) GetUserGrants(ctx context.Context, username string) ([]*auth.Grant, error) {
	query := `
		SELECT id, username, role_id, assigned_at, assigned_by
		FROM grants
		WHERE username = $1
		ORDER BY assigned_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []*auth.Grant
	for rows.Next() {
		grant := &auth.Grant{}
		err := rows.Scan(
			&grant.ID, &grant.Username, &grant.RoleID,
			&grant.AssignedAt, &grant.AssignedBy,
		)
		if err != nil {
			return nil, err
		}
		grants = append(grants, grant)
	}
	return grants, rows.Err()
}

func (s *grantStore) GetRoleGrants(ctx context.Context, roleID uuid.UUID) ([]*auth.Grant, error) {
	query := `
		SELECT id, username, role_id, assigned_at, assigned_by
		FROM grants
		WHERE role_id = $1
		ORDER BY assigned_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []*auth.Grant
	for rows.Next() {
		grant := &auth.Grant{}
		err := rows.Scan(
			&grant.ID, &grant.Username, &grant.RoleID,
			&grant.AssignedAt, &grant.AssignedBy,
		)
		if err != nil {
			return nil, err
		}
		grants = append(grants, grant)
	}
	return grants, rows.Err()
}

func (s *grantStore) GetUserRoles(ctx context.Context, username string) ([]*auth.Role, error) {
	query := `
		SELECT r.id, r.name, r.description, r.permissions, r.status,
			r.created_at, r.created_by, r.updated_at, r.updated_by
		FROM roles r
		INNER JOIN grants g ON g.role_id = r.id
		WHERE g.username = $1
		ORDER BY r.name ASC
	`
	rows, err := s.db.QueryContext(ctx, query, username)
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

func (s *grantStore) HasRole(ctx context.Context, username string, roleName string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM grants g
			INNER JOIN roles r ON r.id = g.role_id
			WHERE g.username = $1 AND r.name = $2
		)
	`
	var exists bool
	err := s.db.QueryRowContext(ctx, query, username, roleName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

var _ auth.GrantStore = (*grantStore)(nil)
