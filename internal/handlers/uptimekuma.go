package handlers

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/uptimekuma"
)

// UptimeKumaHandler handles Uptime Kuma HTTP requests.
type UptimeKumaHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *UptimeKumaHandler) getUptimeKumaClient(ctx context.Context) (*uptimekuma.Client, error) {
	svc, err := h.DB.GetServiceByType(ctx, "uptimekuma")
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	client := uptimekuma.NewClient(svc.URL, nil)
	return client, nil
}

// GetMonitors godoc
// @Summary Get monitors for an Uptime Kuma status page
// @Tags uptimekuma
// @Produce json
// @Param slug path string true "Status page slug"
// @Success 200 {object} uptimekuma.MonitorsOverview
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /uptimekuma/monitors/{slug} [get]
// @Security BearerAuth
func (h *UptimeKumaHandler) GetMonitors(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		slug = "default"
	}

	cacheKey := "kuma:monitors:" + slug
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getUptimeKumaClient(r.Context())
	if err != nil {
		slog.Error("uptimekuma: failed to connect", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to Uptime Kuma"})
		return
	}
	if client == nil {
		slog.Warn("uptimekuma: not configured")
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Uptime Kuma not configured"})
		return
	}

	overview, err := client.GetMonitors(r.Context(), slug)
	if err != nil {
		slog.Error("uptimekuma: failed to fetch monitors", "slug", slug, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch monitors"})
		return
	}

	h.Cache.Set(cacheKey, overview)
	writeJSON(w, http.StatusOK, overview)
}
