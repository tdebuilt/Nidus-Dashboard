package mediaserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPlexGetSessions(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/status/sessions", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Plex-Token") != "test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		resp := map[string]any{
			"MediaContainer": map[string]any{
				"size": 1,
				"Metadata": []map[string]any{
					{
						"title":            "Test Movie",
						"type":             "movie",
						"thumb":            "/library/metadata/1/thumb/123",
						"duration":         7200000,
						"viewOffset":       3600000,
						"year":             2024,
						"User":             map[string]any{"title": "testuser"},
						"Player":           map[string]any{"state": "playing", "product": "Plex Web", "platform": "Chrome"},
						"Session":          map[string]any{"id": "abc123"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewPlexClient(srv.URL, "test-token", nil)
	sessions, err := client.GetSessions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}

	s := sessions[0]
	if s.Title != "Test Movie" {
		t.Errorf("expected title 'Test Movie', got '%s'", s.Title)
	}
	if s.UserName != "testuser" {
		t.Errorf("expected user 'testuser', got '%s'", s.UserName)
	}
	if s.State != "playing" {
		t.Errorf("expected state 'playing', got '%s'", s.State)
	}
	if s.Duration != 7200 {
		t.Errorf("expected duration 7200s, got %d", s.Duration)
	}
	if s.Position != 3600 {
		t.Errorf("expected position 3600s, got %d", s.Position)
	}
	if s.Progress < 0.49 || s.Progress > 0.51 {
		t.Errorf("expected progress ~0.5, got %f", s.Progress)
	}
	if s.MediaType != "movie" {
		t.Errorf("expected type 'movie', got '%s'", s.MediaType)
	}
}

func TestPlexGetSessionsEpisode(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/status/sessions", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"MediaContainer": map[string]any{
				"size": 1,
				"Metadata": []map[string]any{
					{
						"title":            "Pilot",
						"type":             "episode",
						"grandparentTitle": "Breaking Bad",
						"parentIndex":      1,
						"index":            1,
						"duration":         3600000,
						"viewOffset":       900000,
						"User":             map[string]any{"title": "user1"},
						"Player":           map[string]any{"state": "paused", "product": "Plex iOS"},
						"Session":          map[string]any{"id": "def456"},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewPlexClient(srv.URL, "", nil)
	sessions, err := client.GetSessions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := sessions[0]
	if s.Subtitle != "Breaking Bad — S01E01" {
		t.Errorf("expected subtitle 'Breaking Bad — S01E01', got '%s'", s.Subtitle)
	}
	if s.State != "paused" {
		t.Errorf("expected state 'paused', got '%s'", s.State)
	}
}

func TestPlexGetLibraries(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/library/sections", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"MediaContainer": map[string]any{
				"Directory": []map[string]any{
					{"key": "1", "title": "Movies", "type": "movie", "count": 150},
					{"key": "2", "title": "TV Shows", "type": "show", "count": 42},
					{"key": "3", "title": "Music", "type": "artist", "count": 300},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewPlexClient(srv.URL, "", nil)
	libraries, err := client.GetLibraries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(libraries) != 3 {
		t.Fatalf("expected 3 libraries, got %d", len(libraries))
	}
	if libraries[0].Name != "Movies" {
		t.Errorf("expected 'Movies', got '%s'", libraries[0].Name)
	}
	if libraries[2].Type != "music" {
		t.Errorf("expected music type for 'artist', got '%s'", libraries[2].Type)
	}
}

func TestPlexGetServerName(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"MediaContainer": map[string]any{
				"friendlyName": "My Plex Server",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewPlexClient(srv.URL, "", nil)
	name, err := client.GetServerName(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "My Plex Server" {
		t.Errorf("expected 'My Plex Server', got '%s'", name)
	}
}

func TestPlexServerError(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/status/sessions", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewPlexClient(srv.URL, "", nil)
	_, err := client.GetSessions(context.Background())
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestJellyfinGetSessions(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/Sessions", func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != `MediaBrowser Token="test-key"` {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		resp := []map[string]any{
			{
				"Id":         "session-1",
				"UserName":   "testuser",
				"Client":     "Jellyfin Web",
				"DeviceName": "Chrome",
				"NowPlayingItem": map[string]any{
					"Name":              "Test Movie",
					"Type":              "Movie",
					"Id":                "item-1",
					"ProductionYear":    2024,
					"RunTimeTicks":      72000000000,
					"ImageTags":         map[string]string{"Primary": "tag1"},
				},
				"PlayState": map[string]any{
					"PositionTicks": 36000000000,
					"IsPaused":      false,
				},
			},
			{
				"Id":       "session-2",
				"UserName": "idle-user",
				"Client":   "Jellyfin Web",
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewJellyfinClient(srv.URL, "test-key", nil)
	sessions, err := client.GetSessions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only include sessions with NowPlayingItem
	if len(sessions) != 1 {
		t.Fatalf("expected 1 active session, got %d", len(sessions))
	}

	s := sessions[0]
	if s.Title != "Test Movie" {
		t.Errorf("expected title 'Test Movie', got '%s'", s.Title)
	}
	if s.UserName != "testuser" {
		t.Errorf("expected user 'testuser', got '%s'", s.UserName)
	}
	if s.State != "playing" {
		t.Errorf("expected state 'playing', got '%s'", s.State)
	}
	if s.Duration != 7200 {
		t.Errorf("expected duration 7200s, got %d", s.Duration)
	}
	if s.Position != 3600 {
		t.Errorf("expected position 3600s, got %d", s.Position)
	}
	if s.ThumbPath != "/Items/item-1/Images/Primary" {
		t.Errorf("expected thumb path '/Items/item-1/Images/Primary', got '%s'", s.ThumbPath)
	}
}

func TestJellyfinGetSessionsEpisode(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/Sessions", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{
				"Id":         "session-1",
				"UserName":   "user1",
				"Client":     "Jellyfin Mobile",
				"DeviceName": "iPhone",
				"NowPlayingItem": map[string]any{
					"Name":              "Pilot",
					"Type":              "Episode",
					"Id":                "ep-1",
					"SeriesName":        "Breaking Bad",
					"ParentIndexNumber": 1,
					"IndexNumber":       1,
					"RunTimeTicks":      36000000000,
				},
				"PlayState": map[string]any{
					"PositionTicks": 9000000000,
					"IsPaused":      true,
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewJellyfinClient(srv.URL, "", nil)
	sessions, err := client.GetSessions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	s := sessions[0]
	if s.Subtitle != "Breaking Bad — S01E01" {
		t.Errorf("expected subtitle 'Breaking Bad — S01E01', got '%s'", s.Subtitle)
	}
	if s.State != "paused" {
		t.Errorf("expected state 'paused', got '%s'", s.State)
	}
	if s.MediaType != "episode" {
		t.Errorf("expected type 'episode', got '%s'", s.MediaType)
	}
}

func TestJellyfinGetServerName(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/System/Info/Public", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"ServerName": "My Jellyfin", "Version": "10.9.0"}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewJellyfinClient(srv.URL, "", nil)
	name, err := client.GetServerName(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if name != "My Jellyfin" {
		t.Errorf("expected 'My Jellyfin', got '%s'", name)
	}
}

func TestJellyfinGetLibraries(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/Library/VirtualFolders", func(w http.ResponseWriter, r *http.Request) {
		resp := []map[string]any{
			{"Name": "Movies", "ItemId": "lib-1", "CollectionType": "movies"},
			{"Name": "TV Shows", "ItemId": "lib-2", "CollectionType": "tvshows"},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
	mux.HandleFunc("/Items", func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]int{"TotalRecordCount": 42}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewJellyfinClient(srv.URL, "", nil)
	libraries, err := client.GetLibraries(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(libraries) != 2 {
		t.Fatalf("expected 2 libraries, got %d", len(libraries))
	}
	if libraries[0].Type != "movie" {
		t.Errorf("expected 'movie' type, got '%s'", libraries[0].Type)
	}
	if libraries[1].Type != "show" {
		t.Errorf("expected 'show' type, got '%s'", libraries[1].Type)
	}
}

func TestJellyfinServerError(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc("/Sessions", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewJellyfinClient(srv.URL, "", nil)
	_, err := client.GetSessions(context.Background())
	if err == nil {
		t.Fatal("expected error for server error response")
	}
}

func TestTicksToSeconds(t *testing.T) {
	t.Parallel()
	tests := []struct {
		ticks    int64
		expected int64
	}{
		{0, 0},
		{10_000_000, 1},
		{36_000_000_000, 3600},
		{72_000_000_000, 7200},
	}

	for _, tt := range tests {
		got := ticksToSeconds(tt.ticks)
		if got != tt.expected {
			t.Errorf("ticksToSeconds(%d) = %d, want %d", tt.ticks, got, tt.expected)
		}
	}
}
