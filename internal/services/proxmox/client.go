package proxmox

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client communicates with the Proxmox VE API.
type Client struct {
	baseURL    string
	ticket     string
	csrfToken  string
	apiToken   string
	httpClient *http.Client
}

// NewClient creates a Proxmox API client.
// When insecureSkipTLS is true and no custom httpClient is provided,
// TLS certificate verification is skipped (common for self-signed certs).
func NewClient(baseURL string, httpClient *http.Client, insecureSkipTLS bool) *Client {
	if httpClient == nil {
		transport := &http.Transport{}
		if insecureSkipTLS {
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // user-configurable
		}
		httpClient = &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
		}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// Authenticate obtains a ticket and CSRF token via username/password.
func (c *Client) Authenticate(ctx context.Context, username, password string) error {
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", password)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api2/json/access/ticket",
		strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed: status %d", resp.StatusCode)
	}

	var result APIResponse[TicketResponse]
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("decoding auth response: %w", err)
	}

	if result.Data.Ticket == "" {
		return fmt.Errorf("empty ticket in response")
	}

	c.ticket = result.Data.Ticket
	c.csrfToken = result.Data.CSRFPreventionToken
	return nil
}

// SetAPIToken sets an API token for authentication (format: USER@REALM!TOKENID=UUID).
func (c *Client) SetAPIToken(token string) {
	c.apiToken = token
}

// ListNodes returns all cluster nodes.
func (c *Client) ListNodes(ctx context.Context) ([]Node, error) {
	var nodes []Node
	if err := c.get(ctx, "/api2/json/nodes", &nodes); err != nil {
		return nil, fmt.Errorf("listing nodes: %w", err)
	}
	return nodes, nil
}

// ListVMs returns all VMs (qemu) for a given node.
func (c *Client) ListVMs(ctx context.Context, node string) ([]VM, error) {
	path := fmt.Sprintf("/api2/json/nodes/%s/qemu", node)
	var vms []VM
	if err := c.get(ctx, path, &vms); err != nil {
		return nil, fmt.Errorf("listing VMs: %w", err)
	}
	for i := range vms {
		vms[i].Type = "qemu"
		vms[i].Node = node
	}
	return vms, nil
}

// ListLXCs returns all LXC containers for a given node.
func (c *Client) ListLXCs(ctx context.Context, node string) ([]VM, error) {
	path := fmt.Sprintf("/api2/json/nodes/%s/lxc", node)
	var lxcs []VM
	if err := c.get(ctx, path, &lxcs); err != nil {
		return nil, fmt.Errorf("listing LXCs: %w", err)
	}
	for i := range lxcs {
		lxcs[i].Type = "lxc"
		lxcs[i].Node = node
	}
	return lxcs, nil
}

// ListAllVMs returns all VMs and LXCs across all nodes.
func (c *Client) ListAllVMs(ctx context.Context) ([]VM, error) {
	nodes, err := c.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	var all []VM
	for _, n := range nodes {
		vms, err := c.ListVMs(ctx, n.Node)
		if err != nil {
			return nil, err
		}
		all = append(all, vms...)

		lxcs, err := c.ListLXCs(ctx, n.Node)
		if err != nil {
			return nil, err
		}
		all = append(all, lxcs...)
	}
	return all, nil
}

func (c *Client) get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	return c.doRequest(req, result)
}

func (c *Client) post(ctx context.Context, path string, data url.Values, result any) error {
	var body io.Reader
	if data != nil {
		body = strings.NewReader(data.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, body)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if data != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result any) error {
	if c.apiToken != "" {
		req.Header.Set("Authorization", "PVEAPIToken="+c.apiToken)
	} else if c.ticket != "" {
		req.AddCookie(&http.Cookie{Name: "PVEAuthCookie", Value: c.ticket})
		if req.Method != http.MethodGet {
			req.Header.Set("CSRFPreventionToken", c.csrfToken)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("unauthorized: invalid or expired credentials")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		var apiResp APIResponse[json.RawMessage]
		if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
		if err := json.Unmarshal(apiResp.Data, result); err != nil {
			return fmt.Errorf("decoding data: %w", err)
		}
	}
	return nil
}
