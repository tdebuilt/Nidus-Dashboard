package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/mediaserver"
)

// MediaServerHandler handles Plex/Jellyfin HTTP requests.
type MediaServerHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

var validMediaTypes = map[string]bool{
	"plex":     true,
	"jellyfin": true,
}

func (h *MediaServerHandler) getMediaClient(serverType string) (mediaserver.Client, error) {
	svc, err := h.DB.GetServiceByType(serverType)
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	// Decrypt credentials to get the token/API key
	token := ""
	if svc.Credentials != "" {
		encKey, err := h.DB.GetSystemSetting("encryption_key")
		if err == nil && encKey != "" {
			decrypted, err := crypto.Decrypt(svc.Credentials, encKey)
			if err == nil {
				var creds struct {
					Token  string `json:"token"`
					APIKey string `json:"api_key"`
				}
				if json.Unmarshal([]byte(decrypted), &creds) == nil {
					if creds.Token != "" {
						token = creds.Token
					} else if creds.APIKey != "" {
						token = creds.APIKey
					}
				}
			}
		}
	}

	switch serverType {
	case "plex":
		return mediaserver.NewPlexClient(svc.URL, token, nil), nil
	case "jellyfin":
		return mediaserver.NewJellyfinClient(svc.URL, token, nil), nil
	default:
		return nil, nil
	}
}

// GetSessions godoc
// @Summary Get active media server sessions
// @Tags mediaserver
// @Produce json
// @Param type path string true "Media server type" Enums(plex, jellyfin)
// @Success 200 {object} mediaserver.MediaOverview
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /mediaserver/{type}/sessions [get]
// @Security BearerAuth
func (h *MediaServerHandler) GetSessions(w http.ResponseWriter, r *http.Request) {
	serverType := chi.URLParam(r, "type")
	if !validMediaTypes[serverType] {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid media server type, must be 'plex' or 'jellyfin'"})
		return
	}

	cacheKey := serverType + ":sessions"
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getMediaClient(serverType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to media server"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: serverType + " not configured"})
		return
	}

	sessions, err := client.GetSessions()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch sessions"})
		return
	}

	serverName, _ := client.GetServerName()

	overview := mediaserver.MediaOverview{
		Sessions:     sessions,
		SessionCount: len(sessions),
		ServerName:   serverName,
		ServerType:   serverType,
	}

	h.Cache.Set(cacheKey, overview)
	writeJSON(w, http.StatusOK, overview)
}

// GetLibraries godoc
// @Summary Get media server libraries
// @Tags mediaserver
// @Produce json
// @Param type path string true "Media server type" Enums(plex, jellyfin)
// @Success 200 {array} mediaserver.Library
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /mediaserver/{type}/libraries [get]
// @Security BearerAuth
func (h *MediaServerHandler) GetLibraries(w http.ResponseWriter, r *http.Request) {
	serverType := chi.URLParam(r, "type")
	if !validMediaTypes[serverType] {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid media server type"})
		return
	}

	cacheKey := serverType + ":libraries"
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getMediaClient(serverType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to media server"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: serverType + " not configured"})
		return
	}

	libraries, err := client.GetLibraries()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch libraries"})
		return
	}

	h.Cache.Set(cacheKey, libraries)
	writeJSON(w, http.StatusOK, libraries)
}

// ProxyImage godoc
// @Summary Proxy an image from the media server
// @Tags mediaserver
// @Produce image/*
// @Param type path string true "Media server type" Enums(plex, jellyfin)
// @Param path query string true "Image path on the media server"
// @Success 200 {file} binary
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /mediaserver/{type}/proxy [get]
// @Security BearerAuth
func (h *MediaServerHandler) ProxyImage(w http.ResponseWriter, r *http.Request) {
	serverType := chi.URLParam(r, "type")
	if !validMediaTypes[serverType] {
		http.Error(w, "invalid media server type", http.StatusBadRequest)
		return
	}

	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "path parameter required", http.StatusBadRequest)
		return
	}
	if !strings.HasPrefix(path, "/") || strings.Contains(path, "..") || strings.Contains(path, "@") {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}

	client, err := h.getMediaClient(serverType)
	if err != nil || client == nil {
		http.Error(w, "media server not available", http.StatusNotFound)
		return
	}

	body, contentType, err := client.ProxyImage(path)
	if err != nil {
		http.Error(w, "failed to fetch image", http.StatusBadGateway)
		return
	}

	if contentType != "" {
		w.Header().Set("Content-Type", contentType)
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(body)
}
