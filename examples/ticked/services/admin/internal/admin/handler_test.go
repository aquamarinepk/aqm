package admin

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
)

func TestNewHandler(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := validTestConfig(t)
	httpAuthnClient := httpclient.New("http://localhost:8082", logger)
	httpAuthzClient := httpclient.New("http://localhost:8083", logger)

	authnClient := NewAuthNClient(httpAuthnClient)
	authzClient := NewAuthZClient(httpAuthzClient)

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

	if handler.templates == nil {
		t.Error("templates is nil")
	}
}

func TestHandleIndex(t *testing.T) {
	logger := log.NewLogger("error")
	cfg := validTestConfig(t)
	httpAuthnClient := httpclient.New("http://localhost:8082", logger)
	httpAuthzClient := httpclient.New("http://localhost:8083", logger)

	authnClient := NewAuthNClient(httpAuthnClient)
	authzClient := NewAuthZClient(httpAuthzClient)

	handler := NewHandler(authnClient, authzClient, cfg, logger)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.handleIndex(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}

	location := rec.Header().Get("Location")
	if location != "/admin" {
		t.Errorf("Location = %s, want /admin", location)
	}
}

