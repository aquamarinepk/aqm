package auth

import (
	"context"

	"github.com/google/uuid"
)

type UserStore interface {
	Create(ctx context.Context, user *User) error
	Get(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmailLookup(ctx context.Context, lookup []byte) (*User, error)
	GetByUsername(ctx context.Context, username string) (*User, error)
	GetByPINLookup(ctx context.Context, lookup []byte) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*User, error)
	ListByStatus(ctx context.Context, status UserStatus) ([]*User, error)
}

type RoleStore interface {
	Create(ctx context.Context, role *Role) error
	Get(ctx context.Context, id uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, name string) (*Role, error)
	Update(ctx context.Context, role *Role) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context) ([]*Role, error)
	ListByStatus(ctx context.Context, status RoleStatus) ([]*Role, error)
}

type GrantStore interface {
	Create(ctx context.Context, grant *Grant) error
	Delete(ctx context.Context, username string, roleID uuid.UUID) error
	GetUserGrants(ctx context.Context, username string) ([]*Grant, error)
	GetRoleGrants(ctx context.Context, roleID uuid.UUID) ([]*Grant, error)
	GetUserRoles(ctx context.Context, username string) ([]*Role, error)
	HasRole(ctx context.Context, username string, roleName string) (bool, error)
}
