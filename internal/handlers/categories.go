package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
)

// CategoriesHandler handles category-related HTTP requests.
type CategoriesHandler struct {
	DB *database.DB
}

// List godoc
// @Summary List all categories
// @Description Returns all dashboard categories ordered by position.
// @Tags categories
// @Produce json
// @Success 200 {array} models.Category "List of categories"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /categories [get]
// @Security BearerAuth
func (h *CategoriesHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.DB.GetCategories()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list categories"})
		return
	}
	if categories == nil {
		categories = []models.Category{}
	}
	writeJSON(w, http.StatusOK, categories)
}

// Create godoc
// @Summary Create a category
// @Description Creates a new dashboard category.
// @Tags categories
// @Accept json
// @Produce json
// @Param request body models.CreateCategoryRequest true "Category details"
// @Success 201 {object} models.Category "Created category"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /categories [post]
// @Security BearerAuth
func (h *CategoriesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "name is required"})
		return
	}
	if req.Icon == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "icon is required"})
		return
	}

	cat, err := h.DB.CreateCategory(req.Name, req.Icon)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create category"})
		return
	}

	writeJSON(w, http.StatusCreated, cat)
}

// Get godoc
// @Summary Get a category
// @Description Returns a single category by ID.
// @Tags categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} models.Category "Category details"
// @Failure 400 {object} models.ErrorResponse "Invalid category ID"
// @Failure 404 {object} models.ErrorResponse "Category not found"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /categories/{id} [get]
// @Security BearerAuth
func (h *CategoriesHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid category ID"})
		return
	}

	cat, err := h.DB.GetCategory(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to get category"})
		return
	}
	if cat == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "category not found"})
		return
	}

	writeJSON(w, http.StatusOK, cat)
}

// Update godoc
// @Summary Update a category
// @Description Updates an existing category's name and icon.
// @Tags categories
// @Accept json
// @Produce json
// @Param id path int true "Category ID"
// @Param request body models.UpdateCategoryRequest true "Updated category details"
// @Success 200 {object} models.Category "Updated category"
// @Failure 400 {object} models.ErrorResponse "Invalid request body or category ID"
// @Failure 404 {object} models.ErrorResponse "Category not found"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /categories/{id} [put]
// @Security BearerAuth
func (h *CategoriesHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid category ID"})
		return
	}

	var req models.UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "name is required"})
		return
	}
	if req.Icon == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "icon is required"})
		return
	}

	cat, err := h.DB.UpdateCategory(id, req.Name, req.Icon)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update category"})
		return
	}
	if cat == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "category not found"})
		return
	}

	writeJSON(w, http.StatusOK, cat)
}

// Delete godoc
// @Summary Delete a category
// @Description Deletes a category and its associated widgets.
// @Tags categories
// @Produce json
// @Param id path int true "Category ID"
// @Success 200 {object} map[string]string "Category deleted"
// @Failure 400 {object} models.ErrorResponse "Invalid category ID"
// @Failure 404 {object} models.ErrorResponse "Category not found"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /categories/{id} [delete]
// @Security BearerAuth
func (h *CategoriesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid category ID"})
		return
	}

	deleted, err := h.DB.DeleteCategory(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to delete category"})
		return
	}
	if !deleted {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "category not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "category deleted"})
}

// Reorder godoc
// @Summary Reorder categories
// @Description Updates the display order of categories.
// @Tags categories
// @Accept json
// @Produce json
// @Param request body models.ReorderRequest true "Ordered list of category IDs"
// @Success 200 {array} models.Category "Reordered categories"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /categories/reorder [put]
// @Security BearerAuth
func (h *CategoriesHandler) Reorder(w http.ResponseWriter, r *http.Request) {
	var req models.ReorderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if len(req.IDs) == 0 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "ids is required"})
		return
	}

	if err := h.DB.ReorderCategories(req.IDs); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to reorder categories"})
		return
	}

	categories, err := h.DB.GetCategories()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list categories"})
		return
	}

	writeJSON(w, http.StatusOK, categories)
}
