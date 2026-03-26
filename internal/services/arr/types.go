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
