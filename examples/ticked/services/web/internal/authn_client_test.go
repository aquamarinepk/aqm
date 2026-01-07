package internal

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
	"github.com/google/uuid"
)

func TestSignIn(t *testing.T) {
	tests := []struct {
		name           string
		email          string
		password       string
		responseCode   int
		responseBody   string
		wantUserID     uuid.UUID
		wantToken      string
		wantErr        bool
		wantErrMessage string
	}{
		{
			name:         "successful signin",
			email:        "user@example.com",
			password:     "password123",
			responseCode: http.StatusOK,
			responseBody: `{
				"user": {
					"id": "550e8400-e29b-41d4-a716-446655440000",
					"email": "user@example.com",
					"username": "testuser",
					"created_at": "2024-01-01T00:00:00Z"
				},
				"token": "jwt-token-123"
			}`,
			wantUserID: uuid.MustParse("550e8400-e29b-41d4-a716-446655440000"),
			wantToken:  "jwt-token-123",
			wantErr:    false,
		},
		{
			name:           "invalid credentials",
			email:          "user@example.com",
			password:       "wrong-password",
			responseCode:   http.StatusUnauthorized,
			responseBody:   `{"error": "invalid credentials"}`,
			wantErr:        true,
			wantErrMessage: "invalid credentials",
		},
		{
			name:         "server error",
			email:        "user@example.com",
			password:     "password123",
			responseCode: http.StatusInternalServerError,
			responseBody: `{"error": "internal server error"}`,
			wantErr:      true,
		},
		{
			name:         "invalid json response",
			email:        "user@example.com",
			password:     "password123",
			responseCode: http.StatusOK,
			responseBody: `{invalid json}`,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/auth/signin" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("unexpected method: %s", r.Method)
				}

				w.WriteHeader(tt.responseCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			logger := log.NewNoopLogger()
			httpClient := httpclient.New(server.URL, logger)
			client := NewAuthNClient(httpClient)

			user, token, err := client.SignIn(context.Background(), tt.email, tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("SignIn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.wantErrMessage != "" && err.Error() != tt.wantErrMessage {
					t.Errorf("SignIn() error message = %v, want %v", err.Error(), tt.wantErrMessage)
				}
				return
			}

			if user.ID != tt.wantUserID {
				t.Errorf("SignIn() user.ID = %v, want %v", user.ID, tt.wantUserID)
			}

			if token != tt.wantToken {
				t.Errorf("SignIn() token = %v, want %v", token, tt.wantToken)
			}
		})
	}
}

func TestSignInNetworkError(t *testing.T) {
	logger := log.NewNoopLogger()
	httpClient := httpclient.New("http://localhost:99999", logger, httpclient.WithTimeout(100*time.Millisecond))
	client := NewAuthNClient(httpClient)

	_, _, err := client.SignIn(context.Background(), "test@example.com", "password")

	if err == nil {
		t.Error("SignIn() expected network error, got nil")
	}
}
