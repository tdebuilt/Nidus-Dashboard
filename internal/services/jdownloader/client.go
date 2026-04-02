package jdownloader

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// ErrAuth is returned when authentication with MyJDownloader fails.
var ErrAuth = errors.New("authentication failed")

const (
	apiURL = "https://api.jdownloader.org"
	appKey = "nidus"
)

// Client communicates with the MyJDownloader cloud API.
type Client struct {
	httpClient            *http.Client
	email                 string
	password              string
	loginSecret           []byte
	deviceSecret          []byte
	mu                    sync.Mutex
	sessionToken          string
	regainToken           string
	serverEncryptionToken []byte
	deviceEncryptionToken []byte
	deviceID              string
	rid                   int64
}

// NewClient creates a MyJDownloader cloud API client.
func NewClient(email, password string) *Client {
	return &Client{
		httpClient:   &http.Client{Timeout: 15 * time.Second},
		email:        email,
		password:     password,
		loginSecret:  deriveLoginSecret(email, password),
		deviceSecret: deriveDeviceSecret(email, password),
	}
}

func (c *Client) nextRid() int64 {
	return atomic.AddInt64(&c.rid, 1) + time.Now().UnixMilli()
}

// Connect authenticates with the MyJDownloader API.
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	query := fmt.Sprintf("/my/connect?email=%s&appkey=%s",
		url.QueryEscape(c.email), url.QueryEscape(appKey))

	var resp connectResponse
	if err := c.callServer(ctx, query, c.loginSecret, &resp); err != nil {
		return fmt.Errorf("connect: %w", err)
	}

	c.sessionToken = resp.SessionToken
	c.regainToken = resp.RegainToken

	var err error
	c.serverEncryptionToken, err = updateEncryptionToken(c.loginSecret, resp.SessionToken)
	if err != nil {
		return fmt.Errorf("deriving server token: %w", err)
	}
	c.deviceEncryptionToken, err = updateEncryptionToken(c.deviceSecret, resp.SessionToken)
	if err != nil {
		return fmt.Errorf("deriving device token: %w", err)
	}

	return nil
}

// Disconnect invalidates the current session.
func (c *Client) Disconnect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sessionToken == "" {
		return nil
	}

	query := fmt.Sprintf("/my/disconnect?sessiontoken=%s",
		url.QueryEscape(c.sessionToken))
	_ = c.callServer(ctx, query, c.serverEncryptionToken, nil)

	c.sessionToken = ""
	c.regainToken = ""
	c.serverEncryptionToken = nil
	c.deviceEncryptionToken = nil
	c.deviceID = ""
	return nil
}

// ListDevices returns all connected JDownloader devices.
func (c *Client) ListDevices(ctx context.Context) ([]Device, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sessionToken == "" {
		return nil, fmt.Errorf("not connected")
	}

	query := fmt.Sprintf("/my/listdevices?sessiontoken=%s",
		url.QueryEscape(c.sessionToken))

	var resp listDevicesResponse
	if err := c.callServer(ctx, query, c.serverEncryptionToken, &resp); err != nil {
		return nil, fmt.Errorf("listing devices: %w", err)
	}
	return resp.List, nil
}

// ensureConnected connects and selects the first available device.
func (c *Client) ensureConnected(ctx context.Context) error {
	c.mu.Lock()
	hasSession := c.sessionToken != "" && c.deviceID != ""
	c.mu.Unlock()

	if hasSession {
		return nil
	}

	if err := c.Connect(ctx); err != nil {
		return err
	}

	devices, err := c.ListDevices(ctx)
	if err != nil {
		return fmt.Errorf("listing devices after connect: %w", err)
	}
	if len(devices) == 0 {
		return fmt.Errorf("no JDownloader devices connected to MyJDownloader")
	}

	c.mu.Lock()
	c.deviceID = devices[0].ID
	c.mu.Unlock()
	return nil
}

// ListPackages returns all download packages in the queue.
func (c *Client) ListPackages(ctx context.Context) ([]DownloadPackage, error) {
	queryParams := map[string]any{
		"bytesTotal":  true,
		"bytesLoaded": true,
		"speed":       true,
		"eta":         true,
		"finished":    true,
		"enabled":     true,
		"status":      true,
		"childCount":  true,
	}
	params := serializeParams(queryParams)

	var packages []DownloadPackage
	if err := c.callAction(ctx, "downloadsV2/queryPackages", params, &packages); err != nil {
		return nil, fmt.Errorf("listing packages: %w", err)
	}
	return packages, nil
}

// AddLinks adds download links to the queue.
func (c *Client) AddLinks(ctx context.Context, urls []string) error {
	params := serializeParams(map[string]any{
		"links": strings.Join(urls, "\n"),
	})
	if err := c.callAction(ctx, "linkgrabberv2/addLinks", params, nil); err != nil {
		return fmt.Errorf("adding links: %w", err)
	}
	return nil
}

// StartQueue starts all downloads.
func (c *Client) StartQueue(ctx context.Context) error {
	if err := c.callAction(ctx, "downloadcontroller/start", nil, nil); err != nil {
		return fmt.Errorf("starting queue: %w", err)
	}
	return nil
}

// PauseQueue pauses all downloads.
func (c *Client) PauseQueue(ctx context.Context) error {
	if err := c.callAction(ctx, "downloadcontroller/pause", serializeParams(true), nil); err != nil {
		return fmt.Errorf("pausing queue: %w", err)
	}
	return nil
}

// CleanupFinished removes all finished packages from the download list (keeps files on disk).
func (c *Client) CleanupFinished(ctx context.Context) (int, error) {
	// First get all packages to find finished ones
	packages, err := c.ListPackages(ctx)
	if err != nil {
		return 0, fmt.Errorf("listing packages: %w", err)
	}

	var finishedIDs []int64
	for _, p := range packages {
		if p.Finished {
			finishedIDs = append(finishedIDs, p.UUID)
		}
	}

	if len(finishedIDs) == 0 {
		return 0, nil
	}

	linkIDs := make([]int64, 0)
	params := serializeParams(linkIDs, finishedIDs)
	if err := c.callAction(ctx, "downloadsV2/removeLinks", params, nil); err != nil {
		return 0, fmt.Errorf("removing finished packages: %w", err)
	}
	return len(finishedIDs), nil
}

// GetSpeed returns the current download speed.
func (c *Client) GetSpeed(ctx context.Context) (int64, error) {
	var speed int64
	if err := c.callAction(ctx, "downloadcontroller/getSpeedInBps", nil, &speed); err != nil {
		return 0, fmt.Errorf("getting speed: %w", err)
	}
	return speed, nil
}

// IsRunning checks if the download queue is running.
func (c *Client) IsRunning(ctx context.Context) (bool, error) {
	var state string
	if err := c.callAction(ctx, "downloadcontroller/getCurrentState", nil, &state); err != nil {
		return false, fmt.Errorf("getting state: %w", err)
	}
	return state == "RUNNING", nil
}

// serializeParams converts each param to a JSON string for the JDownloader API.
func serializeParams(params ...any) []any {
	result := make([]any, len(params))
	for i, p := range params {
		data, _ := json.Marshal(p)
		result[i] = string(data)
	}
	return result
}

// ToPackageInfo converts a DownloadPackage to a frontend PackageInfo.
func ToPackageInfo(p DownloadPackage) PackageInfo {
	progress := float64(0)
	if p.BytesTotal > 0 {
		progress = float64(p.BytesLoaded) / float64(p.BytesTotal) * 100
	}

	status := p.Status
	switch {
	case p.Finished:
		status = "finished"
	case p.Speed > 0:
		status = "downloading"
	case status == "":
		status = "queued"
	}

	return PackageInfo{
		UUID:       p.UUID,
		Name:       p.Name,
		Status:     status,
		Progress:   progress,
		Size:       p.BytesTotal,
		Downloaded: p.BytesLoaded,
		Speed:      p.Speed,
		ETA:        p.ETA,
		Finished:   p.Finished,
		LinkCount:  p.ChildCount,
	}
}

