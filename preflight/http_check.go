package preflight

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type httpCheck struct {
	name   string
	url    string
	client *http.Client
}

func HTTPCheck(name, url string) Check {
	return &httpCheck{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

func (h *httpCheck) Name() string {
	return h.name
}

func (h *httpCheck) Run(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, h.url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
