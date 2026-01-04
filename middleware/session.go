package middleware

import (
	"context"
	"crypto/ed25519"
	"net/http"

	"github.com/aquamarinepk/aqm/crypto"
)

const (
	SessionCookieName = "session_id"
	UserIDKey         = contextKey("user_id")
	SessionIDKey      = contextKey("session_id")
)

// SessionValidator validates session tokens and returns user ID on success.
type SessionValidator interface {
	ValidateToken(token string) (userID string, sessionID string, err error)
}

// TokenValidator implements SessionValidator using PASETO tokens.
type TokenValidator struct {
	publicKey ed25519.PublicKey
}

// NewTokenValidator creates a new token validator with the given public key.
func NewTokenValidator(publicKey ed25519.PublicKey) *TokenValidator {
	return &TokenValidator{
		publicKey: publicKey,
	}
}

// ValidateToken validates a PASETO token and extracts the user ID and session ID.
func (v *TokenValidator) ValidateToken(token string) (string, string, error) {
	claims, err := crypto.VerifyToken(token, v.publicKey)
	if err != nil {
		return "", "", err
	}
	return claims.Subject, claims.SessionID, nil
}

// Session validates session cookies and injects user context.
// If the session is invalid, it clears the cookie and redirects to /signin.
func Session(validator SessionValidator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie(SessionCookieName)
			if err != nil {
				http.Redirect(w, r, "/signin", http.StatusSeeOther)
				return
			}

			userID, sessionID, err := validator.ValidateToken(cookie.Value)
			if err != nil {
				clearSessionCookie(w)
				http.Redirect(w, r, "/signin", http.StatusSeeOther)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, UserIDKey, userID)
			ctx = context.WithValue(ctx, SessionIDKey, sessionID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// clearSessionCookie clears the session cookie by setting MaxAge to -1.
func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     SessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// GetUserID extracts the user ID from the context.
// Returns an empty string if no user ID is found.
func GetUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(UserIDKey).(string); ok {
		return id
	}
	return ""
}

// GetSessionID extracts the session ID from the context.
// Returns an empty string if no session ID is found.
func GetSessionID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(SessionIDKey).(string); ok {
		return id
	}
	return ""
}

// RoleChecker is an interface for checking user roles and permissions.
type RoleChecker interface {
	HasRole(ctx context.Context, userID string, roleName string) (bool, error)
	CheckPermission(ctx context.Context, userID string, permission string) (bool, error)
	CheckAnyPermission(ctx context.Context, userID string, permissions []string) (bool, error)
	CheckAllPermissions(ctx context.Context, userID string, permissions []string) (bool, error)
}

// RequireRole creates middleware that requires a specific role.
func RequireRole(checker RoleChecker, roleName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			hasRole, err := checker.HasRole(r.Context(), userID, roleName)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !hasRole {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequirePermission creates middleware that requires a specific permission.
func RequirePermission(checker RoleChecker, permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			hasPermission, err := checker.CheckPermission(r.Context(), userID, permission)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAnyPermission creates middleware that requires any of the specified permissions.
func RequireAnyPermission(checker RoleChecker, permissions []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			hasPermission, err := checker.CheckAnyPermission(r.Context(), userID, permissions)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !hasPermission {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// RequireAllPermissions creates middleware that requires all of the specified permissions.
func RequireAllPermissions(checker RoleChecker, permissions []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r.Context())
			if userID == "" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			hasPermissions, err := checker.CheckAllPermissions(r.Context(), userID, permissions)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			if !hasPermissions {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
