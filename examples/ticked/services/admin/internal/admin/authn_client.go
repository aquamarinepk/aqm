package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AuthNClient struct {
	client *httpclient.Client
}

func NewAuthNClient(client *httpclient.Client) *AuthNClient {
	return &AuthNClient{
		client: client,
	}
}

func (c *AuthNClient) ListUsers(ctx context.Context) ([]*User, error) {
	resp, err := c.client.Get(ctx, "/users")
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("authn returned status %d", resp.StatusCode)
	}

	var result struct {
		Data []*User `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Data, nil
}

func (c *AuthNClient) GetUser(ctx context.Context, userID uuid.UUID) (*User, error) {
	path := fmt.Sprintf("/users/%s", userID.String())
	resp, err := c.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if resp.StatusCode == 404 {
		return nil, fmt.Errorf("user not found")
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("authn returned status %d", resp.StatusCode)
	}

	var result struct {
		Data *User `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Data, nil
}
