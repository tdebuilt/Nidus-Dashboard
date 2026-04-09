package qbittorrent

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

const testSID = "test-sid-abc123"

func mockQBittorrentServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	requestCount := &atomic.Int32{}

	mux := http.NewServeMux()

	// Auth endpoint
	mux.HandleFunc("/api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		if r.FormValue("username") != "admin" || r.FormValue("password") != "secret" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Fails."))
			return
		}
		http.SetCookie(w, &http.Cookie{Name: "SID", Value: testSID})
		w.Write([]byte("Ok."))
	})

	// Check SID cookie helper
	checkSID := func(w http.ResponseWriter, r *http.Request) bool {
		cookie, err := r.Cookie("SID")
		if err != nil || cookie.Value != testSID {
			w.WriteHeader(http.StatusForbidden)
			return false
		}
		return true
	}

	// Torrents info
	mux.HandleFunc("/api/v2/torrents/info", func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		if !checkSID(w, r) {
			return
		}
		torrents := []Torrent{
			{
				Hash:      "abc123def456",
				Name:      "Ubuntu 24.04",
				Size:      4000000000,
				Progress:  0.75,
				Dlspeed:   5000000,
				Upspeed:   1000000,
				ETA:       200,
				Ratio:     0.5,
				State:     "downloading",
				AddedOn:   1700000000,
				NumSeeds:  42,
				NumLeechs: 10,
				Category:  "linux",
			},
			{
				Hash:     "fed654cba321",
				Name:     "Debian 12",
				Size:     3000000000,
				Progress: 1.0,
				Upspeed:  500000,
				Ratio:    2.5,
				State:    "uploading",
				AddedOn:  1699000000,
			},
		}
		json.NewEncoder(w).Encode(torrents)
	})

	// Transfer info
	mux.HandleFunc("/api/v2/transfer/info", func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		if !checkSID(w, r) {
			return
		}
		json.NewEncoder(w).Encode(TransferInfo{
			DlInfoSpeed: 5000000,
			UpInfoSpeed: 1500000,
		})
	})

	// Action endpoints
	for _, path := range []string{"/api/v2/torrents/resume", "/api/v2/torrents/pause", "/api/v2/torrents/delete", "/api/v2/torrents/add"} {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			if !checkSID(w, r) {
				return
			}
			w.WriteHeader(http.StatusOK)
		})
	}

	// Categories endpoint
	mux.HandleFunc("/api/v2/torrents/categories", func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		if !checkSID(w, r) {
			return
		}
		json.NewEncoder(w).Encode(map[string]Category{
			"linux": {Name: "linux", SavePath: "/downloads/linux"},
			"tv":    {Name: "tv", SavePath: "/downloads/tv"},
		})
	})

	return httptest.NewServer(mux), requestCount
}

func TestAuthentication(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	torrents, err := client.ListTorrents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(torrents) != 2 {
		t.Fatalf("expected 2 torrents, got %d", len(torrents))
	}
}

func TestAuthRetryOn403(t *testing.T) {
	t.Parallel()

	callCount := &atomic.Int32{}
	authCount := &atomic.Int32{}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		authCount.Add(1)
		http.SetCookie(w, &http.Cookie{Name: "SID", Value: "new-sid"})
		w.Write([]byte("Ok."))
	})
	mux.HandleFunc("/api/v2/torrents/info", func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		cookie, _ := r.Cookie("SID")
		// First call with old SID fails, retry with new SID works
		if cookie != nil && cookie.Value == "new-sid" && callCount.Load() > 1 {
			json.NewEncoder(w).Encode([]Torrent{})
			return
		}
		w.WriteHeader(http.StatusForbidden)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	_, err := client.ListTorrents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if authCount.Load() < 2 {
		t.Fatalf("expected at least 2 auth calls (initial + retry), got %d", authCount.Load())
	}
}

func TestListTorrents(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	torrents, err := client.ListTorrents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(torrents) != 2 {
		t.Fatalf("expected 2 torrents, got %d", len(torrents))
	}
	if torrents[0].Name != "Ubuntu 24.04" {
		t.Fatalf("expected 'Ubuntu 24.04', got '%s'", torrents[0].Name)
	}
	if torrents[0].Progress != 0.75 {
		t.Fatalf("expected 0.75 progress, got %f", torrents[0].Progress)
	}
}

func TestGetTransferInfo(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	info, err := client.GetTransferInfo(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.DlInfoSpeed != 5000000 {
		t.Fatalf("expected dl speed 5000000, got %d", info.DlInfoSpeed)
	}
	if info.UpInfoSpeed != 1500000 {
		t.Fatalf("expected up speed 1500000, got %d", info.UpInfoSpeed)
	}
}

func TestResumeTorrents(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	if err := client.ResumeTorrents(context.Background(), []string{"abc123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPauseTorrents(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	if err := client.PauseTorrents(context.Background(), []string{"abc123"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteTorrents(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	if err := client.DeleteTorrents(context.Background(), []string{"abc123"}, true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddTorrentMagnet(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	err := client.AddTorrent(context.Background(), AddOptions{URL: "magnet:?xt=urn:btih:test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAddTorrentWithCategoryAndSavePath(t *testing.T) {
	t.Parallel()

	var gotBody string
	var gotContentType string
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "SID", Value: testSID})
		w.Write([]byte("Ok."))
	})
	mux.HandleFunc("/api/v2/torrents/add", func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		gotBody = string(body)
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	err := client.AddTorrent(context.Background(), AddOptions{
		URL:      "magnet:?xt=urn:btih:test",
		Category: "linux",
		SavePath: "/mnt/linux",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(gotContentType, "application/x-www-form-urlencoded") {
		t.Fatalf("expected form content-type, got %q", gotContentType)
	}
	if !strings.Contains(gotBody, "category=linux") {
		t.Fatalf("expected category in body, got %q", gotBody)
	}
	if !strings.Contains(gotBody, "savepath=") {
		t.Fatalf("expected savepath in body, got %q", gotBody)
	}
}

func TestAddTorrentFile(t *testing.T) {
	t.Parallel()

	var gotContentType string
	var gotBody []byte
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{Name: "SID", Value: testSID})
		w.Write([]byte("Ok."))
	})
	mux.HandleFunc("/api/v2/torrents/add", func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	err := client.AddTorrent(context.Background(), AddOptions{
		File:     []byte("d8:announce..."),
		Category: "tv",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(gotContentType, "multipart/form-data") {
		t.Fatalf("expected multipart content-type, got %q", gotContentType)
	}
	body := string(gotBody)
	if !strings.Contains(body, `name="torrents"`) {
		t.Fatalf("expected torrents part, got %q", body)
	}
	if !strings.Contains(body, `name="category"`) || !strings.Contains(body, "tv") {
		t.Fatalf("expected category field, got %q", body)
	}
}

func TestAddTorrentRequiresInput(t *testing.T) {
	t.Parallel()
	client := NewClient("http://example.invalid", nil)
	client.SetCredentials("admin", "secret")

	err := client.AddTorrent(context.Background(), AddOptions{})
	if err == nil {
		t.Fatal("expected error for empty options")
	}
	if !strings.Contains(err.Error(), "url or file required") {
		t.Fatalf("expected 'url or file required', got: %v", err)
	}
}

func TestGetCategories(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	cats, err := client.GetCategories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
	if cats["linux"].SavePath != "/downloads/linux" {
		t.Fatalf("unexpected save path: %v", cats["linux"])
	}
}

func TestResumeAll(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	if err := client.ResumeAll(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPauseAll(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	if err := client.PauseAll(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthFailed(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "wrong")

	_, err := client.ListTorrents(context.Background())
	if err == nil {
		t.Fatal("expected authentication error")
	}
	if !strings.Contains(err.Error(), "authentication failed") {
		t.Fatalf("expected 'authentication failed' in error, got: %v", err)
	}
}

func TestToTorrentInfoDownloading(t *testing.T) {
	t.Parallel()
	torrent := Torrent{
		Hash:      "abc123",
		Name:      "Test Torrent",
		State:     "downloading",
		Size:      1000000,
		Progress:  0.75,
		Dlspeed:   100000,
		Upspeed:   50000,
		ETA:       30,
		Ratio:     1.5,
		NumSeeds:  10,
		NumLeechs: 5,
		Category:  "test",
		Downloaded: 750000,
	}

	info := ToTorrentInfo(torrent)

	if info.Status != "downloading" {
		t.Fatalf("expected 'downloading', got '%s'", info.Status)
	}
	if info.Progress != 75.0 {
		t.Fatalf("expected 75.0, got %f", info.Progress)
	}
	if info.Seeds != 10 {
		t.Fatalf("expected 10 seeds, got %d", info.Seeds)
	}
}

func TestToTorrentInfoSeeding(t *testing.T) {
	t.Parallel()
	info := ToTorrentInfo(Torrent{State: "uploading"})
	if info.Status != "seeding" {
		t.Fatalf("expected 'seeding', got '%s'", info.Status)
	}
}

func TestToTorrentInfoPaused(t *testing.T) {
	t.Parallel()
	for _, state := range []string{"pausedDL", "pausedUP", "stoppedDL", "stoppedUP"} {
		info := ToTorrentInfo(Torrent{State: state})
		if info.Status != "paused" {
			t.Fatalf("state %s: expected 'paused', got '%s'", state, info.Status)
		}
	}
}

func TestToTorrentInfoError(t *testing.T) {
	t.Parallel()
	for _, state := range []string{"error", "missingFiles"} {
		info := ToTorrentInfo(Torrent{State: state})
		if info.Status != "error" {
			t.Fatalf("state %s: expected 'error', got '%s'", state, info.Status)
		}
	}
}

func TestNetworkError(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:1", nil)
	client.SetCredentials("admin", "secret")

	_, err := client.ListTorrents(context.Background())
	if err == nil {
		t.Fatal("expected network error")
	}
}

func TestTrailingSlash(t *testing.T) {
	t.Parallel()
	srv, _ := mockQBittorrentServer(t)
	defer srv.Close()

	client := NewClient(srv.URL+"/", srv.Client())
	client.SetCredentials("admin", "secret")

	torrents, err := client.ListTorrents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(torrents) != 2 {
		t.Fatalf("expected 2 torrents, got %d", len(torrents))
	}
}
