package telemetry

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type fakeMetrics struct {
	observedPath     string
	observedMethod   string
	observedStatus   int
	observedDuration time.Duration
	counterCalls     int
}

func (f *fakeMetrics) Counter(ctx context.Context, name string, value float64, labels map[string]string) {
	f.counterCalls++
}

func (f *fakeMetrics) ObserveHTTPRequest(path, method string, status int, duration time.Duration) {
	f.observedPath = path
	f.observedMethod = method
	f.observedStatus = status
	f.observedDuration = duration
}

func TestMetricsMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		metrics        Metrics
		path           string
		method         string
		expectedStatus int
		wantObserved   bool
	}{
		{
			name:           "with real metrics",
			metrics:        &fakeMetrics{},
			path:           "/test",
			method:         "GET",
			expectedStatus: 200,
			wantObserved:   true,
		},
		{
			name:           "with nil metrics defaults to noop",
			metrics:        nil,
			path:           "/test",
			method:         "POST",
			expectedStatus: 201,
			wantObserved:   false,
		},
		{
			name:           "with noop metrics",
			metrics:        NoopMetrics{},
			path:           "/api/users",
			method:         "DELETE",
			expectedStatus: 204,
			wantObserved:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.expectedStatus)
			})

			middleware := MetricsMiddleware(tt.metrics)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			rec := httptest.NewRecorder()

			wrappedHandler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.wantObserved {
				fake, ok := tt.metrics.(*fakeMetrics)
				if !ok {
					t.Fatal("expected fakeMetrics")
				}

				if fake.observedPath != tt.path {
					t.Errorf("expected path %s, got %s", tt.path, fake.observedPath)
				}

				if fake.observedMethod != tt.method {
					t.Errorf("expected method %s, got %s", tt.method, fake.observedMethod)
				}

				if fake.observedStatus != tt.expectedStatus {
					t.Errorf("expected status %d, got %d", tt.expectedStatus, fake.observedStatus)
				}

				if fake.observedDuration == 0 {
					t.Error("expected non-zero duration")
				}
			}
		})
	}
}

func TestMetricsMiddlewareWithError(t *testing.T) {
	fake := &fakeMetrics{}
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	middleware := MetricsMiddleware(fake)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if fake.observedStatus != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", fake.observedStatus)
	}
}
