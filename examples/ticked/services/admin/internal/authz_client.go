package internal

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Permissions []string  `json:"permissions"`
}

type AuthZClient struct {
	client *httpclient.Client
}

func NewAuthZClient(client *httpclient.Client) *AuthZClient {
	return &AuthZClient{
		client: client,
	}
}

func (c *AuthZClient) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*Role, error) {
	path := fmt.Sprintf("/grants/user/%s", userID.String())
	resp, err := c.client.Get(ctx, path)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	if !resp.IsSuccess() {
		return nil, fmt.Errorf("authz returned status %d", resp.StatusCode)
	}

	var result struct {
		Data []*Role `json:"data"`
	}

	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return result.Data, nil
}
