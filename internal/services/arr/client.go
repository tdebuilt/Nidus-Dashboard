package arr

import (
	"context"
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
func (c *Client) GetSystemStatus(ctx context.Context) (*SystemStatus, error) {
	var status SystemStatus
	if err := c.get(ctx, "/system/status", &status); err != nil {
		return nil, fmt.Errorf("getting system status: %w", err)
	}
	return &status, nil
}

// GetQueue returns the download queue with pagination.
func (c *Client) GetQueue(ctx context.Context, pageSize int) (*QueueResponse, error) {
	path := fmt.Sprintf("/queue?pageSize=%d&sortDirection=descending&sortKey=progress", pageSize)
	var queue QueueResponse
	if err := c.get(ctx, path, &queue); err != nil {
		return nil, fmt.Errorf("getting queue: %w", err)
	}
	return &queue, nil
}

// GetCalendar returns upcoming media items for the next 7 days.
func (c *Client) GetCalendar(ctx context.Context, start, end time.Time) ([]CalendarItem, error) {
	path := fmt.Sprintf("/calendar?start=%s&end=%s",
		start.Format(time.RFC3339),
		end.Format(time.RFC3339),
	)
	var items []CalendarItem
	if err := c.get(ctx, path, &items); err != nil {
		return nil, fmt.Errorf("getting calendar: %w", err)
	}
	return items, nil
}

// GetLibraryCount returns the number of items in the library.
// It decodes the response as a raw JSON array to avoid defining full model structs.
func (c *Client) GetLibraryCount(ctx context.Context, libraryPath string) (int, error) {
	var items []json.RawMessage
	if err := c.get(ctx, libraryPath, &items); err != nil {
		return 0, fmt.Errorf("getting library count: %w", err)
	}
	return len(items), nil
}

// GetRadarrLibrary returns all movies from Radarr.
func (c *Client) GetRadarrLibrary(ctx context.Context) ([]RadarrMovie, error) {
	var movies []RadarrMovie
	if err := c.get(ctx, "/movie", &movies); err != nil {
		return nil, fmt.Errorf("getting radarr library: %w", err)
	}
	return movies, nil
}

// GetSonarrLibrary returns all series from Sonarr.
func (c *Client) GetSonarrLibrary(ctx context.Context) ([]SonarrSeries, error) {
	var series []SonarrSeries
	if err := c.get(ctx, "/series", &series); err != nil {
		return nil, fmt.Errorf("getting sonarr library: %w", err)
	}
	return series, nil
}

// GetSonarrEpisodes returns episodes for a given series ID.
func (c *Client) GetSonarrEpisodes(ctx context.Context, seriesID int) ([]SonarrEpisode, error) {
	path := fmt.Sprintf("/episode?seriesId=%d", seriesID)
	var episodes []SonarrEpisode
	if err := c.get(ctx, path, &episodes); err != nil {
		return nil, fmt.Errorf("getting sonarr episodes: %w", err)
	}
	return episodes, nil
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	url := fmt.Sprintf("%s/api/%s%s", c.baseURL, c.apiVersion, path)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
