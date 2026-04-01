package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/arr"
)

const addCacheTTL = 5 * time.Minute

// GetQualityProfiles returns available quality profiles for a radarr or sonarr service.
func (h *ArrHandler) GetQualityProfiles(w http.ResponseWriter, r *http.Request) {
	serviceType := chi.URLParam(r, "type")
	if serviceType != serviceRadarr && serviceType != serviceSonarr {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid service type"})
		return
	}

	cacheKey := fmt.Sprintf("arr:%s:profiles", serviceType)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getArrClient(r.Context(), serviceType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "connection error"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "not_configured"})
		return
	}

	profiles, err := client.GetQualityProfiles(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch quality profiles"})
		return
	}

	h.Cache.SetWithTTL(cacheKey, profiles, addCacheTTL)
	writeJSON(w, http.StatusOK, profiles)
}

// GetRootFolders returns available root folders for a radarr or sonarr service.
func (h *ArrHandler) GetRootFolders(w http.ResponseWriter, r *http.Request) {
	serviceType := chi.URLParam(r, "type")
	if serviceType != serviceRadarr && serviceType != serviceSonarr {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid service type"})
		return
	}

	cacheKey := fmt.Sprintf("arr:%s:rootfolders", serviceType)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getArrClient(r.Context(), serviceType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "connection error"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "not_configured"})
		return
	}

	folders, err := client.GetRootFolders(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch root folders"})
		return
	}

	h.Cache.SetWithTTL(cacheKey, folders, addCacheTTL)
	writeJSON(w, http.StatusOK, folders)
}

// LookupMedia searches for movies or series by term.
func (h *ArrHandler) LookupMedia(w http.ResponseWriter, r *http.Request) {
	serviceType := chi.URLParam(r, "type")
	if serviceType != serviceRadarr && serviceType != serviceSonarr {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid service type"})
		return
	}

	term := r.URL.Query().Get("term")
	if term == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "term is required"})
		return
	}

	client, err := h.getArrClient(r.Context(), serviceType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "connection error"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "not_configured"})
		return
	}

	lookupPath := arr.Configs[serviceType].LibraryPath
	results, err := client.LookupMedia(r.Context(), lookupPath, term)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "search failed"})
		return
	}

	writeJSON(w, http.StatusOK, results)
}

// AddMedia adds a movie to Radarr or a series to Sonarr.
func (h *ArrHandler) AddMedia(w http.ResponseWriter, r *http.Request) {
	serviceType := chi.URLParam(r, "type")
	if serviceType != serviceRadarr && serviceType != serviceSonarr {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid service type"})
		return
	}

	client, err := h.getArrClient(r.Context(), serviceType)
	if err != nil || client == nil {
		slog.Error("arr: not available for AddMedia", "type", serviceType, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "service not available"})
		return
	}

	mediaPath := arr.Configs[serviceType].LibraryPath
	var body any

	if serviceType == serviceRadarr {
		var req arr.AddMovieRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TmdbID == 0 {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
			return
		}
		body = req
	} else {
		var req arr.AddSeriesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.TvdbID == 0 {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
			return
		}
		body = req
	}

	if err := client.AddMedia(r.Context(), mediaPath, body); err != nil {
		slog.Error("arr: failed to add media", "type", serviceType, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: err.Error()})
		return
	}

	h.Cache.InvalidatePrefix("arr:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
