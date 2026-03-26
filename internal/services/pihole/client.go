package pihole

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Client communicates with the Pi-hole v6 API.
type Client struct {
	baseURL    string
	password   string
	httpClient *http.Client
	sid        string
	csrf       string
	expiry     time.Time
	mu         sync.Mutex
}

// NewClient creates a Pi-hole API client.
func NewClient(baseURL, password string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		password:   password,
		httpClient: httpClient,
	}
}

// authenticate obtains a new session from the Pi-hole API.
func (c *Client) authenticate() error {
	payload, err := json.Marshal(AuthRequest{Password: c.password})
	if err != nil {
		return fmt.Errorf("marshaling auth request: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/api/auth", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing auth request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf("unauthorized: invalid password")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return fmt.Errorf("decoding auth response: %w", err)
	}

	if !authResp.Session.Valid {
		return fmt.Errorf("unauthorized: invalid password")
	}

	c.sid = authResp.Session.SID
	c.csrf = authResp.Session.CSRF
	c.expiry = time.Now().Add(time.Duration(authResp.Session.Validity)*time.Second - 10*time.Second)
	return nil
}

// isSessionValid checks whether the current session is still usable.
func (c *Client) isSessionValid() bool {
	return c.sid != "" && time.Now().Before(c.expiry)
}

// getCredentials ensures a valid session exists and returns a copy of the credentials.
func (c *Client) getCredentials() (sid, csrf string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isSessionValid() {
		if err := c.authenticate(); err != nil {
			return "", "", err
		}
	}
	return c.sid, c.csrf, nil
}

// refreshCredentials forces a re-authentication and returns new credentials.
func (c *Client) refreshCredentials() (sid, csrf string, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.authenticate(); err != nil {
		return "", "", err
	}
	return c.sid, c.csrf, nil
}

// authenticatedGet performs a GET request with session authentication.
// On 401, it re-authenticates once and retries.
func (c *Client) authenticatedGet(path string, result any) error {
	sid, csrf, err := c.getCredentials()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("X-FTL-SID", sid)
	req.Header.Set("X-FTL-CSRF", csrf)

	status, err := c.doRequest(req, result)
	if err != nil && status == http.StatusUnauthorized {
		sid, csrf, err = c.refreshCredentials()
		if err != nil {
			return err
		}

		req, err = http.NewRequest(http.MethodGet, c.baseURL+path, nil)
		if err != nil {
			return fmt.Errorf("creating retry request: %w", err)
		}
		req.Header.Set("X-FTL-SID", sid)
		req.Header.Set("X-FTL-CSRF", csrf)

		_, err = c.doRequest(req, result)
		return err
	}
	return err
}

// authenticatedPost performs a POST request with session authentication.
// On 401, it re-authenticates once and retries.
func (c *Client) authenticatedPost(path string, body any, result any) error {
	sid, csrf, err := c.getCredentials()
	if err != nil {
		return err
	}

	req, err := c.buildPostRequest(path, body, sid, csrf)
	if err != nil {
		return err
	}

	status, err := c.doRequest(req, result)
	if err != nil && status == http.StatusUnauthorized {
		sid, csrf, err = c.refreshCredentials()
		if err != nil {
			return err
		}

		req, err = c.buildPostRequest(path, body, sid, csrf)
		if err != nil {
			return err
		}

		_, err = c.doRequest(req, result)
		return err
	}
	return err
}

func (c *Client) buildPostRequest(path string, body any, sid, csrf string) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshaling request body: %w", err)
		}
		reader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+path, reader)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("X-FTL-SID", sid)
	req.Header.Set("X-FTL-CSRF", csrf)
	return req, nil
}

// doRequest executes the request and decodes the JSON response.
// It returns the HTTP status code alongside any error.
func (c *Client) doRequest(req *http.Request, result any) (int, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		return http.StatusUnauthorized, fmt.Errorf("unauthorized: session expired or invalid")
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return resp.StatusCode, fmt.Errorf("decoding response: %w", err)
		}
	}
	return resp.StatusCode, nil
}

// GetStats returns combined DNS statistics and blocking status.
func (c *Client) GetStats() (*StatsInfo, error) {
	var statsResp StatsResponse
	if err := c.authenticatedGet("/api/stats/summary", &statsResp); err != nil {
		return nil, fmt.Errorf("getting stats: %w", err)
	}

	var blockResp BlockingResponse
	if err := c.authenticatedGet("/api/dns/blocking", &blockResp); err != nil {
		return nil, fmt.Errorf("getting blocking status: %w", err)
	}

	return &StatsInfo{
		TotalQueries:     statsResp.Queries.Total,
		BlockedQueries:   statsResp.Queries.Blocked,
		BlockedPercent:   statsResp.Queries.PercentBlocked,
		UniqueDomains:    statsResp.Queries.UniqueDomains,
		CachedQueries:    statsResp.Queries.Cached,
		ForwardedQueries: statsResp.Queries.Forwarded,
		BlockingEnabled:  blockResp.Blocking,
	}, nil
}

// SetBlocking enables or disables Pi-hole DNS blocking.
func (c *Client) SetBlocking(enabled bool) error {
	body := map[string]bool{"blocking": enabled}
	if err := c.authenticatedPost("/api/dns/blocking", body, nil); err != nil {
		return fmt.Errorf("setting blocking: %w", err)
	}
	return nil
}
