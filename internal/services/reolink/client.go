package reolink

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Client communicates with a Reolink camera via its HTTP API.
type Client struct {
	ip         string
	port       int
	username   string
	password   string
	channel    int
	httpClient *http.Client

	mu     sync.Mutex
	scheme string // cached working scheme ("http" or "https")
	token  string // cached auth token for newer cameras
	tokenT time.Time
}

const (
	tokenTTL        = 30 * time.Minute
	contentTypeJPEG = "image/jpeg"
)

// NewClient creates a new Reolink camera client.
// When insecureSkipTLS is true and no custom httpClient is provided,
// TLS certificate verification is skipped (common for camera self-signed certs).
func NewClient(ip, username, password string, channel int, httpClient *http.Client, insecureSkipTLS bool) *Client {
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
		ip:         ip,
		port:       80,
		username:   username,
		password:   password,
		channel:    channel,
		httpClient: httpClient,
	}
}

// GetSnapshot fetches a JPEG snapshot from the camera.
func (c *Client) GetSnapshot(ctx context.Context) ([]byte, string, error) {
	c.mu.Lock()
	scheme := c.scheme
	token := c.token
	tokenAge := time.Since(c.tokenT)
	c.mu.Unlock()

	rs := time.Now().UnixMilli()

	// If we have a cached token that's still fresh, use it directly
	if token != "" && tokenAge < tokenTTL && scheme != "" {
		body, err := c.snapWithToken(ctx, scheme, rs, token)
		if err == nil && isJPEG(body) {
			return body, contentTypeJPEG, nil
		}
		// Token may have expired, clear and retry
		c.mu.Lock()
		c.token = ""
		c.mu.Unlock()
	}

	// If we don't know the scheme, discover from scratch
	if scheme == "" {
		return c.discoverScheme(ctx, rs)
	}

	// Try direct auth with cached scheme
	body, err := c.snapDirect(ctx, scheme, rs)
	if err == nil && isJPEG(body) {
		return body, contentTypeJPEG, nil
	}

	// Direct auth returned non-JPEG; try token-based auth
	if err == nil {
		data, ct, loginErr := c.loginAndSnap(ctx, scheme, rs)
		if loginErr == nil {
			return data, ct, nil
		}
	}

	// Cached scheme failed completely, reset and rediscover
	c.mu.Lock()
	c.scheme = ""
	c.token = ""
	c.mu.Unlock()

	return c.discoverScheme(ctx, rs)
}

// discoverScheme tries each scheme (http, https) with direct and token auth
// to find a working combination, caching the result for future calls.
func (c *Client) discoverScheme(ctx context.Context, rs int64) ([]byte, string, error) {
	for _, s := range []string{"http", "https"} {
		body, err := c.snapDirect(ctx, s, rs)
		if err != nil {
			continue
		}
		if isJPEG(body) {
			c.mu.Lock()
			c.scheme = s
			c.mu.Unlock()
			return body, contentTypeJPEG, nil
		}
		data, ct, loginErr := c.loginAndSnap(ctx, s, rs)
		if loginErr == nil {
			c.mu.Lock()
			c.scheme = s
			c.mu.Unlock()
			return data, ct, nil
		}
	}
	return nil, "", fmt.Errorf("snapshot failed: camera at %s unreachable", c.ip)
}

func (c *Client) snapDirect(ctx context.Context, scheme string, rs int64) ([]byte, error) {
	u := fmt.Sprintf("%s://%s/cgi-bin/api.cgi?cmd=Snap&channel=%d&rs=%d&user=%s&password=%s",
		scheme, c.ip, c.channel, rs, url.QueryEscape(c.username), url.QueryEscape(c.password))
	return c.fetchSnapshot(ctx, u)
}

func (c *Client) snapWithToken(ctx context.Context, scheme string, rs int64, token string) ([]byte, error) {
	u := fmt.Sprintf("%s://%s/cgi-bin/api.cgi?cmd=Snap&channel=%d&rs=%d&token=%s",
		scheme, c.ip, c.channel, rs, token)
	return c.fetchSnapshot(ctx, u)
}

func (c *Client) loginAndSnap(ctx context.Context, scheme string, rs int64) ([]byte, string, error) {
	token, err := c.ensureToken(ctx, scheme)
	if err != nil {
		return nil, "", fmt.Errorf("login failed: %w", err)
	}

	body, err := c.snapWithToken(ctx, scheme, rs, token)
	if err != nil {
		return nil, "", fmt.Errorf("token snapshot failed: %w", err)
	}
	if isJPEG(body) {
		return body, contentTypeJPEG, nil
	}
	return nil, "", fmt.Errorf("camera returned non-image response after token auth")
}

// ensureToken returns a valid token, reusing a cached one or logging in under mutex.
func (c *Client) ensureToken(ctx context.Context, scheme string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Another goroutine may have refreshed the token while we waited
	if c.token != "" && time.Since(c.tokenT) < tokenTTL {
		return c.token, nil
	}

	token, err := c.login(ctx, scheme)
	if err != nil {
		return "", err
	}
	c.token = token
	c.tokenT = time.Now()
	return token, nil
}

func (c *Client) fetchSnapshot(ctx context.Context, u string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) login(ctx context.Context, scheme string) (string, error) {
	payload := fmt.Sprintf(`[{"cmd":"Login","action":0,"param":{"User":{"userName":%q,"password":%q}}}]`,
		c.username, c.password)

	u := fmt.Sprintf("%s://%s/cgi-bin/api.cgi?cmd=Login", scheme, c.ip)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewBufferString(payload))
	if err != nil {
		return "", fmt.Errorf("creating login request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("login request: %w", err)
	}
	defer resp.Body.Close()

	var result []struct {
		Code  int `json:"code"`
		Value struct {
			Token struct {
				Name string `json:"name"`
			} `json:"Token"`
		} `json:"value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("login decode: %w", err)
	}
	if len(result) == 0 || result[0].Code != 0 || result[0].Value.Token.Name == "" {
		return "", fmt.Errorf("login failed")
	}
	return result[0].Value.Token.Name, nil
}

func isJPEG(data []byte) bool {
	return len(data) >= 2 && data[0] == 0xFF && data[1] == 0xD8
}

// GetRTSPURL returns the RTSP stream URL for this camera.
// streamType: "main" for high quality, "sub" for low quality.
func (c *Client) GetRTSPURL(streamType string) string {
	return FormatRTSPURL(c.username, c.password, c.ip, c.channel, streamType)
}

// FormatRTSPURL builds an RTSP URL without needing a Client instance.
func FormatRTSPURL(username, password, ip string, channel int, streamType string) string {
	if streamType == "" {
		streamType = "main"
	}
	ch := fmt.Sprintf("%02d", channel+1)
	return fmt.Sprintf("rtsp://%s:%s@%s/Preview_%s_%s",
		url.PathEscape(username), url.PathEscape(password), ip, ch, streamType)
}
