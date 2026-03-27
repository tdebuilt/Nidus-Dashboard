package reolink

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient_DefaultHTTPClient(t *testing.T) {
	t.Parallel()
	c := NewClient("192.168.1.1", "admin", "pass", 0, nil, true)
	if c.httpClient == nil {
		t.Fatal("expected default HTTP client to be created, got nil")
	}
	if c.httpClient.Timeout == 0 {
		t.Error("expected non-zero timeout on default HTTP client")
	}
}

func TestNewClient_CustomHTTPClient(t *testing.T) {
	t.Parallel()
	custom := &http.Client{}
	c := NewClient("192.168.1.1", "admin", "pass", 0, custom, false)
	if c.httpClient != custom {
		t.Error("expected custom HTTP client to be used")
	}
}

func TestIsJPEG(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{
			name: "valid JPEG header",
			data: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10},
			want: true,
		},
		{
			name: "minimal valid JPEG",
			data: []byte{0xFF, 0xD8},
			want: true,
		},
		{
			name: "non-JPEG data",
			data: []byte{0x89, 0x50, 0x4E, 0x47}, // PNG header
			want: false,
		},
		{
			name: "empty data",
			data: []byte{},
			want: false,
		},
		{
			name: "single byte",
			data: []byte{0xFF},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := isJPEG(tt.data)
			if got != tt.want {
				t.Errorf("isJPEG(%v) = %v, want %v", tt.data, got, tt.want)
			}
		})
	}
}

func TestFormatRTSPURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		username   string
		password   string
		ip         string
		channel    int
		streamType string
		want       string
	}{
		{
			name:       "default stream type",
			username:   "admin",
			password:   "pass123",
			ip:         "192.168.1.100",
			channel:    0,
			streamType: "",
			want:       "rtsp://admin:pass123@192.168.1.100/Preview_01_main",
		},
		{
			name:       "main stream explicit",
			username:   "admin",
			password:   "pass",
			ip:         "10.0.0.5",
			channel:    0,
			streamType: "main",
			want:       "rtsp://admin:pass@10.0.0.5/Preview_01_main",
		},
		{
			name:       "sub stream",
			username:   "user",
			password:   "secret",
			ip:         "172.16.0.10",
			channel:    0,
			streamType: "sub",
			want:       "rtsp://user:secret@172.16.0.10/Preview_01_sub",
		},
		{
			name:       "channel 1 formats as 02",
			username:   "admin",
			password:   "pass",
			ip:         "192.168.1.1",
			channel:    1,
			streamType: "main",
			want:       "rtsp://admin:pass@192.168.1.1/Preview_02_main",
		},
		{
			name:       "channel 9 formats as 10",
			username:   "admin",
			password:   "pass",
			ip:         "192.168.1.1",
			channel:    9,
			streamType: "main",
			want:       "rtsp://admin:pass@192.168.1.1/Preview_10_main",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := FormatRTSPURL(tt.username, tt.password, tt.ip, tt.channel, tt.streamType)
			if got != tt.want {
				t.Errorf("FormatRTSPURL() = %q, want %q", got, tt.want)
			}
		})
	}
}

var mockJPEG = []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}

func TestGetSnapshot_DirectAuth(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/cgi-bin/api.cgi") {
			http.NotFound(w, r)
			return
		}
		cmd := r.URL.Query().Get("cmd")
		user := r.URL.Query().Get("user")
		password := r.URL.Query().Get("password")

		if cmd == "Snap" && user == "admin" && password == "pass" {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(mockJPEG)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	// Extract host from server URL (strip http://)
	host := strings.TrimPrefix(server.URL, "http://")

	c := NewClient(host, "admin", "pass", 0, server.Client(), false)
	data, contentType, err := c.GetSnapshot(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "image/jpeg" {
		t.Errorf("expected content-type 'image/jpeg', got %q", contentType)
	}
	if !isJPEG(data) {
		t.Error("expected JPEG data")
	}
}

func TestGetSnapshot_TokenAuth(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cmd := r.URL.Query().Get("cmd")

		if cmd == "Login" {
			resp := []struct {
				Code  int `json:"code"`
				Value struct {
					Token struct {
						Name string `json:"name"`
					} `json:"Token"`
				} `json:"value"`
			}{
				{
					Code: 0,
				},
			}
			resp[0].Value.Token.Name = "abc123"
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}

		if cmd == "Snap" {
			token := r.URL.Query().Get("token")
			user := r.URL.Query().Get("user")

			// Direct auth returns non-JPEG (simulating a camera that needs token auth)
			if user != "" {
				w.Write([]byte(`{"error": "need token auth"}`))
				return
			}

			// Token auth returns JPEG
			if token == "abc123" {
				w.Header().Set("Content-Type", "image/jpeg")
				w.Write(mockJPEG)
				return
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	c := NewClient(host, "admin", "pass", 0, server.Client(), false)

	data, contentType, err := c.GetSnapshot(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "image/jpeg" {
		t.Errorf("expected content-type 'image/jpeg', got %q", contentType)
	}
	if !isJPEG(data) {
		t.Error("expected JPEG data")
	}
}

func TestGetSnapshot_SchemeDiscovery(t *testing.T) {
	t.Parallel()
	// Server only responds on the scheme it was started on (http via httptest).
	// The client should discover the working scheme.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cmd := r.URL.Query().Get("cmd")
		if cmd == "Snap" {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(mockJPEG)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	c := NewClient(host, "admin", "pass", 0, server.Client(), false)
	// Clear scheme to force discovery
	c.scheme = ""

	data, contentType, err := c.GetSnapshot(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if contentType != "image/jpeg" {
		t.Errorf("expected 'image/jpeg', got %q", contentType)
	}
	if !isJPEG(data) {
		t.Error("expected JPEG data")
	}

	// After successful snapshot, scheme should be cached
	c.mu.Lock()
	cachedScheme := c.scheme
	c.mu.Unlock()
	if cachedScheme != "http" {
		t.Errorf("expected cached scheme 'http', got %q", cachedScheme)
	}
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cmd") != "Login" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `[{"code":0,"value":{"Token":{"name":"abc123"}}}]`)
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	c := NewClient(host, "admin", "pass", 0, server.Client(), false)

	token, err := c.login(context.Background(), "http")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "abc123" {
		t.Errorf("expected token 'abc123', got %q", token)
	}
}

func TestLogin_Failure(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cmd") != "Login" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		// Return error code
		fmt.Fprint(w, `[{"code":1,"value":{"Token":{"name":""}}}]`)
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	c := NewClient(host, "admin", "wrongpass", 0, server.Client(), false)

	_, err := c.login(context.Background(), "http")
	if err == nil {
		t.Fatal("expected error for failed login, got nil")
	}
	if !strings.Contains(err.Error(), "login failed") {
		t.Errorf("expected 'login failed' in error, got %q", err.Error())
	}
}
