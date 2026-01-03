package telemetry

import (
	"context"
	"time"
)

// Metrics models a minimal counter/measure emission interface with HTTP-specific observations.
type Metrics interface {
	Counter(ctx context.Context, name string, value float64, labels map[string]string)
	ObserveHTTPRequest(path, method string, status int, duration time.Duration)
}

// Tracer models an instrumentation provider capable of creating spans.
type Tracer interface {
	Start(ctx context.Context, name string, attrs map[string]any) (context.Context, Span)
}

// Span is the handle returned by Tracer.Start.
type Span interface {
	End(err error)
}

// NoopMetrics is a no-op implementation of Metrics.
type NoopMetrics struct{}

func (NoopMetrics) Counter(context.Context, string, float64, map[string]string) {}
func (NoopMetrics) ObserveHTTPRequest(string, string, int, time.Duration)       {}

// NoopTracer is a no-op implementation of Tracer.
type NoopTracer struct{}

func (NoopTracer) Start(ctx context.Context, _ string, _ map[string]any) (context.Context, Span) {
	return ctx, NoopSpan{}
}

// NoopSpan is a no-op implementation of Span.
type NoopSpan struct{}

func (NoopSpan) End(error) {}
