package uptimekuma

import (
	"context"
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
func (c *Client) GetStatusPage(ctx context.Context, slug string) (*StatusPageResponse, error) {
	var resp StatusPageResponse
	if err := c.get(ctx, "/api/status-page/"+slug, &resp); err != nil {
		return nil, fmt.Errorf("getting status page: %w", err)
	}
	return &resp, nil
}

// GetHeartbeats returns heartbeats and uptime for all monitors on a status page.
func (c *Client) GetHeartbeats(ctx context.Context, slug string) (*HeartbeatResponse, error) {
	var resp HeartbeatResponse
	if err := c.get(ctx, "/api/status-page/heartbeat/"+slug, &resp); err != nil {
		return nil, fmt.Errorf("getting heartbeats: %w", err)
	}
	return &resp, nil
}

// GetMonitors returns the combined monitor overview for a status page.
func (c *Client) GetMonitors(ctx context.Context, slug string) (*MonitorsOverview, error) {
	statusPage, err := c.GetStatusPage(ctx, slug)
	if err != nil {
		return nil, err
	}

	heartbeats, err := c.GetHeartbeats(ctx, slug)
	if err != nil {
		return nil, err
	}

	monitors, totalUp, totalDown := buildMonitorInfos(statusPage, heartbeats)

	return &MonitorsOverview{
		Monitors:   monitors,
		TotalUp:    totalUp,
		TotalDown:  totalDown,
		TotalCount: len(monitors),
		StatusPage: slug,
	}, nil
}

// buildMonitorInfos aggregates status page groups with heartbeat data
// and returns monitor details with up/down counts.
func buildMonitorInfos(statusPage *StatusPageResponse, heartbeats *HeartbeatResponse) ([]MonitorInfo, int, int) {
	var monitors []MonitorInfo
	totalUp := 0
	totalDown := 0

	for _, group := range statusPage.PublicGroupList {
		for _, m := range group.MonitorList {
			info := buildMonitorInfo(m, heartbeats)

			switch info.Status {
			case 1:
				totalUp++
			case 0:
				totalDown++
			}

			monitors = append(monitors, info)
		}
	}
	return monitors, totalUp, totalDown
}

// buildMonitorInfo enriches a single monitor entry with heartbeat and uptime data.
func buildMonitorInfo(m MonitorSummary, heartbeats *HeartbeatResponse) MonitorInfo {
	info := MonitorInfo{
		ID:   m.ID,
		Name: m.Name,
		Type: m.Type,
	}

	idStr := strconv.Itoa(m.ID)
	if beats, ok := heartbeats.HeartbeatList[idStr]; ok && len(beats) > 0 {
		latest := beats[len(beats)-1]
		info.Status = latest.Status
		info.Latency = latest.Ping
		info.Message = latest.Msg
	}

	uptimeKey := fmt.Sprintf("%d_24", m.ID)
	if uptime, ok := heartbeats.UptimeList[uptimeKey]; ok {
		info.Uptime24h = uptime
	}

	return info
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
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
