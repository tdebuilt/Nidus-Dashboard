package arr

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func mockArrServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	mux.HandleFunc("/api/v3/system/status", func(w http.ResponseWriter, r *http.Request) {
		if !checkAPIKey(w, r) {
			return
		}
		json.NewEncoder(w).Encode(SystemStatus{
			Version:      "4.3.1.7361",
			AppName:      "Sonarr",
			InstanceName: "sonarr-main",
		})
	})

	mux.HandleFunc("/api/v3/queue", func(w http.ResponseWriter, r *http.Request) {
		if !checkAPIKey(w, r) {
			return
		}
		json.NewEncoder(w).Encode(QueueResponse{
			Page:         1,
			PageSize:     20,
			TotalRecords: 2,
			Records: []QueueItem{
				{
					ID:                   1,
					Title:                "Some.Show.S01E01",
					Status:               "downloading",
					TrackedDownloadState: "downloading",
					Size:                 1500000000,
					Sizeleft:             500000000,
					Timeleft:             "00:15:00",
				},
				{
					ID:                   2,
					Title:                "Some.Show.S01E02",
					Status:               "queued",
					TrackedDownloadState: "importPending",
					Size:                 1200000000,
					Sizeleft:             1200000000,
					Timeleft:             "01:00:00",
				},
			},
		})
	})

	mux.HandleFunc("/api/v3/calendar", func(w http.ResponseWriter, r *http.Request) {
		if !checkAPIKey(w, r) {
			return
		}
		if r.URL.Query().Get("start") == "" || r.URL.Query().Get("end") == "" {
			http.Error(w, "missing start/end params", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode([]CalendarItem{
			{
				ID:          101,
				Title:       "Episode 5",
				Overview:    "Something happens.",
				AirDateUtc:  "2026-03-25T20:00:00Z",
				HasFile:     false,
				SeriesTitle: "Great Show",
			},
			{
				ID:              201,
				Title:           "Cool Movie",
				InCinemas:       "2026-03-20T00:00:00Z",
				PhysicalRelease: "2026-06-01T00:00:00Z",
				HasFile:         false,
			},
		})
	})

	mux.HandleFunc("/api/v3/series", func(w http.ResponseWriter, r *http.Request) {
		if !checkAPIKey(w, r) {
			return
		}
		items := make([]map[string]any, 5)
		for i := range items {
			items[i] = map[string]any{"id": i + 1, "title": "Series"}
		}
		json.NewEncoder(w).Encode(items)
	})

	return httptest.NewServer(mux)
}

func checkAPIKey(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("X-Api-Key") != "test-api-key-123" {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return false
	}
	return true
}

func TestGetSystemStatus(t *testing.T) {
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	status, err := client.GetSystemStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Version != "4.3.1.7361" {
		t.Fatalf("expected version '4.3.1.7361', got '%s'", status.Version)
	}
	if status.AppName != "Sonarr" {
		t.Fatalf("expected appName 'Sonarr', got '%s'", status.AppName)
	}
	if status.InstanceName != "sonarr-main" {
		t.Fatalf("expected instanceName 'sonarr-main', got '%s'", status.InstanceName)
	}
}

func TestGetQueue(t *testing.T) {
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	queue, err := client.GetQueue(20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if queue.TotalRecords != 2 {
		t.Fatalf("expected 2 total records, got %d", queue.TotalRecords)
	}
	if len(queue.Records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(queue.Records))
	}
	if queue.Records[0].Title != "Some.Show.S01E01" {
		t.Fatalf("expected 'Some.Show.S01E01', got '%s'", queue.Records[0].Title)
	}
	if queue.Records[0].Size != 1500000000 {
		t.Fatalf("expected size 1500000000, got %d", queue.Records[0].Size)
	}
	if queue.Records[1].TrackedDownloadState != "importPending" {
		t.Fatalf("expected 'importPending', got '%s'", queue.Records[1].TrackedDownloadState)
	}
}

func TestGetCalendar(t *testing.T) {
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	start := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)

	items, err := client.GetCalendar(start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 calendar items, got %d", len(items))
	}
	if items[0].SeriesTitle != "Great Show" {
		t.Fatalf("expected 'Great Show', got '%s'", items[0].SeriesTitle)
	}
	if items[0].AirDateUtc != "2026-03-25T20:00:00Z" {
		t.Fatalf("expected '2026-03-25T20:00:00Z', got '%s'", items[0].AirDateUtc)
	}
	if items[1].Title != "Cool Movie" {
		t.Fatalf("expected 'Cool Movie', got '%s'", items[1].Title)
	}
}

func TestGetLibraryCount(t *testing.T) {
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	count, err := client.GetLibraryCount("/series")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected library count 5, got %d", count)
	}
}

func TestApiKeyHeader(t *testing.T) {
	var receivedKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = r.Header.Get("X-Api-Key")
		json.NewEncoder(w).Encode(SystemStatus{Version: "1.0"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "my-secret-key", "v3", srv.Client())

	_, err := client.GetSystemStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedKey != "my-secret-key" {
		t.Fatalf("expected API key 'my-secret-key', got '%s'", receivedKey)
	}
}

func TestUnauthorized(t *testing.T) {
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "wrong-key", "v3", srv.Client())

	_, err := client.GetSystemStatus()
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("expected 'unauthorized' in error, got: %v", err)
	}
}
