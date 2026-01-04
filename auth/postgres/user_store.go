package postgres

import (
	"context"
	"database/sql"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/google/uuid"
)

type userStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) auth.UserStore {
	return &userStore{db: db}
}

func (s *userStore) Create(ctx context.Context, user *auth.User) error {
	query := `
		INSERT INTO users (
			id, username, name,
			email_ct, email_iv, email_tag, email_lookup,
			password_hash, password_salt,
			mfa_secret_ct, pin_ct, pin_iv, pin_tag, pin_lookup,
			status, created_at, created_by, updated_at, updated_by
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
		)
	`
	_, err := s.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Name,
		user.EmailCT, user.EmailIV, user.EmailTag, user.EmailLookup,
		user.PasswordHash, user.PasswordSalt,
		user.MFASecretCT, user.PINCT, user.PINIV, user.PINTag, user.PINLookup,
		user.Status, user.CreatedAt, user.CreatedBy, user.UpdatedAt, user.UpdatedBy,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *userStore) Get(ctx context.Context, id uuid.UUID) (*auth.User, error) {
	query := `
		SELECT id, username, name,
			email_ct, email_iv, email_tag, email_lookup,
			password_hash, password_salt,
			mfa_secret_ct, pin_ct, pin_iv, pin_tag, pin_lookup,
			status, created_at, created_by, updated_at, updated_by
		FROM users
		WHERE id = $1
	`
	user := &auth.User{}
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Name,
		&user.EmailCT, &user.EmailIV, &user.EmailTag, &user.EmailLookup,
		&user.PasswordHash, &user.PasswordSalt,
		&user.MFASecretCT, &user.PINCT, &user.PINIV, &user.PINTag, &user.PINLookup,
		&user.Status, &user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, auth.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userStore) GetByEmailLookup(ctx context.Context, lookup []byte) (*auth.User, error) {
	query := `
		SELECT id, username, name,
			email_ct, email_iv, email_tag, email_lookup,
			password_hash, password_salt,
			mfa_secret_ct, pin_ct, pin_iv, pin_tag, pin_lookup,
			status, created_at, created_by, updated_at, updated_by
		FROM users
		WHERE email_lookup = $1
	`
	user := &auth.User{}
	err := s.db.QueryRowContext(ctx, query, lookup).Scan(
		&user.ID, &user.Username, &user.Name,
		&user.EmailCT, &user.EmailIV, &user.EmailTag, &user.EmailLookup,
		&user.PasswordHash, &user.PasswordSalt,
		&user.MFASecretCT, &user.PINCT, &user.PINIV, &user.PINTag, &user.PINLookup,
		&user.Status, &user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, auth.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userStore) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	query := `
		SELECT id, username, name,
			email_ct, email_iv, email_tag, email_lookup,
			password_hash, password_salt,
			mfa_secret_ct, pin_ct, pin_iv, pin_tag, pin_lookup,
			status, created_at, created_by, updated_at, updated_by
		FROM users
		WHERE username = $1
	`
	user := &auth.User{}
	err := s.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.Username, &user.Name,
		&user.EmailCT, &user.EmailIV, &user.EmailTag, &user.EmailLookup,
		&user.PasswordHash, &user.PasswordSalt,
		&user.MFASecretCT, &user.PINCT, &user.PINIV, &user.PINTag, &user.PINLookup,
		&user.Status, &user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, auth.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userStore) GetByPINLookup(ctx context.Context, lookup []byte) (*auth.User, error) {
	query := `
		SELECT id, username, name,
			email_ct, email_iv, email_tag, email_lookup,
			password_hash, password_salt,
			mfa_secret_ct, pin_ct, pin_iv, pin_tag, pin_lookup,
			status, created_at, created_by, updated_at, updated_by
		FROM users
		WHERE pin_lookup = $1
	`
	user := &auth.User{}
	err := s.db.QueryRowContext(ctx, query, lookup).Scan(
		&user.ID, &user.Username, &user.Name,
		&user.EmailCT, &user.EmailIV, &user.EmailTag, &user.EmailLookup,
		&user.PasswordHash, &user.PasswordSalt,
		&user.MFASecretCT, &user.PINCT, &user.PINIV, &user.PINTag, &user.PINLookup,
		&user.Status, &user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
	)
	if err == sql.ErrNoRows {
		return nil, auth.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userStore) Update(ctx context.Context, user *auth.User) error {
	query := `
		UPDATE users SET
			username = $2, name = $3,
			email_ct = $4, email_iv = $5, email_tag = $6, email_lookup = $7,
			password_hash = $8, password_salt = $9,
			mfa_secret_ct = $10, pin_ct = $11, pin_iv = $12, pin_tag = $13, pin_lookup = $14,
			status = $15, updated_at = $16, updated_by = $17
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Name,
		user.EmailCT, user.EmailIV, user.EmailTag, user.EmailLookup,
		user.PasswordHash, user.PasswordSalt,
		user.MFASecretCT, user.PINCT, user.PINIV, user.PINTag, user.PINLookup,
		user.Status, user.UpdatedAt, user.UpdatedBy,
	)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return auth.ErrUserNotFound
	}
	return nil
}

func (s *userStore) Delete(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE users SET status = 'deleted', updated_at = NOW() WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return auth.ErrUserNotFound
	}
	return nil
}

func (s *userStore) List(ctx context.Context) ([]*auth.User, error) {
	query := `
		SELECT id, username, name,
			email_ct, email_iv, email_tag, email_lookup,
			password_hash, password_salt,
			mfa_secret_ct, pin_ct, pin_iv, pin_tag, pin_lookup,
			status, created_at, created_by, updated_at, updated_by
		FROM users
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*auth.User
	for rows.Next() {
		user := &auth.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Name,
			&user.EmailCT, &user.EmailIV, &user.EmailTag, &user.EmailLookup,
			&user.PasswordHash, &user.PasswordSalt,
			&user.MFASecretCT, &user.PINCT, &user.PINIV, &user.PINTag, &user.PINLookup,
			&user.Status, &user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *userStore) ListByStatus(ctx context.Context, status auth.UserStatus) ([]*auth.User, error) {
	query := `
		SELECT id, username, name,
			email_ct, email_iv, email_tag, email_lookup,
			password_hash, password_salt,
			mfa_secret_ct, pin_ct, pin_iv, pin_tag, pin_lookup,
			status, created_at, created_by, updated_at, updated_by
		FROM users
		WHERE status = $1
		ORDER BY created_at DESC
	`
	rows, err := s.db.QueryContext(ctx, query, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*auth.User
	for rows.Next() {
		user := &auth.User{}
		err := rows.Scan(
			&user.ID, &user.Username, &user.Name,
			&user.EmailCT, &user.EmailIV, &user.EmailTag, &user.EmailLookup,
			&user.PasswordHash, &user.PasswordSalt,
			&user.MFASecretCT, &user.PINCT, &user.PINIV, &user.PINTag, &user.PINLookup,
			&user.Status, &user.CreatedAt, &user.CreatedBy, &user.UpdatedAt, &user.UpdatedBy,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

var _ auth.UserStore = (*userStore)(nil)
