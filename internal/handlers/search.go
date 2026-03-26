package handlers

import (
	"net/http"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/database"
)

// SearchHandler handles search-related HTTP requests.
type SearchHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

// SearchResult represents a single search result item.
type SearchResult struct {
	Type         string `json:"type"`
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	CategoryID   int64  `json:"category_id,omitempty"`
	CategoryName string `json:"category_name,omitempty"`
}

// SearchResponse is the response envelope for search results.
type SearchResponse struct {
	Results []SearchResult `json:"results"`
}

// Search godoc
// @Summary Search widgets and categories by name
// @Tags search
// @Produce json
// @Param q query string true "Search query (minimum 2 characters)"
// @Success 200 {object} SearchResponse "Search results"
// @Failure 500 {object} SearchResponse "Internal server error (empty results)"
// @Router /search [get]
// @Security BearerAuth
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		writeJSON(w, http.StatusOK, SearchResponse{Results: []SearchResult{}})
		return
	}

	results := make([]SearchResult, 0)

	widgets, err := h.DB.SearchWidgets(query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, SearchResponse{Results: []SearchResult{}})
		return
	}
	for _, w := range widgets {
		results = append(results, SearchResult{
			Type:         "widget",
			ID:           w.ID,
			Name:         w.Title,
			CategoryID:   w.CategoryID,
			CategoryName: w.CategoryName,
		})
	}

	categories, err := h.DB.SearchCategories(query)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, SearchResponse{Results: []SearchResult{}})
		return
	}
	for _, c := range categories {
		results = append(results, SearchResult{
			Type: "category",
			ID:   c.ID,
			Name: c.Name,
		})
	}

	writeJSON(w, http.StatusOK, SearchResponse{Results: results})
}
