package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/aquamarinepk/aqm/config"
	"github.com/aquamarinepk/aqm/log"
	"github.com/go-chi/chi/v5"
)

func validTestConfig(t *testing.T) *config.Config {
	t.Helper()

	clearTestEnvVars(t)
	os.Setenv("ADMIN_DATABASE_DRIVER", "fake")
	os.Setenv("ADMIN_SERVER_PORT", ":8084")
	os.Setenv("ADMIN_SERVICES_AUTHN_URL", "http://localhost:8082")
	os.Setenv("ADMIN_SERVICES_AUTHZ_URL", "http://localhost:8083")
	os.Setenv("ADMIN_PREFLIGHT_ENABLED", "false")

	logger := log.NewLogger("error")
	cfg, err := config.New(logger,
		config.WithPrefix("ADMIN_"),
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	return cfg
}

func clearTestEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{
		"ADMIN_SERVER_PORT",
		"ADMIN_DATABASE_DRIVER",
		"ADMIN_SERVICES_AUTHN_URL",
		"ADMIN_SERVICES_AUTHZ_URL",
		"ADMIN_PREFLIGHT_ENABLED",
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
			name: "valid config",
			setup: func(t *testing.T) *config.Config {
				return validTestConfig(t)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setup(t)
			logger := log.NewLogger("error")

			svc, err := New(cfg, logger)

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

			if svc.authnClient == nil {
				t.Error("authnClient is nil")
			}

			if svc.authzClient == nil {
				t.Error("authzClient is nil")
			}

			if svc.handler == nil {
				t.Error("handler is nil")
			}
		})
	}
}

func TestStartStop(t *testing.T) {
	cfg := validTestConfig(t)
	logger := log.NewLogger("error")

	svc, err := New(cfg, logger)
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

func TestStartWithPreflight(t *testing.T) {
	authnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}
	}))
	defer authnServer.Close()

	authzServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}
	}))
	defer authzServer.Close()

	clearTestEnvVars(t)
	os.Setenv("ADMIN_DATABASE_DRIVER", "fake")
	os.Setenv("ADMIN_SERVER_PORT", ":8084")
	os.Setenv("ADMIN_SERVICES_AUTHN_URL", authnServer.URL)
	os.Setenv("ADMIN_SERVICES_AUTHZ_URL", authzServer.URL)
	os.Setenv("ADMIN_PREFLIGHT_ENABLED", "true")

	logger := log.NewLogger("error")
	cfg, err := config.New(logger, config.WithPrefix("ADMIN_"))
	if err != nil {
		t.Fatalf("config.New() error: %v", err)
	}

	svc, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx := context.Background()
	if err := svc.Start(ctx); err != nil {
		t.Errorf("Start() with preflight error: %v", err)
	}
}

func TestRegisterRoutes(t *testing.T) {
	cfg := validTestConfig(t)
	logger := log.NewLogger("error")

	svc, err := New(cfg, logger)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	router := chi.NewRouter()
	svc.RegisterRoutes(router)

	if router == nil {
		t.Error("router is nil after RegisterRoutes")
	}
}
