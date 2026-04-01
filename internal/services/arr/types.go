package arr

// ArrConfig defines API settings for each *arr service type.
type ArrConfig struct {
	APIVersion   string
	LibraryPath  string
	CalendarPath string // empty for services without calendar
	Label        string
}

// Configs maps service type to API configuration.
var Configs = map[string]ArrConfig{
	"sonarr":   {APIVersion: "v3", LibraryPath: "/series", CalendarPath: "/calendar", Label: "Sonarr"},
	"radarr":   {APIVersion: "v3", LibraryPath: "/movie", CalendarPath: "/calendar", Label: "Radarr"},
	"lidarr":   {APIVersion: "v1", LibraryPath: "/artist", CalendarPath: "", Label: "Lidarr"},
	"prowlarr": {APIVersion: "v1", LibraryPath: "/indexer", CalendarPath: "", Label: "Prowlarr"},
}

// SystemStatus represents the /system/status response.
type SystemStatus struct {
	Version      string `json:"version"`
	AppName      string `json:"appName"`
	InstanceName string `json:"instanceName"`
}

// QueueItem represents a single item in the download queue.
type QueueItem struct {
	ID                    int    `json:"id"`
	Title                 string `json:"title"`
	Status                string `json:"status"`
	TrackedDownloadState  string `json:"trackedDownloadState"`
	Size                  int64  `json:"size"`
	Sizeleft              int64  `json:"sizeleft"`
	Timeleft              string `json:"timeleft"`
}

// QueueResponse represents the paginated queue response.
type QueueResponse struct {
	Page         int         `json:"page"`
	PageSize     int         `json:"pageSize"`
	TotalRecords int         `json:"totalRecords"`
	Records      []QueueItem `json:"records"`
}

// CalendarItem represents an upcoming media item.
type CalendarItem struct {
	ID              int    `json:"id"`
	Title           string `json:"title"`
	Overview        string `json:"overview"`
	AirDateUtc      string `json:"airDateUtc"`
	InCinemas       string `json:"inCinemas"`
	PhysicalRelease string `json:"physicalRelease"`
	HasFile         bool   `json:"hasFile"`
	SeriesTitle     string `json:"seriesTitle"`
}

// MediaFile holds quality and media info for a downloaded file.
type MediaFile struct {
	Quality   MediaQuality `json:"quality"`
	MediaInfo *MediaInfo   `json:"mediaInfo,omitempty"`
}

// MediaQuality wraps the quality detail.
type MediaQuality struct {
	Quality QualityDetail `json:"quality"`
}

// QualityDetail holds the quality name (e.g. "HDTV-1080p").
type QualityDetail struct {
	Name string `json:"name"`
}

// MediaInfo holds audio/video metadata.
type MediaInfo struct {
	AudioCodec string `json:"audioCodec"`
	VideoCodec string `json:"videoCodec"`
}

// RadarrMovie represents a movie from the Radarr library.
type RadarrMovie struct {
	ID         int        `json:"id"`
	Title      string     `json:"title"`
	Year       int        `json:"year"`
	HasFile    bool       `json:"hasFile"`
	Monitored  bool       `json:"monitored"`
	SizeOnDisk int64      `json:"sizeOnDisk"`
	Runtime    int        `json:"runtime"`
	Status     string     `json:"status"`
	MovieFile  *MediaFile `json:"movieFile,omitempty"`
}

// SonarrSeries represents a series from the Sonarr library.
type SonarrSeries struct {
	ID          int              `json:"id"`
	Title       string           `json:"title"`
	Year        int              `json:"year"`
	SeasonCount int              `json:"seasonCount"`
	Monitored   bool             `json:"monitored"`
	Status      string           `json:"status"`
	Seasons     []SonarrSeason   `json:"seasons"`
	Statistics  SonarrStatistics `json:"statistics"`
}

// SonarrSeason represents a season's monitoring status.
type SonarrSeason struct {
	SeasonNumber int  `json:"seasonNumber"`
	Monitored    bool `json:"monitored"`
}

// SonarrStatistics holds episode/file stats for a series.
type SonarrStatistics struct {
	EpisodeFileCount  int     `json:"episodeFileCount"`
	EpisodeCount      int     `json:"episodeCount"`
	PercentOfEpisodes float64 `json:"percentOfEpisodes"`
	SizeOnDisk        int64   `json:"sizeOnDisk"`
}

// SonarrEpisode represents a single episode from Sonarr.
type SonarrEpisode struct {
	ID            int        `json:"id"`
	Title         string     `json:"title"`
	EpisodeNumber int        `json:"episodeNumber"`
	SeasonNumber  int        `json:"seasonNumber"`
	HasFile       bool       `json:"hasFile"`
	Monitored     bool       `json:"monitored"`
	AirDateUtc    string     `json:"airDateUtc"`
	EpisodeFile   *MediaFile `json:"episodeFile,omitempty"`
}

// QualityProfile represents a quality profile from Radarr/Sonarr.
type QualityProfile struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// RootFolder represents a root folder from Radarr/Sonarr.
type RootFolder struct {
	ID        int    `json:"id"`
	Path      string `json:"path"`
	FreeSpace int64  `json:"freeSpace"`
}

// LookupResult represents a search result from the movie/series lookup API.
type LookupResult struct {
	Title       string `json:"title"`
	Year        int    `json:"year"`
	TmdbID      int    `json:"tmdbId,omitempty"`
	TvdbID      int    `json:"tvdbId,omitempty"`
	Overview    string `json:"overview"`
	Runtime     int    `json:"runtime,omitempty"`
	SeasonCount int    `json:"seasonCount,omitempty"`
	ID          int    `json:"id,omitempty"` // Non-zero means already in library
}

// AddMovieRequest is the payload for adding a movie to Radarr.
type AddMovieRequest struct {
	Title            string       `json:"title"`
	TmdbID           int          `json:"tmdbId"`
	Year             int          `json:"year"`
	QualityProfileID int          `json:"qualityProfileId"`
	RootFolderPath   string       `json:"rootFolderPath"`
	Monitored        bool         `json:"monitored"`
	AddOptions       MovieAddOpts `json:"addOptions"`
}

// MovieAddOpts holds add-specific options for movies.
type MovieAddOpts struct {
	SearchForMovie bool `json:"searchForMovie"`
}

// AddSeriesRequest is the payload for adding a series to Sonarr.
type AddSeriesRequest struct {
	Title            string        `json:"title"`
	TvdbID           int           `json:"tvdbId"`
	Year             int           `json:"year"`
	QualityProfileID int           `json:"qualityProfileId"`
	RootFolderPath   string        `json:"rootFolderPath"`
	Monitored        bool          `json:"monitored"`
	SeasonFolder     bool          `json:"seasonFolder"`
	AddOptions       SeriesAddOpts `json:"addOptions"`
}

// SeriesAddOpts holds add-specific options for series.
type SeriesAddOpts struct {
	SearchForMissingEpisodes bool `json:"searchForMissingEpisodes"`
}

// ProwlarrIndexer represents an indexer from Prowlarr with merged status.
type ProwlarrIndexer struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Enable   bool   `json:"enable"`
	Protocol string `json:"protocol"`
	Priority int    `json:"priority"`
	Status   string `json:"status"` // "ok", "error", "disabled"
}

// prowlarrIndexerRaw is the raw response from the Prowlarr /indexer endpoint.
type prowlarrIndexerRaw struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Enable   bool   `json:"enable"`
	Protocol string `json:"protocol"`
	Priority int    `json:"priority"`
}

// ProwlarrIndexerStatus represents a failing indexer from /indexerstatus.
type ProwlarrIndexerStatus struct {
	IndexerID         int    `json:"indexerId"`
	DisabledTill      string `json:"disabledTill,omitempty"`
	MostRecentFailure string `json:"mostRecentFailure,omitempty"`
}

// ArrOverview is the per-service frontend response.
type ArrOverview struct {
	Type          string         `json:"type"`
	Label         string         `json:"label"`
	Status        *SystemStatus  `json:"status,omitempty"`
	QueueCount    int            `json:"queue_count"`
	QueueItems    []QueueItem    `json:"queue_items"`
	CalendarItems []CalendarItem `json:"calendar_items"`
	LibraryCount  int            `json:"library_count"`
	Error         string         `json:"error,omitempty"`
}
