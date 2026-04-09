package qbittorrent

// Torrent represents a torrent from the qBittorrent Web API (/api/v2/torrents/info).
type Torrent struct {
	Hash         string  `json:"hash"`
	Name         string  `json:"name"`
	Size         int64   `json:"size"`
	TotalSize    int64   `json:"total_size"`
	Progress     float64 `json:"progress"`
	Dlspeed      int64   `json:"dlspeed"`
	Upspeed      int64   `json:"upspeed"`
	ETA          int64   `json:"eta"`
	Ratio        float64 `json:"ratio"`
	State        string  `json:"state"`
	AddedOn      int64   `json:"added_on"`
	CompletionOn int64   `json:"completion_on"`
	NumSeeds     int     `json:"num_seeds"`
	NumLeechs    int     `json:"num_leechs"`
	Category     string  `json:"category"`
	Tags         string  `json:"tags"`
	Downloaded   int64   `json:"downloaded"`
	Uploaded     int64   `json:"uploaded"`
	Error        string  `json:"error_string,omitempty"`
}

// TransferInfo represents global transfer stats from /api/v2/transfer/info.
type TransferInfo struct {
	DlInfoSpeed int64 `json:"dl_info_speed"`
	UpInfoSpeed int64 `json:"up_info_speed"`
	DlInfoData  int64 `json:"dl_info_data"`
	UpInfoData  int64 `json:"up_info_data"`
}

// TorrentInfo is the frontend DTO for a single torrent.
type TorrentInfo struct {
	Hash       string  `json:"hash"`
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	Progress   float64 `json:"progress"`
	Size       int64   `json:"size"`
	Downloaded int64   `json:"downloaded"`
	SpeedDown  int64   `json:"speed_down"`
	SpeedUp    int64   `json:"speed_up"`
	ETA        int64   `json:"eta"`
	Ratio      float64 `json:"ratio"`
	Seeds      int     `json:"seeds"`
	Leechers   int     `json:"leechers"`
	Category   string  `json:"category,omitempty"`
	Tags       string  `json:"tags,omitempty"`
	AddedOn    int64   `json:"added_on"`
	Error      string  `json:"error,omitempty"`
}

// TorrentsInfo combines the torrent list and global stats for the frontend.
type TorrentsInfo struct {
	Torrents      []TorrentInfo `json:"torrents"`
	DownloadSpeed int64         `json:"download_speed"`
	UploadSpeed   int64         `json:"upload_speed"`
	TotalCount    int           `json:"total_count"`
	ActiveCount   int           `json:"active_count"`
}

// Category represents a qBittorrent category from /api/v2/torrents/categories.
type Category struct {
	Name     string `json:"name"`
	SavePath string `json:"savePath"`
}

// AddOptions carries the parameters accepted by the qBittorrent "add torrent"
// endpoint. Exactly one of URL or File must be set.
type AddOptions struct {
	URL      string // magnet link or HTTP URL to a .torrent
	File     []byte // raw .torrent file contents
	Category string // optional category name
	SavePath string // optional save path (overrides category's savePath)
}

// stateToStatus maps qBittorrent states to simplified Nidus statuses.
func stateToStatus(state string) string {
	switch state {
	case "downloading", "stalledDL", "forcedDL", "metaDL", "allocating":
		return "downloading"
	case "uploading", "stalledUP", "forcedUP":
		return "seeding"
	case "pausedDL", "pausedUP", "stoppedDL", "stoppedUP":
		return "paused"
	case "checkingDL", "checkingUP", "checkingResumeData", "moving":
		return "checking"
	case "queuedDL", "queuedUP":
		return "queued"
	case "error", "missingFiles":
		return "error"
	default:
		return "unknown"
	}
}

// ToTorrentInfo converts a raw API Torrent to a frontend TorrentInfo.
func ToTorrentInfo(t Torrent) TorrentInfo {
	status := stateToStatus(t.State)
	return TorrentInfo{
		Hash:       t.Hash,
		Name:       t.Name,
		Status:     status,
		Progress:   t.Progress * 100,
		Size:       t.Size,
		Downloaded: t.Downloaded,
		SpeedDown:  t.Dlspeed,
		SpeedUp:    t.Upspeed,
		ETA:        t.ETA,
		Ratio:      t.Ratio,
		Seeds:      t.NumSeeds,
		Leechers:   t.NumLeechs,
		Category:   t.Category,
		Tags:       t.Tags,
		AddedOn:    t.AddedOn,
		Error:      t.Error,
	}
}
