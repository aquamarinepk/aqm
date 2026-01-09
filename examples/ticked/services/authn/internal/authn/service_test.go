package authn

import (
	"context"
	"embed"
	"encoding/base64"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/config"
	log "github.com/aquamarinepk/aqm/log"
)

//go:embed testdata/.gitkeep
var testMigrationsFS embed.FS

func validTestConfig(t *testing.T) *config.Config {
	t.Helper()

	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	// Clear and set env vars
	clearTestEnvVars(t)
	os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
	os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
	os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
	os.Setenv("AUTHN_DATABASE_DRIVER", "fake")
	os.Setenv("AUTHN_SERVER_PORT", ":8080")
	os.Setenv("AUTHN_AUTH_ENABLEBOOTSTRAP", "false")

	logger := log.NewLogger("info")
	cfg, err := config.New(logger,
		config.WithPrefix("AUTHN_"),
		config.WithDefaults(map[string]interface{}{
			"crypto.encryptionkey":   "",
			"crypto.signingkey":      "",
			"crypto.tokenprivatekey": "",
			"auth.tokenttl":          "24h",
			"auth.passwordlength":    32,
			"auth.enablebootstrap":   true,
		}),
	)
	if err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	return cfg
}

func clearTestEnvVars(t *testing.T) {
	t.Helper()
	envVars := []string{
		"AUTHN_SERVER_PORT",
		"AUTHN_DATABASE_DRIVER",
		"AUTHN_DATABASE_HOST",
		"AUTHN_DATABASE_DATABASE",
		"AUTHN_CRYPTO_ENCRYPTIONKEY",
		"AUTHN_CRYPTO_SIGNINGKEY",
		"AUTHN_CRYPTO_TOKENPRIVATEKEY",
		"AUTHN_AUTH_TOKENTTL",
		"AUTHN_AUTH_PASSWORDLENGTH",
		"AUTHN_AUTH_ENABLEBOOTSTRAP",
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
				validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
				validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
				validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

				clearTestEnvVars(t)
				os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
				os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
				os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
				os.Setenv("AUTHN_DATABASE_DRIVER", "postgres")
				os.Setenv("AUTHN_DATABASE_HOST", "localhost")
				os.Setenv("AUTHN_DATABASE_DATABASE", "test")

				logger := log.NewLogger("info")
				cfg, err := config.New(logger)
				if err != nil {
					t.Fatalf("failed to create config: %v", err)
				}
				return cfg
			},
			wantErr: false, // Will skip if no DB available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := tt.setup(t)
			svc, err := New(testMigrationsFS, cfg, log.NewNoopLogger())

			if cfg.Database.Driver == "postgres" && err != nil {
				// Skip if postgres is not available
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

			if svc.userStore == nil {
				t.Error("userStore is nil")
			}

			if svc.roleStore == nil {
				t.Error("roleStore is nil")
			}

			if svc.grantStore == nil {
				t.Error("grantStore is nil")
			}

			if svc.crypto == nil {
				t.Error("crypto is nil")
			}

			if svc.tokenGen == nil {
				t.Error("tokenGen is nil")
			}

			if svc.pwdGen == nil {
				t.Error("pwdGen is nil")
			}

			if svc.pinGen == nil {
				t.Error("pinGen is nil")
			}

			if svc.authnHandler == nil {
				t.Error("authnHandler is nil")
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

func TestNewInvalidCryptoKeys(t *testing.T) {
	tests := []struct {
		name  string
		setup func(*testing.T)
	}{
		{
			name: "invalid encryption key",
			setup: func(t *testing.T) {
				validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
				validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

				clearTestEnvVars(t)
				os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", "invalid")
				os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
				os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
			},
		},
		{
			name: "invalid signing key",
			setup: func(t *testing.T) {
				validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
				validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

				clearTestEnvVars(t)
				os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
				os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", "invalid")
				os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
			},
		},
		{
			name: "invalid token key",
			setup: func(t *testing.T) {
				validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
				validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))

				clearTestEnvVars(t)
				os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
				os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
				os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", "invalid")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup(t)
			logger := log.NewLogger("info")
			cfg, err := config.New(logger)
			if err == nil {
				t.Error("config.New() expected error for invalid key, got nil")
				return
			}

			// If config creation failed (validation), that's expected
			// If config was created, service creation should fail
			if cfg != nil {
				_, err := New(testMigrationsFS, cfg, log.NewNoopLogger())
				if err == nil {
					t.Error("New() expected error for invalid key, got nil")
				}
			}
		})
	}
}

func TestBootstrap(t *testing.T) {
	cfg := validTestConfig(t)

	// Re-create config with bootstrap enabled
	clearTestEnvVars(t)
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
	os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
	os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
	os.Setenv("AUTHN_AUTH_ENABLEBOOTSTRAP", "true")

	logger := log.NewLogger("info")
	cfg, err := config.New(logger)
	if err != nil {
		t.Fatalf("config.New() error: %v", err)
	}

	svc, err := New(testMigrationsFS, cfg, log.NewNoopLogger())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx := context.Background()
	err = svc.bootstrap(ctx)
	if err != nil {
		t.Errorf("bootstrap() error: %v", err)
	}

	// Run bootstrap again - should be idempotent
	err = svc.bootstrap(ctx)
	if err != nil {
		t.Errorf("bootstrap() second call error: %v", err)
	}
}

func TestStartStop(t *testing.T) {
	cfg := validTestConfig(t)

	// Re-create config with dynamic port
	clearTestEnvVars(t)
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
	os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
	os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
	os.Setenv("AUTHN_SERVER_PORT", ":0") // Let OS assign a port

	logger := log.NewLogger("info")
	cfg, err := config.New(logger)
	if err != nil {
		t.Fatalf("config.New() error: %v", err)
	}

	svc, err := New(testMigrationsFS, cfg, log.NewNoopLogger())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	// Start server in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- svc.Start(ctx)
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Stop server
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	if err := svc.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error: %v", err)
	}

	// Check that Start() exited cleanly
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Start() error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Start() did not exit after Stop()")
	}
}

func TestStartWithBootstrap(t *testing.T) {
	// Redirect log output to suppress bootstrap messages in test output
	oldLogger := os.Stdout
	os.Stdout, _ = os.Open(os.DevNull)
	defer func() { os.Stdout = oldLogger }()

	clearTestEnvVars(t)
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	os.Setenv("AUTHN_CRYPTO_ENCRYPTIONKEY", validEncKey)
	os.Setenv("AUTHN_CRYPTO_SIGNINGKEY", validSignKey)
	os.Setenv("AUTHN_CRYPTO_TOKENPRIVATEKEY", validTokenKey)
	os.Setenv("AUTHN_SERVER_PORT", ":0")
	os.Setenv("AUTHN_AUTH_ENABLEBOOTSTRAP", "true")

	logger := log.NewLogger("info")
	cfg, err := config.New(logger)
	if err != nil {
		t.Fatalf("config.New() error: %v", err)
	}

	svc, err := New(testMigrationsFS, cfg, log.NewNoopLogger())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errCh := make(chan error, 1)
	go func() {
		errCh <- svc.Start(ctx)
	}()

	// Give server time to start and bootstrap
	time.Sleep(200 * time.Millisecond)

	// Stop server
	stopCtx, stopCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer stopCancel()

	if err := svc.Stop(stopCtx); err != nil {
		t.Errorf("Stop() error: %v", err)
	}

	// Check that Start() exited cleanly
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Start() error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("Start() did not exit after Stop()")
	}
}
