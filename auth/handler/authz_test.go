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

func setupAuthZHandler() *AuthZHandler {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)
	return NewAuthZHandler(roleStore, grantStore)
}

func TestHandleCreateRole(t *testing.T) {
	handler := setupAuthZHandler()

	tests := []struct {
		name       string
		body       CreateRoleRequest
		wantStatus int
		wantCode   string
	}{
		{
			name: "valid role",
			body: CreateRoleRequest{
				Name:        "editor",
				Description: "Can edit content",
				Permissions: []string{"content.read", "content.write"},
				CreatedBy:   "admin",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid role name",
			body: CreateRoleRequest{
				Name:        "a",
				Description: "Too short",
				Permissions: []string{"test"},
				CreatedBy:   "admin",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ROLE_NAME",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.handleCreateRole(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleCreateRole() status = %v, want %v, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleCreateRole() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleGetRole(t *testing.T) {
	handler := setupAuthZHandler()

	// Create a role first
	createBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "viewer",
		Description: "Can view content",
		Permissions: []string{"content.read"},
		CreatedBy:   "admin",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(createBody))
	createW := httptest.NewRecorder()
	handler.handleCreateRole(createW, createReq)

	var createResp RoleResponse
	json.NewDecoder(createW.Body).Decode(&createResp)
	roleID := createResp.Role.ID

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "existing role",
			id:         roleID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid role ID",
			id:         "invalid",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ROLE_ID",
		},
		{
			name:       "non-existing role",
			id:         "00000000-0000-0000-0000-000000000000",
			wantStatus: http.StatusNotFound,
			wantCode:   "ROLE_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/roles/"+tt.id, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.handleGetRole(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleGetRole() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleGetRole() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleListRoles(t *testing.T) {
	handler := setupAuthZHandler()

	// Create test roles
	for i := 1; i <= 3; i++ {
		createBody, _ := json.Marshal(CreateRoleRequest{
			Name:        "role" + string(rune('0'+i)),
			Description: "Test role",
			Permissions: []string{"test.read"},
			CreatedBy:   "admin",
		})
		createReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(createBody))
		createW := httptest.NewRecorder()
		handler.handleCreateRole(createW, createReq)
	}

	tests := []struct {
		name       string
		status     string
		wantStatus int
		wantMin    int
	}{
		{
			name:       "list all roles",
			status:     "",
			wantStatus: http.StatusOK,
			wantMin:    3,
		},
		{
			name:       "list active roles",
			status:     "active",
			wantStatus: http.StatusOK,
			wantMin:    3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := "/roles"
			if tt.status != "" {
				url += "?status=" + tt.status
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handler.handleListRoles(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleListRoles() status = %v, want %v", w.Code, tt.wantStatus)
			}

			var resp ListRolesResponse
			json.NewDecoder(w.Body).Decode(&resp)

			if len(resp.Roles) < tt.wantMin {
				t.Errorf("handleListRoles() count = %v, want at least %v", len(resp.Roles), tt.wantMin)
			}
		})
	}
}

func TestHandleAssignRole(t *testing.T) {
	authNHandler := setupAuthNHandler()
	authZHandler := setupAuthZHandler()

	// Create a user
	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "test@example.com",
		Password:    "Password123!",
		Username:    "testuser",
		DisplayName: "Test User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	authNHandler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	// Create a role
	roleBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "tester",
		Description: "Testing role",
		Permissions: []string{"test.run"},
		CreatedBy:   "admin",
	})
	roleReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(roleBody))
	roleW := httptest.NewRecorder()
	authZHandler.handleCreateRole(roleW, roleReq)

	var roleResp RoleResponse
	json.NewDecoder(roleW.Body).Decode(&roleResp)
	roleID := roleResp.Role.ID

	tests := []struct {
		name       string
		body       AssignRoleRequest
		wantStatus int
		wantCode   string
	}{
		{
			name: "valid assignment",
			body: AssignRoleRequest{
				UserID:     userID.String(),
				RoleID:     roleID.String(),
				AssignedBy: "admin",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "invalid user ID",
			body: AssignRoleRequest{
				UserID:     "invalid",
				RoleID:     roleID.String(),
				AssignedBy: "admin",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
		{
			name: "invalid role ID",
			body: AssignRoleRequest{
				UserID:     userID.String(),
				RoleID:     "invalid",
				AssignedBy: "admin",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ROLE_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPost, "/grants", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			authZHandler.handleAssignRole(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleAssignRole() status = %v, want %v, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleAssignRole() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleCheckPermission(t *testing.T) {
	authNHandler := setupAuthNHandler()
	authZHandler := setupAuthZHandler()

	// Create a user
	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "permtest@example.com",
		Password:    "Password123!",
		Username:    "permuser",
		DisplayName: "Permission User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	authNHandler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	// Create a role with permission
	roleBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "writer",
		Description: "Can write",
		Permissions: []string{"content.write"},
		CreatedBy:   "admin",
	})
	roleReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(roleBody))
	roleW := httptest.NewRecorder()
	authZHandler.handleCreateRole(roleW, roleReq)

	var roleResp RoleResponse
	json.NewDecoder(roleW.Body).Decode(&roleResp)
	roleID := roleResp.Role.ID

	// Assign role to user
	assignBody, _ := json.Marshal(AssignRoleRequest{
		UserID:     userID.String(),
		RoleID:     roleID.String(),
		AssignedBy: "admin",
	})
	assignReq := httptest.NewRequest(http.MethodPost, "/grants", bytes.NewReader(assignBody))
	assignW := httptest.NewRecorder()
	authZHandler.handleAssignRole(assignW, assignReq)

	tests := []struct {
		name           string
		userID         string
		permission     string
		wantStatus     int
		wantPermission bool
		wantCode       string
	}{
		{
			name:           "has permission",
			userID:         userID.String(),
			permission:     "content.write",
			wantStatus:     http.StatusOK,
			wantPermission: true,
		},
		{
			name:           "no permission",
			userID:         userID.String(),
			permission:     "content.delete",
			wantStatus:     http.StatusOK,
			wantPermission: false,
		},
		{
			name:       "invalid user ID",
			userID:     "invalid",
			permission: "test",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID+"/permissions/"+tt.permission, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("user_id", tt.userID)
			rctx.URLParams.Add("permission", tt.permission)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			authZHandler.handleCheckPermission(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleCheckPermission() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleCheckPermission() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			} else {
				var resp PermissionCheckResponse
				json.NewDecoder(w.Body).Decode(&resp)
				if resp.HasPermission != tt.wantPermission {
					t.Errorf("handleCheckPermission() hasPermission = %v, want %v", resp.HasPermission, tt.wantPermission)
				}
			}
		})
	}
}

func TestHandleInvalidJSONAuthZ(t *testing.T) {
	handler := setupAuthZHandler()

	tests := []struct {
		name       string
		endpoint   string
		body       string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "invalid create role JSON",
			endpoint:   "/roles",
			body:       "{invalid json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "invalid assign role JSON",
			endpoint:   "/grants",
			body:       "{invalid json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "invalid revoke role JSON",
			endpoint:   "/grants-revoke",
			body:       "{invalid json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "invalid update role JSON",
			endpoint:   "/roles-update",
			body:       "{invalid json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "invalid check any permission JSON",
			endpoint:   "/check-any",
			body:       "{invalid json",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_REQUEST",
		},
		{
			name:       "invalid check all permissions JSON",
			endpoint:   "/check-all",
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
			case "/roles":
				handler.handleCreateRole(w, req)
			case "/grants":
				handler.handleAssignRole(w, req)
			case "/grants-revoke":
				handler.handleRevokeRole(w, req)
			case "/roles-update":
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("id", "00000000-0000-0000-0000-000000000000")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				handler.handleUpdateRole(w, req)
			case "/check-any":
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("user_id", "00000000-0000-0000-0000-000000000000")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				handler.handleCheckAnyPermission(w, req)
			case "/check-all":
				rctx := chi.NewRouteContext()
				rctx.URLParams.Add("user_id", "00000000-0000-0000-0000-000000000000")
				req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
				handler.handleCheckAllPermissions(w, req)
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

func TestHandleAuthZErrorPaths(t *testing.T) {
	authNHandler := setupAuthNHandler()
	authZHandler := setupAuthZHandler()

	// Create role and user
	roleBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "test-role",
		Description: "Test role",
		Permissions: []string{"test"},
		CreatedBy:   "admin",
	})
	roleReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(roleBody))
	roleW := httptest.NewRecorder()
	authZHandler.handleCreateRole(roleW, roleReq)

	var roleResp RoleResponse
	json.NewDecoder(roleW.Body).Decode(&roleResp)
	roleID := roleResp.Role.ID

	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "error@example.com",
		Password:    "Password123!",
		Username:    "erroruser",
		DisplayName: "Error User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	authNHandler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	// Assign role
	assignBody, _ := json.Marshal(AssignRoleRequest{
		UserID:     userID.String(),
		RoleID:     roleID.String(),
		AssignedBy: "admin",
	})
	assignReq := httptest.NewRequest(http.MethodPost, "/grants", bytes.NewReader(assignBody))
	assignW := httptest.NewRecorder()
	authZHandler.handleAssignRole(assignW, assignReq)

	// Test duplicate role assignment
	t.Run("duplicate role assignment", func(t *testing.T) {
		dupAssignBody, _ := json.Marshal(AssignRoleRequest{
			UserID:     userID.String(),
			RoleID:     roleID.String(),
			AssignedBy: "admin",
		})
		dupAssignReq := httptest.NewRequest(http.MethodPost, "/grants", bytes.NewReader(dupAssignBody))
		dupAssignW := httptest.NewRecorder()
		authZHandler.handleAssignRole(dupAssignW, dupAssignReq)

		if dupAssignW.Code != http.StatusConflict {
			t.Errorf("duplicate assign status = %v, want %v", dupAssignW.Code, http.StatusConflict)
		}

		var errResp ErrorResponse
		json.NewDecoder(dupAssignW.Body).Decode(&errResp)
		if errResp.Code != "GRANT_ALREADY_EXISTS" {
			t.Errorf("duplicate assign error code = %v, want GRANT_ALREADY_EXISTS", errResp.Code)
		}
	})

	// Test duplicate role creation
	t.Run("duplicate role creation", func(t *testing.T) {
		dupRoleBody, _ := json.Marshal(CreateRoleRequest{
			Name:        "test-role",
			Description: "Duplicate",
			Permissions: []string{"test"},
			CreatedBy:   "admin",
		})
		dupRoleReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(dupRoleBody))
		dupRoleW := httptest.NewRecorder()
		authZHandler.handleCreateRole(dupRoleW, dupRoleReq)

		if dupRoleW.Code != http.StatusConflict {
			t.Errorf("duplicate role status = %v, want %v", dupRoleW.Code, http.StatusConflict)
		}

		var errResp ErrorResponse
		json.NewDecoder(dupRoleW.Body).Decode(&errResp)
		if errResp.Code != "ROLE_ALREADY_EXISTS" {
			t.Errorf("duplicate role error code = %v, want ROLE_ALREADY_EXISTS", errResp.Code)
		}
	})

	// Test revoke non-existing grant
	t.Run("revoke non-existing grant", func(t *testing.T) {
		// Create another user without the role
		signupBody2, _ := json.Marshal(SignUpRequest{
			Email:       "nogrant@example.com",
			Password:    "Password123!",
			Username:    "nograntuser",
			DisplayName: "No Grant User",
		})
		signupReq2 := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody2))
		signupW2 := httptest.NewRecorder()
		authNHandler.handleSignUp(signupW2, signupReq2)

		var signupResp2 SignUpResponse
		json.NewDecoder(signupW2.Body).Decode(&signupResp2)
		userID2 := signupResp2.User.ID

		revokeBody, _ := json.Marshal(RevokeRoleRequest{
			UserID: userID2.String(),
			RoleID: roleID.String(),
		})
		revokeReq := httptest.NewRequest(http.MethodDelete, "/grants", bytes.NewReader(revokeBody))
		revokeW := httptest.NewRecorder()
		authZHandler.handleRevokeRole(revokeW, revokeReq)

		if revokeW.Code != http.StatusNotFound {
			t.Errorf("revoke non-existing grant status = %v, want %v", revokeW.Code, http.StatusNotFound)
		}

		var errResp ErrorResponse
		json.NewDecoder(revokeW.Body).Decode(&errResp)
		if errResp.Code != "GRANT_NOT_FOUND" {
			t.Errorf("revoke non-existing grant error code = %v, want GRANT_NOT_FOUND", errResp.Code)
		}
	})
}

func TestHandleGetRoleByName(t *testing.T) {
	handler := setupAuthZHandler()

	// Create a role
	createBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "admin",
		Description: "Admin role",
		Permissions: []string{"admin.all"},
		CreatedBy:   "system",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(createBody))
	createW := httptest.NewRecorder()
	handler.handleCreateRole(createW, createReq)

	tests := []struct {
		name       string
		roleName   string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "existing role",
			roleName:   "admin",
			wantStatus: http.StatusOK,
		},
		{
			name:       "non-existing role",
			roleName:   "nonexistent",
			wantStatus: http.StatusNotFound,
			wantCode:   "ROLE_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/roles/name/"+tt.roleName, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("name", tt.roleName)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.handleGetRoleByName(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleGetRoleByName() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleGetRoleByName() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleUpdateRole(t *testing.T) {
	handler := setupAuthZHandler()

	// Create a role
	createBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "updatable",
		Description: "Original description",
		Permissions: []string{"read"},
		CreatedBy:   "admin",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(createBody))
	createW := httptest.NewRecorder()
	handler.handleCreateRole(createW, createReq)

	var createResp RoleResponse
	json.NewDecoder(createW.Body).Decode(&createResp)
	roleID := createResp.Role.ID

	tests := []struct {
		name       string
		id         string
		body       UpdateRoleRequest
		wantStatus int
		wantCode   string
	}{
		{
			name: "valid update",
			id:   roleID.String(),
			body: UpdateRoleRequest{
				Description: "Updated description",
				Permissions: []string{"read", "write"},
				UpdatedBy:   "admin",
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "invalid role ID",
			id:   "invalid",
			body: UpdateRoleRequest{
				Description: "Test",
				Permissions: []string{"test"},
				UpdatedBy:   "admin",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ROLE_ID",
		},
		{
			name: "non-existing role",
			id:   "00000000-0000-0000-0000-000000000000",
			body: UpdateRoleRequest{
				Description: "Test",
				Permissions: []string{"test"},
				UpdatedBy:   "admin",
			},
			wantStatus: http.StatusNotFound,
			wantCode:   "ROLE_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/roles/"+tt.id, bytes.NewReader(body))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.handleUpdateRole(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleUpdateRole() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleUpdateRole() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleDeleteRole(t *testing.T) {
	handler := setupAuthZHandler()

	// Create a role
	createBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "deletable",
		Description: "Will be deleted",
		Permissions: []string{"test"},
		CreatedBy:   "admin",
	})
	createReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(createBody))
	createW := httptest.NewRecorder()
	handler.handleCreateRole(createW, createReq)

	var createResp RoleResponse
	json.NewDecoder(createW.Body).Decode(&createResp)
	roleID := createResp.Role.ID

	tests := []struct {
		name       string
		id         string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "valid delete",
			id:         roleID.String(),
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "invalid role ID",
			id:         "invalid",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ROLE_ID",
		},
		{
			name:       "non-existing role",
			id:         "00000000-0000-0000-0000-000000000000",
			wantStatus: http.StatusNotFound,
			wantCode:   "ROLE_NOT_FOUND",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/roles/"+tt.id, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", tt.id)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			handler.handleDeleteRole(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleDeleteRole() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleDeleteRole() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleRevokeRole(t *testing.T) {
	authNHandler := setupAuthNHandler()
	authZHandler := setupAuthZHandler()

	// Create user and role
	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "revoke@example.com",
		Password:    "Password123!",
		Username:    "revokeuser",
		DisplayName: "Revoke User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	authNHandler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	roleBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "revokable",
		Description: "Can be revoked",
		Permissions: []string{"test"},
		CreatedBy:   "admin",
	})
	roleReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(roleBody))
	roleW := httptest.NewRecorder()
	authZHandler.handleCreateRole(roleW, roleReq)

	var roleResp RoleResponse
	json.NewDecoder(roleW.Body).Decode(&roleResp)
	roleID := roleResp.Role.ID

	// Assign role
	assignBody, _ := json.Marshal(AssignRoleRequest{
		UserID:     userID.String(),
		RoleID:     roleID.String(),
		AssignedBy: "admin",
	})
	assignReq := httptest.NewRequest(http.MethodPost, "/grants", bytes.NewReader(assignBody))
	assignW := httptest.NewRecorder()
	authZHandler.handleAssignRole(assignW, assignReq)

	tests := []struct {
		name       string
		body       RevokeRoleRequest
		wantStatus int
		wantCode   string
	}{
		{
			name: "valid revoke",
			body: RevokeRoleRequest{
				UserID: userID.String(),
				RoleID: roleID.String(),
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name: "invalid user ID",
			body: RevokeRoleRequest{
				UserID: "invalid",
				RoleID: roleID.String(),
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
		{
			name: "invalid role ID",
			body: RevokeRoleRequest{
				UserID: userID.String(),
				RoleID: "invalid",
			},
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_ROLE_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodDelete, "/grants", bytes.NewReader(body))
			w := httptest.NewRecorder()

			authZHandler.handleRevokeRole(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleRevokeRole() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleRevokeRole() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleGetUserRoles(t *testing.T) {
	authNHandler := setupAuthNHandler()
	authZHandler := setupAuthZHandler()

	// Create user and assign roles
	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "roles@example.com",
		Password:    "Password123!",
		Username:    "rolesuser",
		DisplayName: "Roles User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	authNHandler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	tests := []struct {
		name       string
		userID     string
		wantStatus int
		wantCode   string
	}{
		{
			name:       "valid user",
			userID:     userID.String(),
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid user ID",
			userID:     "invalid",
			wantStatus: http.StatusBadRequest,
			wantCode:   "INVALID_USER_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID+"/roles", nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("user_id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			authZHandler.handleGetUserRoles(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleGetUserRoles() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleGetUserRoles() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			}
		})
	}
}

func TestHandleGetUserGrants(t *testing.T) {
	authNHandler := setupAuthNHandler()
	authZHandler := setupAuthZHandler()

	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "grants@example.com",
		Password:    "Password123!",
		Username:    "grantsuser",
		DisplayName: "Grants User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	authNHandler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	req := httptest.NewRequest(http.MethodGet, "/users/"+userID.String()+"/grants", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("user_id", userID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	authZHandler.handleGetUserGrants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleGetUserGrants() status = %v, want %v", w.Code, http.StatusOK)
	}
}

func TestHandleGetRoleGrants(t *testing.T) {
	authZHandler := setupAuthZHandler()

	roleBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "grantsrole",
		Description: "Has grants",
		Permissions: []string{"test"},
		CreatedBy:   "admin",
	})
	roleReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(roleBody))
	roleW := httptest.NewRecorder()
	authZHandler.handleCreateRole(roleW, roleReq)

	var roleResp RoleResponse
	json.NewDecoder(roleW.Body).Decode(&roleResp)
	roleID := roleResp.Role.ID

	req := httptest.NewRequest(http.MethodGet, "/roles/"+roleID.String()+"/grants", nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("role_id", roleID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	w := httptest.NewRecorder()
	authZHandler.handleGetRoleGrants(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("handleGetRoleGrants() status = %v, want %v", w.Code, http.StatusOK)
	}
}

func TestHandleCheckAnyPermission(t *testing.T) {
	authNHandler := setupAuthNHandler()
	authZHandler := setupAuthZHandler()

	// Create user with role
	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "anyperm@example.com",
		Password:    "Password123!",
		Username:    "anypermuser",
		DisplayName: "Any Perm User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	authNHandler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	roleBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "multi-permissions",
		Description: "Multiple permissions",
		Permissions: []string{"read", "write"},
		CreatedBy:   "admin",
	})
	roleReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(roleBody))
	roleW := httptest.NewRecorder()
	authZHandler.handleCreateRole(roleW, roleReq)

	var roleResp RoleResponse
	json.NewDecoder(roleW.Body).Decode(&roleResp)
	roleID := roleResp.Role.ID

	assignBody, _ := json.Marshal(AssignRoleRequest{
		UserID:     userID.String(),
		RoleID:     roleID.String(),
		AssignedBy: "admin",
	})
	assignReq := httptest.NewRequest(http.MethodPost, "/grants", bytes.NewReader(assignBody))
	assignW := httptest.NewRecorder()
	authZHandler.handleAssignRole(assignW, assignReq)

	tests := []struct {
		name           string
		userID         string
		permissions    []string
		wantStatus     int
		wantPermission bool
		wantCode       string
	}{
		{
			name:           "has one of permissions",
			userID:         userID.String(),
			permissions:    []string{"read", "delete"},
			wantStatus:     http.StatusOK,
			wantPermission: true,
		},
		{
			name:           "has none of permissions",
			userID:         userID.String(),
			permissions:    []string{"delete", "admin"},
			wantStatus:     http.StatusOK,
			wantPermission: false,
		},
		{
			name:        "invalid user ID",
			userID:      "invalid",
			permissions: []string{"test"},
			wantStatus:  http.StatusBadRequest,
			wantCode:    "INVALID_USER_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(CheckAnyPermissionRequest{Permissions: tt.permissions})
			req := httptest.NewRequest(http.MethodPost, "/users/"+tt.userID+"/check-any-permission", bytes.NewReader(reqBody))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("user_id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			authZHandler.handleCheckAnyPermission(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleCheckAnyPermission() status = %v, want %v", w.Code, tt.wantStatus)
			}

			if tt.wantCode != "" {
				var errResp ErrorResponse
				json.NewDecoder(w.Body).Decode(&errResp)
				if errResp.Code != tt.wantCode {
					t.Errorf("handleCheckAnyPermission() error code = %v, want %v", errResp.Code, tt.wantCode)
				}
			} else {
				var resp PermissionCheckResponse
				json.NewDecoder(w.Body).Decode(&resp)
				if resp.HasPermission != tt.wantPermission {
					t.Errorf("handleCheckAnyPermission() hasPermission = %v, want %v", resp.HasPermission, tt.wantPermission)
				}
			}
		})
	}
}

func TestHandleCheckAllPermissions(t *testing.T) {
	authNHandler := setupAuthNHandler()
	authZHandler := setupAuthZHandler()

	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "allperm@example.com",
		Password:    "Password123!",
		Username:    "allpermuser",
		DisplayName: "All Perm User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	authNHandler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	roleBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "fullaccess",
		Description: "Full access",
		Permissions: []string{"read", "write", "delete"},
		CreatedBy:   "admin",
	})
	roleReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(roleBody))
	roleW := httptest.NewRecorder()
	authZHandler.handleCreateRole(roleW, roleReq)

	var roleResp RoleResponse
	json.NewDecoder(roleW.Body).Decode(&roleResp)
	roleID := roleResp.Role.ID

	assignBody, _ := json.Marshal(AssignRoleRequest{
		UserID:     userID.String(),
		RoleID:     roleID.String(),
		AssignedBy: "admin",
	})
	assignReq := httptest.NewRequest(http.MethodPost, "/grants", bytes.NewReader(assignBody))
	assignW := httptest.NewRecorder()
	authZHandler.handleAssignRole(assignW, assignReq)

	tests := []struct {
		name           string
		userID         string
		permissions    []string
		wantStatus     int
		wantPermission bool
	}{
		{
			name:           "has all permissions",
			userID:         userID.String(),
			permissions:    []string{"read", "write"},
			wantStatus:     http.StatusOK,
			wantPermission: true,
		},
		{
			name:           "missing one permission",
			userID:         userID.String(),
			permissions:    []string{"read", "admin"},
			wantStatus:     http.StatusOK,
			wantPermission: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reqBody, _ := json.Marshal(CheckAllPermissionsRequest{Permissions: tt.permissions})
			req := httptest.NewRequest(http.MethodPost, "/users/"+tt.userID+"/check-all-permissions", bytes.NewReader(reqBody))
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("user_id", tt.userID)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			authZHandler.handleCheckAllPermissions(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleCheckAllPermissions() status = %v, want %v", w.Code, tt.wantStatus)
			}

			var resp PermissionCheckResponse
			json.NewDecoder(w.Body).Decode(&resp)
			if resp.HasPermission != tt.wantPermission {
				t.Errorf("handleCheckAllPermissions() hasPermission = %v, want %v", resp.HasPermission, tt.wantPermission)
			}
		})
	}
}

func TestHandleHasRole(t *testing.T) {
	authNHandler := setupAuthNHandler()
	authZHandler := setupAuthZHandler()

	signupBody, _ := json.Marshal(SignUpRequest{
		Email:       "hasrole@example.com",
		Password:    "Password123!",
		Username:    "hasroleuser",
		DisplayName: "Has Role User",
	})
	signupReq := httptest.NewRequest(http.MethodPost, "/auth/signup", bytes.NewReader(signupBody))
	signupW := httptest.NewRecorder()
	authNHandler.handleSignUp(signupW, signupReq)

	var signupResp SignUpResponse
	json.NewDecoder(signupW.Body).Decode(&signupResp)
	userID := signupResp.User.ID

	roleBody, _ := json.Marshal(CreateRoleRequest{
		Name:        "moderator",
		Description: "Moderator role",
		Permissions: []string{"moderate"},
		CreatedBy:   "admin",
	})
	roleReq := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewReader(roleBody))
	roleW := httptest.NewRecorder()
	authZHandler.handleCreateRole(roleW, roleReq)

	var roleResp RoleResponse
	json.NewDecoder(roleW.Body).Decode(&roleResp)
	roleID := roleResp.Role.ID

	assignBody, _ := json.Marshal(AssignRoleRequest{
		UserID:     userID.String(),
		RoleID:     roleID.String(),
		AssignedBy: "admin",
	})
	assignReq := httptest.NewRequest(http.MethodPost, "/grants", bytes.NewReader(assignBody))
	assignW := httptest.NewRecorder()
	authZHandler.handleAssignRole(assignW, assignReq)

	tests := []struct {
		name       string
		userID     string
		roleName   string
		wantStatus int
		hasRole    bool
	}{
		{
			name:       "has role",
			userID:     userID.String(),
			roleName:   "moderator",
			wantStatus: http.StatusOK,
			hasRole:    true,
		},
		{
			name:       "does not have role",
			userID:     userID.String(),
			roleName:   "admin",
			wantStatus: http.StatusOK,
			hasRole:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID+"/has-role/"+tt.roleName, nil)
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("user_id", tt.userID)
			rctx.URLParams.Add("role_name", tt.roleName)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			w := httptest.NewRecorder()
			authZHandler.handleHasRole(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("handleHasRole() status = %v, want %v", w.Code, tt.wantStatus)
			}

			var resp HasRoleResponse
			json.NewDecoder(w.Body).Decode(&resp)
			if resp.HasRole != tt.hasRole {
				t.Errorf("handleHasRole() hasRole = %v, want %v", resp.HasRole, tt.hasRole)
			}
		})
	}
}
