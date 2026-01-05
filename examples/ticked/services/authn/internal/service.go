package internal

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/handler"
	"github.com/aquamarinepk/aqm/auth/service"
	"github.com/aquamarinepk/aqm/examples/ticked/services/authn/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Service coordinates the authn service components and manages the HTTP server lifecycle.
type Service struct {
	cfg    *config.Config
	db     *sql.DB
	server *http.Server

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
	authnHandler *handler.AuthNHandler
	authzHandler *handler.AuthZHandler
}

// New creates a new Service with the given configuration.
// It initializes stores (postgres or fake based on config), crypto services,
// and HTTP handlers. Returns an error if initialization fails.
func New(cfg *config.Config) (*Service, error) {
	s := &Service{
		cfg: cfg,
	}

	// Initialize stores based on driver
	if cfg.Database.Driver == "postgres" {
		connStr := cfg.Database.ConnectionString()
		userStore, roleStore, grantStore, db, err := NewPostgresStores(connStr)
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
	encKey, err := cfg.Crypto.DecodeEncryptionKey()
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	signKey, err := cfg.Crypto.DecodeSigningKey()
	if err != nil {
		return nil, fmt.Errorf("failed to decode signing key: %w", err)
	}

	tokenKey, err := cfg.Crypto.DecodeTokenPrivateKey()
	if err != nil {
		return nil, fmt.Errorf("failed to decode token private key: %w", err)
	}

	tokenTTL, err := cfg.Auth.ParseTokenTTL()
	if err != nil {
		return nil, fmt.Errorf("failed to parse token TTL: %w", err)
	}

	s.crypto = service.NewDefaultCryptoService(encKey, signKey)
	s.tokenGen = service.NewDefaultTokenGenerator(ed25519.PrivateKey(tokenKey), tokenTTL)
	s.pwdGen = service.NewDefaultPasswordGenerator(cfg.Auth.PasswordLength)
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

	return s, nil
}

// Start initializes the HTTP server, optionally bootstraps the superadmin user,
// and starts listening for requests. This method blocks until the server is shutdown.
func (s *Service) Start(ctx context.Context) error {
	// Bootstrap superadmin if enabled
	if s.cfg.Auth.EnableBootstrap {
		if err := s.bootstrap(ctx); err != nil {
			return fmt.Errorf("bootstrap failed: %w", err)
		}
	}

	// Create router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	// Register routes
	s.authnHandler.RegisterRoutes(r)
	s.authzHandler.RegisterRoutes(r)

	// Create HTTP server
	s.server = &http.Server{
		Addr:         s.cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting server on %s", s.cfg.Server.Port)

	// Start server
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the HTTP server and closes database connections.
// It waits for active connections to finish within the context deadline.
func (s *Service) Stop(ctx context.Context) error {
	if s.server != nil {
		if err := s.server.Shutdown(ctx); err != nil {
			return fmt.Errorf("server shutdown error: %w", err)
		}
	}

	if s.db != nil {
		if err := s.db.Close(); err != nil {
			return fmt.Errorf("database close error: %w", err)
		}
	}

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
		log.Println("============================================")
		log.Printf("Superadmin created: %s", service.SuperadminEmail)
		log.Printf("Password: %s", password)
		log.Println("CHANGE THIS PASSWORD IMMEDIATELY!")
		log.Println("============================================")
	} else {
		log.Printf("Superadmin already exists: %s (ID: %s)", service.SuperadminEmail, user.ID)
	}

	return nil
}
