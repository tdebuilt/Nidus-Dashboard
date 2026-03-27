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

// PlexClient communicates with the Plex Media Server API.
type PlexClient struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewPlexClient creates a Plex API client.
func NewPlexClient(baseURL, token string, httpClient *http.Client) *PlexClient {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &PlexClient{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: httpClient,
	}
}

// GetSessions returns active streaming sessions.
func (c *PlexClient) GetSessions(ctx context.Context) ([]Session, error) {
	var resp plexMediaContainer[plexSession]
	if err := c.get(ctx, "/status/sessions", &resp); err != nil {
		return nil, fmt.Errorf("getting sessions: %w", err)
	}

	sessions := make([]Session, 0, len(resp.MediaContainer.Metadata))
	for _, m := range resp.MediaContainer.Metadata {
		s := Session{
			ID:        m.Session.ID,
			UserName:  m.User.Title,
			Title:     m.Title,
			MediaType: m.Type,
			Year:      m.Year,
			State:     m.Player.State,
			Player:    m.Player.Product,
			Platform:  m.Player.Platform,
			Duration:  m.Duration / 1000,   // ms → seconds
			Position:  m.ViewOffset / 1000, // ms → seconds
		}

		if m.Duration > 0 {
			s.Progress = float64(m.ViewOffset) / float64(m.Duration)
		}

		// Build subtitle for TV episodes
		if m.Type == "episode" && m.GrandparentTitle != "" {
			s.Subtitle = fmt.Sprintf("%s — S%02dE%02d", m.GrandparentTitle, m.ParentIndex, m.Index)
		}

		if m.Thumb != "" {
			s.ThumbPath = m.Thumb
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}

// GetLibraries returns the list of media libraries.
func (c *PlexClient) GetLibraries(ctx context.Context) ([]Library, error) {
	var resp plexMediaContainer[any]
	if err := c.get(ctx, "/library/sections", &resp); err != nil {
		return nil, fmt.Errorf("getting libraries: %w", err)
	}

	libraries := make([]Library, 0, len(resp.MediaContainer.Directory))
	for _, d := range resp.MediaContainer.Directory {
		libType := d.Type
		if libType == "artist" {
			libType = "music"
		}
		libraries = append(libraries, Library{
			ID:        d.Key,
			Name:      d.Title,
			Type:      libType,
			ItemCount: d.Count,
		})
	}

	return libraries, nil
}

// GetServerName returns the friendly name of the Plex server.
func (c *PlexClient) GetServerName(ctx context.Context) (string, error) {
	var resp plexMediaContainer[any]
	if err := c.get(ctx, "/", &resp); err != nil {
		return "", fmt.Errorf("getting server info: %w", err)
	}
	return resp.MediaContainer.FriendlyName, nil
}

// ProxyImage fetches an image from the Plex server.
func (c *PlexClient) ProxyImage(ctx context.Context, path string) ([]byte, string, error) {
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

func (c *PlexClient) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	c.setHeaders(req)
	return c.doRequest(req, result)
}

func (c *PlexClient) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("X-Plex-Token", c.token)
	}
}

func (c *PlexClient) doRequest(req *http.Request, result any) error {
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
