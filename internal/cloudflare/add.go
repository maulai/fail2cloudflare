package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type createItemsResponse struct {
	Success bool `json:"success"`
	Result  struct {
		OperationID string `json:"operation_id"`
	} `json:"result"`
}

type CreateItemRequest struct {
	IP      string `json:"ip"`
	Comment string `json:"comment"`
}

func (c *Client) addItems(ctx context.Context, items []CreateItemRequest) (string, error) {
	if len(items) == 0 {
		return "", nil
	}

	dedupedItems := make([]CreateItemRequest, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		if item.IP == "" {
			continue
		}
		if _, ok := seen[item.IP]; ok {
			continue
		}
		seen[item.IP] = struct{}{}
		dedupedItems = append(dedupedItems, item)
	}
	if len(dedupedItems) == 0 {
		return "", nil
	}

	path := fmt.Sprintf(
		"/accounts/%s/rules/lists/%s/items",
		c.accountID,
		c.listID,
	)

	var resp createItemsResponse
	if err := c.doJSON(ctx, http.MethodPost, path, dedupedItems, &resp); err != nil {
		return "", err
	}

	return resp.Result.OperationID, nil
}

func (c *Client) AddIPsAndWait(ctx context.Context, items []CreateItemRequest, interval time.Duration) error {
	operationID, err := c.addItems(ctx, items)
	if err != nil {
		return err
	}

	if operationID == "" {
		return nil
	}

	return c.waitForBulkOperation(ctx, operationID, interval)
}
