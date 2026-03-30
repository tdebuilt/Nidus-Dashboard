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
	"github.com/tdebuilt/nidus/internal/services/pihole"
)

// PiholeHandler handles Pi-hole HTTP requests.
type PiholeHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *PiholeHandler) getPiholeClient(ctx context.Context) (*pihole.Client, error) {
	svc, err := h.DB.GetServiceByType(ctx, "pihole")
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	password := ""
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
			Password string `json:"password"`
		}
		if err := json.Unmarshal([]byte(creds), &authData); err == nil {
			password = authData.Password
		}
	}

	return pihole.NewClient(svc.URL, password, nil), nil
}

// GetStats godoc
// @Summary Get Pi-hole blocking statistics
// @Tags pihole
// @Produce json
// @Success 200 {object} pihole.StatsInfo
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /pihole/stats [get]
// @Security BearerAuth
func (h *PiholeHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.Cache.Get("pihole:stats"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getPiholeClient(r.Context())
	if err != nil {
		slog.Error("pihole: failed to connect", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to Pi-hole"})
		return
	}
	if client == nil {
		slog.Warn("pihole: not configured")
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Pi-hole not configured"})
		return
	}

	stats, err := client.GetStats(r.Context())
	if err != nil {
		slog.Error("pihole: failed to fetch stats", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch stats"})
		return
	}

	h.Cache.Set("pihole:stats", stats)
	writeJSON(w, http.StatusOK, stats)
}

// ToggleBlocking godoc
// @Summary Toggle Pi-hole blocking on or off
// @Tags pihole
// @Accept json
// @Produce json
// @Param body body object true "Blocking state" SchemaExample({"blocking": true})
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /pihole/blocking [post]
// @Security BearerAuth
func (h *PiholeHandler) ToggleBlocking(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Blocking bool `json:"blocking"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	client, err := h.getPiholeClient(r.Context())
	if err != nil || client == nil {
		slog.Error("pihole: not available for toggle blocking", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Pi-hole not available"})
		return
	}

	if err := client.SetBlocking(r.Context(), body.Blocking); err != nil {
		slog.Error("pihole: toggle blocking failed", "blocking", body.Blocking, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "toggle failed"})
		return
	}

	slog.Info("pihole: blocking toggled", "blocking", body.Blocking)
	h.Cache.InvalidatePrefix("pihole:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
