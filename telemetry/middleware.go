package telemetry

import (
	"net/http"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

// MetricsMiddleware measures request durations and reports them through the provided Metrics implementation.
func MetricsMiddleware(metrics Metrics) func(http.Handler) http.Handler {
	if metrics == nil {
		metrics = NoopMetrics{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			rw := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(rw, r)

			duration := time.Since(start)
			metrics.ObserveHTTPRequest(r.URL.Path, r.Method, rw.Status(), duration)
		})
	}
}
