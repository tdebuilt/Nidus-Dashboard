package arr

import (
	"context"
	"encoding/json"
	"io"
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
		series := []SonarrSeries{
			{ID: 1, Title: "Breaking Bad", Year: 2008, SeasonCount: 5, Monitored: true, Status: "ended", Statistics: SonarrStatistics{EpisodeFileCount: 62, EpisodeCount: 62, PercentOfEpisodes: 100, SizeOnDisk: 28500000000}},
			{ID: 2, Title: "Better Call Saul", Year: 2015, SeasonCount: 6, Monitored: true, Status: "ended", Statistics: SonarrStatistics{EpisodeFileCount: 50, EpisodeCount: 63, PercentOfEpisodes: 79.37, SizeOnDisk: 22100000000}},
			{ID: 3, Title: "The Bear", Year: 2022, SeasonCount: 2, Monitored: true, Status: "continuing", Statistics: SonarrStatistics{EpisodeFileCount: 18, EpisodeCount: 18, PercentOfEpisodes: 100, SizeOnDisk: 9200000000}},
			{ID: 4, Title: "Severance", Year: 2022, SeasonCount: 2, Monitored: false, Status: "continuing", Statistics: SonarrStatistics{EpisodeFileCount: 10, EpisodeCount: 19, PercentOfEpisodes: 52.63, SizeOnDisk: 5400000000}},
			{ID: 5, Title: "Shogun", Year: 2024, SeasonCount: 1, Monitored: true, Status: "ended", Statistics: SonarrStatistics{EpisodeFileCount: 10, EpisodeCount: 10, PercentOfEpisodes: 100, SizeOnDisk: 12000000000}},
		}
		json.NewEncoder(w).Encode(series)
	})

	mux.HandleFunc("/api/v3/movie", func(w http.ResponseWriter, r *http.Request) {
		if !checkAPIKey(w, r) {
			return
		}
		movies := []RadarrMovie{
			{ID: 1, Title: "The Dark Knight", Year: 2008, HasFile: true, Monitored: true, SizeOnDisk: 3200000000, Runtime: 152, Status: "released"},
			{ID: 2, Title: "Inception", Year: 2010, HasFile: true, Monitored: true, SizeOnDisk: 4100000000, Runtime: 148, Status: "released"},
			{ID: 3, Title: "Dune: Part Two", Year: 2024, HasFile: false, Monitored: true, SizeOnDisk: 0, Runtime: 166, Status: "released"},
		}
		json.NewEncoder(w).Encode(movies)
	})

	mux.HandleFunc("/api/v3/episode", func(w http.ResponseWriter, r *http.Request) {
		if !checkAPIKey(w, r) {
			return
		}
		seriesID := r.URL.Query().Get("seriesId")
		if seriesID != "1" {
			json.NewEncoder(w).Encode([]SonarrEpisode{})
			return
		}
		episodes := []SonarrEpisode{
			{ID: 1, Title: "Pilot", EpisodeNumber: 1, SeasonNumber: 1, HasFile: true, Monitored: true, AirDateUtc: "2008-01-20T02:00:00Z"},
			{ID: 2, Title: "Cat's in the Bag...", EpisodeNumber: 2, SeasonNumber: 1, HasFile: true, Monitored: true, AirDateUtc: "2008-01-27T02:00:00Z"},
			{ID: 3, Title: "...And the Bag's in the River", EpisodeNumber: 3, SeasonNumber: 1, HasFile: false, Monitored: true, AirDateUtc: "2008-02-10T02:00:00Z"},
		}
		json.NewEncoder(w).Encode(episodes)
	})

	mux.HandleFunc("/api/v3/qualityprofile", func(w http.ResponseWriter, r *http.Request) {
		if !checkAPIKey(w, r) {
			return
		}
		json.NewEncoder(w).Encode([]QualityProfile{
			{ID: 1, Name: "HD-1080p"},
			{ID: 2, Name: "Ultra-HD"},
		})
	})

	mux.HandleFunc("/api/v3/rootfolder", func(w http.ResponseWriter, r *http.Request) {
		if !checkAPIKey(w, r) {
			return
		}
		json.NewEncoder(w).Encode([]RootFolder{
			{ID: 1, Path: "/movies", FreeSpace: 500000000000},
			{ID: 2, Path: "/tv", FreeSpace: 300000000000},
		})
	})

	mux.HandleFunc("/api/v3/movie/lookup", func(w http.ResponseWriter, r *http.Request) {
		if !checkAPIKey(w, r) {
			return
		}
		json.NewEncoder(w).Encode([]LookupResult{
			{Title: "The Dark Knight", Year: 2008, TmdbID: 155, Overview: "Batman raises the stakes.", Runtime: 152, ID: 1},
			{Title: "The Dark Knight Rises", Year: 2012, TmdbID: 49026, Overview: "Eight years after.", Runtime: 165},
		})
	})

	mux.HandleFunc("/api/v3/series/lookup", func(w http.ResponseWriter, r *http.Request) {
		if !checkAPIKey(w, r) {
			return
		}
		json.NewEncoder(w).Encode([]LookupResult{
			{Title: "Breaking Bad", Year: 2008, TvdbID: 81189, Overview: "A chemistry teacher.", SeasonCount: 5, ID: 1},
			{Title: "Breaking Bread", Year: 2020, TvdbID: 99999, Overview: "A cooking show.", SeasonCount: 1},
		})
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
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	status, err := client.GetSystemStatus(context.Background())
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
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	queue, err := client.GetQueue(context.Background(), 20)
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
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	start := time.Date(2026, 3, 23, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 3, 30, 0, 0, 0, 0, time.UTC)

	items, err := client.GetCalendar(context.Background(), start, end)
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
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	count, err := client.GetLibraryCount(context.Background(), "/series")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected library count 5, got %d", count)
	}
}

func TestApiKeyHeader(t *testing.T) {
	t.Parallel()
	var receivedKey string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKey = r.Header.Get("X-Api-Key")
		json.NewEncoder(w).Encode(SystemStatus{Version: "1.0"})
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "my-secret-key", "v3", srv.Client())

	_, err := client.GetSystemStatus(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedKey != "my-secret-key" {
		t.Fatalf("expected API key 'my-secret-key', got '%s'", receivedKey)
	}
}

func TestUnauthorized(t *testing.T) {
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "wrong-key", "v3", srv.Client())

	_, err := client.GetSystemStatus(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("expected 'unauthorized' in error, got: %v", err)
	}
}

func TestGetRadarrLibrary(t *testing.T) {
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	movies, err := client.GetRadarrLibrary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(movies) != 3 {
		t.Fatalf("expected 3 movies, got %d", len(movies))
	}
	if movies[0].Title != "The Dark Knight" {
		t.Fatalf("expected 'The Dark Knight', got '%s'", movies[0].Title)
	}
	if !movies[0].HasFile {
		t.Fatal("expected first movie to have file")
	}
	if movies[2].HasFile {
		t.Fatal("expected third movie to not have file")
	}
	if movies[0].SizeOnDisk != 3200000000 {
		t.Fatalf("expected size 3200000000, got %d", movies[0].SizeOnDisk)
	}
}

func TestGetSonarrLibrary(t *testing.T) {
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	series, err := client.GetSonarrLibrary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(series) != 5 {
		t.Fatalf("expected 5 series, got %d", len(series))
	}
	if series[0].Title != "Breaking Bad" {
		t.Fatalf("expected 'Breaking Bad', got '%s'", series[0].Title)
	}
	if series[0].Statistics.EpisodeFileCount != 62 {
		t.Fatalf("expected 62 episode files, got %d", series[0].Statistics.EpisodeFileCount)
	}
	if series[0].Statistics.PercentOfEpisodes != 100 {
		t.Fatalf("expected 100%% completion, got %f", series[0].Statistics.PercentOfEpisodes)
	}
	if series[1].Statistics.EpisodeCount != 63 {
		t.Fatalf("expected 63 episodes, got %d", series[1].Statistics.EpisodeCount)
	}
}

func TestGetSonarrEpisodes(t *testing.T) {
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	episodes, err := client.GetSonarrEpisodes(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(episodes) != 3 {
		t.Fatalf("expected 3 episodes, got %d", len(episodes))
	}
	if episodes[0].Title != "Pilot" {
		t.Fatalf("expected 'Pilot', got '%s'", episodes[0].Title)
	}
	if episodes[0].SeasonNumber != 1 || episodes[0].EpisodeNumber != 1 {
		t.Fatalf("expected S01E01, got S%02dE%02d", episodes[0].SeasonNumber, episodes[0].EpisodeNumber)
	}
	if !episodes[0].HasFile {
		t.Fatal("expected first episode to have file")
	}
	if episodes[2].HasFile {
		t.Fatal("expected third episode to not have file")
	}

	// Test with non-existent series
	empty, err := client.GetSonarrEpisodes(context.Background(), 999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(empty) != 0 {
		t.Fatalf("expected 0 episodes for unknown series, got %d", len(empty))
	}
}

func TestGetQualityProfiles(t *testing.T) {
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	profiles, err := client.GetQualityProfiles(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(profiles) != 2 {
		t.Fatalf("expected 2 profiles, got %d", len(profiles))
	}
	if profiles[0].Name != "HD-1080p" {
		t.Fatalf("expected 'HD-1080p', got '%s'", profiles[0].Name)
	}
}

func TestGetRootFolders(t *testing.T) {
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	folders, err := client.GetRootFolders(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(folders) != 2 {
		t.Fatalf("expected 2 folders, got %d", len(folders))
	}
	if folders[0].Path != "/movies" {
		t.Fatalf("expected '/movies', got '%s'", folders[0].Path)
	}
	if folders[0].FreeSpace != 500000000000 {
		t.Fatalf("expected 500000000000 free space, got %d", folders[0].FreeSpace)
	}
}

func TestLookupMedia(t *testing.T) {
	t.Parallel()
	srv := mockArrServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	// Radarr lookup
	movies, err := client.LookupMedia(context.Background(), "/movie", "dark knight")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(movies) != 2 {
		t.Fatalf("expected 2 results, got %d", len(movies))
	}
	if movies[0].TmdbID != 155 {
		t.Fatalf("expected tmdbId 155, got %d", movies[0].TmdbID)
	}
	if movies[0].ID != 1 {
		t.Fatal("expected first result to be in library (ID=1)")
	}
	if movies[1].ID != 0 {
		t.Fatal("expected second result to not be in library (ID=0)")
	}

	// Sonarr lookup
	series, err := client.LookupMedia(context.Background(), "/series", "breaking")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(series) != 2 {
		t.Fatalf("expected 2 results, got %d", len(series))
	}
	if series[0].TvdbID != 81189 {
		t.Fatalf("expected tvdbId 81189, got %d", series[0].TvdbID)
	}
}

func TestAddMedia(t *testing.T) {
	t.Parallel()

	var receivedBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "test-api-key-123" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "wrong content type", http.StatusBadRequest)
			return
		}
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id": 42}`))
	}))
	defer srv.Close()

	client := NewClient(srv.URL, "test-api-key-123", "v3", srv.Client())

	body := AddMovieRequest{
		Title:            "Inception",
		TmdbID:           27205,
		Year:             2010,
		QualityProfileID: 1,
		RootFolderPath:   "/movies",
		Monitored:        true,
		AddOptions:       MovieAddOpts{SearchForMovie: true},
	}
	err := client.AddMedia(context.Background(), "/movie", body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receivedBody) == 0 {
		t.Fatal("expected request body to be sent")
	}

	var decoded AddMovieRequest
	if err := json.Unmarshal(receivedBody, &decoded); err != nil {
		t.Fatalf("failed to decode request body: %v", err)
	}
	if decoded.TmdbID != 27205 {
		t.Fatalf("expected tmdbId 27205, got %d", decoded.TmdbID)
	}
}
