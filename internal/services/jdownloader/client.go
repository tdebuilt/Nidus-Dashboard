package jdownloader

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

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
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	query := fmt.Sprintf("/my/connect?email=%s&appkey=%s",
		url.QueryEscape(c.email), url.QueryEscape(appKey))

	var resp connectResponse
	if err := c.callServer(query, c.loginSecret, &resp); err != nil {
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
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sessionToken == "" {
		return nil
	}

	query := fmt.Sprintf("/my/disconnect?sessiontoken=%s",
		url.QueryEscape(c.sessionToken))
	_ = c.callServer(query, c.serverEncryptionToken, nil)

	c.sessionToken = ""
	c.regainToken = ""
	c.serverEncryptionToken = nil
	c.deviceEncryptionToken = nil
	c.deviceID = ""
	return nil
}

// ListDevices returns all connected JDownloader devices.
func (c *Client) ListDevices() ([]Device, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sessionToken == "" {
		return nil, fmt.Errorf("not connected")
	}

	query := fmt.Sprintf("/my/listdevices?sessiontoken=%s",
		url.QueryEscape(c.sessionToken))

	var resp listDevicesResponse
	if err := c.callServer(query, c.serverEncryptionToken, &resp); err != nil {
		return nil, fmt.Errorf("listing devices: %w", err)
	}
	return resp.List, nil
}

// ensureConnected connects and selects the first available device.
func (c *Client) ensureConnected() error {
	c.mu.Lock()
	hasSession := c.sessionToken != "" && c.deviceID != ""
	c.mu.Unlock()

	if hasSession {
		return nil
	}

	if err := c.Connect(); err != nil {
		return err
	}

	devices, err := c.ListDevices()
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
func (c *Client) ListPackages() ([]DownloadPackage, error) {
	queryParams := map[string]any{
		"bytesTotal":  true,
		"bytesLoaded": true,
		"speed":       true,
		"eta":         true,
		"finished":    true,
		"enabled":     true,
		"status":      true,
		"childCount":  true,
		"saveTo":      true,
		"comment":     true,
	}
	params := serializeParams(queryParams)

	var packages []DownloadPackage
	if err := c.callAction("downloadsV2/queryPackages", params, &packages); err != nil {
		return nil, fmt.Errorf("listing packages: %w", err)
	}
	return packages, nil
}

// AddLinks adds download links to the queue.
func (c *Client) AddLinks(urls []string) error {
	params := serializeParams(map[string]any{
		"links": strings.Join(urls, "\n"),
	})
	if err := c.callAction("linkgrabberv2/addLinks", params, nil); err != nil {
		return fmt.Errorf("adding links: %w", err)
	}
	return nil
}

// StartQueue starts all downloads.
func (c *Client) StartQueue() error {
	if err := c.callAction("downloadcontroller/start", nil, nil); err != nil {
		return fmt.Errorf("starting queue: %w", err)
	}
	return nil
}

// PauseQueue pauses all downloads.
func (c *Client) PauseQueue() error {
	if err := c.callAction("downloadcontroller/pause", serializeParams(true), nil); err != nil {
		return fmt.Errorf("pausing queue: %w", err)
	}
	return nil
}

// CleanupFinished removes all finished packages from the download list (keeps files on disk).
func (c *Client) CleanupFinished() (int, error) {
	// First get all packages to find finished ones
	packages, err := c.ListPackages()
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
	if err := c.callAction("downloadsV2/removeLinks", params, nil); err != nil {
		return 0, fmt.Errorf("removing finished packages: %w", err)
	}
	return len(finishedIDs), nil
}

// GetSpeed returns the current download speed.
func (c *Client) GetSpeed() (int64, error) {
	var speed int64
	if err := c.callAction("downloadcontroller/getSpeedInBps", nil, &speed); err != nil {
		return 0, fmt.Errorf("getting speed: %w", err)
	}
	return speed, nil
}

// IsRunning checks if the download queue is running.
func (c *Client) IsRunning() (bool, error) {
	var state string
	if err := c.callAction("downloadcontroller/getCurrentState", nil, &state); err != nil {
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

// callAction calls a device action through the MyJDownloader cloud relay.
// It auto-connects if no session exists and retries once on auth failure.
func (c *Client) callAction(action string, params []any, result any) error {
	if err := c.ensureConnected(); err != nil {
		return err
	}

	err := c.doDeviceCall(action, params, result)
	if err == nil {
		return nil
	}

	// Retry once after reconnect
	c.mu.Lock()
	c.sessionToken = ""
	c.deviceID = ""
	c.mu.Unlock()

	if err := c.ensureConnected(); err != nil {
		return fmt.Errorf("reconnect failed: %w", err)
	}
	return c.doDeviceCall(action, params, result)
}

func (c *Client) doDeviceCall(action string, params []any, result any) error {
	c.mu.Lock()
	session := c.sessionToken
	device := c.deviceID
	devToken := c.deviceEncryptionToken
	c.mu.Unlock()

	rid := c.nextRid()

	actionPath := "/" + action
	httpPath := fmt.Sprintf("/t_%s_%s%s",
		url.QueryEscape(session), url.QueryEscape(device), actionPath)

	body := map[string]any{
		"url":    actionPath,
		"rid":    rid,
		"apiVer": 1,
	}
	if params != nil {
		body["params"] = params
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling body: %w", err)
	}

	encBody, err := encrypt(jsonBody, devToken)
	if err != nil {
		return fmt.Errorf("encrypting body: %w", err)
	}
	b64Body := base64.StdEncoding.EncodeToString(encBody)

	reqURL := apiURL + httpPath
	req, err := http.NewRequest(http.MethodPost, reqURL, strings.NewReader(b64Body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/aesjson-jd; charset=utf-8")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// Try to decrypt error response
		if ciphertext, err := base64.StdEncoding.DecodeString(string(respBytes)); err == nil {
			if plaintext, err := decrypt(ciphertext, devToken); err == nil {
				return fmt.Errorf("status %d: %s", resp.StatusCode, string(plaintext))
			}
		}
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(respBytes))
	}

	if result == nil {
		return nil
	}

	return c.decryptDeviceResponse(respBytes, devToken, result)
}

func (c *Client) decryptDeviceResponse(respBytes, token []byte, result any) error {
	ciphertext, err := base64.StdEncoding.DecodeString(string(respBytes))
	if err != nil {
		return fmt.Errorf("base64 decoding response: %w", err)
	}

	plaintext, err := decrypt(ciphertext, token)
	if err != nil {
		return fmt.Errorf("decrypting response: %w", err)
	}

	var apiResp struct {
		Data json.RawMessage `json:"data"`
		Rid  int64           `json:"rid"`
	}
	if err := json.Unmarshal(plaintext, &apiResp); err != nil {
		return fmt.Errorf("decoding response JSON: %w", err)
	}

	if apiResp.Data != nil {
		if err := json.Unmarshal(apiResp.Data, result); err != nil {
			return fmt.Errorf("decoding data: %w", err)
		}
	}
	return nil
}

// callServer makes a signed server call (connect, listdevices, disconnect).
func (c *Client) callServer(query string, secret []byte, result any) error {
	rid := c.nextRid()
	query += fmt.Sprintf("&rid=%d", rid)
	signature := sign(secret, query)
	query += "&signature=" + signature

	req, err := http.NewRequest(http.MethodPost, apiURL+query, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
	}

	if result == nil {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if len(body) == 0 {
		return nil
	}

	ciphertext, err := base64.StdEncoding.DecodeString(string(body))
	if err != nil {
		return fmt.Errorf("base64 decoding: %w", err)
	}

	plaintext, err := decrypt(ciphertext, secret)
	if err != nil {
		return fmt.Errorf("decrypting response: %w", err)
	}

	if err := json.Unmarshal(plaintext, result); err != nil {
		return fmt.Errorf("decoding JSON: %w", err)
	}
	return nil
}
