package grafana

// DashboardSearchResult represents a dashboard from Grafana search API.
type DashboardSearchResult struct {
	UID   string   `json:"uid"`
	Title string   `json:"title"`
	URL   string   `json:"url"`
	Type  string   `json:"type"`
	Tags  []string `json:"tags"`
}

// DashboardDetail is the full dashboard response from /api/dashboards/uid/{uid}.
type DashboardDetail struct {
	Dashboard DashboardData `json:"dashboard"`
	Meta      DashboardMeta `json:"meta"`
}

// DashboardData holds the dashboard content.
type DashboardData struct {
	UID    string  `json:"uid"`
	Title  string  `json:"title"`
	Panels []Panel `json:"panels"`
}

// DashboardMeta holds dashboard metadata.
type DashboardMeta struct {
	Slug string `json:"slug"`
}

// Panel represents a single panel inside a Grafana dashboard.
type Panel struct {
	ID     int     `json:"id"`
	Title  string  `json:"title"`
	Type   string  `json:"type"`
	Panels []Panel `json:"panels,omitempty"`
}

// HealthResponse is Grafana's /api/health response.
type HealthResponse struct {
	Commit   string `json:"commit"`
	Database string `json:"database"`
	Version  string `json:"version"`
}

// DashboardInfo is the simplified response sent to the Nidus frontend.
type DashboardInfo struct {
	UID    string      `json:"uid"`
	Title  string      `json:"title"`
	Slug   string      `json:"slug"`
	Panels []PanelInfo `json:"panels"`
}

// PanelInfo is the simplified panel sent to the Nidus frontend.
type PanelInfo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"`
}
