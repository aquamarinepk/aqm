package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aquamarinepk/aqm/log"
)

func TestClientGet(t *testing.T) {
	tests := []struct {
		name           string
		serverHandler  http.HandlerFunc
		wantStatusCode int
		wantBody       string
		wantErr        bool
	}{
		{
			name: "successful GET request",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"status":"ok"}`))
			},
			wantStatusCode: http.StatusOK,
			wantBody:       `{"status":"ok"}`,
			wantErr:        false,
		},
		{
			name: "404 not found",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":"not found"}`))
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `{"error":"not found"}`,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			logger := log.NewLogger("error")
			client := New(server.URL, logger)

			resp, err := client.Get(context.Background(), "/test")

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if resp.StatusCode != tt.wantStatusCode {
					t.Errorf("StatusCode = %d, want %d", resp.StatusCode, tt.wantStatusCode)
				}
				if string(resp.Body) != tt.wantBody {
					t.Errorf("Body = %s, want %s", string(resp.Body), tt.wantBody)
				}
			}
		})
	}
}

func TestClientPost(t *testing.T) {
	tests := []struct {
		name           string
		serverHandler  http.HandlerFunc
		requestBody    interface{}
		wantStatusCode int
		wantErr        bool
	}{
		{
			name: "successful POST with JSON body",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}
				w.WriteHeader(http.StatusCreated)
				w.Write([]byte(`{"id":"123"}`))
			},
			requestBody:    map[string]string{"name": "test"},
			wantStatusCode: http.StatusCreated,
			wantErr:        false,
		},
		{
			name: "POST with nil body",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			requestBody:    nil,
			wantStatusCode: http.StatusOK,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.serverHandler)
			defer server.Close()

			logger := log.NewLogger("error")
			client := New(server.URL, logger)

			resp, err := client.Post(context.Background(), "/test", tt.requestBody)

			if (err != nil) != tt.wantErr {
				t.Errorf("Post() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err == nil {
				if resp.StatusCode != tt.wantStatusCode {
					t.Errorf("StatusCode = %d, want %d", resp.StatusCode, tt.wantStatusCode)
				}
			}
		})
	}
}

func TestClientRetry(t *testing.T) {
	tests := []struct {
		name        string
		serverFunc  func() *httptest.Server
		retryMax    int
		wantAttempts int
		wantErr     bool
	}{
		{
			name: "retry on 500 error",
			serverFunc: func() *httptest.Server {
				attempts := 0
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					attempts++
					if attempts < 3 {
						w.WriteHeader(http.StatusInternalServerError)
						return
					}
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(`{"status":"ok"}`))
				}))
			},
			retryMax:    3,
			wantErr:     false,
		},
		{
			name: "fail after max retries",
			serverFunc: func() *httptest.Server {
				return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
			},
			retryMax:    2,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := tt.serverFunc()
			defer server.Close()

			logger := log.NewLogger("error")
			client := New(server.URL, logger, WithRetryMax(tt.retryMax))

			_, err := client.Get(context.Background(), "/test")

			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	logger := log.NewLogger("error")
	client := New(server.URL, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.Get(ctx, "/test")

	if err == nil {
		t.Error("expected context deadline exceeded error, got nil")
	}
}

func TestClientOptions(t *testing.T) {
	tests := []struct {
		name    string
		opts    []Option
		check   func(*Client) bool
	}{
		{
			name: "WithRetryMax sets retry max",
			opts: []Option{WithRetryMax(5)},
			check: func(c *Client) bool {
				return c.retryMax == 5
			},
		},
		{
			name: "WithRetryDelay sets retry delay",
			opts: []Option{WithRetryDelay(500 * time.Millisecond)},
			check: func(c *Client) bool {
				return c.retryDelay == 500*time.Millisecond
			},
		},
		{
			name: "WithTimeout sets timeout",
			opts: []Option{WithTimeout(10 * time.Second)},
			check: func(c *Client) bool {
				return c.httpClient.Timeout == 10*time.Second
			},
		},
		{
			name: "WithHTTPClient sets custom client",
			opts: []Option{WithHTTPClient(&http.Client{Timeout: 5 * time.Second})},
			check: func(c *Client) bool {
				return c.httpClient.Timeout == 5*time.Second
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := log.NewLogger("error")
			client := New("http://localhost", logger, tt.opts...)

			if !tt.check(client) {
				t.Errorf("option not applied correctly")
			}
		})
	}
}

func TestResponseHelpers(t *testing.T) {
	tests := []struct {
		name       string
		response   *Response
		wantJSON   map[string]string
		wantString string
		wantSuccess bool
		wantError  bool
	}{
		{
			name: "successful JSON response",
			response: &Response{
				StatusCode: http.StatusOK,
				Body:       []byte(`{"key":"value"}`),
			},
			wantJSON:    map[string]string{"key": "value"},
			wantString:  `{"key":"value"}`,
			wantSuccess: true,
			wantError:   false,
		},
		{
			name: "error response",
			response: &Response{
				StatusCode: http.StatusBadRequest,
				Body:       []byte(`{"error":"bad request"}`),
			},
			wantString:  `{"error":"bad request"}`,
			wantSuccess: false,
			wantError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.response.String() != tt.wantString {
				t.Errorf("String() = %s, want %s", tt.response.String(), tt.wantString)
			}

			if tt.response.IsSuccess() != tt.wantSuccess {
				t.Errorf("IsSuccess() = %v, want %v", tt.response.IsSuccess(), tt.wantSuccess)
			}

			if tt.response.IsError() != tt.wantError {
				t.Errorf("IsError() = %v, want %v", tt.response.IsError(), tt.wantError)
			}

			if tt.wantJSON != nil {
				var result map[string]string
				if err := tt.response.JSON(&result); err != nil {
					t.Errorf("JSON() error = %v", err)
				}
				if result["key"] != tt.wantJSON["key"] {
					t.Errorf("JSON() key = %s, want %s", result["key"], tt.wantJSON["key"])
				}
			}
		})
	}
}

func TestResponseJSONError(t *testing.T) {
	response := &Response{
		StatusCode: http.StatusOK,
		Body:       []byte(`invalid json`),
	}

	var result map[string]string
	err := response.JSON(&result)

	if err == nil {
		t.Error("expected JSON() to return error for invalid JSON, got nil")
	}
}
