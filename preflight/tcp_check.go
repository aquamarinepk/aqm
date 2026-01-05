package preflight

import (
	"context"
	"fmt"
	"net"
	"time"
)

type tcpCheck struct {
	name    string
	address string
	timeout time.Duration
}

func TCPCheck(name, address string) Check {
	return &tcpCheck{
		name:    name,
		address: address,
		timeout: 5 * time.Second,
	}
}

func (t *tcpCheck) Name() string {
	return t.name
}

func (t *tcpCheck) Run(ctx context.Context) error {
	dialer := net.Dialer{
		Timeout: t.timeout,
	}

	conn, err := dialer.DialContext(ctx, "tcp", t.address)
	if err != nil {
		return fmt.Errorf("TCP connection failed: %w", err)
	}
	defer conn.Close()

	return nil
}
