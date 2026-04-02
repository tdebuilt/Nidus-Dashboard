package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/qbittorrent"
)

// QBittorrentHandler handles qBittorrent HTTP requests.
type QBittorrentHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *QBittorrentHandler) getClient(ctx context.Context) (*qbittorrent.Client, error) {
	svc, err := h.DB.GetServiceByType(ctx, "qbittorrent")
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	client := qbittorrent.NewClient(svc.URL, nil)

	if svc.Credentials != "" {
		encKey, err := h.DB.GetSystemSetting(ctx, "encryption_key")
		if err != nil || encKey == "" {
			return nil, err
		}
		creds, err := crypto.Decrypt(svc.Credentials, encKey)
		if err != nil {
			return nil, err
		}
		var authData struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal([]byte(creds), &authData); err == nil && authData.Username != "" {
			client.SetCredentials(authData.Username, authData.Password)
		}
	}

	return client, nil
}

// ListTorrents godoc
// @Summary List all torrents with transfer stats
// @Tags qbittorrent
// @Produce json
// @Success 200 {object} qbittorrent.TorrentsInfo
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /qbittorrent/torrents [get]
// @Security BearerAuth
func (h *QBittorrentHandler) ListTorrents(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.Cache.Get("qbt:torrents"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getClient(r.Context())
	if err != nil {
		slog.Error("qbittorrent: failed to connect", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to qBittorrent"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "qBittorrent not configured"})
		return
	}

	torrents, err := client.ListTorrents(r.Context())
	if err != nil {
		slog.Error("qbittorrent: failed to fetch torrents", "error", err)
		if errors.Is(err, qbittorrent.ErrAuth) {
			writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "authentication_failed"})
			return
		}
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch torrents"})
		return
	}

	transfer, err := client.GetTransferInfo(r.Context())
	if err != nil {
		slog.Warn("qbittorrent: failed to get transfer info", "error", err)
	}

	infos := make([]qbittorrent.TorrentInfo, 0, len(torrents))
	activeCount := 0
	for _, t := range torrents {
		info := qbittorrent.ToTorrentInfo(t)
		infos = append(infos, info)
		if info.Status == "downloading" || info.Status == "seeding" {
			activeCount++
		}
	}

	result := qbittorrent.TorrentsInfo{
		Torrents:   infos,
		TotalCount: len(infos),
		ActiveCount: activeCount,
	}
	if transfer != nil {
		result.DownloadSpeed = transfer.DlInfoSpeed
		result.UploadSpeed = transfer.UpInfoSpeed
	}

	h.Cache.Set("qbt:torrents", result)
	writeJSON(w, http.StatusOK, result)
}

// AddTorrent godoc
// @Summary Add a torrent by URL/magnet
// @Tags qbittorrent
// @Accept json
// @Produce json
// @Param body body object true "Torrent to add" SchemaExample({"url": "magnet:?xt=..."})
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /qbittorrent/torrents [post]
// @Security BearerAuth
func (h *QBittorrentHandler) AddTorrent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "url is required"})
		return
	}

	client, err := h.getClient(r.Context())
	if err != nil || client == nil {
		slog.Error("qbittorrent: not available for AddTorrent", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "qBittorrent not available"})
		return
	}

	if err := client.AddTorrent(r.Context(), body.URL); err != nil {
		slog.Error("qbittorrent: failed to add torrent", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to add torrent"})
		return
	}

	h.Cache.InvalidatePrefix("qbt:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ResumeTorrent godoc
// @Summary Resume a torrent
// @Tags qbittorrent
// @Produce json
// @Param hash path string true "Torrent hash"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /qbittorrent/torrents/{hash}/resume [post]
// @Security BearerAuth
func (h *QBittorrentHandler) ResumeTorrent(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "hash is required"})
		return
	}

	client, err := h.getClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "qBittorrent not available"})
		return
	}

	if err := client.ResumeTorrents(r.Context(), []string{hash}); err != nil {
		slog.Error("qbittorrent: resume failed", "hash", hash, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "resume failed"})
		return
	}

	h.Cache.InvalidatePrefix("qbt:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// PauseTorrent godoc
// @Summary Pause a torrent
// @Tags qbittorrent
// @Produce json
// @Param hash path string true "Torrent hash"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /qbittorrent/torrents/{hash}/pause [post]
// @Security BearerAuth
func (h *QBittorrentHandler) PauseTorrent(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "hash is required"})
		return
	}

	client, err := h.getClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "qBittorrent not available"})
		return
	}

	if err := client.PauseTorrents(r.Context(), []string{hash}); err != nil {
		slog.Error("qbittorrent: pause failed", "hash", hash, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "pause failed"})
		return
	}

	h.Cache.InvalidatePrefix("qbt:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// DeleteTorrent godoc
// @Summary Delete a torrent
// @Tags qbittorrent
// @Produce json
// @Param hash path string true "Torrent hash"
// @Param deleteFiles query bool false "Also delete downloaded files"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /qbittorrent/torrents/{hash}/delete [post]
// @Security BearerAuth
func (h *QBittorrentHandler) DeleteTorrent(w http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "hash is required"})
		return
	}

	deleteFiles := r.URL.Query().Get("deleteFiles") == "true"

	client, err := h.getClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "qBittorrent not available"})
		return
	}

	if err := client.DeleteTorrents(r.Context(), []string{hash}, deleteFiles); err != nil {
		slog.Error("qbittorrent: delete failed", "hash", hash, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "delete failed"})
		return
	}

	h.Cache.InvalidatePrefix("qbt:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ResumeAllTorrents godoc
// @Summary Resume all torrents
// @Tags qbittorrent
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 502 {object} models.ErrorResponse
// @Router /qbittorrent/torrents/resume-all [post]
// @Security BearerAuth
func (h *QBittorrentHandler) ResumeAllTorrents(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "qBittorrent not available"})
		return
	}

	if err := client.ResumeAll(r.Context()); err != nil {
		slog.Error("qbittorrent: resume all failed", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "resume all failed"})
		return
	}

	h.Cache.InvalidatePrefix("qbt:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// PauseAllTorrents godoc
// @Summary Pause all torrents
// @Tags qbittorrent
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 502 {object} models.ErrorResponse
// @Router /qbittorrent/torrents/pause-all [post]
// @Security BearerAuth
func (h *QBittorrentHandler) PauseAllTorrents(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "qBittorrent not available"})
		return
	}

	if err := client.PauseAll(r.Context()); err != nil {
		slog.Error("qbittorrent: pause all failed", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "pause all failed"})
		return
	}

	h.Cache.InvalidatePrefix("qbt:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CleanupCompleted godoc
// @Summary Remove all completed torrents without deleting files
// @Tags qbittorrent
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} models.ErrorResponse
// @Router /qbittorrent/torrents/cleanup [post]
// @Security BearerAuth
func (h *QBittorrentHandler) CleanupCompleted(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "qBittorrent not available"})
		return
	}

	torrents, err := client.ListTorrents(r.Context())
	if err != nil {
		slog.Error("qbittorrent: failed to list for cleanup", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to list torrents"})
		return
	}

	var hashes []string
	for _, t := range torrents {
		if t.Progress >= 1.0 {
			hashes = append(hashes, t.Hash)
		}
	}

	if len(hashes) == 0 {
		writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "removed": 0})
		return
	}

	if err := client.DeleteTorrents(r.Context(), hashes, false); err != nil {
		slog.Error("qbittorrent: cleanup failed", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "cleanup failed"})
		return
	}

	h.Cache.InvalidatePrefix("qbt:")
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "removed": len(hashes)})
}
