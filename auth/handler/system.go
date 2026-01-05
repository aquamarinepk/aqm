package handler

import (
	"context"
	"net/http"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/service"
	"github.com/go-chi/chi/v5"
)

// SystemHandler handles system-level operations like bootstrap and user lookup.
// These endpoints are designed for inter-service communication during system initialization.
type SystemHandler struct {
	userStore auth.UserStore
	crypto    service.CryptoService
	pwdGen    service.PasswordGenerator
}

// SystemBootstrapStatusResponse represents the current bootstrap status
type SystemBootstrapStatusResponse struct {
	NeedsBootstrap bool   `json:"needs_bootstrap"`
	SuperadminID   string `json:"superadmin_id,omitempty"`
}

// SystemBootstrapResponse represents the result of bootstrap operation
type SystemBootstrapResponse struct {
	SuperadminID string `json:"superadmin_id"`
	Email        string `json:"email"`
	Password     string `json:"password"`
}

// SystemUserIDResponse represents user lookup response
type SystemUserIDResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

// NewSystemHandler creates a new system handler with required dependencies.
func NewSystemHandler(userStore auth.UserStore, crypto service.CryptoService, pwdGen service.PasswordGenerator) *SystemHandler {
	return &SystemHandler{
		userStore: userStore,
		crypto:    crypto,
		pwdGen:    pwdGen,
	}
}

// RegisterRoutes registers system management routes
func (h *SystemHandler) RegisterRoutes(r chi.Router) {
	r.Get("/system/bootstrap-status", h.GetBootstrapStatus)
	r.Post("/system/bootstrap", h.Bootstrap)
	r.Get("/system/users/by-email/{email}", h.GetUserIDByEmail)
}

// GetBootstrapStatus checks if the system needs bootstrap.
// Returns needs_bootstrap=true if superadmin doesn't exist, false otherwise.
func (h *SystemHandler) GetBootstrapStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	superadmin, err := h.findSuperadmin(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BOOTSTRAP_CHECK_FAILED", "Failed to check bootstrap state")
		return
	}

	if superadmin == nil {
		writeJSON(w, http.StatusOK, SystemBootstrapStatusResponse{NeedsBootstrap: true})
		return
	}

	writeJSON(w, http.StatusOK, SystemBootstrapStatusResponse{
		NeedsBootstrap: false,
		SuperadminID:   superadmin.ID.String(),
	})
}

// Bootstrap creates the superadmin user if it doesn't exist.
// Returns superadmin details with password only if newly created.
// This endpoint is idempotent - multiple calls are safe.
func (h *SystemHandler) Bootstrap(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, password, err := service.Bootstrap(ctx, h.userStore, h.crypto, h.pwdGen)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "BOOTSTRAP_FAILED", "Failed to bootstrap superadmin")
		return
	}

	writeJSON(w, http.StatusOK, SystemBootstrapResponse{
		SuperadminID: user.ID.String(),
		Email:        service.SuperadminEmail,
		Password:     password,
	})
}

// GetUserIDByEmail retrieves user ID by email.
// This is a system endpoint used by other services (like authz) during bootstrap.
func (h *SystemHandler) GetUserIDByEmail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	email := chi.URLParam(r, "email")
	if email == "" {
		writeError(w, http.StatusBadRequest, "MISSING_EMAIL", "Email parameter is required")
		return
	}

	normalizedEmail := auth.NormalizeEmail(email)
	emailLookup := h.crypto.ComputeLookupHash(normalizedEmail)

	user, err := h.userStore.GetByEmailLookup(ctx, emailLookup)
	if err != nil {
		if err == auth.ErrUserNotFound {
			writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "User not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "LOOKUP_FAILED", "Failed to lookup user")
		return
	}

	writeJSON(w, http.StatusOK, SystemUserIDResponse{
		UserID: user.ID.String(),
		Email:  email,
	})
}

// findSuperadmin looks up the superadmin user by email
func (h *SystemHandler) findSuperadmin(ctx context.Context) (*auth.User, error) {
	emailLookup := h.crypto.ComputeLookupHash(service.SuperadminEmail)

	user, err := h.userStore.GetByEmailLookup(ctx, emailLookup)
	if err != nil {
		if err == auth.ErrUserNotFound {
			return nil, nil
		}
		return nil, err
	}

	return user, nil
}
