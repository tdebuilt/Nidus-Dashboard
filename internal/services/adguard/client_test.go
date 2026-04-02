package adguard

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockAdGuardServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	checkAuth := func(w http.ResponseWriter, r *http.Request) bool {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return false
		}
		return true
	}

	mux.HandleFunc("/control/stats", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}
		stats := Stats{
			NumDNSQueries:       50000,
			NumBlockedFiltering: 12500,
			AvgProcessingTime:   0.005,
			TopBlockedDomains:   []map[string]int{{"ads.example.com": 500}},
		}
		json.NewEncoder(w).Encode(stats)
	})

	mux.HandleFunc("/control/filtering/status", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}
		status := FilteringStatus{
			Enabled:  true,
			Interval: 24,
			Filters: []Filter{
				{ID: 1, Name: "AdGuard Base", Enabled: true, RulesCount: 50000, URL: "https://example.com/filter.txt"},
				{ID: 2, Name: "Tracking", Enabled: true, RulesCount: 20000, URL: "https://example.com/tracking.txt"},
				{ID: 3, Name: "Social", Enabled: false, RulesCount: 5000, URL: "https://example.com/social.txt"},
			},
		}
		json.NewEncoder(w).Encode(status)
	})

	mux.HandleFunc("/control/filtering/config", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(w, r) {
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	return httptest.NewServer(mux)
}

func TestGetStats(t *testing.T) {
	t.Parallel()
	srv := mockAdGuardServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	stats, err := client.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.NumDNSQueries != 50000 {
		t.Fatalf("expected 50000 queries, got %d", stats.NumDNSQueries)
	}
	if stats.NumBlockedFiltering != 12500 {
		t.Fatalf("expected 12500 blocked, got %d", stats.NumBlockedFiltering)
	}
}

func TestGetStatsUnauthorized(t *testing.T) {
	t.Parallel()
	srv := mockAdGuardServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())

	_, err := client.GetStats(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if !errors.Is(err, ErrAuth) {
		t.Fatalf("expected ErrAuth sentinel, got: %v", err)
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Fatalf("expected 'authentication failed' in error, got: %v", err)
	}
}

func TestGetFilteringStatus(t *testing.T) {
	t.Parallel()
	srv := mockAdGuardServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	status, err := client.GetFilteringStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !status.Enabled {
		t.Fatal("expected filtering to be enabled")
	}
	if len(status.Filters) != 3 {
		t.Fatalf("expected 3 filters, got %d", len(status.Filters))
	}
	if status.Filters[0].Name != "AdGuard Base" {
		t.Fatalf("expected 'AdGuard Base', got '%s'", status.Filters[0].Name)
	}
}

func TestSetFilteringEnabled(t *testing.T) {
	t.Parallel()
	srv := mockAdGuardServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	if err := client.SetFilteringEnabled(context.Background(), false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetFilteringEnabledUnauthorized(t *testing.T) {
	t.Parallel()
	srv := mockAdGuardServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())

	err := client.SetFilteringEnabled(context.Background(), false)
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestTrailingSlash(t *testing.T) {
	t.Parallel()
	srv := mockAdGuardServer(t)
	defer srv.Close()

	client := NewClient(srv.URL+"/", srv.Client())
	client.SetCredentials("admin", "secret")

	stats, err := client.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.NumDNSQueries != 50000 {
		t.Fatalf("expected 50000 queries, got %d", stats.NumDNSQueries)
	}
}

func TestNetworkError(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:1", nil)
	client.SetCredentials("admin", "secret")

	_, err := client.GetStats(context.Background())
	if err == nil {
		t.Fatal("expected network error")
	}
}
