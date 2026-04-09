package qbittorrent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// ErrAuth is returned when authentication with qBittorrent fails.
var ErrAuth = errors.New("authentication failed")

// Client communicates with the qBittorrent Web API (v2).
type Client struct {
	baseURL    string
	username   string
	password   string
	httpClient *http.Client
	sid        string
	mu         sync.Mutex
}

// NewClient creates a qBittorrent API client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// SetCredentials sets the login credentials.
func (c *Client) SetCredentials(username, password string) {
	c.username = username
	c.password = password
}

// authenticate logs in and stores the SID cookie.
func (c *Client) authenticate(ctx context.Context) error {
	form := url.Values{
		"username": {c.username},
		"password": {c.password},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/api/v2/auth/login", strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing auth request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || strings.TrimSpace(string(body)) != "Ok." {
		return fmt.Errorf("%w (HTTP %d): %s", ErrAuth, resp.StatusCode, string(body))
	}

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "SID" {
			c.sid = cookie.Value
			return nil
		}
	}

	return fmt.Errorf("authentication succeeded but no SID cookie received")
}

// getSID ensures a valid session exists and returns the SID.
func (c *Client) getSID(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sid == "" {
		if err := c.authenticate(ctx); err != nil {
			return "", err
		}
	}
	return c.sid, nil
}

// refreshSID forces a re-authentication.
func (c *Client) refreshSID(ctx context.Context) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.authenticate(ctx); err != nil {
		return "", err
	}
	return c.sid, nil
}

// doGet performs an authenticated GET request. On 403, it re-authenticates once and retries.
func (c *Client) doGet(ctx context.Context, path string, params url.Values, result any) error {
	sid, err := c.getSID(ctx)
	if err != nil {
		return err
	}

	fullURL := c.baseURL + path
	if len(params) > 0 {
		fullURL += "?" + params.Encode()
	}

	status, err := c.executeGet(ctx, fullURL, sid, result)
	if err != nil && status == http.StatusForbidden {
		sid, err = c.refreshSID(ctx)
		if err != nil {
			return err
		}
		_, err = c.executeGet(ctx, fullURL, sid, result)
		return err
	}
	return err
}

func (c *Client) executeGet(ctx context.Context, fullURL, sid string, result any) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.AddCookie(&http.Cookie{Name: "SID", Value: sid})

	return c.doRequest(req, result)
}

// doPost performs an authenticated form POST. On 403, it re-authenticates once and retries.
func (c *Client) doPost(ctx context.Context, path string, form url.Values) error {
	sid, err := c.getSID(ctx)
	if err != nil {
		return err
	}

	status, err := c.executePost(ctx, path, form, sid)
	if err != nil && status == http.StatusForbidden {
		sid, err = c.refreshSID(ctx)
		if err != nil {
			return err
		}
		_, err = c.executePost(ctx, path, form, sid)
		return err
	}
	return err
}

func (c *Client) executePost(ctx context.Context, path string, form url.Values, sid string) (int, error) {
	var body io.Reader
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, body)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "SID", Value: sid})

	return c.doRequest(req, nil)
}

// doMultipartPost performs an authenticated multipart/form-data POST. On 403,
// it re-authenticates once and retries with the same pre-built body.
func (c *Client) doMultipartPost(ctx context.Context, path, contentType string, body []byte) error {
	sid, err := c.getSID(ctx)
	if err != nil {
		return err
	}

	status, err := c.executeMultipartPost(ctx, path, contentType, body, sid)
	if err != nil && status == http.StatusForbidden {
		sid, err = c.refreshSID(ctx)
		if err != nil {
			return err
		}
		_, err = c.executeMultipartPost(ctx, path, contentType, body, sid)
		return err
	}
	return err
}

func (c *Client) executeMultipartPost(ctx context.Context, path, contentType string, body []byte, sid string) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", contentType)
	req.AddCookie(&http.Cookie{Name: "SID", Value: sid})

	return c.doRequest(req, nil)
}

// doRequest executes the request and optionally decodes JSON.
func (c *Client) doRequest(req *http.Request, result any) (int, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return http.StatusForbidden, fmt.Errorf("forbidden: session expired or invalid")
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

// ListTorrents returns all torrents.
func (c *Client) ListTorrents(ctx context.Context) ([]Torrent, error) {
	var torrents []Torrent
	if err := c.doGet(ctx, "/api/v2/torrents/info", nil, &torrents); err != nil {
		return nil, fmt.Errorf("listing torrents: %w", err)
	}
	return torrents, nil
}

// GetTransferInfo returns global transfer statistics.
func (c *Client) GetTransferInfo(ctx context.Context) (*TransferInfo, error) {
	var info TransferInfo
	if err := c.doGet(ctx, "/api/v2/transfer/info", nil, &info); err != nil {
		return nil, fmt.Errorf("getting transfer info: %w", err)
	}
	return &info, nil
}

// ResumeTorrents resumes one or more torrents by hash. Use "all" for all.
func (c *Client) ResumeTorrents(ctx context.Context, hashes []string) error {
	form := url.Values{"hashes": {strings.Join(hashes, "|")}}
	if err := c.doPost(ctx, "/api/v2/torrents/resume", form); err != nil {
		return fmt.Errorf("resuming torrents: %w", err)
	}
	return nil
}

// PauseTorrents pauses one or more torrents by hash. Use "all" for all.
func (c *Client) PauseTorrents(ctx context.Context, hashes []string) error {
	form := url.Values{"hashes": {strings.Join(hashes, "|")}}
	if err := c.doPost(ctx, "/api/v2/torrents/pause", form); err != nil {
		return fmt.Errorf("pausing torrents: %w", err)
	}
	return nil
}

// DeleteTorrents deletes one or more torrents. If deleteFiles is true, downloaded data is also removed.
func (c *Client) DeleteTorrents(ctx context.Context, hashes []string, deleteFiles bool) error {
	delFiles := "false"
	if deleteFiles {
		delFiles = "true"
	}
	form := url.Values{
		"hashes":      {strings.Join(hashes, "|")},
		"deleteFiles": {delFiles},
	}
	if err := c.doPost(ctx, "/api/v2/torrents/delete", form); err != nil {
		return fmt.Errorf("deleting torrents: %w", err)
	}
	return nil
}

// AddTorrent adds a torrent from a URL/magnet link or raw .torrent bytes.
// Category and SavePath are optional and forwarded to qBittorrent as-is.
func (c *Client) AddTorrent(ctx context.Context, opts AddOptions) error {
	if len(opts.File) == 0 && opts.URL == "" {
		return errors.New("add torrent: url or file required")
	}
	if len(opts.File) > 0 {
		return c.addTorrentMultipart(ctx, opts)
	}
	form := url.Values{"urls": {opts.URL}}
	if opts.Category != "" {
		form.Set("category", opts.Category)
	}
	if opts.SavePath != "" {
		form.Set("savepath", opts.SavePath)
	}
	if err := c.doPost(ctx, "/api/v2/torrents/add", form); err != nil {
		return fmt.Errorf("adding torrent: %w", err)
	}
	return nil
}

// addTorrentMultipart uploads a .torrent file via multipart/form-data.
func (c *Client) addTorrentMultipart(ctx context.Context, opts AddOptions) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("torrents", "upload.torrent")
	if err != nil {
		return fmt.Errorf("building multipart: %w", err)
	}
	if _, err := part.Write(opts.File); err != nil {
		return fmt.Errorf("writing torrent bytes: %w", err)
	}
	if opts.Category != "" {
		_ = writer.WriteField("category", opts.Category)
	}
	if opts.SavePath != "" {
		_ = writer.WriteField("savepath", opts.SavePath)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("closing multipart: %w", err)
	}

	if err := c.doMultipartPost(ctx, "/api/v2/torrents/add", writer.FormDataContentType(), body.Bytes()); err != nil {
		return fmt.Errorf("adding torrent file: %w", err)
	}
	return nil
}

// GetCategories returns the list of categories configured in qBittorrent.
func (c *Client) GetCategories(ctx context.Context) (map[string]Category, error) {
	categories := map[string]Category{}
	if err := c.doGet(ctx, "/api/v2/torrents/categories", nil, &categories); err != nil {
		return nil, fmt.Errorf("fetching categories: %w", err)
	}
	return categories, nil
}

// ResumeAll resumes all torrents.
func (c *Client) ResumeAll(ctx context.Context) error {
	return c.ResumeTorrents(ctx, []string{"all"})
}

// PauseAll pauses all torrents.
func (c *Client) PauseAll(ctx context.Context) error {
	return c.PauseTorrents(ctx, []string{"all"})
}
