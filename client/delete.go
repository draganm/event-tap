package client

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

func (c *Client) Delete(ctx context.Context, id string) error {

	tapURL := c.tapsURL.JoinPath(id)

	req, err := http.NewRequest("DELETE", tapURL.String(), nil)

	if err != nil {
		return fmt.Errorf("could not create DELETE request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not perform DELETE request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		rd, _ := io.ReadAll(res.Body)
		return fmt.Errorf("unexpected status %s: %s", res.Status, string(rd))
	}

	return nil

}
