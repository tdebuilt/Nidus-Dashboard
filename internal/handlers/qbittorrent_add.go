package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/qbittorrent"
)

// maxTorrentFileBytes caps the decoded size of an uploaded .torrent file.
// Real-world .torrent files are typically under 1 MB; 10 MB is a safe margin.
const maxTorrentFileBytes = 10 << 20

// maxAddTorrentRequestBytes caps the raw JSON request body for AddTorrent.
// Base64 inflates bytes by ~33%, plus JSON wrapping, so allow ~15 MB.
const maxAddTorrentRequestBytes = 15 << 20

// addTorrentRequest is the JSON body accepted by AddTorrent. Exactly one of
// URL or Metainfo must be set.
type addTorrentRequest struct {
	URL      string `json:"url,omitempty"`
	Metainfo string `json:"metainfo,omitempty"` // base64-encoded .torrent file
	Category string `json:"category,omitempty"`
	SavePath string `json:"save_path,omitempty"`
}

// AddTorrent godoc
// @Summary Add a torrent by URL/magnet or base64 .torrent file
// @Tags qbittorrent
// @Accept json
// @Produce json
// @Param body body object true "Torrent to add" SchemaExample({"url": "magnet:?xt=..."})
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 413 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /qbittorrent/torrents [post]
// @Security BearerAuth
func (h *QBittorrentHandler) AddTorrent(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxAddTorrentRequestBytes)

	var body addTorrentRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		if _, ok := err.(*http.MaxBytesError); ok {
			writeJSON(w, http.StatusRequestEntityTooLarge, models.ErrorResponse{Error: "file_too_large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	opts, err := buildAddOptions(body)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	client, err := h.getClient(r.Context())
	if err != nil || client == nil {
		slog.Error("qbittorrent: not available for AddTorrent", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "qBittorrent not available"})
		return
	}

	if err := client.AddTorrent(r.Context(), opts); err != nil {
		slog.Error("qbittorrent: failed to add torrent", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to add torrent"})
		return
	}

	h.Cache.InvalidatePrefix("qbt:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// buildAddOptions validates the JSON request and converts it into the client
// AddOptions struct, decoding the base64 metainfo and enforcing the file size
// cap.
func buildAddOptions(body addTorrentRequest) (qbittorrent.AddOptions, error) {
	if body.URL == "" && body.Metainfo == "" {
		return qbittorrent.AddOptions{}, errors.New("url or metainfo is required")
	}
	if body.URL != "" && body.Metainfo != "" {
		return qbittorrent.AddOptions{}, errors.New("provide either url or metainfo, not both")
	}
	opts := qbittorrent.AddOptions{
		URL:      body.URL,
		Category: body.Category,
		SavePath: body.SavePath,
	}
	if body.Metainfo != "" {
		data, err := base64.StdEncoding.DecodeString(body.Metainfo)
		if err != nil {
			return qbittorrent.AddOptions{}, errors.New("invalid base64 metainfo")
		}
		if len(data) > maxTorrentFileBytes {
			return qbittorrent.AddOptions{}, errors.New("file_too_large")
		}
		opts.File = data
	}
	return opts, nil
}

// GetCategories godoc
// @Summary List qBittorrent categories
// @Tags qbittorrent
// @Produce json
// @Success 200 {object} map[string]qbittorrent.Category
// @Failure 404 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /qbittorrent/categories [get]
// @Security BearerAuth
func (h *QBittorrentHandler) GetCategories(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r.Context())
	if err != nil {
		slog.Error("qbittorrent: failed to connect for GetCategories", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to qBittorrent"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "qBittorrent not configured"})
		return
	}

	categories, err := client.GetCategories(r.Context())
	if err != nil {
		slog.Error("qbittorrent: failed to fetch categories", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch categories"})
		return
	}

	writeJSON(w, http.StatusOK, categories)
}
