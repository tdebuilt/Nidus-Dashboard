package grafana

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client communicates with the Grafana HTTP API.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a Grafana API client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// SetToken sets the Service Account Token for authentication.
func (c *Client) SetToken(apiKey string) {
	c.apiKey = apiKey
}

// SearchDashboards returns all dashboards.
func (c *Client) SearchDashboards(ctx context.Context) ([]DashboardSearchResult, error) {
	var results []DashboardSearchResult
	if err := c.get(ctx, "/api/search?type=dash-db", &results); err != nil {
		return nil, fmt.Errorf("searching dashboards: %w", err)
	}
	return results, nil
}

// GetDashboard returns the full dashboard detail by UID.
func (c *Client) GetDashboard(ctx context.Context, uid string) (*DashboardDetail, error) {
	var detail DashboardDetail
	if err := c.get(ctx, "/api/dashboards/uid/"+uid, &detail); err != nil {
		return nil, fmt.Errorf("getting dashboard: %w", err)
	}
	return &detail, nil
}

// GetHealth checks the Grafana server health.
func (c *Client) GetHealth(ctx context.Context) (*HealthResponse, error) {
	var health HealthResponse
	if err := c.get(ctx, "/api/health", &health); err != nil {
		return nil, fmt.Errorf("checking health: %w", err)
	}
	return &health, nil
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result any) error {
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: invalid or expired token")
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
