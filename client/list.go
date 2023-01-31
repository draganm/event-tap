package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/draganm/event-tap/data"
)

func (c *Client) List(ctx context.Context) ([]data.TapListEntry, error) {

	entries := []data.TapListEntry{}
	cursor := ""
	for {
		page, err := c.getListPage(ctx, cursor)
		if err != nil {
			return nil, fmt.Errorf("could not get list page: %w", err)
		}
		entries = append(entries, page.Entries...)
		if page.Cursor == "" {
			break
		}

		cursor = page.Cursor

	}

	return entries, nil

}

func (c *Client) getListPage(ctx context.Context, cursor string) (*data.TapListPage, error) {

	u := *c.tapsURL

	tu := &u
	q := tu.Query()
	q.Set("cursor", cursor)
	tu.RawPath = q.Encode()

	req, err := http.NewRequest("GET", tu.String(), nil)

	if err != nil {
		return nil, fmt.Errorf("could not create GET request: %w", err)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not perform GET request: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		rd, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("unexpected status %s: %s", res.Status, string(rd))
	}

	resObj := data.TapListPage{}

	err = json.NewDecoder(res.Body).Decode(&resObj)
	if err != nil {
		return nil, fmt.Errorf("could nod unmarshal response object: %w", err)
	}

	return &resObj, nil
}
