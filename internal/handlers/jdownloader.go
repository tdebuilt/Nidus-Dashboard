package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/jdownloader"
)

// JDownloaderHandler handles JDownloader HTTP requests.
type JDownloaderHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *JDownloaderHandler) getJDClient(ctx context.Context) (*jdownloader.Client, error) {
	// Return cached client if available
	if val, ok := h.Cache.Get("jd:client"); ok {
		if client, ok := val.(*jdownloader.Client); ok {
			return client, nil
		}
	}

	svc, err := h.DB.GetServiceByType(ctx, "jdownloader")
	if err != nil {
		return nil, err
	}
	if svc == nil || svc.Credentials == "" {
		return nil, nil
	}

	encKey, err := h.DB.GetSystemSetting(ctx, "encryption_key")
	if err != nil || encKey == "" {
		return nil, err
	}
	creds, err := crypto.Decrypt(svc.Credentials, encKey)
	if err != nil {
		return nil, err
	}

	var config struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal([]byte(creds), &config); err != nil {
		return nil, nil
	}
	if config.Email == "" || config.Password == "" {
		return nil, nil
	}

	client := jdownloader.NewClient(config.Email, config.Password)
	h.Cache.Set("jd:client", client)
	return client, nil
}

// GetQueue godoc
// @Summary Get JDownloader download queue
// @Tags jdownloader
// @Produce json
// @Success 200 {object} jdownloader.QueueInfo
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /jdownloader/queue [get]
// @Security BearerAuth
func (h *JDownloaderHandler) GetQueue(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.Cache.Get("jd:queue"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getJDClient(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to JDownloader"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "JDownloader not configured"})
		return
	}

	packages, err := client.ListPackages(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch queue: " + err.Error()})
		return
	}

	speed, err := client.GetSpeed(r.Context())
	if err != nil {
		slog.Warn("failed to get JDownloader speed", "error", err)
	}
	running, err := client.IsRunning(r.Context())
	if err != nil {
		slog.Warn("failed to check JDownloader running state", "error", err)
	}

	infos := make([]jdownloader.PackageInfo, 0, len(packages))
	for _, p := range packages {
		infos = append(infos, jdownloader.ToPackageInfo(p))
	}

	result := jdownloader.QueueInfo{
		Packages:   infos,
		TotalSpeed: speed,
		Running:    running,
	}

	h.Cache.Set("jd:queue", result)
	writeJSON(w, http.StatusOK, result)
}

// AddLinks godoc
// @Summary Add download links to JDownloader
// @Tags jdownloader
// @Accept json
// @Produce json
// @Param body body object true "Links to add" SchemaExample({"links": ["https://example.com/file.zip"]})
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /jdownloader/links [post]
// @Security BearerAuth
func (h *JDownloaderHandler) AddLinks(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Links []string `json:"links"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}
	if len(body.Links) == 0 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "no links provided"})
		return
	}

	client, err := h.getJDClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "JDownloader not available"})
		return
	}

	if err := client.AddLinks(r.Context(), body.Links); err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to add links: " + err.Error()})
		return
	}

	h.Cache.InvalidatePrefix("jd:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// StartQueue godoc
// @Summary Start the JDownloader download queue
// @Tags jdownloader
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 502 {object} models.ErrorResponse
// @Router /jdownloader/queue/start [post]
// @Security BearerAuth
func (h *JDownloaderHandler) StartQueue(w http.ResponseWriter, r *http.Request) {
	client, err := h.getJDClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "JDownloader not available"})
		return
	}

	if err := client.StartQueue(r.Context()); err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to start queue: " + err.Error()})
		return
	}

	h.Cache.InvalidatePrefix("jd:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CleanupFinished godoc
// @Summary Remove finished packages from the JDownloader queue
// @Tags jdownloader
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 502 {object} models.ErrorResponse
// @Router /jdownloader/queue/cleanup [post]
// @Security BearerAuth
func (h *JDownloaderHandler) CleanupFinished(w http.ResponseWriter, r *http.Request) {
	client, err := h.getJDClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "JDownloader not available"})
		return
	}

	count, err := client.CleanupFinished(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to cleanup: " + err.Error()})
		return
	}

	h.Cache.InvalidatePrefix("jd:")
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "removed": count})
}

// PauseQueue godoc
// @Summary Pause the JDownloader download queue
// @Tags jdownloader
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 502 {object} models.ErrorResponse
// @Router /jdownloader/queue/pause [post]
// @Security BearerAuth
func (h *JDownloaderHandler) PauseQueue(w http.ResponseWriter, r *http.Request) {
	client, err := h.getJDClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "JDownloader not available"})
		return
	}

	if err := client.PauseQueue(r.Context()); err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to pause queue: " + err.Error()})
		return
	}

	h.Cache.InvalidatePrefix("jd:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
