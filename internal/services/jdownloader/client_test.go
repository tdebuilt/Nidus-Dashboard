package jdownloader

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// mockCloudServer simulates the MyJDownloader cloud API with real crypto.
func mockCloudServer(t *testing.T, email, password string) *httptest.Server {
	t.Helper()

	loginSec := deriveLoginSecret(email, password)
	deviceSec := deriveDeviceSecret(email, password)

	sessionToken := "abcdef0123456789abcdef0123456789"
	deviceID := "test-device-001"

	serverToken, _ := updateEncryptionToken(loginSec, sessionToken)
	deviceToken, _ := updateEncryptionToken(deviceSec, sessionToken)

	mux := http.NewServeMux()

	mux.HandleFunc("/my/connect", func(w http.ResponseWriter, r *http.Request) {
		resp := connectResponse{
			SessionToken: sessionToken,
			RegainToken:  "regain-token-placeholder",
		}
		writeEncryptedResponse(t, w, resp, loginSec)
	})

	mux.HandleFunc("/my/listdevices", func(w http.ResponseWriter, r *http.Request) {
		resp := listDevicesResponse{
			List: []Device{
				{ID: deviceID, Name: "Test JDownloader", Status: "CONNECTED"},
			},
		}
		writeEncryptedResponse(t, w, resp, serverToken)
	})

	mux.HandleFunc("/my/disconnect", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Device call handler — matches /t_{session}_{device}/*
	mux.HandleFunc("/t_"+url.QueryEscape(sessionToken)+"_"+url.QueryEscape(deviceID)+"/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "read error", 500)
			return
		}

		ciphertext, err := base64.StdEncoding.DecodeString(string(body))
		if err != nil {
			http.Error(w, "base64 error", 400)
			return
		}

		plaintext, err := decrypt(ciphertext, deviceToken)
		if err != nil {
			http.Error(w, "decrypt error", 400)
			return
		}

		var req map[string]any
		if err := json.Unmarshal(plaintext, &req); err != nil {
			http.Error(w, "json error", 400)
			return
		}

		urlStr, _ := req["url"].(string)

		var data any
		switch {
		case strings.Contains(urlStr, "queryPackages"):
			data = []DownloadPackage{
				{UUID: 1001, Name: "Test Package", BytesTotal: 1000000, BytesLoaded: 500000, Speed: 50000},
			}
		case strings.Contains(urlStr, "getSpeedInBps"):
			data = int64(50000)
		case strings.Contains(urlStr, "getCurrentState"):
			data = "RUNNING"
		case strings.Contains(urlStr, "addLinks"):
			data = map[string]any{}
		case strings.Contains(urlStr, "start"), strings.Contains(urlStr, "pause"):
			data = true
		default:
			data = nil
		}

		rid, _ := req["rid"].(float64)
		resp := map[string]any{"data": data, "rid": int64(rid)}
		writeEncryptedResponse(t, w, resp, deviceToken)
	})

	return httptest.NewServer(mux)
}

func writeEncryptedResponse(t *testing.T, w http.ResponseWriter, data any, token []byte) {
	t.Helper()
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, "marshal error", 500)
		return
	}
	ciphertext, err := encrypt(jsonBytes, token)
	if err != nil {
		http.Error(w, "encrypt error", 500)
		return
	}
	w.Write([]byte(base64.StdEncoding.EncodeToString(ciphertext)))
}

func newTestClient(t *testing.T, serverURL, email, password string) *Client {
	t.Helper()
	c := NewClient(email, password)
	c.httpClient = &http.Client{Timeout: 5 * 1e9}
	// Override apiURL by patching the client's request building
	// We do this by wrapping the transport to rewrite URLs
	c.httpClient.Transport = &urlRewriter{base: serverURL}
	return c
}

// urlRewriter intercepts requests and rewrites api.jdownloader.org to the test server.
type urlRewriter struct {
	base string
}

func (u *urlRewriter) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = "http"
	req.URL.Host = strings.TrimPrefix(u.base, "http://")
	return http.DefaultTransport.RoundTrip(req)
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	c := NewClient("test@test.com", "password")
	if c.email != "test@test.com" {
		t.Fatal("email not set")
	}
	if len(c.loginSecret) != 32 {
		t.Fatal("loginSecret not derived")
	}
	if len(c.deviceSecret) != 32 {
		t.Fatal("deviceSecret not derived")
	}
}

func TestConnect(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if c.sessionToken == "" {
		t.Fatal("sessionToken not set after Connect")
	}
	if c.regainToken == "" {
		t.Fatal("regainToken not set after Connect")
	}
	if c.serverEncryptionToken == nil {
		t.Fatal("serverEncryptionToken not derived")
	}
	if c.deviceEncryptionToken == nil {
		t.Fatal("deviceEncryptionToken not derived")
	}
}

func TestListDevices(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	devices, err := c.ListDevices(context.Background())
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}
	if len(devices) != 1 {
		t.Fatalf("expected 1 device, got %d", len(devices))
	}
	if devices[0].Name != "Test JDownloader" {
		t.Fatalf("unexpected device name: %s", devices[0].Name)
	}
}

func TestDisconnect(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	if err := c.Connect(context.Background()); err != nil {
		t.Fatalf("Connect failed: %v", err)
	}

	if err := c.Disconnect(context.Background()); err != nil {
		t.Fatalf("Disconnect failed: %v", err)
	}

	if c.sessionToken != "" {
		t.Fatal("sessionToken should be empty after Disconnect")
	}
}

func TestListPackages(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	packages, err := c.ListPackages(context.Background())
	if err != nil {
		t.Fatalf("ListPackages failed: %v", err)
	}
	if len(packages) != 1 {
		t.Fatalf("expected 1 package, got %d", len(packages))
	}
	if packages[0].Name != "Test Package" {
		t.Fatalf("unexpected package name: %s", packages[0].Name)
	}
	if packages[0].BytesTotal != 1000000 {
		t.Fatalf("unexpected bytes total: %d", packages[0].BytesTotal)
	}
}

func TestGetSpeed(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	speed, err := c.GetSpeed(context.Background())
	if err != nil {
		t.Fatalf("GetSpeed failed: %v", err)
	}
	if speed != 50000 {
		t.Fatalf("expected speed 50000, got %d", speed)
	}
}

func TestIsRunning(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	running, err := c.IsRunning(context.Background())
	if err != nil {
		t.Fatalf("IsRunning failed: %v", err)
	}
	if !running {
		t.Fatal("expected queue to be running")
	}
}

func TestAddLinks(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	err := c.AddLinks(context.Background(), []string{"https://example.com/file.zip"})
	if err != nil {
		t.Fatalf("AddLinks failed: %v", err)
	}
}

func TestStartQueue(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	if err := c.StartQueue(context.Background()); err != nil {
		t.Fatalf("StartQueue failed: %v", err)
	}
}

func TestPauseQueue(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	if err := c.PauseQueue(context.Background()); err != nil {
		t.Fatalf("PauseQueue failed: %v", err)
	}
}

func TestEnsureConnectedAutoSelectDevice(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	if err := c.ensureConnected(context.Background()); err != nil {
		t.Fatalf("ensureConnected failed: %v", err)
	}

	if c.deviceID == "" {
		t.Fatal("deviceID not set after ensureConnected")
	}
	if c.sessionToken == "" {
		t.Fatal("sessionToken not set after ensureConnected")
	}
}

func TestReconnectOnError(t *testing.T) {
	t.Parallel()
	email, password := "user@example.com", "secret123"
	srv := mockCloudServer(t, email, password)
	defer srv.Close()

	c := newTestClient(t, srv.URL, email, password)

	ctx := context.Background()

	// First call connects and works
	_, err := c.ListPackages(ctx)
	if err != nil {
		t.Fatalf("first ListPackages failed: %v", err)
	}

	// Invalidate session to force reconnect
	c.mu.Lock()
	c.sessionToken = "invalid-session"
	c.deviceID = "invalid-device"
	c.mu.Unlock()

	// callAction should reconnect and succeed
	_, err = c.ListPackages(ctx)
	if err != nil {
		t.Fatalf("ListPackages after invalidation failed: %v", err)
	}
}

func TestToPackageInfo(t *testing.T) {
	t.Parallel()
	pkg := DownloadPackage{
		UUID:        1001,
		Name:        "Test Package",
		BytesTotal:  1000000,
		BytesLoaded: 500000,
		Speed:       100000,
		ETA:         5,
		Finished:    false,
		ChildCount:  3,
	}

	info := ToPackageInfo(pkg)
	if info.Progress != 50.0 {
		t.Fatalf("expected progress 50.0, got %f", info.Progress)
	}
	if info.Status != "downloading" {
		t.Fatalf("expected status 'downloading', got '%s'", info.Status)
	}
	if info.LinkCount != 3 {
		t.Fatalf("expected 3 links, got %d", info.LinkCount)
	}
}

func TestToPackageInfoFinished(t *testing.T) {
	t.Parallel()
	pkg := DownloadPackage{
		UUID:        1002,
		Name:        "Done Package",
		BytesTotal:  1000,
		BytesLoaded: 1000,
		Finished:    true,
	}

	info := ToPackageInfo(pkg)
	if info.Status != "finished" {
		t.Fatalf("expected status 'finished', got '%s'", info.Status)
	}
}
