package homeassistant

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client communicates with the Home Assistant REST API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a Home Assistant API client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// SetToken sets the long-lived access token.
func (c *Client) SetToken(token string) {
	c.token = token
}

// ListStates returns all entity states.
func (c *Client) ListStates(ctx context.Context) ([]Entity, error) {
	var entities []Entity
	if err := c.get(ctx, "/api/states", &entities); err != nil {
		return nil, fmt.Errorf("listing states: %w", err)
	}
	return entities, nil
}

// GetState returns the state of a single entity.
func (c *Client) GetState(ctx context.Context, entityID string) (*Entity, error) {
	var entity Entity
	if err := c.get(ctx, "/api/states/"+entityID, &entity); err != nil {
		return nil, fmt.Errorf("getting state: %w", err)
	}
	return &entity, nil
}

// CallService calls a Home Assistant service.
func (c *Client) CallService(ctx context.Context, domain, service string, data ServiceCallRequest) (ServiceCallResponse, error) {
	var resp ServiceCallResponse
	path := fmt.Sprintf("/api/services/%s/%s", domain, service)
	if err := c.post(ctx, path, data, &resp); err != nil {
		return nil, fmt.Errorf("calling service: %w", err)
	}
	return resp, nil
}

// GetCameraSnapshot returns a camera entity snapshot as raw bytes.
func (c *Client) GetCameraSnapshot(ctx context.Context, entityID string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/api/camera_proxy/"+entityID, nil)
	if err != nil {
		return nil, "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("reading response: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	return data, contentType, nil
}

// ToEntityInfo converts a raw Entity to a simplified EntityInfo.
func ToEntityInfo(e Entity) EntityInfo {
	parts := strings.SplitN(e.EntityID, ".", 2)
	domain := ""
	if len(parts) > 0 {
		domain = parts[0]
	}

	name := e.EntityID
	if fn, ok := e.Attributes["friendly_name"].(string); ok {
		name = fn
	}

	icon := ""
	if ic, ok := e.Attributes["icon"].(string); ok {
		icon = ic
	}

	unit := ""
	if u, ok := e.Attributes["unit_of_measurement"].(string); ok {
		unit = u
	}

	return EntityInfo{
		EntityID:      e.EntityID,
		Domain:        domain,
		Name:          name,
		State:         e.State,
		Attributes:    e.Attributes,
		Icon:          icon,
		UnitOfMeasure: unit,
		LastChanged:   e.LastChanged,
	}
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
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
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
