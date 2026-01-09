package nats

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/pubsub"
	"github.com/testcontainers/testcontainers-go/modules/nats"
)

func setupNATS(t *testing.T) (string, func()) {
	t.Helper()
	ctx := context.Background()

	container, err := nats.Run(ctx, "nats:2.10-alpine")
	if err != nil {
		t.Fatalf("cannot start NATS container: %v", err)
	}

	url, err := container.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("cannot get connection string: %v", err)
	}

	cleanup := func() {
		if err := container.Terminate(context.Background()); err != nil {
			t.Logf("cannot terminate container: %v", err)
		}
	}

	return url, cleanup
}

func testLogger() log.Logger {
	return log.NewNoopLogger()
}

func TestBrokerStartStop(t *testing.T) {
	url, cleanup := setupNATS(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.URL = url

	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	if err := broker.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := broker.Stop(ctx); err != nil {
		t.Fatalf("Stop failed: %v", err)
	}
}

func TestBrokerPublishSubscribe(t *testing.T) {
	url, cleanup := setupNATS(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.URL = url

	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	if err := broker.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer broker.Close()

	var received []pubsub.Envelope
	var mu sync.Mutex
	done := make(chan struct{})

	handler := func(ctx context.Context, env pubsub.Envelope) error {
		mu.Lock()
		received = append(received, env)
		mu.Unlock()
		close(done)
		return nil
	}

	if err := broker.Subscribe(ctx, "test-topic", handler, pubsub.SubscribeOptions{
		SubscriberID: "test-sub",
	}); err != nil {
		t.Fatalf("Subscribe failed: %v", err)
	}

	env := pubsub.NewEnvelope("test-topic", "test-payload")
	if err := broker.Publish(ctx, "test-topic", env); err != nil {
		t.Fatalf("Publish failed: %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	mu.Lock()
	defer mu.Unlock()

	if len(received) != 1 {
		t.Fatalf("expected 1 received message, got %d", len(received))
	}

	if received[0].ID != env.ID {
		t.Errorf("expected ID %s, got %s", env.ID, received[0].ID)
	}
}

func TestBrokerFanOut(t *testing.T) {
	url, cleanup := setupNATS(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.URL = url

	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	if err := broker.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer broker.Close()

	var count1, count2 int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	handler1 := func(ctx context.Context, env pubsub.Envelope) error {
		mu.Lock()
		count1++
		mu.Unlock()
		wg.Done()
		return nil
	}

	handler2 := func(ctx context.Context, env pubsub.Envelope) error {
		mu.Lock()
		count2++
		mu.Unlock()
		wg.Done()
		return nil
	}

	broker.Subscribe(ctx, "test-topic", handler1, pubsub.SubscribeOptions{SubscriberID: "sub1"})
	broker.Subscribe(ctx, "test-topic", handler2, pubsub.SubscribeOptions{SubscriberID: "sub2"})

	broker.Publish(ctx, "test-topic", pubsub.NewEnvelope("test-topic", "payload"))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for fan-out")
	}

	mu.Lock()
	defer mu.Unlock()

	if count1 != 1 {
		t.Errorf("handler1 expected 1 call, got %d", count1)
	}
	if count2 != 1 {
		t.Errorf("handler2 expected 1 call, got %d", count2)
	}
}

func TestBrokerDifferentTopics(t *testing.T) {
	url, cleanup := setupNATS(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.URL = url

	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	if err := broker.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer broker.Close()

	var received1, received2 int
	var mu sync.Mutex
	done := make(chan struct{})

	broker.Subscribe(ctx, "topic1", func(ctx context.Context, env pubsub.Envelope) error {
		mu.Lock()
		received1++
		mu.Unlock()
		close(done)
		return nil
	}, pubsub.SubscribeOptions{SubscriberID: "sub1"})

	broker.Subscribe(ctx, "topic2", func(ctx context.Context, env pubsub.Envelope) error {
		mu.Lock()
		received2++
		mu.Unlock()
		return nil
	}, pubsub.SubscribeOptions{SubscriberID: "sub2"})

	broker.Publish(ctx, "topic1", pubsub.NewEnvelope("topic1", "payload"))

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()

	if received1 != 1 {
		t.Errorf("topic1 handler expected 1 call, got %d", received1)
	}
	if received2 != 0 {
		t.Errorf("topic2 handler expected 0 calls, got %d", received2)
	}
}

func TestBrokerDuplicateSubscriberIDFails(t *testing.T) {
	url, cleanup := setupNATS(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.URL = url

	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	if err := broker.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer broker.Close()

	handler := func(ctx context.Context, env pubsub.Envelope) error { return nil }

	if err := broker.Subscribe(ctx, "test-topic", handler, pubsub.SubscribeOptions{
		SubscriberID: "same-id",
	}); err != nil {
		t.Fatalf("First subscribe failed: %v", err)
	}

	err := broker.Subscribe(ctx, "test-topic", handler, pubsub.SubscribeOptions{
		SubscriberID: "same-id",
	})
	if err == nil {
		t.Error("expected error for duplicate subscriber ID")
	}
}

func TestBrokerPublishBeforeStartFails(t *testing.T) {
	cfg := DefaultConfig()
	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	err := broker.Publish(ctx, "test-topic", pubsub.NewEnvelope("test-topic", "payload"))
	if err == nil {
		t.Error("expected error publishing before Start")
	}
}

func TestBrokerSubscribeBeforeStartFails(t *testing.T) {
	cfg := DefaultConfig()
	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	err := broker.Subscribe(ctx, "test-topic", func(ctx context.Context, env pubsub.Envelope) error {
		return nil
	}, pubsub.SubscribeOptions{})
	if err == nil {
		t.Error("expected error subscribing before Start")
	}
}

func TestBrokerPublishAfterCloseFails(t *testing.T) {
	url, cleanup := setupNATS(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.URL = url

	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	if err := broker.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	broker.Close()

	err := broker.Publish(ctx, "test-topic", pubsub.NewEnvelope("test-topic", "payload"))
	if err == nil {
		t.Error("expected error publishing to closed broker")
	}
}

func TestBrokerSubscribeAfterCloseFails(t *testing.T) {
	url, cleanup := setupNATS(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.URL = url

	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	if err := broker.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	broker.Close()

	err := broker.Subscribe(ctx, "test-topic", func(ctx context.Context, env pubsub.Envelope) error {
		return nil
	}, pubsub.SubscribeOptions{})
	if err == nil {
		t.Error("expected error subscribing to closed broker")
	}
}

func TestBrokerCloseIdempotent(t *testing.T) {
	url, cleanup := setupNATS(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.URL = url

	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	if err := broker.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if err := broker.Close(); err != nil {
		t.Fatalf("First close failed: %v", err)
	}

	if err := broker.Close(); err != nil {
		t.Fatalf("Second close failed: %v", err)
	}
}

func TestBrokerHandlerError(t *testing.T) {
	url, cleanup := setupNATS(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.URL = url

	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	if err := broker.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer broker.Close()

	var callCount int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	broker.Subscribe(ctx, "test-topic", func(ctx context.Context, env pubsub.Envelope) error {
		mu.Lock()
		callCount++
		mu.Unlock()
		wg.Done()
		return fmt.Errorf("handler error")
	}, pubsub.SubscribeOptions{SubscriberID: "test-sub"})

	broker.Publish(ctx, "test-topic", pubsub.NewEnvelope("test-topic", "msg1"))
	broker.Publish(ctx, "test-topic", pubsub.NewEnvelope("test-topic", "msg2"))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for messages")
	}

	mu.Lock()
	defer mu.Unlock()

	if callCount != 2 {
		t.Errorf("expected 2 calls despite errors, got %d", callCount)
	}
}

func TestBrokerWithMetadata(t *testing.T) {
	url, cleanup := setupNATS(t)
	defer cleanup()

	cfg := DefaultConfig()
	cfg.URL = url

	broker := NewBroker(cfg, testLogger())
	ctx := context.Background()

	if err := broker.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}
	defer broker.Close()

	var received pubsub.Envelope
	var mu sync.Mutex
	done := make(chan struct{})

	broker.Subscribe(ctx, "test-topic", func(ctx context.Context, env pubsub.Envelope) error {
		mu.Lock()
		received = env
		mu.Unlock()
		close(done)
		return nil
	}, pubsub.SubscribeOptions{SubscriberID: "test-sub"})

	env := pubsub.NewEnvelope("test-topic", "payload").
		WithMetadata("key1", "value1").
		WithMetadata("key2", "value2")

	broker.Publish(ctx, "test-topic", env)

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for message")
	}

	mu.Lock()
	defer mu.Unlock()

	if received.Metadata["key1"] != "value1" {
		t.Errorf("expected key1=value1, got %s", received.Metadata["key1"])
	}
	if received.Metadata["key2"] != "value2" {
		t.Errorf("expected key2=value2, got %s", received.Metadata["key2"])
	}
}
