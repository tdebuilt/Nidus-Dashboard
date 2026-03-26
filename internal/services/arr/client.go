package arr

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client communicates with a *arr service API (Sonarr, Radarr, Lidarr, Prowlarr).
type Client struct {
	baseURL    string
	apiKey     string
	apiVersion string
	httpClient *http.Client
}

// NewClient creates a new *arr API client.
func NewClient(baseURL, apiKey, apiVersion string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		apiVersion: apiVersion,
		httpClient: httpClient,
	}
}

// GetSystemStatus returns version and instance information.
func (c *Client) GetSystemStatus() (*SystemStatus, error) {
	var status SystemStatus
	if err := c.get("/system/status", &status); err != nil {
		return nil, fmt.Errorf("getting system status: %w", err)
	}
	return &status, nil
}

// GetQueue returns the download queue with pagination.
func (c *Client) GetQueue(pageSize int) (*QueueResponse, error) {
	path := fmt.Sprintf("/queue?pageSize=%d&sortDirection=descending&sortKey=progress", pageSize)
	var queue QueueResponse
	if err := c.get(path, &queue); err != nil {
		return nil, fmt.Errorf("getting queue: %w", err)
	}
	return &queue, nil
}

// GetCalendar returns upcoming media items for the next 7 days.
func (c *Client) GetCalendar(start, end time.Time) ([]CalendarItem, error) {
	path := fmt.Sprintf("/calendar?start=%s&end=%s",
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
	)
	var items []CalendarItem
	if err := c.get(path, &items); err != nil {
		return nil, fmt.Errorf("getting calendar: %w", err)
	}
	return items, nil
}

// GetLibraryCount returns the number of items in the library.
// It decodes the response as a raw JSON array to avoid defining full model structs.
func (c *Client) GetLibraryCount(libraryPath string) (int, error) {
	var items []json.RawMessage
	if err := c.get(libraryPath, &items); err != nil {
		return 0, fmt.Errorf("getting library count: %w", err)
	}
	return len(items), nil
}

func (c *Client) get(path string, result any) error {
	url := fmt.Sprintf("%s/api/%s%s", c.baseURL, c.apiVersion, path)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-Api-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("unauthorized: invalid API key")
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
