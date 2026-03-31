package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type deleteItemsResponse struct {
	Success bool `json:"success"`
	Result  struct {
		OperationID string `json:"operation_id"`
	} `json:"result"`
}

type deleteItemRequest struct {
	ID string `json:"id"`
}

func (c *Client) deleteItems(ctx context.Context, itemIDs []string) (string, error) {
	if len(itemIDs) == 0 {
		return "", nil
	}

	items := make([]deleteItemRequest, 0, len(itemIDs))
	seen := make(map[string]struct{}, len(itemIDs))

	for _, id := range itemIDs {
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}

		items = append(items, deleteItemRequest{
			ID: id,
		})
	}

	if len(items) == 0 {
		return "", nil
	}

	path := fmt.Sprintf(
		"/accounts/%s/rules/lists/%s/items",
		c.accountID,
		c.listID,
	)

	var resp deleteItemsResponse
	if err := c.doJSON(ctx, http.MethodDelete, path, map[string]any{
		"items": items,
	}, &resp); err != nil {
		return "", err
	}

	return resp.Result.OperationID, nil
}

func (c *Client) DeleteItemsAndWait(ctx context.Context, itemIDs []string, interval time.Duration) error {
	operationID, err := c.deleteItems(ctx, itemIDs)
	if err != nil {
		return err
	}

	if operationID == "" {
		return nil
	}

	return c.waitForBulkOperation(ctx, operationID, interval)
}
