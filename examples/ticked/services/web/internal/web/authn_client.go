package web

import (
	"context"
	"fmt"
	"time"

	"github.com/aquamarinepk/aqm/httpclient"
	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type SignInRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type SignInResponse struct {
	User  *User  `json:"user"`
	Token string `json:"token"`
}

type AuthNClient struct {
	client *httpclient.Client
}

func NewAuthNClient(client *httpclient.Client) *AuthNClient {
	return &AuthNClient{
		client: client,
	}
}

func (c *AuthNClient) SignIn(ctx context.Context, email, password string) (*User, string, error) {
	req := SignInRequest{
		Email:    email,
		Password: password,
	}

	resp, err := c.client.Post(ctx, "/auth/signin", req)
	if err != nil {
		return nil, "", fmt.Errorf("signin request: %w", err)
	}

	if resp.StatusCode == 401 {
		return nil, "", fmt.Errorf("invalid credentials")
	}

	if resp.StatusCode >= 400 {
		return nil, "", fmt.Errorf("signin failed: status %d", resp.StatusCode)
	}

	var result SignInResponse
	if err := resp.JSON(&result); err != nil {
		return nil, "", fmt.Errorf("parse signin response: %w", err)
	}

	return result.User, result.Token, nil
}
