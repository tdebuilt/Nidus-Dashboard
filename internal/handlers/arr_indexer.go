package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// GetProwlarrIndexers returns the list of Prowlarr indexers with status.
func (h *ArrHandler) GetProwlarrIndexers(w http.ResponseWriter, r *http.Request) {
	cacheKey := "arr:prowlarr:indexers"
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getArrClient(r.Context(), "prowlarr")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "connection error"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not configured"})
		return
	}

	indexers, err := client.GetProwlarrIndexers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	h.Cache.Set(cacheKey, indexers)
	writeJSON(w, http.StatusOK, indexers)
}

// TestProwlarrIndexer tests a single Prowlarr indexer.
func (h *ArrHandler) TestProwlarrIndexer(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid indexer ID"})
		return
	}

	client, err := h.getArrClient(r.Context(), "prowlarr")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "connection error"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "not configured"})
		return
	}

	testErr := client.TestProwlarrIndexer(r.Context(), id)

	h.Cache.Invalidate("arr:prowlarr:indexers")
	h.Cache.Invalidate("arr:overview")

	if testErr != nil {
		writeJSON(w, http.StatusOK, map[string]any{"success": false, "error": testErr.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"success": true})
}
