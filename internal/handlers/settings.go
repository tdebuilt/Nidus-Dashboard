package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	nidusmw "github.com/tdebuilt/nidus/internal/middleware"

	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
)

// SettingsHandler handles settings HTTP requests.
type SettingsHandler struct {
	DB *database.DB
}

// UserPreferencesHandler handles per-user preference HTTP requests.
type UserPreferencesHandler struct {
	DB *database.DB
}

// Get godoc
// @Summary Get application settings
// @Description Returns the current application settings (theme, language, refresh interval, etc.).
// @Tags settings
// @Produce json
// @Success 200 {object} models.Settings "Application settings"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /settings [get]
// @Security BearerAuth
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetSettings()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to read settings"})
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

// Update godoc
// @Summary Update application settings
// @Description Updates application settings (admin only). Supports partial updates.
// @Tags settings
// @Accept json
// @Produce json
// @Param request body models.UpdateSettingsRequest true "Settings to update"
// @Success 200 {object} models.Settings "Updated settings"
// @Failure 400 {object} models.ErrorResponse "Invalid settings values"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /settings [put]
// @Security BearerAuth
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	var req models.UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	// Validate refresh_interval (5-300 seconds)
	if req.RefreshInterval != nil {
		if *req.RefreshInterval < 5 || *req.RefreshInterval > 300 {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "refresh_interval must be between 5 and 300 seconds"})
			return
		}
	}

	if err := h.DB.SaveSettings(req); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save settings"})
		return
	}

	// Return updated settings
	settings, err := h.DB.GetSettings()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to read settings"})
		return
	}
	writeJSON(w, http.StatusOK, settings)
}

// GetPreferences godoc
// @Summary Get user preferences
// @Description Returns the current user's preferences, with global defaults as fallback.
// @Tags settings
// @Produce json
// @Success 200 {object} models.UserPreferences "User preferences"
// @Failure 401 {object} models.ErrorResponse "Authentication required"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /preferences [get]
// @Security BearerAuth
func (h *UserPreferencesHandler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := nidusmw.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "authentication required"})
		return
	}

	prefs, err := h.DB.GetUserPreferences(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to read preferences"})
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}

// UpdatePreferences godoc
// @Summary Update user preferences
// @Description Updates the current user's preferences. Supports partial updates.
// @Tags settings
// @Accept json
// @Produce json
// @Param request body models.UpdateUserPreferencesRequest true "Preferences to update"
// @Success 200 {object} models.UserPreferences "Updated preferences"
// @Failure 400 {object} models.ErrorResponse "Invalid values"
// @Failure 401 {object} models.ErrorResponse "Authentication required"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /preferences [put]
// @Security BearerAuth
func (h *UserPreferencesHandler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	userID, ok := nidusmw.GetUserID(r.Context())
	if !ok {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "authentication required"})
		return
	}

	var req models.UpdateUserPreferencesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.RefreshInterval != nil {
		if *req.RefreshInterval < 5 || *req.RefreshInterval > 300 {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "refresh_interval must be between 5 and 300 seconds"})
			return
		}
	}

	if req.Theme != nil {
		if err := h.DB.SaveUserPreference(userID, "setting_theme", *req.Theme); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save preference"})
			return
		}
	}
	if req.Language != nil {
		if err := h.DB.SaveUserPreference(userID, "setting_language", *req.Language); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save preference"})
			return
		}
	}
	if req.RefreshInterval != nil {
		if err := h.DB.SaveUserPreference(userID, "setting_refresh_interval", strconv.Itoa(*req.RefreshInterval)); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save preference"})
			return
		}
	}
	if req.AccentColor != nil {
		if err := h.DB.SaveUserPreference(userID, "setting_accent_color", *req.AccentColor); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save preference"})
			return
		}
	}
	if req.EnableKeyboardShortcuts != nil {
		val := "false"
		if *req.EnableKeyboardShortcuts {
			val = "true"
		}
		if err := h.DB.SaveUserPreference(userID, "setting_keyboard_shortcuts", val); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save preference"})
			return
		}
	}

	prefs, err := h.DB.GetUserPreferences(userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to read preferences"})
		return
	}
	writeJSON(w, http.StatusOK, prefs)
}
