package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
	"github.com/google/uuid"
)

func TestAuthZClientGetUserRoles(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		statusCode int
		response   interface{}
		wantErr    bool
		wantLen    int
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			response: map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":          uuid.New().String(),
						"name":        "admin",
						"description": "Administrator role",
						"permissions": []string{"users:read", "users:write"},
					},
					{
						"id":          uuid.New().String(),
						"name":        "viewer",
						"description": "Viewer role",
						"permissions": []string{"users:read"},
					},
				},
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:       "empty roles",
			statusCode: http.StatusOK,
			response: map[string]interface{}{
				"data": []map[string]interface{}{},
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			response:   map[string]interface{}{"error": "internal error"},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				expectedPath := "/grants/user/" + userID.String()
				if r.URL.Path != expectedPath {
					t.Errorf("expected path %s, got %s", expectedPath, r.URL.Path)
				}
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}

				w.WriteHeader(tt.statusCode)
				json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			logger := log.NewLogger("error")
			httpClient := httpclient.New(server.URL, logger)
			client := NewAuthZClient(httpClient)

			ctx := context.Background()
			roles, err := client.GetUserRoles(ctx, userID)

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(roles) != tt.wantLen {
				t.Errorf("expected %d roles, got %d", tt.wantLen, len(roles))
			}

			if tt.wantLen > 0 {
				if roles[0].Name != "admin" {
					t.Errorf("expected first role name admin, got %s", roles[0].Name)
				}
				if len(roles[0].Permissions) != 2 {
					t.Errorf("expected 2 permissions for admin role, got %d", len(roles[0].Permissions))
				}
			}
		})
	}
}
