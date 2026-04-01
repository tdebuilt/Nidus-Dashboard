package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/arr"
)

const libraryCacheTTL = 5 * time.Minute

// GetLibrary returns the full library for a radarr or sonarr service.
func (h *ArrHandler) GetLibrary(w http.ResponseWriter, r *http.Request) {
	serviceType := chi.URLParam(r, "type")
	if serviceType != "radarr" && serviceType != "sonarr" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid service type"})
		return
	}

	cacheKey := fmt.Sprintf("arr:%s:library", serviceType)
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

	switch serviceType {
	case "radarr":
		h.serveRadarrLibrary(w, r, client, cacheKey)
	case "sonarr":
		h.serveSonarrLibrary(w, r, client, cacheKey)
	}
}

func (h *ArrHandler) serveRadarrLibrary(w http.ResponseWriter, r *http.Request, client *arr.Client, cacheKey string) {
	movies, err := client.GetRadarrLibrary(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch library"})
		return
	}
	result := map[string]any{"items": movies, "total": len(movies)}
	h.Cache.SetWithTTL(cacheKey, result, libraryCacheTTL)
	writeJSON(w, http.StatusOK, result)
}

func (h *ArrHandler) serveSonarrLibrary(w http.ResponseWriter, r *http.Request, client *arr.Client, cacheKey string) {
	series, err := client.GetSonarrLibrary(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch library"})
		return
	}
	result := map[string]any{"items": series, "total": len(series)}
	h.Cache.SetWithTTL(cacheKey, result, libraryCacheTTL)
	writeJSON(w, http.StatusOK, result)
}

// GetEpisodes returns episodes for a Sonarr series.
func (h *ArrHandler) GetEpisodes(w http.ResponseWriter, r *http.Request) {
	seriesID, ok := parseIntIDParam(w, r, "seriesId")
	if !ok {
		return
	}

	cacheKey := fmt.Sprintf("arr:sonarr:episodes:%d", seriesID)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getArrClient(r.Context(), "sonarr")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "connection error"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "not_configured"})
		return
	}

	episodes, err := client.GetSonarrEpisodes(r.Context(), seriesID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch episodes"})
		return
	}

	result := map[string]any{"items": episodes, "total": len(episodes)}
	h.Cache.SetWithTTL(cacheKey, result, libraryCacheTTL)
	writeJSON(w, http.StatusOK, result)
}
