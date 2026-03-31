package portainer

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client communicates with the Portainer API.
type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// NewClient creates a Portainer API client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// Authenticate logs in and stores the JWT token for subsequent requests.
func (c *Client) Authenticate(ctx context.Context, username, password string) error {
	body, err := json.Marshal(AuthRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("marshaling auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/auth", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed: status %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("decoding auth response: %w", err)
	}

	if authResp.JWT == "" {
		return fmt.Errorf("empty JWT token in response")
	}

	c.token = authResp.JWT
	return nil
}

// SetToken sets the JWT token directly (useful when using API keys).
func (c *Client) SetToken(token string) {
	c.token = token
}

// HasToken reports whether the client has an authentication token set.
func (c *Client) HasToken() bool {
	return c.token != ""
}

// ListEnvironments returns all Portainer environments (endpoints).
func (c *Client) ListEnvironments(ctx context.Context) ([]Environment, error) {
	var envs []Environment
	if err := c.get(ctx, "/api/endpoints", &envs); err != nil {
		return nil, fmt.Errorf("listing environments: %w", err)
	}
	return envs, nil
}

// ListStacks returns all stacks, optionally filtered by environment ID.
func (c *Client) ListStacks(ctx context.Context, envID int) ([]Stack, error) {
	path := "/api/stacks"
	if envID > 0 {
		path = fmt.Sprintf("/api/stacks?filters={\"EndpointID\":%d}", envID)
	}

	var stacks []Stack
	if err := c.get(ctx, path, &stacks); err != nil {
		return nil, fmt.Errorf("listing stacks: %w", err)
	}
	return stacks, nil
}

// ListContainers returns all containers for a given environment.
func (c *Client) ListContainers(ctx context.Context, envID int) ([]Container, error) {
	path := fmt.Sprintf("/api/endpoints/%d/docker/containers/json?all=true", envID)

	var containers []Container
	if err := c.get(ctx, path, &containers); err != nil {
		return nil, fmt.Errorf("listing containers: %w", err)
	}
	return containers, nil
}

// InspectContainer returns detailed info for a container.
func (c *Client) InspectContainer(ctx context.Context, envID int, containerID string) (*ContainerDetails, error) {
	path := fmt.Sprintf("/api/endpoints/%d/docker/containers/%s/json", envID, containerID)

	var details ContainerDetails
	if err := c.get(ctx, path, &details); err != nil {
		return nil, fmt.Errorf("inspecting container: %w", err)
	}
	return &details, nil
}

// InspectImage returns inspect data for a local image.
func (c *Client) InspectImage(ctx context.Context, envID int, imageID string) (*ImageInspect, error) {
	path := fmt.Sprintf("/api/endpoints/%d/docker/images/%s/json", envID, imageID)
	var img ImageInspect
	if err := c.get(ctx, path, &img); err != nil {
		return nil, fmt.Errorf("inspecting image: %w", err)
	}
	return &img, nil
}

// GetDistribution returns the remote registry digest for an image without pulling.
func (c *Client) GetDistribution(ctx context.Context, envID int, imageName string) (*DistributionInfo, error) {
	path := fmt.Sprintf("/api/endpoints/%d/docker/distribution/%s/json", envID, imageName)
	var info DistributionInfo
	if err := c.get(ctx, path, &info); err != nil {
		return nil, fmt.Errorf("getting distribution: %w", err)
	}
	return &info, nil
}

// GetContainerStats returns a single stats snapshot for a container.
func (c *Client) GetContainerStats(ctx context.Context, envID int, containerID string) (*DockerStats, error) {
	path := fmt.Sprintf("/api/endpoints/%d/docker/containers/%s/stats?stream=false&one-shot=true", envID, containerID)

	var stats DockerStats
	if err := c.get(ctx, path, &stats); err != nil {
		return nil, fmt.Errorf("getting container stats: %w", err)
	}
	return &stats, nil
}

// CalculateContainerStats fetches raw stats and computes CPU% and memory usage.
func (c *Client) CalculateContainerStats(ctx context.Context, envID int, containerID string) (*ContainerStats, error) {
	raw, err := c.GetContainerStats(ctx, envID, containerID)
	if err != nil {
		return nil, err
	}

	// CPU percentage calculation (same formula as docker stats CLI)
	cpuPercent := 0.0
	cpuDelta := float64(raw.CPUStats.CPUUsage.TotalUsage - raw.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(raw.CPUStats.SystemUsage - raw.PreCPUStats.SystemUsage)
	if sysDelta > 0 && cpuDelta > 0 {
		cpus := raw.CPUStats.OnlineCPUs
		if cpus == 0 {
			cpus = 1
		}
		cpuPercent = (cpuDelta / sysDelta) * float64(cpus) * 100.0
	}

	// Memory: usage minus cache (if available)
	memUsage := raw.MemoryStats.Usage
	if cache, ok := raw.MemoryStats.Stats["cache"]; ok {
		memUsage -= cache
	}
	memLimit := raw.MemoryStats.Limit
	memPercent := 0.0
	if memLimit > 0 {
		memPercent = float64(memUsage) / float64(memLimit) * 100.0
	}

	return &ContainerStats{
		ContainerID: containerID,
		CPUPercent:  cpuPercent,
		MemUsage:    memUsage,
		MemLimit:    memLimit,
		MemPercent:  memPercent,
	}, nil
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

func (c *Client) put(ctx context.Context, path string, body any, result any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		reader = strings.NewReader(string(data))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+path, reader)
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
		if strings.HasPrefix(c.token, "ptr_") {
			req.Header.Set("X-API-Key", c.token)
		} else {
			req.Header.Set("Authorization", "Bearer "+c.token)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: invalid or expired token")
	}

	// 409 Conflict is OK for idempotent actions (e.g. "stack already running")
	if resp.StatusCode == http.StatusConflict {
		_, _ = io.ReadAll(resp.Body)
		return nil
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
