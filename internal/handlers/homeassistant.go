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
	"github.com/tdebuilt/nidus/internal/services/homeassistant"
	nidusws "github.com/tdebuilt/nidus/internal/websocket"
)

// HomeAssistantHandler handles Home Assistant HTTP requests.
type HomeAssistantHandler struct {
	DB    *database.DB
	Cache *cache.Cache
	Hub   *nidusws.Hub
	wsClient *homeassistant.WSClient
}

func (h *HomeAssistantHandler) getHAClient(ctx context.Context) (*homeassistant.Client, string, error) {
	svc, err := h.DB.GetServiceByType(ctx, "homeassistant")
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, "", err
	}
	if svc == nil {
		return nil, "", nil
	}

	client := homeassistant.NewClient(svc.URL, nil)
	token := ""

	if svc.Credentials != "" {
		encKey, err := h.DB.GetSystemSetting(ctx, "encryption_key")
		if err != nil || encKey == "" {
			return nil, "", err
		}
		creds, err := crypto.Decrypt(svc.Credentials, encKey)
		if err != nil {
			return nil, "", err
		}
		var authData struct {
			Token string `json:"token"`
		}
		if err := json.Unmarshal([]byte(creds), &authData); err != nil {
			// Not JSON — treat as raw token
			client.SetToken(creds)
			return client, creds, nil
		}
		if authData.Token != "" {
			client.SetToken(authData.Token)
			token = authData.Token
		}
	}

	return client, token, nil
}

func (h *HomeAssistantHandler) ensureWSClient(ctx context.Context, baseURL, token string) {
	if h.wsClient != nil || h.Hub == nil || token == "" {
		return
	}
	svc, err := h.DB.GetServiceByType(ctx, "homeassistant")
	if (err != nil && !errors.Is(err, database.ErrNotFound)) || svc == nil {
		return
	}
	h.wsClient = homeassistant.NewWSClient(svc.URL, token, h.Hub)
	h.wsClient.OnStateChanged = func() {
		h.Cache.Invalidate("ha:entities")
	}
	go func() {
		if err := h.wsClient.Connect(); err != nil {
			slog.Error("homeassistant ws connect failed", "error", err)
		}
	}()
}

// ListEntities godoc
// @Summary List all Home Assistant entities
// @Tags homeassistant
// @Produce json
// @Success 200 {array} homeassistant.EntityInfo
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /homeassistant/entities [get]
// @Security BearerAuth
func (h *HomeAssistantHandler) ListEntities(w http.ResponseWriter, r *http.Request) {
	// Always ensure WS client is connected for real-time updates
	if h.wsClient == nil {
		if _, token, err := h.getHAClient(r.Context()); err == nil && token != "" {
			h.ensureWSClient(r.Context(), "", token)
		}
	}

	if val, ok := h.Cache.Get("ha:entities"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, token, err := h.getHAClient(r.Context())
	if err != nil {
		slog.Error("homeassistant: failed to connect", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to Home Assistant"})
		return
	}
	if client == nil {
		slog.Warn("homeassistant: not configured")
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Home Assistant not configured"})
		return
	}

	entities, err := client.ListStates(r.Context())
	if err != nil {
		slog.Error("homeassistant: failed to fetch entities", "error", err)
		if errors.Is(err, homeassistant.ErrAuth) {
			writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "authentication_failed"})
			return
		}
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch entities"})
		return
	}

	result := make([]homeassistant.EntityInfo, 0, len(entities))
	for _, e := range entities {
		result = append(result, homeassistant.ToEntityInfo(e))
	}

	h.Cache.Set("ha:entities", result)
	h.ensureWSClient(r.Context(), "", token)
	writeJSON(w, http.StatusOK, result)
}

// GetEntity godoc
// @Summary Get a single Home Assistant entity state
// @Tags homeassistant
// @Produce json
// @Param entityId path string true "Entity ID (e.g. sensor.temperature)"
// @Success 200 {object} homeassistant.EntityInfo
// @Failure 502 {object} models.ErrorResponse
// @Router /homeassistant/entities/{entityId} [get]
// @Security BearerAuth
func (h *HomeAssistantHandler) GetEntity(w http.ResponseWriter, r *http.Request) {
	entityID := chi.URLParam(r, "entityId")

	cacheKey := "ha:entity:" + entityID
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, _, err := h.getHAClient(r.Context())
	if err != nil || client == nil {
		slog.Error("homeassistant: not available for GetEntity", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Home Assistant not available"})
		return
	}

	entity, err := client.GetState(r.Context(), entityID)
	if err != nil {
		slog.Error("homeassistant: failed to fetch entity", "entity_id", entityID, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch entity"})
		return
	}

	info := homeassistant.ToEntityInfo(*entity)
	h.Cache.Set(cacheKey, info)
	writeJSON(w, http.StatusOK, info)
}

// CallService godoc
// @Summary Call a Home Assistant service
// @Tags homeassistant
// @Accept json
// @Produce json
// @Param domain path string true "Service domain (e.g. light, switch)"
// @Param service path string true "Service name (e.g. turn_on, turn_off)"
// @Param body body homeassistant.ServiceCallRequest true "Service call payload"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /homeassistant/services/{domain}/{service} [post]
// @Security BearerAuth
func (h *HomeAssistantHandler) CallService(w http.ResponseWriter, r *http.Request) {
	domain := chi.URLParam(r, "domain")
	service := chi.URLParam(r, "service")

	var req homeassistant.ServiceCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	client, _, err := h.getHAClient(r.Context())
	if err != nil || client == nil {
		slog.Error("homeassistant: not available for CallService", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Home Assistant not available"})
		return
	}

	if _, err := client.CallService(r.Context(), domain, service, req); err != nil {
		slog.Error("homeassistant: service call failed", "domain", domain, "service", service, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "service call failed"})
		return
	}

	slog.Info("homeassistant: service called", "domain", domain, "service", service)
	h.Cache.InvalidatePrefix("ha:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// CameraSnapshot godoc
// @Summary Get a camera entity snapshot image
// @Tags homeassistant
// @Produce image/jpeg
// @Param entityId path string true "Camera entity ID"
// @Success 200 {file} binary
// @Failure 502 {object} models.ErrorResponse
// @Router /homeassistant/camera/{entityId}/snapshot [get]
// @Security BearerAuth
func (h *HomeAssistantHandler) CameraSnapshot(w http.ResponseWriter, r *http.Request) {
	entityID := chi.URLParam(r, "entityId")

	client, _, err := h.getHAClient(r.Context())
	if err != nil || client == nil {
		slog.Error("homeassistant: not available for CameraSnapshot", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Home Assistant not available"})
		return
	}

	data, contentType, err := client.GetCameraSnapshot(r.Context(), entityID)
	if err != nil {
		slog.Error("homeassistant: failed to get camera snapshot", "entity_id", entityID, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to get snapshot"})
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}
