package client

import (
	"context"
	"fmt"
	"net/url"
)

type Client struct {
	tapsURL *url.URL
}

func New(baseURL string) (*Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("could not parse base URL: %w", err)
	}

	return &Client{tapsURL: u.JoinPath("taps")}, nil
}

type contextKeyType string

const contextKey contextKeyType = "tapClient"

func ContextWithClient(ctx context.Context, client *Client) context.Context {
	return context.WithValue(ctx, contextKey, client)
}

func FromContext(ctx context.Context) *Client {
	return ctx.Value(contextKey).(*Client)
}
