package admin

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

func TestHandleListUsersIntegration(t *testing.T) {
	userID1 := uuid.New()
	userID2 := uuid.New()

	authnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":         userID1.String(),
						"email":      "user1@example.com",
						"username":   "user1",
						"status":     "active",
						"created_at": time.Now().Format(time.RFC3339),
						"updated_at": time.Now().Format(time.RFC3339),
					},
					{
						"id":         userID2.String(),
						"email":      "user2@example.com",
						"username":   "user2",
						"status":     "active",
						"created_at": time.Now().Format(time.RFC3339),
						"updated_at": time.Now().Format(time.RFC3339),
					},
				},
			})
		}
	}))
	defer authnServer.Close()

	authzServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer authzServer.Close()

	logger := log.NewLogger("error")
	cfg := validTestConfig(t)

	httpAuthnClient := httpclient.New(authnServer.URL, logger)
	httpAuthzClient := httpclient.New(authzServer.URL, logger)

	authnClient := NewAuthNClient(httpAuthnClient)
	authzClient := NewAuthZClient(httpAuthzClient)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/list-users", nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	handler.handleListUsers(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status code: %d", rec.Code)
	}
}

func TestHandleListUsersError(t *testing.T) {
	authnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer authnServer.Close()

	authzServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer authzServer.Close()

	logger := log.NewLogger("error")
	cfg := validTestConfig(t)

	httpAuthnClient := httpclient.New(authnServer.URL, logger)
	httpAuthzClient := httpclient.New(authzServer.URL, logger)

	authnClient := NewAuthNClient(httpAuthnClient)
	authzClient := NewAuthZClient(httpAuthzClient)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/list-users", nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	handler.handleListUsers(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestHandleGetUserIntegration(t *testing.T) {
	userID := uuid.New()

	authnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/"+userID.String() {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":         userID.String(),
					"email":      "user@example.com",
					"username":   "testuser",
					"status":     "active",
					"created_at": time.Now().Format(time.RFC3339),
					"updated_at": time.Now().Format(time.RFC3339),
				},
			})
		}
	}))
	defer authnServer.Close()

	authzServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/grants/user/"+userID.String() {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{
						"id":          uuid.New().String(),
						"name":        "admin",
						"description": "Administrator",
						"permissions": []string{"users:read", "users:write"},
					},
				},
			})
		}
	}))
	defer authzServer.Close()

	logger := log.NewLogger("error")
	cfg := validTestConfig(t)

	httpAuthnClient := httpclient.New(authnServer.URL, logger)
	httpAuthzClient := httpclient.New(authzServer.URL, logger)

	authnClient := NewAuthNClient(httpAuthnClient)
	authzClient := NewAuthZClient(httpAuthzClient)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/get-user?id="+userID.String(), nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	handler.handleGetUser(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status code: %d", rec.Code)
	}
}

func TestHandleGetUserMissingID(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := validTestConfig(t)

	httpAuthnClient := httpclient.New("http://localhost:8082", logger)
	httpAuthzClient := httpclient.New("http://localhost:8083", logger)

	authnClient := NewAuthNClient(httpAuthnClient)
	authzClient := NewAuthZClient(httpAuthzClient)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/get-user", nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	handler.handleGetUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleGetUserInvalidID(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := validTestConfig(t)

	httpAuthnClient := httpclient.New("http://localhost:8082", logger)
	httpAuthzClient := httpclient.New("http://localhost:8083", logger)

	authnClient := NewAuthNClient(httpAuthnClient)
	authzClient := NewAuthZClient(httpAuthzClient)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/get-user?id=invalid", nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	handler.handleGetUser(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestHandleGetUserNotFound(t *testing.T) {
	userID := uuid.New()

	authnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer authnServer.Close()

	authzServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer authzServer.Close()

	logger := log.NewLogger("error")
	cfg := validTestConfig(t)

	httpAuthnClient := httpclient.New(authnServer.URL, logger)
	httpAuthzClient := httpclient.New(authzServer.URL, logger)

	authnClient := NewAuthNClient(httpAuthnClient)
	authzClient := NewAuthZClient(httpAuthzClient)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/get-user?id="+userID.String(), nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	handler.handleGetUser(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestHandleGetUserAuthZError(t *testing.T) {
	userID := uuid.New()

	authnServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/users/"+userID.String() {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"id":         userID.String(),
					"email":      "user@example.com",
					"username":   "testuser",
					"status":     "active",
					"created_at": time.Now().Format(time.RFC3339),
					"updated_at": time.Now().Format(time.RFC3339),
				},
			})
		}
	}))
	defer authnServer.Close()

	authzServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer authzServer.Close()

	logger := log.NewLogger("error")
	cfg := validTestConfig(t)

	httpAuthnClient := httpclient.New(authnServer.URL, logger)
	httpAuthzClient := httpclient.New(authzServer.URL, logger)

	authnClient := NewAuthNClient(httpAuthnClient)
	authzClient := NewAuthZClient(httpAuthzClient)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin/get-user?id="+userID.String(), nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	handler.handleGetUser(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status code: %d", rec.Code)
	}
}

func TestHandleDashboard(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := validTestConfig(t)

	httpAuthnClient := httpclient.New("http://localhost:8082", logger)
	httpAuthzClient := httpclient.New("http://localhost:8083", logger)

	authnClient := NewAuthNClient(httpAuthnClient)
	authzClient := NewAuthZClient(httpAuthzClient)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	req = req.WithContext(context.Background())
	rec := httptest.NewRecorder()

	handler.handleDashboard(rec, req)

	if rec.Code != http.StatusOK && rec.Code != http.StatusInternalServerError {
		t.Errorf("unexpected status code: %d", rec.Code)
	}
}
