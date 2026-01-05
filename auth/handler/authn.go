package handler

import (
	"encoding/json"
	"net/http"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/service"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type AuthNHandler struct {
	userStore auth.UserStore
	crypto    service.CryptoService
	tokenGen  service.TokenGenerator
	pwdGen    service.PasswordGenerator
	pinGen    service.PINGenerator
}

func NewAuthNHandler(
	userStore auth.UserStore,
	crypto service.CryptoService,
	tokenGen service.TokenGenerator,
	pwdGen service.PasswordGenerator,
	pinGen service.PINGenerator,
) *AuthNHandler {
	return &AuthNHandler{
		userStore: userStore,
		crypto:    crypto,
		tokenGen:  tokenGen,
		pwdGen:    pwdGen,
		pinGen:    pinGen,
	}
}

func (h *AuthNHandler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/signup", h.handleSignUp)
	r.Post("/auth/signin", h.handleSignIn)
	r.Post("/auth/signin-pin", h.handleSignInByPIN)
	r.Post("/auth/bootstrap", h.handleBootstrap)
	r.Post("/auth/generate-pin", h.handleGeneratePIN)

	r.Get("/users/{id}", h.handleGetUser)
	r.Get("/users/username/{username}", h.handleGetUserByUsername)
	r.Get("/users", h.handleListUsers)
	r.Put("/users/{id}", h.handleUpdateUser)
	r.Delete("/users/{id}", h.handleDeleteUser)
}

type SignUpRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
}

type SignUpResponse struct {
	User *auth.User `json:"user"`
}

func (h *AuthNHandler) handleSignUp(w http.ResponseWriter, r *http.Request) {
	var req SignUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, err := service.SignUp(
		r.Context(),
		h.userStore,
		h.crypto,
		req.Email,
		req.Password,
		req.Username,
		req.DisplayName,
	)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, SignUpResponse{User: user})
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInResponse struct {
	User  *auth.User `json:"user"`
	Token string     `json:"token"`
}

func (h *AuthNHandler) handleSignIn(w http.ResponseWriter, r *http.Request) {
	var req SignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, token, err := service.SignIn(
		r.Context(),
		h.userStore,
		h.crypto,
		h.tokenGen,
		req.Email,
		req.Password,
	)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, SignInResponse{User: user, Token: token})
}

type SignInByPINRequest struct {
	PIN string `json:"pin"`
}

type SignInByPINResponse struct {
	User *auth.User `json:"user"`
}

func (h *AuthNHandler) handleSignInByPIN(w http.ResponseWriter, r *http.Request) {
	var req SignInByPINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, err := service.SignInByPIN(r.Context(), h.userStore, h.crypto, req.PIN)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, SignInByPINResponse{User: user})
}

type BootstrapResponse struct {
	User     *auth.User `json:"user"`
	Password string     `json:"password,omitempty"`
}

func (h *AuthNHandler) handleBootstrap(w http.ResponseWriter, r *http.Request) {
	user, password, err := service.Bootstrap(r.Context(), h.userStore, h.crypto, h.pwdGen)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, BootstrapResponse{User: user, Password: password})
}

type GeneratePINRequest struct {
	UserID string `json:"user_id"`
}

type GeneratePINResponse struct {
	PIN string `json:"pin"`
}

func (h *AuthNHandler) handleGeneratePIN(w http.ResponseWriter, r *http.Request) {
	var req GeneratePINRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	user, err := service.GetUserByID(r.Context(), h.userStore, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	pin, err := service.GeneratePIN(r.Context(), h.userStore, h.crypto, h.pinGen, user)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, GeneratePINResponse{PIN: pin})
}

type UserResponse struct {
	User *auth.User `json:"data"`
}

func (h *AuthNHandler) handleGetUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	user, err := service.GetUserByID(r.Context(), h.userStore, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, UserResponse{User: user})
}

func (h *AuthNHandler) handleGetUserByUsername(w http.ResponseWriter, r *http.Request) {
	username := chi.URLParam(r, "username")

	user, err := service.GetUserByUsername(r.Context(), h.userStore, username)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, UserResponse{User: user})
}

type ListUsersResponse struct {
	Users []*auth.User `json:"data"`
}

func (h *AuthNHandler) handleListUsers(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")

	var users []*auth.User
	var err error

	if status != "" {
		users, err = service.ListUsersByStatus(r.Context(), h.userStore, auth.UserStatus(status))
	} else {
		users, err = service.ListUsers(r.Context(), h.userStore)
	}

	if err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, ListUsersResponse{Users: users})
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

func (h *AuthNHandler) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	user, err := service.GetUserByID(r.Context(), h.userStore, userID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	user.Name = req.Name
	if err := service.UpdateUser(r.Context(), h.userStore, user); err != nil {
		handleServiceError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, UserResponse{User: user})
}

func (h *AuthNHandler) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_USER_ID", "Invalid user ID format")
		return
	}

	if err := service.DeleteUser(r.Context(), h.userStore, userID); err != nil {
		handleServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
