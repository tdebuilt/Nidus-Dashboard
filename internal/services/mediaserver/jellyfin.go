package mediaserver

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// JellyfinClient communicates with the Jellyfin API.
type JellyfinClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewJellyfinClient creates a Jellyfin API client.
func NewJellyfinClient(baseURL, apiKey string, httpClient *http.Client) *JellyfinClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &JellyfinClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

// GetSessions returns active streaming sessions.
func (c *JellyfinClient) GetSessions(ctx context.Context) ([]Session, error) {
	var jSessions []jellyfinSession
	if err := c.get(ctx, "/Sessions", &jSessions); err != nil {
		return nil, fmt.Errorf("getting sessions: %w", err)
	}

	sessions := make([]Session, 0)
	for _, js := range jSessions {
		// Only include sessions that are actively playing something
		if js.NowPlayingItem == nil {
			continue
		}

		np := js.NowPlayingItem
		s := Session{
			ID:        js.ID,
			UserName:  js.UserName,
			Title:     np.Name,
			MediaType: strings.ToLower(np.Type),
			Year:      np.ProductionYear,
			Player:    js.Client,
			Platform:  js.DeviceName,
			Duration:  ticksToSeconds(np.RunTimeTicks),
		}

		if js.PlayState != nil {
			s.Position = ticksToSeconds(js.PlayState.PositionTicks)
			if js.PlayState.IsPaused {
				s.State = "paused"
			} else {
				s.State = "playing"
			}
		}

		if np.RunTimeTicks > 0 && js.PlayState != nil {
			s.Progress = float64(js.PlayState.PositionTicks) / float64(np.RunTimeTicks)
		}

		// Build subtitle for TV episodes
		if np.Type == "Episode" && np.SeriesName != "" {
			s.Subtitle = fmt.Sprintf("%s — S%02dE%02d", np.SeriesName, np.ParentIndexNumber, np.IndexNumber)
		}

		// Build thumb path
		if np.ImageTags != nil {
			if _, ok := np.ImageTags["Primary"]; ok {
				s.ThumbPath = fmt.Sprintf("/Items/%s/Images/Primary", np.ID)
			}
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}

// GetLibraries returns the list of media libraries.
func (c *JellyfinClient) GetLibraries(ctx context.Context) ([]Library, error) {
	var jLibraries []jellyfinLibrary
	if err := c.get(ctx, "/Library/VirtualFolders", &jLibraries); err != nil {
		return nil, fmt.Errorf("getting libraries: %w", err)
	}

	libraries := make([]Library, 0, len(jLibraries))
	for _, jl := range jLibraries {
		libType := jl.CollectionType
		switch libType {
		case "movies":
			libType = "movie"
		case "tvshows":
			libType = "show"
		}

		// Get item count for each library
		count := 0
		var itemsResp jellyfinItemsResponse
		err := c.get(ctx, fmt.Sprintf("/Items?ParentId=%s&Limit=0&Recursive=true", jl.ItemID), &itemsResp)
		if err == nil {
			count = itemsResp.TotalRecordCount
		}

		libraries = append(libraries, Library{
			ID:        jl.ItemID,
			Name:      jl.Name,
			Type:      libType,
			ItemCount: count,
		})
	}

	return libraries, nil
}

// GetServerName returns the name of the Jellyfin server.
func (c *JellyfinClient) GetServerName(ctx context.Context) (string, error) {
	var info jellyfinServerInfo
	if err := c.get(ctx, "/System/Info/Public", &info); err != nil {
		return "", fmt.Errorf("getting server info: %w", err)
	}
	return info.ServerName, nil
}

// ProxyImage fetches an image from the Jellyfin server.
func (c *JellyfinClient) ProxyImage(ctx context.Context, path string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("fetching image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading image: %w", err)
	}

	return body, resp.Header.Get("Content-Type"), nil
}

func (c *JellyfinClient) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)
	return c.doRequest(req, result)
}

func (c *JellyfinClient) setHeaders(req *http.Request) {
	if c.apiKey != "" {
		req.Header.Set("Authorization", fmt.Sprintf("MediaBrowser Token=%q", c.apiKey))
	}
}

func (c *JellyfinClient) doRequest(req *http.Request, result any) error {
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

// ticksToSeconds converts Jellyfin ticks (100-nanosecond intervals) to seconds.
func ticksToSeconds(ticks int64) int64 {
	return ticks / 10_000_000
}
