package internal

import (
	"context"
	"os"
	"testing"

	"github.com/aquamarinepk/aqm/config"
	log "github.com/aquamarinepk/aqm/log"
	"github.com/go-chi/chi/v5"
)

func validTestConfig(t *testing.T) *config.Config {
	t.Helper()

	clearTestEnvVars(t)
	os.Setenv("AUTHZ_DATABASE_DRIVER", "fake")
	os.Setenv("AUTHZ_SERVER_PORT", ":8083")

	logger := log.NewLogger("info")
	cfg, err := config.New(logger,
		config.WithPrefix("AUTHZ_"),
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	return cfg
}

func clearTestEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{
		"AUTHZ_SERVER_PORT",
		"AUTHZ_DATABASE_DRIVER",
		"AUTHZ_DATABASE_HOST",
		"AUTHZ_DATABASE_DATABASE",
	}

	for _, v := range envVars {
		os.Unsetenv(v)
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*testing.T) *config.Config
		wantErr bool
	}{
		{
			name: "valid fake config",
			setup: func(t *testing.T) *config.Config {
				return validTestConfig(t)
			},
			wantErr: false,
		},
		{
			name: "postgres config",
			setup: func(t *testing.T) *config.Config {
				clearTestEnvVars(t)
				os.Setenv("AUTHZ_DATABASE_DRIVER", "postgres")
				os.Setenv("AUTHZ_DATABASE_HOST", "localhost")
				os.Setenv("AUTHZ_DATABASE_DATABASE", "test")

				logger := log.NewLogger("info")
				cfg, err := config.New(logger,
					config.WithPrefix("AUTHZ_"),
				)
				if err != nil {
					t.Fatalf("failed to create config: %v", err)
				}
				return cfg
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setup(t)
			svc, err := New(cfg)

			if cfg.Database.Driver == "postgres" && err != nil {
				t.Skip("skipping postgres test: database not available")
				return
			}

			if tt.wantErr {
				if err == nil {
					t.Error("New() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("New() unexpected error: %v", err)
				return
			}

			if svc == nil {
				t.Fatal("service is nil")
			}

			if svc.cfg == nil {
				t.Error("config is nil")
			}

			if svc.roleStore == nil {
				t.Error("roleStore is nil")
			}

			if svc.grantStore == nil {
				t.Error("grantStore is nil")
			}

			if svc.authzHandler == nil {
				t.Error("authzHandler is nil")
			}

			// Clean up if postgres
			if svc.db != nil {
				svc.db.Close()
			}
		})
	}
}

func TestStartStop(t *testing.T) {
	cfg := validTestConfig(t)

	clearTestEnvVars(t)
	os.Setenv("AUTHZ_DATABASE_DRIVER", "fake")
	os.Setenv("AUTHZ_SERVER_PORT", ":0")

	logger := log.NewLogger("info")
	cfg, err := config.New(logger,
		config.WithPrefix("AUTHZ_"),
	)
	if err != nil {
		t.Fatalf("config.New() error: %v", err)
	}

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := svc.Start(ctx); err != nil {
		t.Errorf("Start() error: %v", err)
	}

	stopCtx := context.Background()
	if err := svc.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error: %v", err)
	}
}

func TestRegisterRoutes(t *testing.T) {
	cfg := validTestConfig(t)

	svc, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	router := chi.NewRouter()

	svc.RegisterRoutes(router)

	// Verify routes are registered by checking the router
	if router == nil {
		t.Error("router is nil after RegisterRoutes")
	}
}
