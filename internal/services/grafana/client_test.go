package grafana

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mockGrafanaServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(HealthResponse{
			Commit:   "abc123",
			Database: "ok",
			Version:  "10.0.0",
		})
	})

	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode([]DashboardSearchResult{
			{UID: "abc", Title: "Dashboard 1", Type: "dash-db"},
			{UID: "def", Title: "Dashboard 2", Type: "dash-db"},
		})
	})

	mux.HandleFunc("/api/dashboards/uid/abc", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(DashboardDetail{
			Dashboard: DashboardData{
				UID:   "abc",
				Title: "Dashboard 1",
				Panels: []Panel{
					{ID: 1, Title: "CPU", Type: "graph"},
					{ID: 2, Title: "Row", Type: "row", Panels: []Panel{
						{ID: 3, Title: "Memory", Type: "gauge"},
					}},
				},
			},
			Meta: DashboardMeta{Slug: "dashboard-1"},
		})
	})

	return httptest.NewServer(mux)
}

func TestGetHealth(t *testing.T) {
	srv := mockGrafanaServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, nil)
	health, err := client.GetHealth()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Version != "10.0.0" {
		t.Errorf("expected version 10.0.0, got %s", health.Version)
	}
}

func TestSearchDashboards(t *testing.T) {
	srv := mockGrafanaServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, nil)
	client.SetToken("test-token")

	results, err := client.SearchDashboards()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 dashboards, got %d", len(results))
	}
}

func TestSearchDashboardsUnauthorized(t *testing.T) {
	srv := mockGrafanaServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, nil)
	// No token set
	_, err := client.SearchDashboards()
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestGetDashboard(t *testing.T) {
	srv := mockGrafanaServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, nil)
	client.SetToken("test-token")

	detail, err := client.GetDashboard("abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if detail.Dashboard.Title != "Dashboard 1" {
		t.Errorf("expected title Dashboard 1, got %s", detail.Dashboard.Title)
	}
	if len(detail.Dashboard.Panels) != 2 {
		t.Errorf("expected 2 top-level panels, got %d", len(detail.Dashboard.Panels))
	}
}

func TestTrailingSlash(t *testing.T) {
	srv := mockGrafanaServer(t)
	defer srv.Close()

	client := NewClient(srv.URL+"/", nil)
	health, err := client.GetHealth()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if health.Database != "ok" {
		t.Errorf("expected database ok, got %s", health.Database)
	}
}
