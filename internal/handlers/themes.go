package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
)

// ThemesHandler handles custom theme HTTP requests.
type ThemesHandler struct {
	DB *database.DB
}

var themeColorKeys = []string{
	"color-bg", "color-bg-primary", "color-bg-secondary", "color-bg-tertiary",
	"color-border",
	"color-text", "color-text-primary", "color-text-secondary", "color-text-muted",
	"color-primary", "color-primary-hover",
	"color-accent", "color-accent-hover",
	"color-danger", "color-danger-hover",
	"color-success", "color-warning",
	"color-sidebar-bg", "color-sidebar-hover",
	"color-error-text", "color-error-border", "color-error-bg",
	"color-success-text", "color-success-border", "color-success-bg",
	"color-info-text", "color-info-border", "color-info-bg",
}

func validateThemeJSON(raw string) error {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return err
	}

	required := []string{"id", "name", "author", "mode", "colors"}
	for _, key := range required {
		if _, ok := obj[key]; !ok {
			return &validationError{field: key, message: "missing required field"}
		}
	}

	mode, ok := obj["mode"].(string)
	if !ok || (mode != "dark" && mode != "light") {
		return &validationError{field: "mode", message: "must be \"dark\" or \"light\""}
	}

	colors, ok := obj["colors"].(map[string]interface{})
	if !ok {
		return &validationError{field: "colors", message: "must be an object"}
	}

	for _, key := range themeColorKeys {
		if _, ok := colors[key]; !ok {
			return &validationError{field: "colors." + key, message: "missing color key"}
		}
	}

	return nil
}

type validationError struct {
	field   string
	message string
}

func (e *validationError) Error() string {
	return e.field + ": " + e.message
}

// List returns all custom themes.
func (h *ThemesHandler) List(w http.ResponseWriter, r *http.Request) {
	themes, err := h.DB.ListCustomThemes(r.Context())
	if err != nil {
		slog.Error("themes: failed to list themes", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list themes"})
		return
	}
	writeJSON(w, http.StatusOK, themes)
}

// Create creates a new custom theme (admin only, max 5).
func (h *ThemesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCustomThemeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" || len(req.Name) > 50 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "name is required (max 50 chars)"})
		return
	}

	if err := validateThemeJSON(req.ThemeJSON); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid theme JSON: " + err.Error()})
		return
	}

	count, err := h.DB.CountCustomThemes(r.Context())
	if err != nil {
		slog.Error("themes: failed to count themes", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to count themes"})
		return
	}
	if count >= 5 {
		writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "maximum of 5 custom themes reached"})
		return
	}

	id, err := h.DB.CreateCustomTheme(r.Context(), req.Name, req.ThemeJSON)
	if err != nil {
		slog.Error("themes: failed to create theme", "name", req.Name, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create theme"})
		return
	}

	theme, err := h.DB.GetCustomTheme(r.Context(), id)
	if err != nil {
		slog.Error("themes: failed to read created theme", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to read created theme"})
		return
	}
	slog.Info("themes: theme created", "id", id, "name", req.Name)
	writeJSON(w, http.StatusCreated, theme)
}

// Update updates a custom theme (admin only).
func (h *ThemesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid theme ID"})
		return
	}

	var req models.UpdateCustomThemeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	existing, err := h.DB.GetCustomTheme(r.Context(), id)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		slog.Error("themes: failed to get theme for update", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to get theme"})
		return
	}
	if existing == nil {
		slog.Warn("themes: theme not found for update", "id", id)
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "theme not found"})
		return
	}

	name := existing.Name
	themeJSON := existing.ThemeJSON

	if req.Name != nil {
		if *req.Name == "" || len(*req.Name) > 50 {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "name is required (max 50 chars)"})
			return
		}
		name = *req.Name
	}

	if req.ThemeJSON != nil {
		if err := validateThemeJSON(*req.ThemeJSON); err != nil {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid theme JSON: " + err.Error()})
			return
		}
		themeJSON = *req.ThemeJSON
	}

	if err := h.DB.UpdateCustomTheme(r.Context(), id, name, themeJSON); err != nil {
		slog.Error("themes: failed to update theme", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update theme"})
		return
	}

	theme, err := h.DB.GetCustomTheme(r.Context(), id)
	if err != nil {
		slog.Error("themes: failed to read updated theme", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to read updated theme"})
		return
	}
	slog.Info("themes: theme updated", "id", id)
	writeJSON(w, http.StatusOK, theme)
}

// Delete deletes a custom theme (admin only).
func (h *ThemesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid theme ID"})
		return
	}

	if err := h.DB.DeleteCustomTheme(r.Context(), id); err != nil {
		slog.Error("themes: failed to delete theme", "id", id, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to delete theme"})
		return
	}
	slog.Info("themes: theme deleted", "id", id)
	w.WriteHeader(http.StatusNoContent)
}
