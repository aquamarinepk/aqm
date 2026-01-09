package authz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/seed"
	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
)

// BootstrapService orchestrates the authz bootstrap process by coordinating with authn.
type BootstrapService struct {
	roleStore  auth.RoleStore
	grantStore auth.GrantStore
	seeder     *seed.Seeder
	authnURL   string
	httpClient *http.Client
	log        log.Logger
}

// NOTE: With username-based grants, AuthZ no longer requires UUID synchronization
// with AuthN. Each service maintains its own internal identifiers.
// The username serves as the natural key for cross-service correlation.

type BootstrapStatusResponse struct {
	NeedsBootstrap     bool   `json:"needs_bootstrap"`
	SuperadminID       string `json:"superadmin_id,omitempty"`       // kept for compatibility
	SuperadminUsername string `json:"superadmin_username,omitempty"` // primary identifier
}

type BootstrapResponse struct {
	SuperadminID       string `json:"superadmin_id"`       // kept for compatibility
	SuperadminUsername string `json:"superadmin_username"` // primary identifier
	Email              string `json:"email"`
	Password           string `json:"password"`
}

// NewBootstrapService creates a new bootstrap service with required dependencies.
func NewBootstrapService(roleStore auth.RoleStore, grantStore auth.GrantStore, seeder *seed.Seeder, cfg *config.Config, logger log.Logger) *BootstrapService {
	authnURL := cfg.GetStringOrDef("authn.url", "http://localhost:8082")

	return &BootstrapService{
		roleStore:  roleStore,
		grantStore: grantStore,
		seeder:     seeder,
		authnURL:   authnURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		log:        logger,
	}
}

// Bootstrap orchestrates the complete bootstrap process.
// This is idempotent and safe to call multiple times.
func (s *BootstrapService) Bootstrap(ctx context.Context) error {
	s.log.Infof("Starting authz bootstrap process...")

	status, err := s.getBootstrapStatus(ctx)
	if err != nil {
		return fmt.Errorf("failed to get bootstrap status from authn: %w", err)
	}

	var superadminUsername string

	if status.NeedsBootstrap {
		s.log.Infof("System needs bootstrap, triggering authn bootstrap...")

		response, err := s.triggerBootstrap(ctx)
		if err != nil {
			return fmt.Errorf("failed to trigger authn bootstrap: %w", err)
		}

		superadminUsername = response.SuperadminUsername
		s.log.Infof("Authn bootstrap completed: username=%s email=%s", response.SuperadminUsername, response.Email)

		if response.Password != "" {
			s.log.Infof("============================================")
			s.log.Infof("Superadmin Password: %s", response.Password)
			s.log.Infof("SAVE THIS PASSWORD SECURELY!")
			s.log.Infof("============================================")
		}
	} else {
		s.log.Infof("System already bootstrapped: username=%s", status.SuperadminUsername)
		superadminUsername = status.SuperadminUsername
	}

	if err := s.bootstrapRoles(ctx); err != nil {
		return fmt.Errorf("failed to bootstrap roles: %w", err)
	}

	if err := s.ensureSuperadminGrant(ctx, superadminUsername); err != nil {
		return fmt.Errorf("failed to ensure superadmin grant: %w", err)
	}

	s.log.Infof("Authz bootstrap process completed successfully")
	return nil
}

// getBootstrapStatus calls authn to check if system needs bootstrap
func (s *BootstrapService) getBootstrapStatus(ctx context.Context) (*BootstrapStatusResponse, error) {
	url := s.authnURL + "/system/bootstrap-status"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bootstrap status request failed: %d", resp.StatusCode)
	}

	var response BootstrapStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// triggerBootstrap calls authn to create superadmin
func (s *BootstrapService) triggerBootstrap(ctx context.Context) (*BootstrapResponse, error) {
	url := s.authnURL + "/system/bootstrap"

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bootstrap request failed: %d", resp.StatusCode)
	}

	var response BootstrapResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

// bootstrapRoles creates default roles (superadmin, admin, user, etc.)
func (s *BootstrapService) bootstrapRoles(ctx context.Context) error {
	roles := []seed.RoleInput{
		{
			Name:        "superadmin",
			Description: "Super administrator with all permissions",
			Permissions: []string{"*"},
			CreatedBy:   "system",
		},
		{
			Name:        "admin",
			Description: "Administrator with management permissions",
			Permissions: []string{"users:*", "roles:*", "grants:*"},
			CreatedBy:   "system",
		},
		{
			Name:        "user",
			Description: "Regular user with basic permissions",
			Permissions: []string{"profile:read", "profile:update"},
			CreatedBy:   "system",
		},
	}

	for _, roleInput := range roles {
		existing, err := s.roleStore.GetByName(ctx, roleInput.Name)
		if err != nil && err != auth.ErrRoleNotFound {
			return fmt.Errorf("failed to check role existence: %w", err)
		}

		if existing != nil {
			s.log.Infof("Role already exists: name=%s id=%s", roleInput.Name, existing.ID)
			continue
		}

		_, err = s.seeder.SeedRole(ctx, roleInput)
		if err != nil {
			return fmt.Errorf("failed to seed role %s: %w", roleInput.Name, err)
		}
	}

	return nil
}

// ensureSuperadminGrant creates grant for superadmin if it doesn't exist
func (s *BootstrapService) ensureSuperadminGrant(ctx context.Context, superadminUsername string) error {
	if superadminUsername == "" {
		return fmt.Errorf("superadmin username is required")
	}

	role, err := s.roleStore.GetByName(ctx, "superadmin")
	if err != nil {
		return fmt.Errorf("superadmin role not found: %w", err)
	}

	grants, err := s.grantStore.GetUserGrants(ctx, superadminUsername)
	if err != nil {
		s.log.Errorf("Failed to check existing grants (continuing): %v", err)
	} else {
		for _, g := range grants {
			if g.RoleID == role.ID {
				s.log.Infof("Superadmin grant already exists: grant_id=%s", g.ID)
				return nil
			}
		}
	}

	_, err = s.seeder.SeedGrant(ctx, seed.GrantInput{
		Username:   superadminUsername,
		RoleID:     role.ID,
		AssignedBy: "system",
	})
	if err != nil {
		return fmt.Errorf("failed to create superadmin grant: %w", err)
	}

	s.log.Infof("Superadmin grant created successfully: username=%s role_id=%s", superadminUsername, role.ID)
	return nil
}
