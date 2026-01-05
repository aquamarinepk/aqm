package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aquamarinepk/aqm/log"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	retryMax   int
	retryDelay time.Duration
	log        log.Logger
}

type Response struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

func New(baseURL string, logger log.Logger, opts ...Option) *Client {
	c := &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		retryMax:   3,
		retryDelay: 100 * time.Millisecond,
		log:        logger,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) Get(ctx context.Context, path string) (*Response, error) {
	return c.Do(ctx, http.MethodGet, path, nil)
}

func (c *Client) Post(ctx context.Context, path string, body interface{}) (*Response, error) {
	return c.Do(ctx, http.MethodPost, path, body)
}

func (c *Client) Do(ctx context.Context, method, path string, body interface{}) (*Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryMax; attempt++ {
		if attempt > 0 {
			delay := c.retryDelay * time.Duration(1<<uint(attempt-1))
			c.log.Debugf("Retrying request after %v (attempt %d/%d)", delay, attempt, c.retryMax)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return nil, ctx.Err()
			}

			if body != nil {
				jsonBody, _ := json.Marshal(body)
				bodyReader = bytes.NewReader(jsonBody)
			}
		}

		req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		c.log.Debugf("HTTP %s %s (attempt %d/%d)", method, url, attempt+1, c.retryMax+1)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if !isRetryable(err) {
				return nil, fmt.Errorf("request failed: %w", err)
			}
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("server error: %d", resp.StatusCode)
			if attempt < c.retryMax {
				continue
			}
			return nil, lastErr
		}

		c.log.Debugf("HTTP %s %s -> %d", method, url, resp.StatusCode)

		return &Response{
			StatusCode: resp.StatusCode,
			Body:       respBody,
			Headers:    resp.Header,
		}, nil
	}

	return nil, fmt.Errorf("request failed after %d attempts: %w", c.retryMax+1, lastErr)
}

func isRetryable(err error) bool {
	return true
}
