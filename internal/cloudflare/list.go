package cloudflare

import (
	"context"
	"fmt"
	"net/http"
)

type ListItem struct {
	ID string `json:"id"`
	IP string `json:"ip"`
}

type listItemsResponse struct {
	Success    bool       `json:"success"`
	Result     []ListItem `json:"result"`
	ResultInfo struct {
		Cursors struct {
			After string `json:"after"`
		} `json:"cursors"`
	} `json:"result_info"`
}

func (c *Client) listItems(ctx context.Context, cursor string) ([]ListItem, string, error) {
	path := fmt.Sprintf(
		"/accounts/%s/rules/lists/%s/items?per_page=500",
		c.accountID,
		c.listID,
	)

	if cursor != "" {
		path += "&cursor=" + cursor
	}

	var resp listItemsResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return nil, "", err
	}

	return resp.Result, resp.ResultInfo.Cursors.After, nil
}

func (c *Client) ListAllItems(ctx context.Context) ([]ListItem, error) {
	var allItems []ListItem
	var cursor string

	for {
		items, nextCursor, err := c.listItems(ctx, cursor)
		if err != nil {
			return allItems, err
		}

		allItems = append(allItems, items...)

		if nextCursor == "" {
			break
		}

		cursor = nextCursor
	}

	return allItems, nil
}
