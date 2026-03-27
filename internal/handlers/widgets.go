package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
)

// WidgetsHandler handles widget-related HTTP requests.
type WidgetsHandler struct {
	DB              *database.DB
	OnReolinkChange func()
}

// ListByCategory godoc
// @Summary List widgets by category
// @Description Returns all widgets belonging to a specific category.
// @Tags widgets
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {array} models.Widget "List of widgets"
// @Failure 400 {object} models.ErrorResponse "Invalid category ID"
// @Failure 404 {object} models.ErrorResponse "Category not found"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /categories/{id}/widgets [get]
// @Security BearerAuth
func (h *WidgetsHandler) ListByCategory(w http.ResponseWriter, r *http.Request) {
	categoryID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid category ID"})
		return
	}

	// Check category exists
	cat, err := h.DB.GetCategory(r.Context(), categoryID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if cat == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "category not found"})
		return
	}

	widgets, err := h.DB.GetWidgetsByCategory(r.Context(), categoryID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list widgets"})
		return
	}
	if widgets == nil {
		widgets = []models.Widget{}
	}
	writeJSON(w, http.StatusOK, widgets)
}

// Create godoc
// @Summary Create a widget
// @Description Creates a new widget in the specified category.
// @Tags widgets
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param request body models.CreateWidgetRequest true "Widget details"
// @Success 201 {object} models.Widget "Created widget"
// @Failure 400 {object} models.ErrorResponse "Invalid request body or category ID"
// @Failure 404 {object} models.ErrorResponse "Category not found"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /categories/{id}/widgets [post]
// @Security BearerAuth
func (h *WidgetsHandler) Create(w http.ResponseWriter, r *http.Request) {
	categoryID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid category ID"})
		return
	}

	// Check category exists
	cat, err := h.DB.GetCategory(r.Context(), categoryID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if cat == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "category not found"})
		return
	}

	var req models.CreateWidgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Type == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "type is required"})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "title is required"})
		return
	}

	widget, err := h.DB.CreateWidget(r.Context(), categoryID, req.Type, req.Title, req.Config, req.PosX, req.PosY, req.Width, req.Height)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create widget"})
		return
	}

	if req.Type == "reolink" && h.OnReolinkChange != nil {
		go h.OnReolinkChange()
	}

	writeJSON(w, http.StatusCreated, widget)
}

// Update godoc
// @Summary Update a widget
// @Description Updates an existing widget's type, title, and config.
// @Tags widgets
// @Accept json
// @Produce json
// @Param id path int true "Widget ID"
// @Param request body models.UpdateWidgetRequest true "Updated widget details"
// @Success 200 {object} models.Widget "Updated widget"
// @Failure 400 {object} models.ErrorResponse "Invalid request body or widget ID"
// @Failure 404 {object} models.ErrorResponse "Widget not found"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /widgets/{id} [put]
// @Security BearerAuth
func (h *WidgetsHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid widget ID"})
		return
	}

	var req models.UpdateWidgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Type == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "type is required"})
		return
	}
	if req.Title == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "title is required"})
		return
	}
	if req.Config == "" {
		req.Config = "{}"
	}

	widget, err := h.DB.UpdateWidget(r.Context(), id, req.Type, req.Title, req.Config)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update widget"})
		return
	}
	if widget == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "widget not found"})
		return
	}

	if req.Type == "reolink" && h.OnReolinkChange != nil {
		go h.OnReolinkChange()
	}

	writeJSON(w, http.StatusOK, widget)
}

// Delete godoc
// @Summary Delete a widget
// @Description Deletes a widget by ID.
// @Tags widgets
// @Produce json
// @Param id path int true "Widget ID"
// @Success 200 {object} map[string]string "Widget deleted"
// @Failure 400 {object} models.ErrorResponse "Invalid widget ID"
// @Failure 404 {object} models.ErrorResponse "Widget not found"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /widgets/{id} [delete]
// @Security BearerAuth
func (h *WidgetsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid widget ID"})
		return
	}

	deleted, err := h.DB.DeleteWidget(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to delete widget"})
		return
	}
	if !deleted {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "widget not found"})
		return
	}

	if h.OnReolinkChange != nil {
		go h.OnReolinkChange()
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "widget deleted"})
}

// ToggleCollapse godoc
// @Summary Toggle widget collapse state
// @Description Toggles whether a widget is collapsed or expanded.
// @Tags widgets
// @Accept json
// @Produce json
// @Param id path int true "Widget ID"
// @Param request body models.ToggleCollapseRequest true "Collapse state"
// @Success 200 {object} models.Widget "Updated widget"
// @Failure 400 {object} models.ErrorResponse "Invalid request body or widget ID"
// @Failure 404 {object} models.ErrorResponse "Widget not found"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /widgets/{id}/toggle-collapse [patch]
// @Security BearerAuth
func (h *WidgetsHandler) ToggleCollapse(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid widget ID"})
		return
	}

	var req models.ToggleCollapseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	widget, err := h.DB.SetWidgetCollapsed(r.Context(), id, req.Collapsed)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update widget"})
		return
	}
	if widget == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "widget not found"})
		return
	}

	writeJSON(w, http.StatusOK, widget)
}

// SaveLayout godoc
// @Summary Save widget layout
// @Description Saves the position and size of all widgets in bulk.
// @Tags widgets
// @Accept json
// @Produce json
// @Param request body models.SaveLayoutRequest true "Widget layout positions"
// @Success 200 {object} map[string]string "Layout saved"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /widgets/layout [put]
// @Security BearerAuth
func (h *WidgetsHandler) SaveLayout(w http.ResponseWriter, r *http.Request) {
	var req models.SaveLayoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if len(req.Widgets) == 0 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "widgets is required"})
		return
	}

	if err := h.DB.SaveWidgetLayout(r.Context(), req.Widgets); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save layout"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "layout saved"})
}
