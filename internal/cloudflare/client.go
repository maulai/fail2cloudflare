package cloudflare

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiToken   string
	accountID  string
	listID     string
}

type Config struct {
	APIToken  string
	AccountID string
	ListID    string
	BaseURL   string
	Timeout   time.Duration
}

func New(cfg Config) *Client {
	baseURL := cfg.BaseURL
	if strings.TrimSpace(baseURL) == "" {
		baseURL = "https://api.cloudflare.com/client/v4"
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: timeout,
		},
		baseURL:   strings.TrimRight(baseURL, "/"),
		apiToken:  cfg.APIToken,
		accountID: cfg.AccountID,
		listID:    cfg.ListID,
	}
}

func (c *Client) newRequest(ctx context.Context, method string, path string, body any) (*http.Request, error) {
	var bodyReader *bytes.Reader

	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(raw)
	} else {
		bodyReader = bytes.NewReader(nil)
	}

	url := c.baseURL + "/" + strings.TrimLeft(path, "/")

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Accept", "application/json")

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return req, nil
}

func (c *Client) doJSON(ctx context.Context, method string, path string, reqBody any, respBody any) error {
	req, err := c.newRequest(ctx, method, path, reqBody)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		return fmt.Errorf("cloudflare api returned status %d: %s", resp.StatusCode, string(raw))
	}

	if respBody == nil {
		io.Copy(io.Discard, resp.Body)
		return nil
	}

	if err := json.NewDecoder(resp.Body).Decode(respBody); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}

	return nil
}
