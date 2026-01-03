package telemetry

import (
	"context"
	"testing"
	"time"
)

func TestNoopMetricsCounter(t *testing.T) {
	m := NoopMetrics{}
	ctx := context.Background()

	m.Counter(ctx, "test.counter", 1.0, map[string]string{"key": "value"})
}

func TestNoopMetricsObserveHTTPRequest(t *testing.T) {
	m := NoopMetrics{}

	m.ObserveHTTPRequest("/test", "GET", 200, time.Second)
}

func TestNoopTracerStart(t *testing.T) {
	tr := NoopTracer{}
	ctx := context.Background()

	newCtx, span := tr.Start(ctx, "test.span", map[string]any{"key": "value"})

	if newCtx != ctx {
		t.Errorf("expected context to be unchanged")
	}

	if span == nil {
		t.Errorf("expected non-nil span")
	}

	if _, ok := span.(NoopSpan); !ok {
		t.Errorf("expected NoopSpan, got %T", span)
	}
}

func TestNoopSpanEnd(t *testing.T) {
	s := NoopSpan{}

	s.End(nil)
	s.End(context.Canceled)
}
