package seed

import (
	"context"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/log"
	"github.com/google/uuid"
)

type Config struct {
	EncryptionKey []byte
	SigningKey    []byte
}

type Seeder struct {
	users  auth.UserStore
	roles  auth.RoleStore
	grants auth.GrantStore
	cfg    *Config
	log    log.Logger
}

func New(users auth.UserStore, roles auth.RoleStore, grants auth.GrantStore, cfg *Config, logger log.Logger) *Seeder {
	return &Seeder{
		users:  users,
		roles:  roles,
		grants: grants,
		cfg:    cfg,
		log:    logger,
	}
}

type RoleInput struct {
	Name        string
	Description string
	Permissions []string
	CreatedBy   string
}

func (s *Seeder) SeedRole(ctx context.Context, input RoleInput) (*auth.Role, error) {
	role := auth.NewRole()
	role.Name = input.Name
	role.Description = input.Description
	role.Permissions = input.Permissions
	role.CreatedBy = input.CreatedBy
	role.UpdatedBy = input.CreatedBy

	role.BeforeCreate()

	if err := role.Validate(); err != nil {
		s.log.Errorf("role validation failed: name=%s error=%v", input.Name, err)
		return nil, err
	}

	if err := s.roles.Create(ctx, role); err != nil {
		s.log.Errorf("failed to create role: name=%s error=%v", input.Name, err)
		return nil, err
	}

	s.log.Infof("seeded role: id=%s name=%s", role.ID, role.Name)
	return role, nil
}

type UserInput struct {
	Username  string
	Name      string
	Email     string
	Password  string
	PIN       string
	CreatedBy string
}

func (s *Seeder) SeedUser(ctx context.Context, input UserInput) (*auth.User, error) {
	user := auth.NewUser()
	user.Username = input.Username
	user.Name = input.Name
	user.CreatedBy = input.CreatedBy
	user.UpdatedBy = input.CreatedBy

	if err := user.SetEmail(input.Email, s.cfg.EncryptionKey, s.cfg.SigningKey); err != nil {
		s.log.Errorf("failed to set email: username=%s error=%v", input.Username, err)
		return nil, err
	}

	if err := user.SetPassword(input.Password); err != nil {
		s.log.Errorf("failed to set password: username=%s error=%v", input.Username, err)
		return nil, err
	}

	if input.PIN != "" {
		if err := user.SetPIN(input.PIN, s.cfg.EncryptionKey, s.cfg.SigningKey); err != nil {
			s.log.Errorf("failed to set PIN: username=%s error=%v", input.Username, err)
			return nil, err
		}
	}

	user.BeforeCreate()

	if err := user.Validate(); err != nil {
		s.log.Errorf("user validation failed: username=%s error=%v", input.Username, err)
		return nil, err
	}

	if err := s.users.Create(ctx, user); err != nil {
		s.log.Errorf("failed to create user: username=%s error=%v", input.Username, err)
		return nil, err
	}

	s.log.Infof("seeded user: id=%s username=%s", user.ID, user.Username)
	return user, nil
}

type GrantInput struct {
	UserID     uuid.UUID
	RoleID     uuid.UUID
	AssignedBy string
}

func (s *Seeder) SeedGrant(ctx context.Context, input GrantInput) (*auth.Grant, error) {
	grant := auth.NewGrant(input.UserID, input.RoleID, input.AssignedBy)

	if err := grant.Validate(); err != nil {
		s.log.Errorf("grant validation failed: user_id=%s role_id=%s error=%v", input.UserID, input.RoleID, err)
		return nil, err
	}

	if err := s.grants.Create(ctx, grant); err != nil {
		s.log.Errorf("failed to create grant: user_id=%s role_id=%s error=%v", input.UserID, input.RoleID, err)
		return nil, err
	}

	s.log.Infof("seeded grant: id=%s user_id=%s role_id=%s", grant.ID, grant.UserID, grant.RoleID)
	return grant, nil
}
