package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/grafana"
)

// GrafanaHandler handles Grafana HTTP requests.
type GrafanaHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *GrafanaHandler) getGrafanaClient(ctx context.Context) (*grafana.Client, error) {
	svc, err := h.DB.GetServiceByType(ctx, "grafana")
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	client := grafana.NewClient(svc.URL, nil)

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
			Token string `json:"token"`
		}
		if err := json.Unmarshal([]byte(creds), &authData); err != nil {
			return client, nil
		}
		if authData.Token != "" {
			client.SetToken(authData.Token)
		}
	}

	return client, nil
}

// ListDashboards godoc
// @Summary List all Grafana dashboards
// @Tags grafana
// @Produce json
// @Success 200 {array} grafana.DashboardSearchResult
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /grafana/dashboards [get]
// @Security BearerAuth
func (h *GrafanaHandler) ListDashboards(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.Cache.Get("grafana:dashboards"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getGrafanaClient(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to Grafana"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Grafana not configured"})
		return
	}

	dashboards, err := client.SearchDashboards(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch dashboards"})
		return
	}

	h.Cache.Set("grafana:dashboards", dashboards)
	writeJSON(w, http.StatusOK, dashboards)
}

// GetDashboardPanels godoc
// @Summary Get panels for a Grafana dashboard
// @Tags grafana
// @Produce json
// @Param uid path string true "Dashboard UID"
// @Success 200 {object} grafana.DashboardInfo
// @Failure 404 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /grafana/dashboards/{uid}/panels [get]
// @Security BearerAuth
func (h *GrafanaHandler) GetDashboardPanels(w http.ResponseWriter, r *http.Request) {
	uid := chi.URLParam(r, "uid")

	cacheKey := "grafana:panels:" + uid
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getGrafanaClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Grafana not configured"})
		return
	}

	detail, err := client.GetDashboard(r.Context(), uid)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch dashboard"})
		return
	}

	info := grafana.BuildDashboardInfo(detail)
	h.Cache.Set(cacheKey, info)
	writeJSON(w, http.StatusOK, info)
}

// GetEmbedURL godoc
// @Summary Get the iframe embed URL for a Grafana panel
// @Tags grafana
// @Produce json
// @Param dashboardUid query string true "Dashboard UID"
// @Param panelId query int true "Panel ID"
// @Param theme query string false "Theme (dark or light)"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /grafana/embed-url [get]
// @Security BearerAuth
func (h *GrafanaHandler) GetEmbedURL(w http.ResponseWriter, r *http.Request) {
	dashboardUID := r.URL.Query().Get("dashboardUid")
	panelIDStr := r.URL.Query().Get("panelId")
	theme := r.URL.Query().Get("theme")

	if dashboardUID == "" || panelIDStr == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "dashboardUid and panelId are required"})
		return
	}

	panelID, err := strconv.Atoi(panelIDStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "panelId must be a number"})
		return
	}

	if theme == "" {
		theme = "dark"
	}

	svc, err := h.DB.GetServiceByType(r.Context(), "grafana")
	if err != nil || svc == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Grafana not configured"})
		return
	}

	url := fmt.Sprintf("%s/d-solo/%s/_?orgId=1&panelId=%d&theme=%s",
		svc.URL, dashboardUID, panelID, theme)

	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}
