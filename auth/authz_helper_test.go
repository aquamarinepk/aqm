package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/log"
)

func TestCheckPermission(t *testing.T) {
	tests := []struct {
		name         string
		userID       string
		permission   string
		resource     string
		responseCode int
		responseBody string
		wantAllowed  bool
		wantErr      bool
		setupCache   func(*AuthzHelper)
	}{
		{
			name:         "permission allowed",
			userID:       "user-123",
			permission:   "read",
			resource:     "documents",
			responseCode: http.StatusOK,
			responseBody: `{"allowed": true}`,
			wantAllowed:  true,
			wantErr:      false,
		},
		{
			name:         "permission denied",
			userID:       "user-123",
			permission:   "delete",
			resource:     "documents",
			responseCode: http.StatusOK,
			responseBody: `{"allowed": false}`,
			wantAllowed:  false,
			wantErr:      false,
		},
		{
			name:         "cached permission allowed",
			userID:       "user-456",
			permission:   "write",
			resource:     "files",
			responseCode: http.StatusOK,
			responseBody: `{"allowed": false}`,
			wantAllowed:  true,
			wantErr:      false,
			setupCache: func(h *AuthzHelper) {
				h.cache.set("user-456:write:files", true, 5*time.Minute)
			},
		},
		{
			name:         "cached permission denied",
			userID:       "user-789",
			permission:   "admin",
			resource:     "system",
			responseCode: http.StatusOK,
			responseBody: `{"allowed": true}`,
			wantAllowed:  false,
			wantErr:      false,
			setupCache: func(h *AuthzHelper) {
				h.cache.set("user-789:admin:system", false, 5*time.Minute)
			},
		},
		{
			name:         "server error",
			userID:       "user-error",
			permission:   "read",
			resource:     "data",
			responseCode: http.StatusInternalServerError,
			responseBody: `{"error": "internal error"}`,
			wantAllowed:  false,
			wantErr:      true,
		},
		{
			name:         "invalid json response",
			userID:       "user-bad",
			permission:   "read",
			resource:     "data",
			responseCode: http.StatusOK,
			responseBody: `{invalid json}`,
			wantAllowed:  false,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			logger := log.NewNoopLogger()
			helper := NewAuthzHelper(server.URL, 5*time.Minute, logger)

			if tt.setupCache != nil {
				tt.setupCache(helper)
			}

			got, err := helper.CheckPermission(context.Background(), tt.userID, tt.permission, tt.resource)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.wantAllowed {
				t.Errorf("CheckPermission() = %v, want %v", got, tt.wantAllowed)
			}
		})
	}
}

func TestPermissionCacheExpiration(t *testing.T) {
	cache := newPermissionCache()

	cache.set("key1", true, 50*time.Millisecond)

	if allowed, ok := cache.get("key1"); !ok || !allowed {
		t.Errorf("expected cached value to exist and be true")
	}

	time.Sleep(100 * time.Millisecond)

	if _, ok := cache.get("key1"); ok {
		t.Errorf("expected cached value to expire")
	}
}

func TestPermissionCacheConcurrency(t *testing.T) {
	cache := newPermissionCache()

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				cache.set("key", true, time.Minute)
				cache.get("key")
			}
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}
