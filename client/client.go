package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/draganm/event-tap/data"
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

type CreateTapOptions data.TapOptions

type createTapResponse data.TapID

func (c *Client) CreateTap(ctx context.Context, options CreateTapOptions) (string, error) {
	d, err := json.Marshal(options)
	if err != nil {
		return "", fmt.Errorf("could not marshal options: %w", err)
	}

	req, err := http.NewRequest("POST", c.tapsURL.String(), bytes.NewReader(d))

	if err != nil {
		return "", fmt.Errorf("could not create POST request: %w", err)
	}

	req.Header.Set("content-type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not perform POST request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusCreated {
		rd, _ := io.ReadAll(res.Body)
		return "", fmt.Errorf("unexpected status %s: %s", res.Status, string(rd))
	}

	resObj := createTapResponse{}

	err = json.NewDecoder(res.Body).Decode(&resObj)
	if err != nil {
		return "", fmt.Errorf("could nod unmarshal response object: %w", err)
	}

	return resObj.ID, nil

}
