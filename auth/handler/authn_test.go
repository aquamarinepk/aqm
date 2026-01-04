package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aquamarinepk/aqm/auth/fake"
	"github.com/go-chi/chi/v5"
)

func setupAuthNHandler() *AuthNHandler {
	userStore := fake.NewUserStore()
	crypto := fake.NewCryptoService()
	tokenGen := fake.NewTokenGenerator()
	pwdGen := fake.NewPasswordGenerator()
	pinGen := fake.NewPINGenerator()

	return NewAuthNHandler(userStore, crypto, tokenGen, pwdGen, pinGen)
}

func TestHandleSignUp(t *testing.T) {
	handler := setupAuthNHandler()

	tests := []struct {
		name       string
		body       SignUpRequest
		wantStatus int
		wantCode   string
	}{
		{
			name: "valid signup",
			body: SignUpRequest{
				Email:       "test@example.com",
				Password:    "Password123!",
				Username:    "testuser",
				DisplayName: "Test User",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid email",
			body: SignUpRequest{
				Email:       "invalid",
				Password:    "Password123!",
				Username:    "testuser2",
				DisplayName: "Test User",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_EMAIL",
		},
		{
			name: "weak password",
			body: SignUpRequest{
				Email:       "test2@example.com",
				Password:    "weak",
				Username:    "testuser3",
				DisplayName: "Test User",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_PASSWORD",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.handleSignUp(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleSignUp() status = %v, want %v, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleSignUp() error code = %v, want %v, message: %s", errResp.Code, tt.wantCode, errResp.Message)
				}
			}
		})
	}
}

func TestHandleSignUpErrorPaths(t *testing.T) {
	handler := setupAuthNHandler()

	// Create a user first
	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "duplicate@example.com",
		Password:    "Password123!",
		Username:    "dupuser",
		DisplayName: "Duplicate User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	handler.handleSignUp(signupW, signupReq)

	tests := []struct {
		name       string
		body       SignUpRequest
		wantStatus int
		wantCode   string
	}{
		{
			name: "duplicate email",
			body: SignUpRequest{
				Email:       "duplicate@example.com",
				Password:    "Password123!",
				Username:    "anotheruser",
				DisplayName: "Another User",
			},
			wantStatus: http.StatusConflict,
			wantCode:   "USER_ALREADY_EXISTS",
		},
		{
			name: "duplicate username",
			body: SignUpRequest{
				Email:       "newemail@example.com",
				Password:    "Password123!",
				Username:    "dupuser",
				DisplayName: "New User",
			},
			wantStatus: http.StatusConflict,
			wantCode:   "USERNAME_EXISTS",
		},
		{
			name: "invalid username",
			body: SignUpRequest{
				Email:       "test3@example.com",
				Password:    "Password123!",
				Username:    "a",
				DisplayName: "Test",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USERNAME",
		},
		{
			name: "invalid display name",
			body: SignUpRequest{
				Email:       "test4@example.com",
				Password:    "Password123!",
				Username:    "testuser4",
				DisplayName: "",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_DISPLAY_NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.handleSignUp(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleSignUp() status = %v, want %v, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			var errResp ErrorResponse
			json.NewDecoder(w.Body).Decode(&errResp)
			if errResp.Code != tt.wantCode {
				t.Errorf("handleSignUp() error code = %v, want %v", errResp.Code, tt.wantCode)
			}
		})
	}
}

func TestHandleSignIn(t *testing.T) {
	handler := setupAuthNHandler()

	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "signin@example.com",
		Password:    "Password123!",
		Username:    "signinuser",
		DisplayName: "Signin User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	handler.handleSignUp(signupW, signupReq)

	tests := []struct {
		name       string
		body       SignInRequest
		wantStatus int
		wantCode   string
	}{
		{
			name: "correct credentials",
			body: SignInRequest{
				Email:    "signin@example.com",
				Password: "Password123!",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "wrong password",
			body: SignInRequest{
				Email:    "signin@example.com",
				Password: "WrongPassword123!",
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_CREDENTIALS",
		},
		{
			name: "non-existing user",
			body: SignInRequest{
				Email:    "notfound@example.com",
				Password: "Password123!",
			},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_CREDENTIALS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/signin", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.handleSignIn(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleSignIn() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleSignIn() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}

			if tt.wantStatus == http.StatusOK {
				var resp SignInResponse
				json.NewDecoder(w.Body).Decode(&resp)
				if resp.Token == "" {
					t.Error("handleSignIn() returned empty token")
				}
			}
		})
	}
}

func TestHandleBootstrap(t *testing.T) {
	handler := setupAuthNHandler()

	req := httptest.NewRequest(http.MethodPost, "/auth/bootstrap", nil)
	w := httptest.NewRecorder()

	handler.handleBootstrap(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleBootstrap() status = %v, want %v", w.Code, http.StatusOK)
	}

	var resp BootstrapResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.User == nil {
		t.Error("handleBootstrap() returned nil user")
	}

	if resp.Password == "" {
		t.Error("handleBootstrap() returned empty password")
	}

	if resp.User.Username != "superadmin" {
		t.Errorf("handleBootstrap() username = %v, want superadmin", resp.User.Username)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/auth/bootstrap", nil)
	w2 := httptest.NewRecorder()
	handler.handleBootstrap(w2, req2)

	var resp2 BootstrapResponse
	json.NewDecoder(w2.Body).Decode(&resp2)

	if resp2.Password != "" {
		t.Error("handleBootstrap() second call should return empty password")
	}
}

func TestHandleGetUser(t *testing.T) {
	handler := setupAuthNHandler()

	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "getuser@example.com",
		Password:    "Password123!",
		Username:    "getusertest",
		DisplayName: "Get User Test",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	handler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "existing user",
			id:         userID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid user ID",
			id:         "invalid",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
		{
			name:       "non-existing user",
			id:         "00000000-0000-0000-0000-000000000000",
			wantStatus: http.StatusNotFound,
			wantCode:   "USER_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.id, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.handleGetUser(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleGetUser() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleGetUser() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleListUsers(t *testing.T) {
	handler := setupAuthNHandler()

	for i := 1; i <= 3; i++ {
		signupBody, _ := json.Marshal(SignUpRequest{
			Email:       "list" + string(rune('0'+i)) + "@example.com",
			Password:    "Password123!",
			Username:    "listuser" + string(rune('0'+i)),
			DisplayName: "List User",
		})
		signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
		signupW := httptest.NewRecorder()
		handler.handleSignUp(signupW, signupReq)

		if signupW.Code != http.StatusCreated {
			t.Fatalf("Failed to create test user %d: %v", i, signupW.Body.String())
		}
	}

	tests := []struct {
		name       string
		status     string
		wantStatus int
		wantMin    int
	}{
		{
			name:       "list all users",
			status:     "",
			wantStatus: http.StatusOK,
			wantMin:    3,
		},
		{
			name:       "list active users",
			status:     "active",
			wantStatus: http.StatusOK,
			wantMin:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/users"
			if tt.status != "" {
				url += "?status=" + tt.status
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handler.handleListUsers(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleListUsers() status = %v, want %v", w.Code, tt.wantStatus)
			}

			var resp ListUsersResponse
			json.NewDecoder(w.Body).Decode(&resp)

			if len(resp.Users) < tt.wantMin {
				t.Errorf("handleListUsers() count = %v, want at least %v", len(resp.Users), tt.wantMin)
			}
		})
	}
}

func TestHandleInvalidJSON(t *testing.T) {
	handler := setupAuthNHandler()

	tests := []struct {
		name       string
		endpoint   string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "invalid signup JSON",
			endpoint:   "/auth/signup",
			body:       "{invalid json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "invalid signin JSON",
			endpoint:   "/auth/signin",
			body:       "{invalid json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "invalid signin-pin JSON",
			endpoint:   "/auth/signin-pin",
			body:       "{invalid json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "invalid generate-pin JSON",
			endpoint:   "/auth/generate-pin",
			body:       "{invalid json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "invalid update user JSON",
			endpoint:   "/users/update",
			body:       "{invalid json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.endpoint, bytes.NewReader([]byte(tt.body)))
			w := httptest.NewRecorder()

			switch tt.endpoint {
			case "/auth/signup":
				handler.handleSignUp(w, req)
			case "/auth/signin":
				handler.handleSignIn(w, req)
			case "/auth/signin-pin":
				handler.handleSignInByPIN(w, req)
			case "/auth/generate-pin":
				handler.handleGeneratePIN(w, req)
			case "/users/update":
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", "00000000-0000-0000-0000-000000000000")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				handler.handleUpdateUser(w, req)
			}

			if w.Code != tt.wantStatus {
				t.Errorf("handler status = %v, want %v", w.Code, tt.wantStatus)
			}

			var errResp ErrorResponse
			json.NewDecoder(w.Body).Decode(&errResp)
			if errResp.Code != tt.wantCode {
				t.Errorf("handler error code = %v, want %v", errResp.Code, tt.wantCode)
			}
		})
	}
}

func TestHandleSignInByPIN(t *testing.T) {
	handler := setupAuthNHandler()

	// Create user and generate PIN
	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "pinuser@example.com",
		Password:    "Password123!",
		Username:    "pinuser",
		DisplayName: "PIN User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	handler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)

	genBody, _ := json.Marshal(GeneratePINRequest{UserID: signupResp.User.ID.String()})
	genReq := httptest.NewRequest(http.MethodPost, "/auth/generate-pin", bytes.NewReader(genBody))
	genW := httptest.NewRecorder()
	handler.handleGeneratePIN(genW, genReq)

	var genResp GeneratePINResponse
	json.NewDecoder(genW.Body).Decode(&genResp)

	tests := []struct {
		name       string
		body       SignInByPINRequest
		wantStatus int
		wantCode   string
	}{
		{
			name:       "valid PIN",
			body:       SignInByPINRequest{PIN: genResp.PIN},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid PIN",
			body:       SignInByPINRequest{PIN: "999999"},
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_CREDENTIALS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/signin-pin", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.handleSignInByPIN(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleSignInByPIN() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleSignInByPIN() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleGeneratePIN(t *testing.T) {
	handler := setupAuthNHandler()

	// Create user
	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "genpin@example.com",
		Password:    "Password123!",
		Username:    "genpinuser",
		DisplayName: "Gen PIN User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	handler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)

	tests := []struct {
		name       string
		body       GeneratePINRequest
		wantStatus int
		wantCode   string
	}{
		{
			name:       "valid user",
			body:       GeneratePINRequest{UserID: signupResp.User.ID.String()},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid user ID",
			body:       GeneratePINRequest{UserID: "invalid"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
		{
			name:       "non-existing user",
			body:       GeneratePINRequest{UserID: "00000000-0000-0000-0000-000000000000"},
			wantStatus: http.StatusNotFound,
			wantCode:   "USER_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/auth/generate-pin", bytes.NewReader(body))
			w := httptest.NewRecorder()

			handler.handleGeneratePIN(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleGeneratePIN() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleGeneratePIN() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleGetUserByUsername(t *testing.T) {
	handler := setupAuthNHandler()

	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "username@example.com",
		Password:    "Password123!",
		Username:    "uniqueuser",
		DisplayName: "Unique User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	handler.handleSignUp(signupW, signupReq)

	tests := []struct {
		name       string
		username   string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "existing username",
			username:   "uniqueuser",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existing username",
			username:   "nonexistent",
			wantStatus: http.StatusNotFound,
			wantCode:   "USER_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/username/"+tt.username, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("username", tt.username)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.handleGetUserByUsername(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleGetUserByUsername() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleGetUserByUsername() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleUpdateUser(t *testing.T) {
	handler := setupAuthNHandler()

	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "update@example.com",
		Password:    "Password123!",
		Username:    "updateuser",
		DisplayName: "Update User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	handler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	tests := []struct {
		name       string
		id         string
		body       UpdateUserRequest
		wantStatus int
		wantCode   string
	}{
		{
			name:       "valid update",
			id:         userID.String(),
			body:       UpdateUserRequest{Name: "Updated Name"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid user ID",
			id:         "invalid",
			body:       UpdateUserRequest{Name: "Test"},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
		{
			name:       "non-existing user",
			id:         "00000000-0000-0000-0000-000000000000",
			body:       UpdateUserRequest{Name: "Test"},
			wantStatus: http.StatusNotFound,
			wantCode:   "USER_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/users/"+tt.id, bytes.NewReader(body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.handleUpdateUser(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleUpdateUser() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleUpdateUser() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleDeleteUser(t *testing.T) {
	handler := setupAuthNHandler()

	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "delete@example.com",
		Password:    "Password123!",
		Username:    "deleteuser",
		DisplayName: "Delete User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	handler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "valid delete",
			id:         userID.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid user ID",
			id:         "invalid",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
		{
			name:       "non-existing user",
			id:         "00000000-0000-0000-0000-000000000000",
			wantStatus: http.StatusNotFound,
			wantCode:   "USER_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.id, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.handleDeleteUser(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleDeleteUser() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleDeleteUser() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}
