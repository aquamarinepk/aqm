package httpclient

import (
	"net/http"
	"time"
)

type Option func(*Client)

func WithRetryMax(max int) Option {
	return func(c *Client) {
		if max >= 0 {
			c.retryMax = max
		}
	}
}

func WithRetryDelay(delay time.Duration) Option {
	return func(c *Client) {
		if delay > 0 {
			c.retryDelay = delay
		}
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if timeout > 0 {
			c.httpClient.Timeout = timeout
		}
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		if client != nil {
			c.httpClient = client
		}
	}
}
