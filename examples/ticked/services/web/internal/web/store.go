package web

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/aquamarinepk/aqm/log"
	"github.com/google/uuid"
)

var (
	ErrListNotFound = errors.New("todo list not found")
)

// TodoList represents a user's todo list aggregate (matches ticked API response).
type TodoList struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	Items     []TodoItem `json:"items"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// TodoItem represents an item within a todo list (matches ticked API response).
type TodoItem struct {
	ID          uuid.UUID  `json:"id"`
	Text        string     `json:"text"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// TodoListStore abstracts persistence operations for todo lists.
// This interface allows for swappable implementations (HTTP, gRPC, Fake for tests).
type TodoListStore interface {
	Get(ctx context.Context, userID uuid.UUID) (*TodoList, error)
	AddItem(ctx context.Context, userID uuid.UUID, text string) (*TodoList, error)
	UpdateItem(ctx context.Context, userID, itemID uuid.UUID, text *string, completed *bool) (*TodoList, error)
	RemoveItem(ctx context.Context, userID, itemID uuid.UUID) (*TodoList, error)
}

// HTTPTodoListStore implements TodoListStore using HTTP calls to ticked API.
// This wraps httpclient.Client which provides automatic retry with exponential backoff.
type HTTPTodoListStore struct {
	client *httpclient.Client
	log    log.Logger
}

// NewHTTPTodoListStore creates a new HTTP-backed todo list store.
func NewHTTPTodoListStore(tickedURL string, logger log.Logger) TodoListStore {
	client := httpclient.New(
		tickedURL,
		logger,
		httpclient.WithRetryMax(3),
		httpclient.WithRetryDelay(100*time.Millisecond),
		httpclient.WithTimeout(30*time.Second),
	)

	return &HTTPTodoListStore{
		client: client,
		log:    logger,
	}
}

// Get retrieves a user's todo list.
func (s *HTTPTodoListStore) Get(ctx context.Context, userID uuid.UUID) (*TodoList, error) {
	path := fmt.Sprintf("/users/%s/list", userID)

	resp, err := s.client.Get(ctx, path)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == 404 {
		return nil, ErrListNotFound
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("ticked API error: status %d", resp.StatusCode)
	}

	var list TodoList
	if err := resp.JSON(&list); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &list, nil
}

// AddItem adds a new item to the user's list.
func (s *HTTPTodoListStore) AddItem(ctx context.Context, userID uuid.UUID, text string) (*TodoList, error) {
	path := fmt.Sprintf("/users/%s/list/items", userID)

	payload := map[string]string{
		"text": text,
	}

	resp, err := s.client.Post(ctx, path, payload)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("ticked API error: status %d", resp.StatusCode)
	}

	var list TodoList
	if err := resp.JSON(&list); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &list, nil
}

// UpdateItem updates an existing item in the user's list.
func (s *HTTPTodoListStore) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, text *string, completed *bool) (*TodoList, error) {
	path := fmt.Sprintf("/users/%s/list/items/%s", userID, itemID)

	payload := make(map[string]interface{})
	if text != nil {
		payload["text"] = *text
	}
	if completed != nil {
		payload["completed"] = *completed
	}

	resp, err := s.client.Do(ctx, "PATCH", path, payload)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("ticked API error: status %d", resp.StatusCode)
	}

	var list TodoList
	if err := resp.JSON(&list); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &list, nil
}

// RemoveItem removes an item from the user's list.
func (s *HTTPTodoListStore) RemoveItem(ctx context.Context, userID, itemID uuid.UUID) (*TodoList, error) {
	path := fmt.Sprintf("/users/%s/list/items/%s", userID, itemID)

	resp, err := s.client.Do(ctx, "DELETE", path, nil)
	if err != nil {
		return nil, err
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("ticked API error: status %d", resp.StatusCode)
	}

	var list TodoList
	if err := resp.JSON(&list); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &list, nil
}
