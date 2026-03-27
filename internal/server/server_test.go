package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/tdebuilt/nidus/internal/config"
)

func testConfig() config.Config {
	return config.Config{
		Server: config.ServerConfig{
			Port:    0, // will use random port in tests
			BaseURL: "http://localhost",
		},
		Database: config.DatabaseConfig{
			Path: "./data/nidus.db",
		},
	}
}

func testServer() *Server {
	return NewServer("test")
}

func TestHealthEndpoint(t *testing.T) {
	cfg := testConfig()
	srv := testServer()
	r := New(srv, cfg, nil)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}

	ct := resp.Header.Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}
}

func TestSecurityHeaders(t *testing.T) {
	cfg := testConfig()
	srv := testServer()
	r := New(srv, cfg, nil)

	ts := httptest.NewServer(r)
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	tests := []struct {
		header   string
		contains string
	}{
		{"X-Content-Type-Options", "nosniff"},
		{"X-Frame-Options", "DENY"},
		{"Referrer-Policy", "strict-origin-when-cross-origin"},
		{"Content-Security-Policy", "script-src 'self' 'nonce-"},
		{"Content-Security-Policy", "media-src 'self' blob:"},
		{"Content-Security-Policy", "connect-src 'self' ws: wss:"},
	}

	for _, tc := range tests {
		value := resp.Header.Get(tc.header)
		if value == "" {
			t.Errorf("missing header %s", tc.header)
		} else if !strings.Contains(value, tc.contains) {
			t.Errorf("header %s should contain '%s', got '%s'", tc.header, tc.contains, value)
		}
	}
}

func TestStaticFilesServed(t *testing.T) {
	cfg := testConfig()
	srv := testServer()

	// Set up in-memory static files
	srv.StaticFiles = fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>Nidus</html>")},
		"app.js":     &fstest.MapFile{Data: []byte("console.log('hello')")},
	}

	r := New(srv, cfg, nil)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Test index.html at root
	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "<html>Nidus</html>" {
		t.Errorf("expected index.html content, got %s", string(body))
	}

	// Test direct file access
	resp2, err := http.Get(ts.URL + "/app.js")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp2.Body.Close()
	body2, _ := io.ReadAll(resp2.Body)
	if string(body2) != "console.log('hello')" {
		t.Errorf("expected app.js content, got %s", string(body2))
	}
}

func TestSPAFallback(t *testing.T) {
	cfg := testConfig()
	srv := testServer()

	srv.StaticFiles = fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>SPA</html>")},
	}

	r := New(srv, cfg, nil)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// Unknown route should fallback to index.html (SPA routing)
	resp, err := http.Get(ts.URL + "/dashboard/some-page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "<html>SPA</html>" {
		t.Errorf("expected SPA fallback to index.html, got %s", string(body))
	}
}

func TestAPIRouteTakesPrecedence(t *testing.T) {
	cfg := testConfig()
	srv := testServer()

	srv.StaticFiles = fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>SPA</html>")},
	}

	r := New(srv, cfg, nil)
	ts := httptest.NewServer(r)
	defer ts.Close()

	// /api/health should return JSON, not index.html
	resp, err := http.Get(ts.URL + "/api/health")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status ok, got %s", body["status"])
	}
}

func TestGracefulShutdown(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	cfg := config.Config{
		Server: config.ServerConfig{
			Port:    port,
			BaseURL: fmt.Sprintf("http://localhost:%d", port),
		},
		Database: config.DatabaseConfig{
			Path: "./data/nidus.db",
		},
	}

	srv := testServer()
	r := New(srv, cfg, nil)

	// Run server with cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		errCh <- RunWithContext(ctx, srv, cfg, r)
	}()

	// Wait for server to be ready
	addr := fmt.Sprintf("http://localhost:%d", port)
	ready := false
	for i := 0; i < 50; i++ {
		resp, err := http.Get(addr + "/api/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				ready = true
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !ready {
		t.Fatal("server did not start in time")
	}

	// Cancel context to trigger graceful shutdown
	cancel()

	// Wait for Run to return
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("expected nil error from graceful shutdown, got: %v", err)
		}
	case <-time.After(15 * time.Second):
		t.Fatal("server did not shut down in time")
	}
}

func TestServerStartsAndServesRequests(t *testing.T) {
	// Find a free port
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		t.Fatalf("failed to find free port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	cfg := config.Config{
		Server: config.ServerConfig{
			Port:    port,
			BaseURL: fmt.Sprintf("http://localhost:%d", port),
		},
		Database: config.DatabaseConfig{
			Path: "./data/nidus.db",
		},
	}

	srv := testServer()
	srv.StaticFiles = fstest.MapFS{
		"index.html": &fstest.MapFile{Data: []byte("<html>Test</html>")},
	}

	r := New(srv, cfg, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		RunWithContext(ctx, srv, cfg, r)
	}()

	// Wait for server to start
	addr := fmt.Sprintf("http://localhost:%d", port)
	ready := false
	for i := 0; i < 50; i++ {
		resp, err := http.Get(addr + "/api/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == 200 {
				ready = true
				break
			}
		}
		time.Sleep(50 * time.Millisecond)
	}
	if !ready {
		t.Fatal("server did not start in time")
	}

	// Verify health endpoint
	resp, err := http.Get(addr + "/api/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	// Verify static files
	resp2, err := http.Get(addr + "/")
	if err != nil {
		t.Fatalf("static request failed: %v", err)
	}
	defer resp2.Body.Close()
	body, _ := io.ReadAll(resp2.Body)
	if string(body) != "<html>Test</html>" {
		t.Errorf("expected static content, got %s", string(body))
	}

	// Clean up: cancel context
	cancel()
	time.Sleep(200 * time.Millisecond)
}
