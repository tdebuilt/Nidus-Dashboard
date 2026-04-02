package transmission

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// ErrAuth is returned when authentication with Transmission fails.
var ErrAuth = errors.New("authentication failed")

const sessionIDHeader = "X-Transmission-Session-Id"

// Client communicates with the Transmission RPC API.
type Client struct {
	baseURL    string
	username   string
	password   string
	sessionID  string
	httpClient *http.Client
	mu         sync.Mutex
}

// NewClient creates a Transmission RPC client.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// SetCredentials sets the HTTP Basic Auth credentials.
func (c *Client) SetCredentials(username, password string) {
	c.username = username
	c.password = password
}

// ListTorrents returns all torrents with detailed fields.
func (c *Client) ListTorrents(ctx context.Context) ([]Torrent, error) {
	args := map[string]interface{}{
		"fields": []string{
			"id", "name", "status", "totalSize", "sizeWhenDone",
			"leftUntilDone", "percentDone", "rateDownload", "rateUpload",
			"eta", "uploadRatio", "peersConnected", "addedDate",
			"error", "errorString",
		},
	}

	var resp TorrentListResponse
	if err := c.doRPC(ctx, "torrent-get", args, &resp); err != nil {
		return nil, fmt.Errorf("listing torrents: %w", err)
	}
	return resp.Torrents, nil
}

// AddTorrent adds a torrent by URL or magnet link.
func (c *Client) AddTorrent(ctx context.Context, url string) error {
	args := map[string]string{
		"filename": url,
	}
	if err := c.doRPC(ctx, "torrent-add", args, nil); err != nil {
		return fmt.Errorf("adding torrent: %w", err)
	}
	return nil
}

// AddTorrentByFile adds a torrent from base64-encoded .torrent file content.
func (c *Client) AddTorrentByFile(ctx context.Context, metainfo string) error {
	args := map[string]string{
		"metainfo": metainfo,
	}
	if err := c.doRPC(ctx, "torrent-add", args, nil); err != nil {
		return fmt.Errorf("adding torrent by file: %w", err)
	}
	return nil
}

// StartTorrent starts torrents by IDs.
func (c *Client) StartTorrent(ctx context.Context, ids []int) error {
	args := map[string]interface{}{
		"ids": ids,
	}
	if err := c.doRPC(ctx, "torrent-start", args, nil); err != nil {
		return fmt.Errorf("starting torrent: %w", err)
	}
	return nil
}

// StopTorrent stops torrents by IDs.
func (c *Client) StopTorrent(ctx context.Context, ids []int) error {
	args := map[string]interface{}{
		"ids": ids,
	}
	if err := c.doRPC(ctx, "torrent-stop", args, nil); err != nil {
		return fmt.Errorf("stopping torrent: %w", err)
	}
	return nil
}

// StartAll starts all torrents.
func (c *Client) StartAll(ctx context.Context) error {
	if err := c.doRPC(ctx, "torrent-start", nil, nil); err != nil {
		return fmt.Errorf("starting all torrents: %w", err)
	}
	return nil
}

// StopAll stops all torrents.
func (c *Client) StopAll(ctx context.Context) error {
	if err := c.doRPC(ctx, "torrent-stop", nil, nil); err != nil {
		return fmt.Errorf("stopping all torrents: %w", err)
	}
	return nil
}

// RemoveCompleted removes all completed torrents without deleting files.
func (c *Client) RemoveCompleted(ctx context.Context) (int, error) {
	torrents, err := c.ListTorrents(ctx)
	if err != nil {
		return 0, err
	}
	var ids []int
	for _, t := range torrents {
		if t.PercentDone >= 1.0 {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) == 0 {
		return 0, nil
	}
	args := map[string]interface{}{
		"ids":               ids,
		"delete-local-data": false,
	}
	if err := c.doRPC(ctx, "torrent-remove", args, nil); err != nil {
		return 0, fmt.Errorf("removing completed torrents: %w", err)
	}
	return len(ids), nil
}

// GetSessionStats returns session statistics.
func (c *Client) GetSessionStats(ctx context.Context) (*SessionStats, error) {
	var stats SessionStats
	if err := c.doRPC(ctx, "session-stats", nil, &stats); err != nil {
		return nil, fmt.Errorf("getting session stats: %w", err)
	}
	return &stats, nil
}

// ToTorrentInfo converts a Torrent to a frontend TorrentInfo.
func ToTorrentInfo(t Torrent) TorrentInfo {
	status := statusString(t.Status)
	downloaded := t.SizeWhenDone - t.LeftUntilDone

	errMsg := ""
	if t.Error > 0 && t.ErrorString != "" {
		errMsg = t.ErrorString
	}

	return TorrentInfo{
		ID:         t.ID,
		Name:       t.Name,
		Status:     status,
		Progress:   t.PercentDone * 100,
		Size:       t.SizeWhenDone,
		Downloaded: downloaded,
		SpeedDown:  t.RateDownload,
		SpeedUp:    t.RateUpload,
		ETA:        t.ETA,
		Ratio:      t.UploadRatio,
		Peers:      t.Peers,
		Error:      errMsg,
	}
}

func statusString(status int) string {
	switch status {
	case StatusStopped:
		return "stopped"
	case StatusCheckWait, StatusChecking:
		return "checking"
	case StatusDownloadWait, StatusDownloading:
		return "downloading"
	case StatusSeedWait, StatusSeeding:
		return "seeding"
	default:
		return "unknown"
	}
}

func (c *Client) doRPC(ctx context.Context, method string, args interface{}, result interface{}) error {
	data, err := json.Marshal(RPCRequest{Method: method, Arguments: args})
	if err != nil {
		return fmt.Errorf("marshaling RPC request: %w", err)
	}

	resp, err := c.sendRequest(ctx, data)
	if err != nil {
		return err
	}

	// Handle 409 — get session ID and retry
	if resp.StatusCode == http.StatusConflict {
		c.mu.Lock()
		c.sessionID = resp.Header.Get(sessionIDHeader)
		c.mu.Unlock()
		resp.Body.Close()

		resp, err = c.sendRequest(ctx, data)
		if err != nil {
			return err
		}
	}
	defer resp.Body.Close()

	return decodeRPCResponse(resp, result)
}

// decodeRPCResponse reads and validates the RPC response body.
func decodeRPCResponse(resp *http.Response, result interface{}) error {
	if resp.StatusCode == http.StatusUnauthorized {
		_, _ = io.Copy(io.Discard, resp.Body)
		return fmt.Errorf("%w: invalid credentials", ErrAuth)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("unexpected status %d", resp.StatusCode)
		}
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(bodyBytes))
	}
	var rpcResp RPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("decoding RPC response: %w", err)
	}
	if rpcResp.Result != "success" {
		return fmt.Errorf("RPC error: %s", rpcResp.Result)
	}
	if result != nil && rpcResp.Arguments != nil {
		if err := json.Unmarshal(rpcResp.Arguments, result); err != nil {
			return fmt.Errorf("decoding arguments: %w", err)
		}
	}
	return nil
}

func (c *Client) sendRequest(ctx context.Context, data []byte) (*http.Response, error) {
	rpcURL := c.baseURL + "/transmission/rpc"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpcURL, strings.NewReader(string(data)))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	c.mu.Lock()
	if c.sessionID != "" {
		req.Header.Set(sessionIDHeader, c.sessionID)
	}
	c.mu.Unlock()

	if c.username != "" {
		req.SetBasicAuth(c.username, c.password)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}

	return resp, nil
}
