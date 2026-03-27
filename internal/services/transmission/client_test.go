package transmission

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

func mockTransmissionServer(t *testing.T) (*httptest.Server, *atomic.Int32) {
	t.Helper()

	requestCount := &atomic.Int32{}
	sessionID := "test-session-id-123"

	mux := http.NewServeMux()
	mux.HandleFunc("/transmission/rpc", func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)

		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Check auth
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "secret" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		// Check session ID — return 409 if missing
		if r.Header.Get(sessionIDHeader) != sessionID {
			w.Header().Set(sessionIDHeader, sessionID)
			w.WriteHeader(http.StatusConflict)
			return
		}

		var rpcReq RPCRequest
		json.NewDecoder(r.Body).Decode(&rpcReq)

		switch rpcReq.Method {
		case "torrent-get":
			torrents := TorrentListResponse{
				Torrents: []Torrent{
					{
						ID:            1,
						Name:          "Ubuntu 24.04",
						Status:        StatusDownloading,
						SizeWhenDone:  4000000000,
						LeftUntilDone: 1000000000,
						PercentDone:   0.75,
						RateDownload:  5000000,
						RateUpload:    1000000,
						ETA:           200,
						UploadRatio:   0.5,
						Peers:         42,
					},
					{
						ID:            2,
						Name:          "Debian 12",
						Status:        StatusSeeding,
						SizeWhenDone:  3000000000,
						LeftUntilDone: 0,
						PercentDone:   1.0,
						RateDownload:  0,
						RateUpload:    500000,
						ETA:           -1,
						UploadRatio:   2.5,
						Peers:         10,
					},
					{
						ID:            3,
						Name:          "Arch Linux",
						Status:        StatusStopped,
						SizeWhenDone:  800000000,
						LeftUntilDone: 800000000,
						PercentDone:   0,
						RateDownload:  0,
						RateUpload:    0,
						ETA:           -1,
					},
				},
			}
			args, _ := json.Marshal(torrents)
			json.NewEncoder(w).Encode(RPCResponse{Result: "success", Arguments: json.RawMessage(args)})

		case "torrent-add":
			json.NewEncoder(w).Encode(RPCResponse{Result: "success"})

		case "torrent-start":
			json.NewEncoder(w).Encode(RPCResponse{Result: "success"})

		case "torrent-stop":
			json.NewEncoder(w).Encode(RPCResponse{Result: "success"})

		case "session-stats":
			stats := SessionStats{
				DownloadSpeed: 5000000,
				UploadSpeed:   1500000,
				TorrentCount:  3,
				ActiveCount:   2,
				PausedCount:   1,
			}
			args, _ := json.Marshal(stats)
			json.NewEncoder(w).Encode(RPCResponse{Result: "success", Arguments: json.RawMessage(args)})

		default:
			json.NewEncoder(w).Encode(RPCResponse{Result: "unknown method"})
		}
	})

	return httptest.NewServer(mux), requestCount
}

func TestSessionIDRetry(t *testing.T) {
	t.Parallel()
	srv, requestCount := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	// First call should trigger 409 then retry
	torrents, err := client.ListTorrents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(torrents) != 3 {
		t.Fatalf("expected 3 torrents, got %d", len(torrents))
	}
	// Should be 2 requests (409 + retry)
	if requestCount.Load() != 2 {
		t.Fatalf("expected 2 requests (409 + retry), got %d", requestCount.Load())
	}
}

func TestSessionIDReuse(t *testing.T) {
	t.Parallel()
	srv, requestCount := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	// First call: 409 + retry = 2 requests
	_, err := client.ListTorrents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second call should reuse session ID: only 1 request
	_, err = client.ListTorrents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if requestCount.Load() != 3 {
		t.Fatalf("expected 3 total requests (2+1), got %d", requestCount.Load())
	}
}

func TestListTorrents(t *testing.T) {
	t.Parallel()
	srv, _ := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	torrents, err := client.ListTorrents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(torrents) != 3 {
		t.Fatalf("expected 3 torrents, got %d", len(torrents))
	}
	if torrents[0].Name != "Ubuntu 24.04" {
		t.Fatalf("expected 'Ubuntu 24.04', got '%s'", torrents[0].Name)
	}
	if torrents[0].PercentDone != 0.75 {
		t.Fatalf("expected 0.75, got %f", torrents[0].PercentDone)
	}
	if torrents[1].Status != StatusSeeding {
		t.Fatalf("expected seeding status, got %d", torrents[1].Status)
	}
}

func TestAddTorrent(t *testing.T) {
	t.Parallel()
	srv, _ := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	err := client.AddTorrent(context.Background(), "magnet:?xt=urn:btih:test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartTorrent(t *testing.T) {
	t.Parallel()
	srv, _ := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	if err := client.StartTorrent(context.Background(), []int{1, 2}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStopTorrent(t *testing.T) {
	t.Parallel()
	srv, _ := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	if err := client.StopTorrent(context.Background(), []int{1}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStartAll(t *testing.T) {
	t.Parallel()
	srv, _ := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	if err := client.StartAll(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestStopAll(t *testing.T) {
	t.Parallel()
	srv, _ := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	if err := client.StopAll(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetSessionStats(t *testing.T) {
	t.Parallel()
	srv, _ := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetCredentials("admin", "secret")

	stats, err := client.GetSessionStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.DownloadSpeed != 5000000 {
		t.Fatalf("expected download speed 5000000, got %d", stats.DownloadSpeed)
	}
	if stats.TorrentCount != 3 {
		t.Fatalf("expected 3 torrents, got %d", stats.TorrentCount)
	}
	if stats.ActiveCount != 2 {
		t.Fatalf("expected 2 active, got %d", stats.ActiveCount)
	}
}

func TestUnauthorized(t *testing.T) {
	t.Parallel()
	srv, _ := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	// No credentials

	_, err := client.ListTorrents(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("expected 'unauthorized' in error, got: %v", err)
	}
}

func TestToTorrentInfo(t *testing.T) {
	t.Parallel()
	torrent := Torrent{
		ID:            1,
		Name:          "Test Torrent",
		Status:        StatusDownloading,
		SizeWhenDone:  1000000,
		LeftUntilDone: 250000,
		PercentDone:   0.75,
		RateDownload:  100000,
		RateUpload:    50000,
		ETA:           30,
		UploadRatio:   1.5,
		Peers:         10,
	}

	info := ToTorrentInfo(torrent)

	if info.Status != "downloading" {
		t.Fatalf("expected 'downloading', got '%s'", info.Status)
	}
	if info.Progress != 75.0 {
		t.Fatalf("expected 75.0, got %f", info.Progress)
	}
	if info.Downloaded != 750000 {
		t.Fatalf("expected 750000, got %d", info.Downloaded)
	}
}

func TestToTorrentInfoStopped(t *testing.T) {
	t.Parallel()
	torrent := Torrent{
		ID:     2,
		Name:   "Stopped",
		Status: StatusStopped,
	}
	info := ToTorrentInfo(torrent)
	if info.Status != "stopped" {
		t.Fatalf("expected 'stopped', got '%s'", info.Status)
	}
}

func TestToTorrentInfoSeeding(t *testing.T) {
	t.Parallel()
	torrent := Torrent{
		ID:     3,
		Name:   "Seeding",
		Status: StatusSeeding,
	}
	info := ToTorrentInfo(torrent)
	if info.Status != "seeding" {
		t.Fatalf("expected 'seeding', got '%s'", info.Status)
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
	srv, _ := mockTransmissionServer(t)
	defer srv.Close()

	client := NewClient(srv.URL+"/", srv.Client())
	client.SetCredentials("admin", "secret")

	torrents, err := client.ListTorrents(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(torrents) != 3 {
		t.Fatalf("expected 3 torrents, got %d", len(torrents))
	}
}
