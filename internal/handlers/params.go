package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/models"
)

// parseIDParam extracts a URL parameter as int64 and validates it is >= 1.
// Returns the parsed value and true on success, or writes a 400 response and returns false.
func parseIDParam(w http.ResponseWriter, r *http.Request, param string) (int64, bool) {
	raw := chi.URLParam(r, param)
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id < 1 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid " + param})
		return 0, false
	}
	return id, true
}

// parseIntIDParam extracts a URL parameter as int and validates it is >= 1.
// Returns the parsed value and true on success, or writes a 400 response and returns false.
func parseIntIDParam(w http.ResponseWriter, r *http.Request, param string) (int, bool) {
	raw := chi.URLParam(r, param)
	id, err := strconv.Atoi(raw)
	if err != nil || id < 1 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid " + param})
		return 0, false
	}
	return id, true
}
