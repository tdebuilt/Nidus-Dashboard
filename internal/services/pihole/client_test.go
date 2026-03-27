package pihole

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const (
	testSID      = "test-sid-abc123"
	testCSRF     = "test-csrf-xyz789"
	testPassword = "pihole-secret"
	testValidity = 300
)

func mockPiholeServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	authCount := &atomic.Int32{}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var authReq AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&authReq); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		if authReq.Password != testPassword {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(AuthResponse{
				Session: SessionInfo{Valid: false},
			})
			return
		}

		authCount.Add(1)
		json.NewEncoder(w).Encode(AuthResponse{
			Session: SessionInfo{
				Valid:    true,
				SID:      testSID,
				CSRF:     testCSRF,
				Validity: testValidity,
			},
		})
	})

	mux.HandleFunc("/api/stats/summary", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-FTL-SID") != testSID {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(StatsResponse{
			Queries: QueriesStats{
				Total:          150000,
				Blocked:        45000,
				PercentBlocked: 30.0,
				UniqueDomains:  8500,
				Forwarded:      75000,
				Cached:         30000,
			},
			Took: 0.012,
		})
	})

	mux.HandleFunc("/api/dns/blocking", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-FTL-SID") != testSID {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodGet:
			json.NewEncoder(w).Encode(BlockingResponse{Blocking: true})
		case http.MethodPost:
			var body map[string]bool
			json.NewDecoder(r.Body).Decode(&body)
			json.NewEncoder(w).Encode(BlockingResponse{Blocking: body["blocking"]})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux), authCount
}

func TestNewClient(t *testing.T) {
	t.Parallel()
	client := NewClient("http://pihole.local", "secret", nil)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.baseURL != "http://pihole.local" {
		t.Fatalf("expected baseURL 'http://pihole.local', got '%s'", client.baseURL)
	}
	if client.password != "secret" {
		t.Fatalf("expected password 'secret', got '%s'", client.password)
	}
	if client.httpClient == nil {
		t.Fatal("expected non-nil httpClient")
	}
}

func TestNewClientTrailingSlash(t *testing.T) {
	t.Parallel()
	client := NewClient("http://pihole.local/", "secret", nil)
	if client.baseURL != "http://pihole.local" {
		t.Fatalf("expected trailing slash stripped, got '%s'", client.baseURL)
	}
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()
	srv, authCount := mockPiholeServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, testPassword, srv.Client())

	client.mu.Lock()
	err := client.authenticate(context.Background())
	client.mu.Unlock()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.sid != testSID {
		t.Fatalf("expected SID '%s', got '%s'", testSID, client.sid)
	}
	if client.csrf != testCSRF {
		t.Fatalf("expected CSRF '%s', got '%s'", testCSRF, client.csrf)
	}
	if authCount.Load() != 1 {
		t.Fatalf("expected 1 auth call, got %d", authCount.Load())
	}
}

func TestGetStats(t *testing.T) {
	t.Parallel()
	srv, _ := mockPiholeServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, testPassword, srv.Client())

	stats, err := client.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.TotalQueries != 150000 {
		t.Fatalf("expected 150000 total queries, got %d", stats.TotalQueries)
	}
	if stats.BlockedQueries != 45000 {
		t.Fatalf("expected 45000 blocked queries, got %d", stats.BlockedQueries)
	}
	if stats.BlockedPercent != 30.0 {
		t.Fatalf("expected 30.0%% blocked, got %f", stats.BlockedPercent)
	}
	if stats.UniqueDomains != 8500 {
		t.Fatalf("expected 8500 unique domains, got %d", stats.UniqueDomains)
	}
	if stats.CachedQueries != 30000 {
		t.Fatalf("expected 30000 cached queries, got %d", stats.CachedQueries)
	}
	if stats.ForwardedQueries != 75000 {
		t.Fatalf("expected 75000 forwarded queries, got %d", stats.ForwardedQueries)
	}
	if !stats.BlockingEnabled {
		t.Fatal("expected blocking to be enabled")
	}
}

func TestSetBlocking(t *testing.T) {
	t.Parallel()
	srv, _ := mockPiholeServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, testPassword, srv.Client())

	if err := client.SetBlocking(context.Background(), false); err != nil {
		t.Fatalf("unexpected error disabling blocking: %v", err)
	}

	if err := client.SetBlocking(context.Background(), true); err != nil {
		t.Fatalf("unexpected error enabling blocking: %v", err)
	}
}

func TestUnauthorized(t *testing.T) {
	t.Parallel()
	srv, _ := mockPiholeServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "wrong-password", srv.Client())

	_, err := client.GetStats(context.Background())
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("expected 'unauthorized' in error, got: %v", err)
	}
}

func TestSessionReuse(t *testing.T) {
	t.Parallel()
	srv, authCount := mockPiholeServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, testPassword, srv.Client())

	// First call triggers authentication
	_, err := client.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error on first call: %v", err)
	}

	// Second call should reuse the session
	_, err = client.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}

	if authCount.Load() != 1 {
		t.Fatalf("expected 1 auth call (session reused), got %d", authCount.Load())
	}
}

func TestNetworkError(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:1", "secret", nil)

	_, err := client.GetStats(context.Background())
	if err == nil {
		t.Fatal("expected network error")
	}
}
