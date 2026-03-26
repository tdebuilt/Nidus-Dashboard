package uptimekuma

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Client communicates with the Uptime Kuma API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates an Uptime Kuma API client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// GetStatusPage returns the status page configuration and monitor list.
func (c *Client) GetStatusPage(slug string) (*StatusPageResponse, error) {
	var resp StatusPageResponse
	if err := c.get("/api/status-page/"+slug, &resp); err != nil {
		return nil, fmt.Errorf("getting status page: %w", err)
	}
	return &resp, nil
}

// GetHeartbeats returns heartbeats and uptime for all monitors on a status page.
func (c *Client) GetHeartbeats(slug string) (*HeartbeatResponse, error) {
	var resp HeartbeatResponse
	if err := c.get("/api/status-page/heartbeat/"+slug, &resp); err != nil {
		return nil, fmt.Errorf("getting heartbeats: %w", err)
	}
	return &resp, nil
}

// GetMonitors returns the combined monitor overview for a status page.
func (c *Client) GetMonitors(slug string) (*MonitorsOverview, error) {
	statusPage, err := c.GetStatusPage(slug)
	if err != nil {
		return nil, err
	}

	heartbeats, err := c.GetHeartbeats(slug)
	if err != nil {
		return nil, err
	}

	// Build a map of monitor ID → summary from the status page
	monitorMap := make(map[int]MonitorSummary)
	for _, group := range statusPage.PublicGroupList {
		for _, m := range group.MonitorList {
			monitorMap[m.ID] = m
		}
	}

	var monitors []MonitorInfo
	totalUp := 0
	totalDown := 0

	for _, group := range statusPage.PublicGroupList {
		for _, m := range group.MonitorList {
			info := MonitorInfo{
				ID:   m.ID,
				Name: m.Name,
				Type: m.Type,
			}

			// Get latest heartbeat
			idStr := strconv.Itoa(m.ID)
			if beats, ok := heartbeats.HeartbeatList[idStr]; ok && len(beats) > 0 {
				latest := beats[len(beats)-1]
				info.Status = latest.Status
				info.Latency = latest.Ping
				info.Message = latest.Msg
			}

			// Get 24h uptime
			uptimeKey := fmt.Sprintf("%d_24", m.ID)
			if uptime, ok := heartbeats.UptimeList[uptimeKey]; ok {
				info.Uptime24h = uptime
			}

			switch info.Status {
			case 1:
				totalUp++
			case 0:
				totalDown++
			}

			monitors = append(monitors, info)
		}
	}

	return &MonitorsOverview{
		Monitors:   monitors,
		TotalUp:    totalUp,
		TotalDown:  totalDown,
		TotalCount: len(monitors),
		StatusPage: slug,
	}, nil
}

func (c *Client) get(path string, result any) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result any) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

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
