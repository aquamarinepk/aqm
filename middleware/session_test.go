package middleware

import (
	"context"
	"crypto/ed25519"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/crypto"
)

// mockValidator is a test validator that returns predefined results.
type mockValidator struct {
	userID    string
	sessionID string
	err       error
}

func (m *mockValidator) ValidateToken(token string) (string, string, error) {
	if m.err != nil {
		return "", "", m.err
	}
	return m.userID, m.sessionID, nil
}

func TestSession(t *testing.T) {
	tests := []struct {
		name               string
		cookieValue        string
		cookiePresent      bool
		validator          SessionValidator
		wantRedirect       bool
		wantUserID         string
		wantSessionID      string
		wantCookieCleared  bool
	}{
		{
			name:          "valid session",
			cookieValue:   "valid-token",
			cookiePresent: true,
			validator: &mockValidator{
				userID:    "user-123",
				sessionID: "session-456",
				err:       nil,
			},
			wantRedirect:      false,
			wantUserID:        "user-123",
			wantSessionID:     "session-456",
			wantCookieCleared: false,
		},
		{
			name:          "missing cookie",
			cookiePresent: false,
			validator: &mockValidator{
				userID:    "user-123",
				sessionID: "session-456",
				err:       nil,
			},
			wantRedirect:      true,
			wantUserID:        "",
			wantSessionID:     "",
			wantCookieCleared: false,
		},
		{
			name:          "invalid token",
			cookieValue:   "invalid-token",
			cookiePresent: true,
			validator: &mockValidator{
				err: errors.New("invalid token"),
			},
			wantRedirect:      true,
			wantUserID:        "",
			wantSessionID:     "",
			wantCookieCleared: true,
		},
		{
			name:          "expired token",
			cookieValue:   "expired-token",
			cookiePresent: true,
			validator: &mockValidator{
				err: crypto.ErrTokenExpired,
			},
			wantRedirect:      true,
			wantUserID:        "",
			wantSessionID:     "",
			wantCookieCleared: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedUserID, capturedSessionID string
			var handlerCalled bool

			handler := Session(tt.validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				handlerCalled = true
				capturedUserID = GetUserID(r.Context())
				capturedSessionID = GetSessionID(r.Context())
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest(http.MethodGet, "/protected", nil)
			if tt.cookiePresent {
				req.AddCookie(&http.Cookie{
					Name:  SessionCookieName,
					Value: tt.cookieValue,
				})
			}

			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if tt.wantRedirect {
				if rec.Code != http.StatusSeeOther {
					t.Errorf("Status code = %d, want %d", rec.Code, http.StatusSeeOther)
				}
				location := rec.Header().Get("Location")
				if location != "/signin" {
					t.Errorf("Redirect location = %v, want /signin", location)
				}
				if handlerCalled {
					t.Error("Handler was called when redirect expected")
				}
			} else {
				if rec.Code != http.StatusOK {
					t.Errorf("Status code = %d, want %d", rec.Code, http.StatusOK)
				}
				if !handlerCalled {
					t.Error("Handler was not called")
				}
				if capturedUserID != tt.wantUserID {
					t.Errorf("User ID = %v, want %v", capturedUserID, tt.wantUserID)
				}
				if capturedSessionID != tt.wantSessionID {
					t.Errorf("Session ID = %v, want %v", capturedSessionID, tt.wantSessionID)
				}
			}

			if tt.wantCookieCleared {
				cookies := rec.Result().Cookies()
				found := false
				for _, cookie := range cookies {
					if cookie.Name == SessionCookieName {
						found = true
						if cookie.MaxAge != -1 {
							t.Errorf("Cookie MaxAge = %d, want -1 (cleared)", cookie.MaxAge)
						}
						if cookie.Value != "" {
							t.Errorf("Cookie Value = %v, want empty (cleared)", cookie.Value)
						}
					}
				}
				if !found {
					t.Error("Session cookie was not set for clearing")
				}
			}
		})
	}
}

func TestTokenValidator(t *testing.T) {
	publicKey, privateKey, _ := ed25519.GenerateKey(nil)

	validClaims := crypto.TokenClaims{
		Subject:   "user-123",
		SessionID: "session-456",
		Audience:  "pulap-lite",
		ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
	}

	validToken, _ := crypto.GenerateToken(validClaims, privateKey)

	expiredClaims := crypto.TokenClaims{
		Subject:   "user-123",
		SessionID: "session-456",
		Audience:  "pulap-lite",
		ExpiresAt: time.Now().Add(-1 * time.Hour).Unix(),
	}

	expiredToken, _ := crypto.GenerateToken(expiredClaims, privateKey)

	tests := []struct {
		name          string
		token         string
		wantUserID    string
		wantSessionID string
		wantErr       error
	}{
		{
			name:          "valid token",
			token:         validToken,
			wantUserID:    "user-123",
			wantSessionID: "session-456",
			wantErr:       nil,
		},
		{
			name:          "expired token",
			token:         expiredToken,
			wantUserID:    "",
			wantSessionID: "",
			wantErr:       crypto.ErrTokenExpired,
		},
		{
			name:          "invalid token",
			token:         "invalid.token.string",
			wantUserID:    "",
			wantSessionID: "",
			wantErr:       crypto.ErrInvalidToken,
		},
		{
			name:          "empty token",
			token:         "",
			wantUserID:    "",
			wantSessionID: "",
			wantErr:       crypto.ErrInvalidToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewTokenValidator(publicKey)

			userID, sessionID, err := validator.ValidateToken(tt.token)

			if err != tt.wantErr {
				t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr == nil {
				if userID != tt.wantUserID {
					t.Errorf("ValidateToken() userID = %v, want %v", userID, tt.wantUserID)
				}
				if sessionID != tt.wantSessionID {
					t.Errorf("ValidateToken() sessionID = %v, want %v", sessionID, tt.wantSessionID)
				}
			}
		})
	}
}

func TestGetUserID(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "valid user ID in context",
			ctx:  context.WithValue(context.Background(), UserIDKey, "user-123"),
			want: "user-123",
		},
		{
			name: "empty context",
			ctx:  context.Background(),
			want: "",
		},
		{
			name: "wrong type in context",
			ctx:  context.WithValue(context.Background(), UserIDKey, 12345),
			want: "",
		},
		{
			name: "nil context",
			ctx:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetUserID(tt.ctx)

			if got != tt.want {
				t.Errorf("GetUserID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetSessionID(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{
			name: "valid session ID in context",
			ctx:  context.WithValue(context.Background(), SessionIDKey, "session-456"),
			want: "session-456",
		},
		{
			name: "empty context",
			ctx:  context.Background(),
			want: "",
		},
		{
			name: "wrong type in context",
			ctx:  context.WithValue(context.Background(), SessionIDKey, 67890),
			want: "",
		},
		{
			name: "nil context",
			ctx:  nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetSessionID(tt.ctx)

			if got != tt.want {
				t.Errorf("GetSessionID() = %v, want %v", got, tt.want)
			}
		})
	}
}
