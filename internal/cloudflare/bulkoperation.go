package cloudflare

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type BulkOperation struct {
	ID        string `json:"id"`
	Status    string `json:"status"`
	Completed string `json:"completed,omitempty"`
	Error     string `json:"error,omitempty"`
}

type getBulkOperationResponse struct {
	Success bool          `json:"success"`
	Result  BulkOperation `json:"result"`
}

func (c *Client) getBulkOperation(ctx context.Context, operationID string) (BulkOperation, error) {
	path := fmt.Sprintf(
		"/accounts/%s/rules/lists/bulk_operations/%s",
		c.accountID,
		operationID,
	)

	var resp getBulkOperationResponse
	if err := c.doJSON(ctx, http.MethodGet, path, nil, &resp); err != nil {
		return BulkOperation{}, err
	}

	return resp.Result, nil
}

func (c *Client) waitForBulkOperation(ctx context.Context, operationID string, interval time.Duration) error {
	if interval <= 0 {
		interval = 2 * time.Second
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		op, err := c.getBulkOperation(ctx, operationID)
		if err != nil {
			return err
		}

		switch op.Status {
		case "completed":
			return nil

		case "failed":
			if op.Error != "" {
				return fmt.Errorf("cloudflare bulk operation failed: %s", op.Error)
			}
			return fmt.Errorf("cloudflare bulk operation failed")

		case "pending", "running":
			// weiter warten

		default:
			return fmt.Errorf("unknown cloudflare bulk operation status: %s", op.Status)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}
	}
}
