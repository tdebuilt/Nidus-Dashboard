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
	"github.com/tdebuilt/nidus/internal/services/adguard"
)

// AdGuardHandler handles AdGuard Home HTTP requests.
type AdGuardHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *AdGuardHandler) getAdGuardClient(ctx context.Context) (*adguard.Client, error) {
	svc, err := h.DB.GetServiceByType(ctx, "adguard")
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	client := adguard.NewClient(svc.URL, nil)

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
			// Not JSON — cannot determine username/password
			return client, nil
		}
		if authData.Username != "" {
			client.SetCredentials(authData.Username, authData.Password)
		}
	}

	return client, nil
}

// GetStats godoc
// @Summary Get AdGuard Home filtering statistics
// @Tags adguard
// @Produce json
// @Success 200 {object} adguard.StatsInfo
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /adguard/stats [get]
// @Security BearerAuth
func (h *AdGuardHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.Cache.Get("adguard:stats"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getAdGuardClient(r.Context())
	if err != nil {
		slog.Error("adguard: failed to connect", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to AdGuard"})
		return
	}
	if client == nil {
		slog.Warn("adguard: not configured")
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "AdGuard not configured"})
		return
	}

	stats, err := client.GetStats(r.Context())
	if err != nil {
		slog.Error("adguard: failed to fetch stats", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch stats"})
		return
	}

	filtering, err := client.GetFilteringStatus(r.Context())
	if err != nil {
		slog.Error("adguard: failed to fetch filtering status", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch filtering status"})
		return
	}

	blockedPercent := float64(0)
	if stats.NumDNSQueries > 0 {
		blockedPercent = float64(stats.NumBlockedFiltering) / float64(stats.NumDNSQueries) * 100
	}

	activeFilters := 0
	totalRules := 0
	for _, f := range filtering.Filters {
		if f.Enabled {
			activeFilters++
			totalRules += f.RulesCount
		}
	}

	result := adguard.StatsInfo{
		TotalQueries:     stats.NumDNSQueries,
		BlockedQueries:   stats.NumBlockedFiltering,
		BlockedPercent:   blockedPercent,
		AvgResponseTime:  stats.AvgProcessingTime,
		FilteringEnabled: filtering.Enabled,
		ActiveFilters:    activeFilters,
		TotalRules:       totalRules,
	}

	h.Cache.Set("adguard:stats", result)
	writeJSON(w, http.StatusOK, result)
}

// ToggleFiltering godoc
// @Summary Toggle AdGuard Home filtering on or off
// @Tags adguard
// @Accept json
// @Produce json
// @Param body body object true "Filtering state" SchemaExample({"enabled": true})
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /adguard/filtering/toggle [post]
// @Security BearerAuth
func (h *AdGuardHandler) ToggleFiltering(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	client, err := h.getAdGuardClient(r.Context())
	if err != nil || client == nil {
		slog.Error("adguard: not available for toggle filtering", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "AdGuard not available"})
		return
	}

	if err := client.SetFilteringEnabled(r.Context(), body.Enabled); err != nil {
		slog.Error("adguard: toggle filtering failed", "enabled", body.Enabled, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "toggle failed: " + err.Error()})
		return
	}

	slog.Info("adguard: filtering toggled", "enabled", body.Enabled)
	h.Cache.InvalidatePrefix("adguard:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
