package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aquamarinepk/aqm/log"
	"github.com/aquamarinepk/aqm/pubsub"
	"github.com/nats-io/nats.go"
)

// Config holds NATS pubsub configuration.
type Config struct {
	URL            string
	MaxReconnect   int
	ReconnectWait  time.Duration
	ConnectTimeout time.Duration
}

// DefaultConfig returns sensible defaults for NATS pubsub.
func DefaultConfig() Config {
	return Config{
		URL:            "nats://localhost:4222",
		MaxReconnect:   60,
		ReconnectWait:  time.Second,
		ConnectTimeout: 5 * time.Second,
	}
}

// subscription holds a registered NATS subscription.
type subscription struct {
	sub     *nats.Subscription
	handler pubsub.Handler
}

// Broker implements pubsub.Broker using NATS.
// NATS provides native fan-out: each subscriber receives all messages.
type Broker struct {
	conn          *nats.Conn
	cfg           Config
	log           log.Logger
	mu            sync.RWMutex
	subscriptions map[string]*subscription
	closed        bool
}

// NewBroker creates a new NATS-backed pubsub broker.
func NewBroker(cfg Config, log log.Logger) *Broker {
	return &Broker{
		cfg:           cfg,
		log:           log.With("component", "pubsub"),
		subscriptions: make(map[string]*subscription),
	}
}

// Start connects to the NATS server.
// Implements app.Startable interface.
func (b *Broker) Start(ctx context.Context) error {
	opts := []nats.Option{
		nats.MaxReconnects(b.cfg.MaxReconnect),
		nats.ReconnectWait(b.cfg.ReconnectWait),
		nats.Timeout(b.cfg.ConnectTimeout),
		nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
			if err != nil {
				b.log.Errorf("NATS disconnected: %v", err)
			}
		}),
		nats.ReconnectHandler(func(nc *nats.Conn) {
			b.log.Info("NATS reconnected")
		}),
		nats.ClosedHandler(func(nc *nats.Conn) {
			b.log.Info("NATS connection closed")
		}),
	}

	conn, err := nats.Connect(b.cfg.URL, opts...)
	if err != nil {
		return fmt.Errorf("cannot connect to NATS: %w", err)
	}

	b.conn = conn
	b.log.Infof("Connected to NATS at %s", b.cfg.URL)
	return nil
}

// Stop gracefully shuts down all subscriptions and closes the connection.
// Implements app.Stoppable interface.
func (b *Broker) Stop(ctx context.Context) error {
	return b.Close()
}

// Publish sends a message to the specified topic.
func (b *Broker) Publish(ctx context.Context, topic string, env pubsub.Envelope) error {
	b.mu.RLock()
	if b.closed {
		b.mu.RUnlock()
		return fmt.Errorf("broker is closed")
	}
	conn := b.conn
	b.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("broker not connected")
	}

	data, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("cannot marshal envelope: %w", err)
	}

	if err := conn.Publish(topic, data); err != nil {
		return fmt.Errorf("cannot publish message: %w", err)
	}

	b.log.Debugf("Published message %s to topic %s", env.ID, topic)
	return nil
}

// Subscribe registers a handler for the given topic.
// NATS provides native fan-out: each subscriber receives all messages.
// SubscriberID is used as the subscription identifier for management.
func (b *Broker) Subscribe(ctx context.Context, topic string, handler pubsub.Handler, opts pubsub.SubscribeOptions) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return fmt.Errorf("broker is closed")
	}

	if b.conn == nil {
		return fmt.Errorf("broker not connected")
	}

	subscriberID := opts.SubscriberID
	if subscriberID == "" {
		subscriberID = fmt.Sprintf("%s-%d", topic, time.Now().UnixNano())
	}

	if _, exists := b.subscriptions[subscriberID]; exists {
		return fmt.Errorf("subscriber %s already registered", subscriberID)
	}

	sub, err := b.conn.Subscribe(topic, func(msg *nats.Msg) {
		var env pubsub.Envelope
		if err := json.Unmarshal(msg.Data, &env); err != nil {
			b.log.Errorf("Cannot unmarshal message on topic %s: %v", topic, err)
			return
		}

		if err := handler(context.Background(), env); err != nil {
			b.log.Errorf("Handler error for message %s: %v", env.ID, err)
		}
	})
	if err != nil {
		return fmt.Errorf("cannot subscribe to topic %s: %w", topic, err)
	}

	b.subscriptions[subscriberID] = &subscription{
		sub:     sub,
		handler: handler,
	}

	b.log.Infof("Subscriber %s registered for topic %s", subscriberID, topic)
	return nil
}

// Close stops all subscriptions and closes the NATS connection.
func (b *Broker) Close() error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.closed {
		return nil
	}
	b.closed = true

	for id, sub := range b.subscriptions {
		if err := sub.sub.Unsubscribe(); err != nil {
			b.log.Errorf("Cannot unsubscribe %s: %v", id, err)
		}
	}
	b.subscriptions = make(map[string]*subscription)

	if b.conn != nil {
		b.conn.Close()
		b.conn = nil
	}

	b.log.Info("PubSub broker closed")
	return nil
}
