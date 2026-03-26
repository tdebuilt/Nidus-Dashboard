package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

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

func (h *TransmissionHandler) getTransmissionClient() (*transmission.Client, error) {
	svc, err := h.DB.GetServiceByType("transmission")
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	client := transmission.NewClient(svc.URL, nil)

	if svc.Credentials != "" {
		encKey, err := h.DB.GetSystemSetting("encryption_key")
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

	client, err := h.getTransmissionClient()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to Transmission"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Transmission not configured"})
		return
	}

	torrents, err := client.ListTorrents()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch torrents"})
		return
	}

	stats, _ := client.GetSessionStats()

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

	client, err := h.getTransmissionClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	if body.Metainfo != "" {
		err = client.AddTorrentByFile(body.Metainfo)
	} else {
		err = client.AddTorrent(body.URL)
	}
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to add torrent: " + err.Error()})
		return
	}

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
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid torrent ID"})
		return
	}

	client, err := h.getTransmissionClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	if err := client.StartTorrent([]int{id}); err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "start failed: " + err.Error()})
		return
	}

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
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid torrent ID"})
		return
	}

	client, err := h.getTransmissionClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	if err := client.StopTorrent([]int{id}); err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "stop failed: " + err.Error()})
		return
	}

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
	client, err := h.getTransmissionClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	if err := client.StartAll(); err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "start all failed: " + err.Error()})
		return
	}

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
	client, err := h.getTransmissionClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	if err := client.StopAll(); err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "stop all failed: " + err.Error()})
		return
	}

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
	client, err := h.getTransmissionClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Transmission not available"})
		return
	}

	count, err := client.RemoveCompleted()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "cleanup failed: " + err.Error()})
		return
	}

	h.Cache.InvalidatePrefix("tx:")
	writeJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "removed": count})
}
