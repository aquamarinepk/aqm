package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/aquamarinepk/aqm/auth"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, ErrorResponse{Code: code, Message: message})
}

func handleServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
	case errors.Is(err, auth.ErrUserAlreadyExists):
		writeError(w, http.StatusConflict, "USER_ALREADY_EXISTS", err.Error())
	case errors.Is(err, auth.ErrUsernameExists):
		writeError(w, http.StatusConflict, "USERNAME_EXISTS", err.Error())
	case errors.Is(err, auth.ErrInvalidEmail):
		writeError(w, http.StatusBadRequest, "INVALID_EMAIL", err.Error())
	case errors.Is(err, auth.ErrInvalidPassword):
		writeError(w, http.StatusBadRequest, "INVALID_PASSWORD", err.Error())
	case errors.Is(err, auth.ErrInvalidUsername):
		writeError(w, http.StatusBadRequest, "INVALID_USERNAME", err.Error())
	case errors.Is(err, auth.ErrInvalidDisplayName):
		writeError(w, http.StatusBadRequest, "INVALID_DISPLAY_NAME", err.Error())
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", err.Error())
	case errors.Is(err, auth.ErrInactiveAccount):
		writeError(w, http.StatusForbidden, "INACTIVE_ACCOUNT", err.Error())
	case errors.Is(err, auth.ErrRoleNotFound):
		writeError(w, http.StatusNotFound, "ROLE_NOT_FOUND", err.Error())
	case errors.Is(err, auth.ErrRoleAlreadyExists):
		writeError(w, http.StatusConflict, "ROLE_ALREADY_EXISTS", err.Error())
	case errors.Is(err, auth.ErrInvalidRoleName):
		writeError(w, http.StatusBadRequest, "INVALID_ROLE_NAME", err.Error())
	case errors.Is(err, auth.ErrGrantNotFound):
		writeError(w, http.StatusNotFound, "GRANT_NOT_FOUND", err.Error())
	case errors.Is(err, auth.ErrGrantAlreadyExists):
		writeError(w, http.StatusConflict, "GRANT_ALREADY_EXISTS", err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")
	}
}
