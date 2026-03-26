package jdownloader

// Device represents a JDownloader device from MyJDownloader.
type Device struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type connectResponse struct {
	SessionToken string `json:"sessiontoken"`
	RegainToken  string `json:"regaintoken"`
}

type listDevicesResponse struct {
	List []Device `json:"list"`
}

// DownloadPackage represents a download package in the queue.
type DownloadPackage struct {
	UUID           int64  `json:"uuid"`
	Name           string `json:"name"`
	Status         string `json:"status"`
	BytesTotal     int64  `json:"bytesTotal"`
	BytesLoaded    int64  `json:"bytesLoaded"`
	Speed          int64  `json:"speed"`
	ETA            int64  `json:"eta"`
	Finished       bool   `json:"finished"`
	Enabled        bool   `json:"enabled"`
	ChildCount     int    `json:"childCount"`
	Comment        string `json:"comment,omitempty"`
	SaveTo         string `json:"saveTo,omitempty"`
}

// DownloadLink represents a single download link within a package.
type DownloadLink struct {
	UUID        int64  `json:"uuid"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Status      string `json:"status"`
	BytesTotal  int64  `json:"bytesTotal"`
	BytesLoaded int64  `json:"bytesLoaded"`
	Speed       int64  `json:"speed"`
	ETA         int64  `json:"eta"`
	Finished    bool   `json:"finished"`
	Enabled     bool   `json:"enabled"`
	PackageUUID int64  `json:"packageUUID"`
}

// PackageInfo represents a package for the Nidus frontend.
type PackageInfo struct {
	UUID       int64   `json:"uuid"`
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	Progress   float64 `json:"progress"`
	Size       int64   `json:"size"`
	Downloaded int64   `json:"downloaded"`
	Speed      int64   `json:"speed"`
	ETA        int64   `json:"eta"`
	Finished   bool    `json:"finished"`
	LinkCount  int     `json:"link_count"`
}

// QueueInfo combines queue data for the frontend.
type QueueInfo struct {
	Packages   []PackageInfo `json:"packages"`
	TotalSpeed int64         `json:"total_speed"`
	Running    bool          `json:"running"`
}

// APIResponse represents a JDownloader API response.
type APIResponse struct {
	Data interface{} `json:"data"`
}

// AddLinksRequest represents a request to add links.
type AddLinksRequest struct {
	Links []string `json:"links"`
}
