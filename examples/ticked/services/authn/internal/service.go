package internal

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"fmt"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/handler"
	"github.com/aquamarinepk/aqm/auth/service"
	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/go-chi/chi/v5"
)

// Service coordinates the authn service components and manages lifecycle.
type Service struct {
	cfg    *config.Config
	logger log.Logger
	db     *sql.DB

	// Stores
	userStore  auth.UserStore
	roleStore  auth.RoleStore
	grantStore auth.GrantStore

	// Crypto services
	crypto   service.CryptoService
	tokenGen service.TokenGenerator
	pwdGen   service.PasswordGenerator
	pinGen   service.PINGenerator

	// Handlers
	authnHandler  *handler.AuthNHandler
	authzHandler  *handler.AuthZHandler
	systemHandler *handler.SystemHandler
}

// New creates a new Service with the given configuration.
// It initializes stores (postgres or fake based on config), crypto services,
// and HTTP handlers. Returns an error if initialization fails.
func New(cfg *config.Config, migrationsFS embed.FS, logger log.Logger) (*Service, error) {
	s := &Service{
		cfg:    cfg,
		logger: logger,
	}

	// Initialize stores based on driver
	if cfg.Database.Driver == "postgres" {
		connStr := cfg.Database.ConnectionString()
		userStore, roleStore, grantStore, db, err := NewPostgresStores(connStr, migrationsFS, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres stores: %w", err)
		}
		s.db = db
		s.userStore = userStore
		s.roleStore = roleStore
		s.grantStore = grantStore
	} else {
		userStore, roleStore, grantStore := NewFakeStores()
		s.userStore = userStore
		s.roleStore = roleStore
		s.grantStore = grantStore
	}

	// Initialize crypto services
	encKeyStr := cfg.GetString("crypto.encryptionkey")
	if encKeyStr == "" {
		return nil, fmt.Errorf("crypto.encryptionkey is required")
	}
	encKey, err := hex.DecodeString(encKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	signKeyStr := cfg.GetString("crypto.signingkey")
	if signKeyStr == "" {
		return nil, fmt.Errorf("crypto.signingkey is required")
	}
	signKey, err := hex.DecodeString(signKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode signing key: %w", err)
	}

	tokenKeyStr := cfg.GetString("crypto.tokenprivatekey")
	if tokenKeyStr == "" {
		return nil, fmt.Errorf("crypto.tokenprivatekey is required")
	}
	tokenKey, err := base64.StdEncoding.DecodeString(tokenKeyStr)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token private key: %w", err)
	}

	tokenTTL, err := cfg.GetDuration("auth.tokenttl")
	if err != nil {
		return nil, fmt.Errorf("failed to parse token TTL: %w", err)
	}

	passwordLength := cfg.GetIntOrDef("auth.passwordlength", 32)

	s.crypto = service.NewDefaultCryptoService(encKey, signKey)
	s.tokenGen = service.NewDefaultTokenGenerator(ed25519.PrivateKey(tokenKey), tokenTTL)

	// Check for dev mode - use fixed password generator for easier development
	if cfg.AQM.DevMode {
		s.pwdGen = service.NewDevPasswordGenerator("")
		logger.Info("WARNING: DEV_MODE enabled, using fixed bootstrap password (Superadmin123!)")
	} else {
		s.pwdGen = service.NewDefaultPasswordGenerator(passwordLength)
	}

	s.pinGen = service.NewDefaultPINGenerator()

	// Initialize handlers
	s.authnHandler = handler.NewAuthNHandler(
		s.userStore,
		s.crypto,
		s.tokenGen,
		s.pwdGen,
		s.pinGen,
	)

	s.authzHandler = handler.NewAuthZHandler(
		s.roleStore,
		s.grantStore,
	)

	s.systemHandler = handler.NewSystemHandler(
		s.userStore,
		s.crypto,
		s.pwdGen,
	)

	return s, nil
}

// Start initializes the service and optionally bootstraps the superadmin user.
func (s *Service) Start(ctx context.Context) error {
	// Bootstrap superadmin if enabled
	if s.cfg.GetBoolOrDef("auth.enablebootstrap", true) {
		if err := s.bootstrap(ctx); err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}
	}

	s.logger.Info("Service started successfully")
	return nil
}

// RegisterRoutes registers all HTTP routes for the service.
func (s *Service) RegisterRoutes(r chi.Router) {
	s.authnHandler.RegisterRoutes(r)
	s.authzHandler.RegisterRoutes(r)
	s.systemHandler.RegisterRoutes(r)
}

// Stop gracefully shuts down the service and closes database connections.
func (s *Service) Stop(ctx context.Context) error {
	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("database close error: %w", err)
		}
	}

	s.logger.Info("Service stopped successfully")
	return nil
}

// bootstrap creates the superadmin user if it doesn't exist.
// It uses the Bootstrap service function which is idempotent.
func (s *Service) bootstrap(ctx context.Context) error {
	user, password, err := service.Bootstrap(ctx, s.userStore, s.crypto, s.pwdGen)
	if err != nil {
		return err
	}

	// Only log if a new superadmin was created (password is returned)
	if password != "" {
		s.logger.Info("============================================")
		s.logger.Infof("Superadmin created: %s", service.SuperadminEmail)
		s.logger.Infof("Password: %s", password)
		s.logger.Info("CHANGE THIS PASSWORD IMMEDIATELY!")
		s.logger.Info("============================================")
	} else {
		s.logger.Infof("Superadmin already exists: %s (ID: %s)", service.SuperadminEmail, user.ID)
	}

	return nil
}
