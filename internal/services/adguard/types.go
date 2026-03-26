package adguard

// Stats represents AdGuard Home DNS statistics.
type Stats struct {
	NumDNSQueries           int                `json:"num_dns_queries"`
	NumBlockedFiltering     int                `json:"num_blocked_filtering"`
	NumReplacedSafebrowsing int                `json:"num_replaced_safebrowsing"`
	NumReplacedParental     int                `json:"num_replaced_parental"`
	NumReplacedSafesearch   int                `json:"num_replaced_safesearch"`
	AvgProcessingTime       float64            `json:"avg_processing_time"`
	TopQueriedDomains       []map[string]int   `json:"top_queried_domains"`
	TopBlockedDomains       []map[string]int   `json:"top_blocked_domains"`
	TopClients              []map[string]int   `json:"top_clients"`
	DNSQueries              []int              `json:"dns_queries"`
	BlockedFiltering        []int              `json:"blocked_filtering"`
}

// FilteringStatus represents the current filtering state.
type FilteringStatus struct {
	Enabled  bool     `json:"enabled"`
	Interval int      `json:"interval"`
	Filters  []Filter `json:"filters"`
}

// Filter represents a single filtering list.
type Filter struct {
	Enabled     bool   `json:"enabled"`
	ID          int    `json:"id"`
	Name        string `json:"name"`
	RulesCount  int    `json:"rules_count"`
	URL         string `json:"url"`
	LastUpdated string `json:"last_updated"`
}

// StatsInfo combines stats and filtering status for the frontend.
type StatsInfo struct {
	TotalQueries    int     `json:"total_queries"`
	BlockedQueries  int     `json:"blocked_queries"`
	BlockedPercent  float64 `json:"blocked_percent"`
	AvgResponseTime float64 `json:"avg_response_time"`
	FilteringEnabled bool   `json:"filtering_enabled"`
	ActiveFilters   int     `json:"active_filters"`
	TotalRules      int     `json:"total_rules"`
}
