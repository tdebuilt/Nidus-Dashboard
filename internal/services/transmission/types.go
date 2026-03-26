package transmission

import "encoding/json"

// Torrent status constants.
const (
	StatusStopped      = 0
	StatusCheckWait    = 1
	StatusChecking     = 2
	StatusDownloadWait = 3
	StatusDownloading  = 4
	StatusSeedWait     = 5
	StatusSeeding      = 6
)

// RPCRequest represents a Transmission RPC request.
type RPCRequest struct {
	Method    string      `json:"method"`
	Arguments interface{} `json:"arguments,omitempty"`
}

// RPCResponse represents a Transmission RPC response.
type RPCResponse struct {
	Result    string          `json:"result"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// Torrent represents a torrent from the RPC API.
type Torrent struct {
	ID            int     `json:"id"`
	Name          string  `json:"name"`
	Status        int     `json:"status"`
	TotalSize     int64   `json:"totalSize"`
	SizeWhenDone  int64   `json:"sizeWhenDone"`
	LeftUntilDone int64   `json:"leftUntilDone"`
	PercentDone   float64 `json:"percentDone"`
	RateDownload  int64   `json:"rateDownload"`
	RateUpload    int64   `json:"rateUpload"`
	ETA           int64   `json:"eta"`
	UploadRatio   float64 `json:"uploadRatio"`
	Peers         int     `json:"peersConnected"`
	AddedDate     int64   `json:"addedDate"`
	Error         int     `json:"error"`
	ErrorString   string  `json:"errorString"`
}

// TorrentListResponse wraps the torrent-get response.
type TorrentListResponse struct {
	Torrents []Torrent `json:"torrents"`
}

// SessionStats represents the session statistics.
type SessionStats struct {
	DownloadSpeed int64 `json:"downloadSpeed"`
	UploadSpeed   int64 `json:"uploadSpeed"`
	TorrentCount  int   `json:"torrentCount"`
	ActiveCount   int   `json:"activeTorrentCount"`
	PausedCount   int   `json:"pausedTorrentCount"`
}

// TorrentInfo represents a torrent for the Nidus frontend.
type TorrentInfo struct {
	ID        int     `json:"id"`
	Name      string  `json:"name"`
	Status    string  `json:"status"`
	Progress  float64 `json:"progress"`
	Size      int64   `json:"size"`
	Downloaded int64  `json:"downloaded"`
	SpeedDown int64   `json:"speed_down"`
	SpeedUp   int64   `json:"speed_up"`
	ETA       int64   `json:"eta"`
	Ratio     float64 `json:"ratio"`
	Peers     int     `json:"peers"`
	Error     string  `json:"error,omitempty"`
}

// TorrentsInfo combines torrent list and session stats for frontend.
type TorrentsInfo struct {
	Torrents      []TorrentInfo `json:"torrents"`
	DownloadSpeed int64         `json:"download_speed"`
	UploadSpeed   int64         `json:"upload_speed"`
	TotalCount    int           `json:"total_count"`
	ActiveCount   int           `json:"active_count"`
}
