package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/transmission"
)

// TransmissionHandler handles Transmission HTTP requests.
type TransmissionHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *TransmissionHandler) getTransmissionClient(ctx context.Context) (*transmission.Client, error) {
	svc, err := h.DB.GetServiceByType(ctx, "transmission")
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	client := transmission.NewClient(svc.URL, nil)

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
		if err := json.Unmarshal([]byte(creds), &authData); err != nil {
			return client, nil
		}
		if authData.Username != "" {
			client.SetCredentials(authData.Username, authData.Password)
		}
	}

	return client, nil
}

// ListTorrents godoc
// @Summary List all torrents with session stats
// @Tags transmission
// @Produce json
// @Success 200 {object} transmission.TorrentsInfo
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /transmission/torrents [get]
// @Security BearerAuth
func (h *TransmissionHandler) ListTorrents(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.Cache.Get("tx:torrents"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getTransmissionClient(r.Context())
	if err != nil {
		slog.Error("transmission: failed to connect", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to Transmission"})
		return
	}
	if client == nil {
		slog.Warn("transmission: not configured")
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Transmission not configured"})
		return
	}

	torrents, err := client.ListTorrents(r.Context())
	if err != nil {
		slog.Error("transmission: failed to fetch torrents", "error", err)
		if errors.Is(err, transmission.ErrAuth) {
			writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "authentication_failed"})
			return
		}
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch torrents"})
		return
	}

	stats, err := client.GetSessionStats(r.Context())
	if err != nil {
		slog.Warn("failed to get session stats", "error", err)
	}

	infos := make([]transmission.TorrentInfo, 0, len(torrents))
	for _, t := range torrents {
		infos = append(infos, transmission.ToTorrentInfo(t))
	}

	result := transmission.TorrentsInfo{
		Torrents: infos,
	}
	if stats != nil {
		result.DownloadSpeed = stats.DownloadSpeed
		result.UploadSpeed = stats.UploadSpeed
		result.TotalCount = stats.TorrentCount
		result.ActiveCount = stats.ActiveCount
	}

	h.Cache.Set("tx:torrents", result)
	writeJSON(w, http.StatusOK, result)
}

// AddTorrent godoc
// @Summary Add a torrent by URL/magnet or base64 .torrent file
// @Tags transmission
// @Accept json
// @Produce json
// @Param body body object true "Torrent to add" SchemaExample({"url": "magnet:?xt=..."})
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /transmission/torrents [post]
// @Security BearerAuth
func (h *TransmissionHandler) AddTorrent(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL      string `json:"url"`
		Metainfo string `json:"metainfo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || (body.URL == "" && body.Metainfo == "") {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "url or metainfo is required"})
		return
	}

	client, err := h.getTransmissionClient(r.Context())
	if err != nil || client == nil {
		slog.Error("transmission: not available for AddTorrent", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	if body.Metainfo != "" {
		err = client.AddTorrentByFile(r.Context(), body.Metainfo)
	} else {
		err = client.AddTorrent(r.Context(), body.URL)
	}
	if err != nil {
		slog.Error("transmission: failed to add torrent", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to add torrent"})
		return
	}

	slog.Info("transmission: torrent added")
	h.Cache.InvalidatePrefix("tx:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// StartTorrent godoc
// @Summary Start a torrent
// @Tags transmission
// @Produce json
// @Param id path int true "Torrent ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /transmission/torrents/{id}/start [post]
// @Security BearerAuth
func (h *TransmissionHandler) StartTorrent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntIDParam(w, r, "id")
	if !ok {
		return
	}

	client, err := h.getTransmissionClient(r.Context())
	if err != nil || client == nil {
		slog.Error("transmission: not available for StartTorrent", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	if err := client.StartTorrent(r.Context(), []int{id}); err != nil {
		slog.Error("transmission: start torrent failed", "id", id, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "start failed"})
		return
	}

	slog.Info("transmission: torrent started", "id", id)
	h.Cache.InvalidatePrefix("tx:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// StopTorrent godoc
// @Summary Stop a torrent
// @Tags transmission
// @Produce json
// @Param id path int true "Torrent ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /transmission/torrents/{id}/stop [post]
// @Security BearerAuth
func (h *TransmissionHandler) StopTorrent(w http.ResponseWriter, r *http.Request) {
	id, ok := parseIntIDParam(w, r, "id")
	if !ok {
		return
	}

	client, err := h.getTransmissionClient(r.Context())
	if err != nil || client == nil {
		slog.Error("transmission: not available for StopTorrent", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	if err := client.StopTorrent(r.Context(), []int{id}); err != nil {
		slog.Error("transmission: stop torrent failed", "id", id, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "stop failed"})
		return
	}

	slog.Info("transmission: torrent stopped", "id", id)
	h.Cache.InvalidatePrefix("tx:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// StartAllTorrents godoc
// @Summary Start all torrents
// @Tags transmission
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 502 {object} models.ErrorResponse
// @Router /transmission/torrents/start-all [post]
// @Security BearerAuth
func (h *TransmissionHandler) StartAllTorrents(w http.ResponseWriter, r *http.Request) {
	client, err := h.getTransmissionClient(r.Context())
	if err != nil || client == nil {
		slog.Error("transmission: not available for StartAll", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	if err := client.StartAll(r.Context()); err != nil {
		slog.Error("transmission: start all failed", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "start all failed"})
		return
	}

	slog.Info("transmission: all torrents started")
	h.Cache.InvalidatePrefix("tx:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// StopAllTorrents godoc
// @Summary Stop all torrents
// @Tags transmission
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 502 {object} models.ErrorResponse
// @Router /transmission/torrents/stop-all [post]
// @Security BearerAuth
func (h *TransmissionHandler) StopAllTorrents(w http.ResponseWriter, r *http.Request) {
	client, err := h.getTransmissionClient(r.Context())
	if err != nil || client == nil {
		slog.Error("transmission: not available for StopAll", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	if err := client.StopAll(r.Context()); err != nil {
		slog.Error("transmission: stop all failed", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "stop all failed"})
		return
	}

	slog.Info("transmission: all torrents stopped")
	h.Cache.InvalidatePrefix("tx:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CleanupCompleted godoc
// @Summary Remove all completed torrents without deleting files
// @Tags transmission
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} models.ErrorResponse
// @Router /transmission/torrents/cleanup [post]
// @Security BearerAuth
func (h *TransmissionHandler) CleanupCompleted(w http.ResponseWriter, r *http.Request) {
	client, err := h.getTransmissionClient(r.Context())
	if err != nil || client == nil {
		slog.Error("transmission: not available for CleanupCompleted", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	count, err := client.RemoveCompleted(r.Context())
	if err != nil {
		slog.Error("transmission: cleanup failed", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "cleanup failed"})
		return
	}

	slog.Info("transmission: cleanup completed", "removed", count)
	h.Cache.InvalidatePrefix("tx:")
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "removed": count})
}
