package pihole

// AuthRequest is the payload sent to authenticate with Pi-hole.
type AuthRequest struct {
	Password string `json:"password"`
}

// AuthResponse is the response from the Pi-hole auth endpoint.
type AuthResponse struct {
	Session SessionInfo `json:"session"`
}

// SessionInfo contains session details returned after authentication.
type SessionInfo struct {
	Valid    bool   `json:"valid"`
	SID      string `json:"sid"`
	CSRF     string `json:"csrf"`
	Validity int    `json:"validity"`
}

// StatsResponse is the raw response from /api/stats/summary.
type StatsResponse struct {
	Queries QueriesStats `json:"queries"`
	Took    float64      `json:"took"`
}

// QueriesStats contains the raw query statistics from Pi-hole.
type QueriesStats struct {
	Total          int     `json:"total"`
	Blocked        int     `json:"blocked"`
	PercentBlocked float64 `json:"percent_blocked"`
	UniqueDomains  int     `json:"unique_domains"`
	Forwarded      int     `json:"forwarded"`
	Cached         int     `json:"cached"`
}

// BlockingResponse is the response from /api/dns/blocking.
type BlockingResponse struct {
	Blocking bool `json:"blocking"`
}

// StatsInfo combines stats and blocking status for the frontend.
type StatsInfo struct {
	TotalQueries     int     `json:"total_queries"`
	BlockedQueries   int     `json:"blocked_queries"`
	BlockedPercent   float64 `json:"blocked_percent"`
	UniqueDomains    int     `json:"unique_domains"`
	CachedQueries    int     `json:"cached_queries"`
	ForwardedQueries int     `json:"forwarded_queries"`
	BlockingEnabled  bool    `json:"blocking_enabled"`
}
