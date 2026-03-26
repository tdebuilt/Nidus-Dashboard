package mediaserver

// Session represents an active streaming session (unified for Plex/Jellyfin).
type Session struct {
	ID        string  `json:"id"`
	UserName  string  `json:"user_name"`
	Title     string  `json:"title"`              // Movie name or episode name
	Subtitle  string  `json:"subtitle,omitempty"`  // For TV: "S01E05" or series info
	MediaType string  `json:"media_type"`          // "movie", "episode", "track"
	Year      int     `json:"year,omitempty"`
	Progress  float64 `json:"progress"`            // 0.0 - 1.0
	State     string  `json:"state"`               // "playing", "paused", "buffering"
	Player    string  `json:"player"`              // Client name (e.g. "Plex Web")
	Platform  string  `json:"platform,omitempty"`  // Platform (e.g. "Chrome")
	ThumbPath string  `json:"thumb_path,omitempty"` // Relative path for image proxy
	Duration  int64   `json:"duration"`            // Total duration in seconds
	Position  int64   `json:"position"`            // Current position in seconds
}

// Library represents a media library.
type Library struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`       // "movie", "show", "music", "photo"
	ItemCount int    `json:"item_count"`
}

// MediaOverview is the aggregated data returned to the frontend.
type MediaOverview struct {
	Sessions     []Session `json:"sessions"`
	SessionCount int       `json:"session_count"`
	Libraries    []Library `json:"libraries,omitempty"`
	ServerName   string    `json:"server_name"`
	ServerType   string    `json:"server_type"` // "plex" or "jellyfin"
}

// Client is the interface for media server API clients.
type Client interface {
	GetSessions() ([]Session, error)
	GetLibraries() ([]Library, error)
	GetServerName() (string, error)
	ProxyImage(path string) ([]byte, string, error) // returns body, content-type, error
}

// --- Plex API response types ---

type plexMediaContainer[T any] struct {
	MediaContainer struct {
		Size     int    `json:"size"`
		Metadata []T    `json:"Metadata"`
		Title1   string `json:"title1"`
		// For identity
		FriendlyName string `json:"friendlyName"`
		// For libraries
		Directory []plexDirectory `json:"Directory"`
	} `json:"MediaContainer"`
}

type plexSession struct {
	Title            string   `json:"title"`
	Type             string   `json:"type"` // "movie", "episode", "track"
	Thumb            string   `json:"thumb"`
	Duration         int64    `json:"duration"`   // milliseconds
	ViewOffset       int64    `json:"viewOffset"` // milliseconds
	GrandparentTitle string   `json:"grandparentTitle"`
	ParentIndex      int      `json:"parentIndex"`
	Index            int      `json:"index"`
	Year             int      `json:"year"`
	User             plexUser `json:"User"`
	Player           plexPlayer `json:"Player"`
	Session          struct {
		ID string `json:"id"`
	} `json:"Session"`
}

type plexUser struct {
	Title string `json:"title"`
}

type plexPlayer struct {
	State    string `json:"state"` // "playing", "paused", "buffering"
	Product  string `json:"product"`
	Platform string `json:"platform"`
}

type plexDirectory struct {
	Key   string `json:"key"`
	Title string `json:"title"`
	Type  string `json:"type"`
	Count int    `json:"count"`
}

// --- Jellyfin API response types ---

type jellyfinSession struct {
	ID             string              `json:"Id"`
	UserName       string              `json:"UserName"`
	Client         string              `json:"Client"`
	DeviceName     string              `json:"DeviceName"`
	NowPlayingItem *jellyfinNowPlaying `json:"NowPlayingItem"`
	PlayState      *jellyfinPlayState  `json:"PlayState"`
}

type jellyfinNowPlaying struct {
	Name              string            `json:"Name"`
	Type              string            `json:"Type"` // "Movie", "Episode", "Audio"
	ID                string            `json:"Id"`
	SeriesName        string            `json:"SeriesName"`
	ParentIndexNumber int               `json:"ParentIndexNumber"`
	IndexNumber       int               `json:"IndexNumber"`
	ProductionYear    int               `json:"ProductionYear"`
	RunTimeTicks      int64             `json:"RunTimeTicks"` // 100-nanosecond intervals
	ImageTags         map[string]string `json:"ImageTags"`
}

type jellyfinPlayState struct {
	PositionTicks int64 `json:"PositionTicks"`
	IsPaused      bool  `json:"IsPaused"`
}

type jellyfinLibrary struct {
	Name            string `json:"Name"`
	ItemID          string `json:"ItemId"`
	CollectionType  string `json:"CollectionType"` // "movies", "tvshows", "music"
}

type jellyfinItemsResponse struct {
	TotalRecordCount int `json:"TotalRecordCount"`
}

type jellyfinServerInfo struct {
	ServerName string `json:"ServerName"`
	Version    string `json:"Version"`
}
