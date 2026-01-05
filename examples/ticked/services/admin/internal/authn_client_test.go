package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
	"github.com/google/uuid"
)

func TestAuthNClient_ListUsers(t *testing.T) {
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
						"id":         uuid.New().String(),
						"email":      "user1@example.com",
						"username":   "user1",
						"status":     "active",
						"created_at": time.Now().Format(time.RFC3339),
						"updated_at": time.Now().Format(time.RFC3339),
					},
					{
						"id":         uuid.New().String(),
						"email":      "user2@example.com",
						"username":   "user2",
						"status":     "active",
						"created_at": time.Now().Format(time.RFC3339),
						"updated_at": time.Now().Format(time.RFC3339),
					},
				},
			},
			wantErr: false,
			wantLen: 2,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			response:   map[string]interface{}{"error": "internal error"},
			wantErr:    true,
		},
		{
			name:       "empty list",
			statusCode: http.StatusOK,
			response: map[string]interface{}{
				"data": []map[string]interface{}{},
			},
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/users" {
					t.Errorf("expected path /users, got %s", r.URL.Path)
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
			client := NewAuthNClient(httpClient)

			ctx := context.Background()
			users, err := client.ListUsers(ctx)

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

			if len(users) != tt.wantLen {
				t.Errorf("expected %d users, got %d", tt.wantLen, len(users))
			}
		})
	}
}

func TestAuthNClient_GetUser(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		statusCode int
		response   interface{}
		wantErr    bool
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			response: map[string]interface{}{
				"data": map[string]interface{}{
					"id":         userID.String(),
					"email":      "user@example.com",
					"username":   "testuser",
					"status":     "active",
					"created_at": time.Now().Format(time.RFC3339),
					"updated_at": time.Now().Format(time.RFC3339),
				},
			},
			wantErr: false,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			response:   map[string]interface{}{"error": "user not found"},
			wantErr:    true,
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
				expectedPath := "/users/" + userID.String()
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
			client := NewAuthNClient(httpClient)

			ctx := context.Background()
			user, err := client.GetUser(ctx, userID)

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

			if user == nil {
				t.Error("expected user, got nil")
				return
			}

			if user.ID != userID {
				t.Errorf("expected ID %s, got %s", userID, user.ID)
			}

			if user.Username != "testuser" {
				t.Errorf("expected username testuser, got %s", user.Username)
			}
		})
	}
}
