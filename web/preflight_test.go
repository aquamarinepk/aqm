package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/log"
)

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name           string
		permission     string
		resource       string
		setupContext   func(*http.Request) *http.Request
		authzResponse  bool
		wantStatusCode int
		wantCalled     bool
	}{
		{
			name:       "permission allowed",
			permission: "read",
			resource:   "documents",
			setupContext: func(r *http.Request) *http.Request {
				ctx := WithUserID(r.Context(), "user-123")
				return r.WithContext(ctx)
			},
			authzResponse:  true,
			wantStatusCode: http.StatusOK,
			wantCalled:     true,
		},
		{
			name:       "permission denied",
			permission: "delete",
			resource:   "documents",
			setupContext: func(r *http.Request) *http.Request {
				ctx := WithUserID(r.Context(), "user-123")
				return r.WithContext(ctx)
			},
			authzResponse:  false,
			wantStatusCode: http.StatusForbidden,
			wantCalled:     false,
		},
		{
			name:       "no user_id in context",
			permission: "read",
			resource:   "documents",
			setupContext: func(r *http.Request) *http.Request {
				return r
			},
			wantStatusCode: http.StatusUnauthorized,
			wantCalled:     false,
		},
		{
			name:       "empty user_id in context",
			permission: "read",
			resource:   "documents",
			setupContext: func(r *http.Request) *http.Request {
				ctx := WithUserID(r.Context(), "")
				return r.WithContext(ctx)
			},
			wantStatusCode: http.StatusUnauthorized,
			wantCalled:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.authzResponse {
					w.Write([]byte(`{"allowed": true}`))
				} else {
					w.Write([]byte(`{"allowed": false}`))
				}
			}))
			defer server.Close()

			logger := log.NewNoopLogger()
			authzHelper := auth.NewAuthzHelper(server.URL, 5*time.Minute, logger)
			checker := NewPreflightChecker(authzHelper, logger)

			handlerCalled := false
			handler := checker.RequirePermission(tt.permission, tt.resource)(
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					handlerCalled = true
					w.WriteHeader(http.StatusOK)
				}),
			)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req = tt.setupContext(req)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("status code = %v, want %v", w.Code, tt.wantStatusCode)
			}

			if handlerCalled != tt.wantCalled {
				t.Errorf("handler called = %v, want %v", handlerCalled, tt.wantCalled)
			}
		})
	}
}

func TestRequirePermissionAuthzError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	logger := log.NewNoopLogger()
	authzHelper := auth.NewAuthzHelper(server.URL, 5*time.Minute, logger)
	checker := NewPreflightChecker(authzHelper, logger)

	handlerCalled := false
	handler := checker.RequirePermission("read", "documents")(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			handlerCalled = true
			w.WriteHeader(http.StatusOK)
		}),
	)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	ctx := WithUserID(req.Context(), "user-123")
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status code = %v, want %v", w.Code, http.StatusInternalServerError)
	}

	if handlerCalled {
		t.Errorf("handler called = true, want false")
	}
}

func TestWithUserIDAndGetUserID(t *testing.T) {
	tests := []struct {
		name       string
		userID     string
		wantUserID string
		wantOk     bool
	}{
		{
			name:       "set and get user_id",
			userID:     "user-123",
			wantUserID: "user-123",
			wantOk:     true,
		},
		{
			name:       "empty user_id",
			userID:     "",
			wantUserID: "",
			wantOk:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			ctx = WithUserID(ctx, tt.userID)

			gotUserID, gotOk := GetUserID(ctx)

			if gotUserID != tt.wantUserID {
				t.Errorf("GetUserID() userID = %v, want %v", gotUserID, tt.wantUserID)
			}

			if gotOk != tt.wantOk {
				t.Errorf("GetUserID() ok = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestGetUserIDNoValue(t *testing.T) {
	ctx := context.Background()
	userID, ok := GetUserID(ctx)

	if ok {
		t.Errorf("GetUserID() ok = true, want false")
	}

	if userID != "" {
		t.Errorf("GetUserID() userID = %v, want empty string", userID)
	}
}
