package internal

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
)

func TestNewHandler(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := validTestConfig(t)
	authnClient := httpclient.New("http://localhost:8082", logger)
	authzClient := httpclient.New("http://localhost:8083", logger)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	if handler == nil {
		t.Fatal("handler is nil")
	}

	if handler.authnClient == nil {
		t.Error("authnClient is nil")
	}

	if handler.authzClient == nil {
		t.Error("authzClient is nil")
	}

	if handler.cfg == nil {
		t.Error("cfg is nil")
	}

	if handler.log == nil {
		t.Error("log is nil")
	}
}

func TestHandleIndex(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := validTestConfig(t)
	authnClient := httpclient.New("http://localhost:8082", logger)
	authzClient := httpclient.New("http://localhost:8083", logger)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.handleIndex(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %s, want text/html; charset=utf-8", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Admin Service") {
		t.Error("body missing 'Admin Service'")
	}
}

func TestHandleDashboard(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := validTestConfig(t)
	authnClient := httpclient.New("http://localhost:8082", logger)
	authzClient := httpclient.New("http://localhost:8083", logger)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()

	handler.handleDashboard(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	if contentType := rec.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %s, want text/html; charset=utf-8", contentType)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Dashboard") {
		t.Error("body missing 'Dashboard'")
	}
}
