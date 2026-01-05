package preflight

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/log"
)

type mockCheck struct {
	name string
	err  error
}

func (m *mockCheck) Name() string {
	return m.name
}

func (m *mockCheck) Run(ctx context.Context) error {
	return m.err
}

func TestCheckerRunAll(t *testing.T) {
	tests := []struct {
		name    string
		checks  []Check
		wantErr bool
	}{
		{
			name: "no checks configured",
			checks: []Check{},
			wantErr: false,
		},
		{
			name: "single passing check",
			checks: []Check{
				&mockCheck{name: "test1", err: nil},
			},
			wantErr: false,
		},
		{
			name: "multiple passing checks",
			checks: []Check{
				&mockCheck{name: "test1", err: nil},
				&mockCheck{name: "test2", err: nil},
				&mockCheck{name: "test3", err: nil},
			},
			wantErr: false,
		},
		{
			name: "single failing check",
			checks: []Check{
				&mockCheck{name: "test1", err: errors.New("check failed")},
			},
			wantErr: true,
		},
		{
			name: "first check fails",
			checks: []Check{
				&mockCheck{name: "test1", err: errors.New("first failed")},
				&mockCheck{name: "test2", err: nil},
			},
			wantErr: true,
		},
		{
			name: "second check fails",
			checks: []Check{
				&mockCheck{name: "test1", err: nil},
				&mockCheck{name: "test2", err: errors.New("second failed")},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewLogger("error")
			checker := New(logger)

			for _, check := range tt.checks {
				checker.Add(check)
			}

			err := checker.RunAll(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("RunAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPCheck(t *testing.T) {
	tests := []struct {
		name          string
		serverHandler http.HandlerFunc
		wantErr       bool
	}{
		{
			name: "successful health check",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			},
			wantErr: false,
		},
		{
			name: "500 error",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			wantErr: true,
		},
		{
			name: "404 not found",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			check := HTTPCheck("test-service", server.URL+"/health")
			err := check.Run(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPCheck.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHTTPCheckContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	check := HTTPCheck("test-service", server.URL+"/health")
	err := check.Run(ctx)

	if err == nil {
		t.Error("expected context deadline exceeded error, got nil")
	}
}

func TestHTTPCheckInvalidURL(t *testing.T) {
	check := HTTPCheck("test-service", "http://localhost:99999/health")
	err := check.Run(context.Background())

	if err == nil {
		t.Error("expected connection error, got nil")
	}
}

func TestTCPCheck(t *testing.T) {
	listener, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		t.Fatalf("failed to create listener: %v", err)
	}
	defer listener.Close()

	tests := []struct {
		name    string
		address string
		wantErr bool
	}{
		{
			name:    "successful TCP connection",
			address: listener.Addr().String(),
			wantErr: false,
		},
		{
			name:    "connection refused",
			address: "localhost:99999",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			check := TCPCheck("test-tcp", tt.address)
			err := check.Run(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("TCPCheck.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTCPCheckContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	check := TCPCheck("test-tcp", "localhost:99999")
	err := check.Run(ctx)

	if err == nil {
		t.Error("expected context canceled error, got nil")
	}
}

func TestCheckName(t *testing.T) {
	tests := []struct {
		name     string
		check    Check
		wantName string
	}{
		{
			name:     "HTTP check name",
			check:    HTTPCheck("authn-service", "http://localhost:8082/health"),
			wantName: "authn-service",
		},
		{
			name:     "TCP check name",
			check:    TCPCheck("postgres-db", "localhost:5432"),
			wantName: "postgres-db",
		},
		{
			name:     "mock check name",
			check:    &mockCheck{name: "custom-check"},
			wantName: "custom-check",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.check.Name(); got != tt.wantName {
				t.Errorf("Check.Name() = %v, want %v", got, tt.wantName)
			}
		})
	}
}
