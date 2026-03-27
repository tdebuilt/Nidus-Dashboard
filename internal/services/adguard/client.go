package adguard

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client communicates with the AdGuard Home API.
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
}

// NewClient creates an AdGuard Home API client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// SetCredentials sets the HTTP Basic Auth credentials.
func (c *Client) SetCredentials(username, password string) {
	c.username = username
	c.password = password
}

// GetStats returns the DNS query statistics.
func (c *Client) GetStats(ctx context.Context) (*Stats, error) {
	var stats Stats
	if err := c.get(ctx, "/control/stats", &stats); err != nil {
		return nil, fmt.Errorf("getting stats: %w", err)
	}
	return &stats, nil
}

// GetFilteringStatus returns the current filtering status.
func (c *Client) GetFilteringStatus(ctx context.Context) (*FilteringStatus, error) {
	var status FilteringStatus
	if err := c.get(ctx, "/control/filtering/status", &status); err != nil {
		return nil, fmt.Errorf("getting filtering status: %w", err)
	}
	return &status, nil
}

// SetFilteringEnabled enables or disables DNS filtering.
func (c *Client) SetFilteringEnabled(ctx context.Context, enabled bool) error {
	body := map[string]interface{}{
		"enabled":  enabled,
		"interval": 0,
	}
	if err := c.post(ctx, "/control/filtering/config", body, nil); err != nil {
		return fmt.Errorf("setting filtering: %w", err)
	}
	return nil
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	return c.doRequest(req, result)
}

func (c *Client) post(ctx context.Context, path string, body any, result any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		reader = strings.NewReader(string(data))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, reader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result any) error {
	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("unauthorized: invalid credentials")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}
