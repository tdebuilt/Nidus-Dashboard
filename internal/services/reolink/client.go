package reolink

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

const tokenTTL = 30 * time.Minute

// NewClient creates a new Reolink camera client.
func NewClient(ip, username, password string, channel int, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
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
func (c *Client) GetSnapshot() ([]byte, string, error) {
	c.mu.Lock()
	scheme := c.scheme
	token := c.token
	tokenAge := time.Since(c.tokenT)
	c.mu.Unlock()

	rs := time.Now().UnixMilli()

	// If we have a cached token that's still fresh, use it directly
	if token != "" && tokenAge < tokenTTL && scheme != "" {
		body, err := c.snapWithToken(scheme, rs, token)
		if err == nil && isJPEG(body) {
			return body, "image/jpeg", nil
		}
		// Token may have expired, clear and retry
		c.mu.Lock()
		c.token = ""
		c.mu.Unlock()
	}

	// If we know the scheme, try direct auth then token auth
	if scheme != "" {
		body, err := c.snapDirect(scheme, rs)
		if err == nil && isJPEG(body) {
			return body, "image/jpeg", nil
		}
		if err == nil {
			data, ct, loginErr := c.loginAndSnap(scheme, rs)
			if loginErr == nil {
				return data, ct, nil
			}
		}
		// Cached scheme failed completely, reset and rediscover
		c.mu.Lock()
		c.scheme = ""
		c.token = ""
		c.mu.Unlock()
	}

	// Discover working scheme: try direct auth, then token auth per scheme
	for _, s := range []string{"http", "https"} {
		body, err := c.snapDirect(s, rs)
		if err != nil {
			continue
		}
		if isJPEG(body) {
			c.mu.Lock()
			c.scheme = s
			c.mu.Unlock()
			return body, "image/jpeg", nil
		}
		// Not JPEG — try token auth on this scheme
		data, ct, loginErr := c.loginAndSnap(s, rs)
		if loginErr == nil {
			c.mu.Lock()
			c.scheme = s
			c.mu.Unlock()
			return data, ct, nil
		}
		// Login failed on this scheme, try next
	}

	return nil, "", fmt.Errorf("snapshot failed: camera at %s unreachable", c.ip)
}

func (c *Client) snapDirect(scheme string, rs int64) ([]byte, error) {
	url := fmt.Sprintf("%s://%s/cgi-bin/api.cgi?cmd=Snap&channel=%d&rs=%d&user=%s&password=%s",
		scheme, c.ip, c.channel, rs, c.username, c.password)
	return c.fetchSnapshot(url)
}

func (c *Client) snapWithToken(scheme string, rs int64, token string) ([]byte, error) {
	url := fmt.Sprintf("%s://%s/cgi-bin/api.cgi?cmd=Snap&channel=%d&rs=%d&token=%s",
		scheme, c.ip, c.channel, rs, token)
	return c.fetchSnapshot(url)
}

func (c *Client) loginAndSnap(scheme string, rs int64) ([]byte, string, error) {
	token, err := c.ensureToken(scheme)
	if err != nil {
		return nil, "", fmt.Errorf("login failed: %w", err)
	}

	body, err := c.snapWithToken(scheme, rs, token)
	if err != nil {
		return nil, "", fmt.Errorf("token snapshot failed: %w", err)
	}
	if isJPEG(body) {
		return body, "image/jpeg", nil
	}
	return nil, "", fmt.Errorf("camera returned non-image response after token auth")
}

// ensureToken returns a valid token, reusing a cached one or logging in under mutex.
func (c *Client) ensureToken(scheme string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Another goroutine may have refreshed the token while we waited
	if c.token != "" && time.Since(c.tokenT) < tokenTTL {
		return c.token, nil
	}

	token, err := c.login(scheme)
	if err != nil {
		return "", err
	}
	c.token = token
	c.tokenT = time.Now()
	return token, nil
}

func (c *Client) fetchSnapshot(url string) ([]byte, error) {
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func (c *Client) login(scheme string) (string, error) {
	payload := fmt.Sprintf(`[{"cmd":"Login","action":0,"param":{"User":{"userName":%q,"password":%q}}}]`,
		c.username, c.password)

	url := fmt.Sprintf("%s://%s/cgi-bin/api.cgi?cmd=Login", scheme, c.ip)
	resp, err := c.httpClient.Post(url, "application/json", bytes.NewBufferString(payload))
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
	return fmt.Sprintf("rtsp://%s:%s@%s/Preview_%s_%s", username, password, ip, ch, streamType)
}
