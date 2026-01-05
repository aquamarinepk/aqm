package internal

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/examples/ticked/services/authn/config"
)

func validTestConfig() *config.Config {
	validEncKey := hex.EncodeToString([]byte("12345678901234567890123456789012"))
	validSignKey := hex.EncodeToString([]byte("abcdefghijklmnopqrstuvwxyz123456"))
	validTokenKey := base64.StdEncoding.EncodeToString(make([]byte, 64))

	return &config.Config{
		Server: config.ServerConfig{
			Port: ":8080",
		},
		Database: config.DatabaseConfig{
			Driver: "fake",
		},
		Crypto: config.CryptoConfig{
			EncryptionKey:   validEncKey,
			SigningKey:      validSignKey,
			TokenPrivateKey: validTokenKey,
		},
		Auth: config.AuthConfig{
			TokenTTL:        "24h",
			PasswordLength:  32,
			EnableBootstrap: false,
		},
		Log: config.LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *config.Config
		wantErr bool
	}{
		{
			name:    "valid fake config",
			cfg:     validTestConfig(),
			wantErr: false,
		},
		{
			name: "postgres config",
			cfg: func() *config.Config {
				cfg := validTestConfig()
				cfg.Database.Driver = "postgres"
				cfg.Database.Host = "localhost"
				cfg.Database.Database = "test"
				return cfg
			}(),
			wantErr: false, // Will skip if no DB available
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, err := New(tt.cfg)

			if tt.cfg.Database.Driver == "postgres" && err != nil {
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

func TestNew_InvalidCryptoKeys(t *testing.T) {
	tests := []struct {
		name string
		cfg  *config.Config
	}{
		{
			name: "invalid encryption key",
			cfg: func() *config.Config {
				cfg := validTestConfig()
				cfg.Crypto.EncryptionKey = "invalid"
				return cfg
			}(),
		},
		{
			name: "invalid signing key",
			cfg: func() *config.Config {
				cfg := validTestConfig()
				cfg.Crypto.SigningKey = "invalid"
				return cfg
			}(),
		},
		{
			name: "invalid token key",
			cfg: func() *config.Config {
				cfg := validTestConfig()
				cfg.Crypto.TokenPrivateKey = "invalid"
				return cfg
			}(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.cfg)
			if err == nil {
				t.Error("New() expected error for invalid key, got nil")
			}
		})
	}
}

func TestBootstrap(t *testing.T) {
	cfg := validTestConfig()
	cfg.Auth.EnableBootstrap = true

	svc, err := New(cfg)
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
	// Use a random available port
	cfg := validTestConfig()
	cfg.Server.Port = ":0" // Let OS assign a port

	svc, err := New(cfg)
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

	cfg := validTestConfig()
	cfg.Server.Port = ":0"
	cfg.Auth.EnableBootstrap = true

	svc, err := New(cfg)
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
