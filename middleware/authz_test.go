package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aquamarinepk/aqm/auth"
	"github.com/aquamarinepk/aqm/auth/fake"
)

func TestAuthzChecker_HasRole(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)

	adminRole := auth.NewRole()
	adminRole.Name = "admin"
	adminRole.Status = auth.RoleStatusActive
	adminRole.BeforeCreate()
	_ = roleStore.Create(context.Background(), adminRole)

	username := "testuser"
	grant := auth.NewGrant(username, adminRole.ID, "system")
	_ = grantStore.Create(context.Background(), grant)

	checker := NewAuthzChecker(grantStore)

	tests := []struct {
		name     string
		username string
		roleName string
		want     bool
		wantErr  bool
	}{
		{
			name:     "user has role",
			username: username,
			roleName: "admin",
			want:     true,
			wantErr:  false,
		},
		{
			name:     "user does not have role",
			username: username,
			roleName: "editor",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "empty username",
			username: "",
			roleName: "admin",
			want:     false,
			wantErr:  false,
		},
		{
			name:     "nonexistent user",
			username: "nonexistent",
			roleName: "admin",
			want:     false,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.HasRole(context.Background(), tt.username, tt.roleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("HasRole() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("HasRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthzChecker_CheckPermission(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)

	adminRole := auth.NewRole()
	adminRole.Name = "admin"
	adminRole.Permissions = []string{"users:read", "users:write"}
	adminRole.Status = auth.RoleStatusActive
	adminRole.BeforeCreate()
	_ = roleStore.Create(context.Background(), adminRole)

	username := "testuser"
	grant := auth.NewGrant(username, adminRole.ID, "system")
	_ = grantStore.Create(context.Background(), grant)

	checker := NewAuthzChecker(grantStore)

	tests := []struct {
		name       string
		username   string
		permission string
		want       bool
		wantErr    bool
	}{
		{
			name:       "user has permission",
			username:   username,
			permission: "users:read",
			want:       true,
			wantErr:    false,
		},
		{
			name:       "user does not have permission",
			username:   username,
			permission: "posts:delete",
			want:       false,
			wantErr:    false,
		},
		{
			name:       "empty username",
			username:   "",
			permission: "users:read",
			want:       false,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.CheckPermission(context.Background(), tt.username, tt.permission)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthzChecker_CheckAnyPermission(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)

	adminRole := auth.NewRole()
	adminRole.Name = "admin"
	adminRole.Permissions = []string{"users:read", "users:write"}
	adminRole.Status = auth.RoleStatusActive
	adminRole.BeforeCreate()
	_ = roleStore.Create(context.Background(), adminRole)

	username := "testuser"
	grant := auth.NewGrant(username, adminRole.ID, "system")
	_ = grantStore.Create(context.Background(), grant)

	checker := NewAuthzChecker(grantStore)

	tests := []struct {
		name        string
		username    string
		permissions []string
		want        bool
		wantErr     bool
	}{
		{
			name:        "user has one of the permissions",
			username:    username,
			permissions: []string{"users:read", "posts:read"},
			want:        true,
			wantErr:     false,
		},
		{
			name:        "user has all permissions",
			username:    username,
			permissions: []string{"users:read", "users:write"},
			want:        true,
			wantErr:     false,
		},
		{
			name:        "user has none of the permissions",
			username:    username,
			permissions: []string{"posts:read", "posts:write"},
			want:        false,
			wantErr:     false,
		},
		{
			name:        "empty username",
			username:    "",
			permissions: []string{"users:read"},
			want:        false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.CheckAnyPermission(context.Background(), tt.username, tt.permissions)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAnyPermission() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckAnyPermission() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthzChecker_CheckAllPermissions(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)

	adminRole := auth.NewRole()
	adminRole.Name = "admin"
	adminRole.Permissions = []string{"users:read", "users:write", "posts:read"}
	adminRole.Status = auth.RoleStatusActive
	adminRole.BeforeCreate()
	_ = roleStore.Create(context.Background(), adminRole)

	username := "testuser"
	grant := auth.NewGrant(username, adminRole.ID, "system")
	_ = grantStore.Create(context.Background(), grant)

	checker := NewAuthzChecker(grantStore)

	tests := []struct {
		name        string
		username    string
		permissions []string
		want        bool
		wantErr     bool
	}{
		{
			name:        "user has all permissions",
			username:    username,
			permissions: []string{"users:read", "users:write"},
			want:        true,
			wantErr:     false,
		},
		{
			name:        "user missing one permission",
			username:    username,
			permissions: []string{"users:read", "posts:write"},
			want:        false,
			wantErr:     false,
		},
		{
			name:        "user has none of the permissions",
			username:    username,
			permissions: []string{"comments:read", "comments:write"},
			want:        false,
			wantErr:     false,
		},
		{
			name:        "empty username",
			username:    "",
			permissions: []string{"users:read"},
			want:        false,
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checker.CheckAllPermissions(context.Background(), tt.username, tt.permissions)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckAllPermissions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("CheckAllPermissions() = %v, want %v", got, tt.want)
			}
		})
	}
}

type fakeRoleChecker struct {
	hasRole             bool
	checkPermission     bool
	checkAnyPermission  bool
	checkAllPermissions bool
	err                 error
}

func (f *fakeRoleChecker) HasRole(ctx context.Context, userID string, roleName string) (bool, error) {
	return f.hasRole, f.err
}

func (f *fakeRoleChecker) CheckPermission(ctx context.Context, userID string, permission string) (bool, error) {
	return f.checkPermission, f.err
}

func (f *fakeRoleChecker) CheckAnyPermission(ctx context.Context, userID string, permissions []string) (bool, error) {
	return f.checkAnyPermission, f.err
}

func (f *fakeRoleChecker) CheckAllPermissions(ctx context.Context, userID string, permissions []string) (bool, error) {
	return f.checkAllPermissions, f.err
}

func TestRequireRole(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		hasRole        bool
		err            error
		expectedStatus int
	}{
		{
			name:           "user has required role",
			username:       "testuser",
			hasRole:        true,
			err:            nil,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "user does not have required role",
			username:       "testuser",
			hasRole:        false,
			err:            nil,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "no username in context",
			username:       "",
			hasRole:        false,
			err:            nil,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "role checker error",
			username:       "testuser",
			hasRole:        false,
			err:            http.ErrServerClosed,
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &fakeRoleChecker{
				hasRole: tt.hasRole,
				err:     tt.err,
			}

			handler := RequireRole(checker, "admin")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.username != "" {
				ctx := context.WithValue(req.Context(), UserIDKey, tt.username)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("RequireRole() status = %v, want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestRequirePermission(t *testing.T) {
	tests := []struct {
		name            string
		username        string
		checkPermission bool
		err             error
		expectedStatus  int
	}{
		{
			name:            "user has required permission",
			username:        "testuser",
			checkPermission: true,
			err:             nil,
			expectedStatus:  http.StatusOK,
		},
		{
			name:            "user does not have required permission",
			username:        "testuser",
			checkPermission: false,
			err:             nil,
			expectedStatus:  http.StatusForbidden,
		},
		{
			name:            "no username in context",
			username:        "",
			checkPermission: false,
			err:             nil,
			expectedStatus:  http.StatusUnauthorized,
		},
		{
			name:            "permission checker error",
			username:        "testuser",
			checkPermission: false,
			err:             http.ErrServerClosed,
			expectedStatus:  http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &fakeRoleChecker{
				checkPermission: tt.checkPermission,
				err:             tt.err,
			}

			handler := RequirePermission(checker, "users:read")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.username != "" {
				ctx := context.WithValue(req.Context(), UserIDKey, tt.username)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("RequirePermission() status = %v, want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestRequireAnyPermission(t *testing.T) {
	tests := []struct {
		name               string
		username           string
		checkAnyPermission bool
		err                error
		expectedStatus     int
	}{
		{
			name:               "user has one of the required permissions",
			username:           "testuser",
			checkAnyPermission: true,
			err:                nil,
			expectedStatus:     http.StatusOK,
		},
		{
			name:               "user has none of the required permissions",
			username:           "testuser",
			checkAnyPermission: false,
			err:                nil,
			expectedStatus:     http.StatusForbidden,
		},
		{
			name:               "no username in context",
			username:           "",
			checkAnyPermission: false,
			err:                nil,
			expectedStatus:     http.StatusUnauthorized,
		},
		{
			name:               "permission checker error",
			username:           "testuser",
			checkAnyPermission: false,
			err:                http.ErrServerClosed,
			expectedStatus:     http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &fakeRoleChecker{
				checkAnyPermission: tt.checkAnyPermission,
				err:                tt.err,
			}

			handler := RequireAnyPermission(checker, []string{"users:read", "posts:read"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.username != "" {
				ctx := context.WithValue(req.Context(), UserIDKey, tt.username)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("RequireAnyPermission() status = %v, want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestRequireAllPermissions(t *testing.T) {
	tests := []struct {
		name                string
		username            string
		checkAllPermissions bool
		err                 error
		expectedStatus      int
	}{
		{
			name:                "user has all required permissions",
			username:            "testuser",
			checkAllPermissions: true,
			err:                 nil,
			expectedStatus:      http.StatusOK,
		},
		{
			name:                "user missing some required permissions",
			username:            "testuser",
			checkAllPermissions: false,
			err:                 nil,
			expectedStatus:      http.StatusForbidden,
		},
		{
			name:                "no username in context",
			username:            "",
			checkAllPermissions: false,
			err:                 nil,
			expectedStatus:      http.StatusUnauthorized,
		},
		{
			name:                "permission checker error",
			username:            "testuser",
			checkAllPermissions: false,
			err:                 http.ErrServerClosed,
			expectedStatus:      http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			checker := &fakeRoleChecker{
				checkAllPermissions: tt.checkAllPermissions,
				err:                 tt.err,
			}

			handler := RequireAllPermissions(checker, []string{"users:read", "users:write"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.username != "" {
				ctx := context.WithValue(req.Context(), UserIDKey, tt.username)
				req = req.WithContext(ctx)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("RequireAllPermissions() status = %v, want %v", rr.Code, tt.expectedStatus)
			}
		})
	}
}

func TestNewAuthzChecker(t *testing.T) {
	roleStore := fake.NewRoleStore()
	grantStore := fake.NewGrantStore(roleStore)

	checker := NewAuthzChecker(grantStore)

	if checker == nil {
		t.Fatal("NewAuthzChecker() returned nil")
	}
	if checker.grantStore != grantStore {
		t.Error("NewAuthzChecker() did not set grantStore correctly")
	}
}
