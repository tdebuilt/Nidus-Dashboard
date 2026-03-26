package uptimekuma

// StatusPageResponse is the response from GET /api/status-page/:slug.
type StatusPageResponse struct {
	Config          StatusPageConfig `json:"config"`
	PublicGroupList []MonitorGroup   `json:"publicGroupList"`
}

// StatusPageConfig holds the status page configuration.
type StatusPageConfig struct {
	ID   int    `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"title"`
}

// MonitorGroup is a group of monitors on a status page.
type MonitorGroup struct {
	ID          int              `json:"id"`
	Name        string           `json:"name"`
	MonitorList []MonitorSummary `json:"monitorList"`
}

// MonitorSummary is a monitor entry from the status page list.
type MonitorSummary struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// HeartbeatResponse is the response from GET /api/status-page/heartbeat/:slug.
type HeartbeatResponse struct {
	HeartbeatList map[string][]Heartbeat `json:"heartbeatList"`
	UptimeList    map[string]float64     `json:"uptimeList"`
}

// Heartbeat represents a single heartbeat entry.
type Heartbeat struct {
	Status int     `json:"status"` // 0=DOWN, 1=UP, 2=PENDING, 3=MAINTENANCE
	Time   string  `json:"time"`
	Msg    string  `json:"msg"`
	Ping   float64 `json:"ping"`
}

// MonitorInfo is the combined monitor data for the frontend.
type MonitorInfo struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Status    int     `json:"status"`    // latest heartbeat status
	Uptime24h float64 `json:"uptime_24h"` // 0.0 - 1.0
	Latency   float64 `json:"latency"`   // latest ping in ms
	Message   string  `json:"message"`   // latest heartbeat message
}

// MonitorsOverview is the aggregated data returned to the frontend.
type MonitorsOverview struct {
	Monitors   []MonitorInfo `json:"monitors"`
	TotalUp    int           `json:"total_up"`
	TotalDown  int           `json:"total_down"`
	TotalCount int           `json:"total_count"`
	StatusPage string        `json:"status_page"`
}
