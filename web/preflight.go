package web

import (
	"context"
	"net/http"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/log"
)

type PreflightChecker struct {
	authz *auth.AuthzHelper
	log   log.Logger
}

func NewPreflightChecker(authz *auth.AuthzHelper, log log.Logger) *PreflightChecker {
	return &PreflightChecker{
		authz: authz,
		log:   log,
	}
}

func (p *PreflightChecker) RequirePermission(permission, resource string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := r.Context().Value("user_id").(string)
			if !ok || userID == "" {
				p.log.Infof("Preflight check failed: no user_id in context")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			allowed, err := p.authz.CheckPermission(r.Context(), userID, permission, resource)
			if err != nil {
				p.log.Errorf("Authorization check error: %v", err)
				http.Error(w, "Authorization check failed", http.StatusInternalServerError)
				return
			}

			if !allowed {
				p.log.Infof("Permission denied for user %s: %s on %s", userID, permission, resource)
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "user_id", userID)
}

func GetUserID(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value("user_id").(string)
	return userID, ok
}
