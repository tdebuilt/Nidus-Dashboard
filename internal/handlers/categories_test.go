package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/models"
)

func categoriesHandler(t *testing.T) *CategoriesHandler {
	t.Helper()
	db := setupTestDB(t)
	return &CategoriesHandler{DB: db}
}

func createCategoryRequest(t *testing.T, body interface{}) *http.Request {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader(b))
}



// serveWithChi creates a minimal chi router to properly resolve URL params.
func serveWithChi(method, pattern string, handler http.HandlerFunc, req *http.Request) *httptest.ResponseRecorder {
	r := chi.NewRouter()
	switch method {
	case http.MethodGet:
		r.Get(pattern, handler)
	case http.MethodPut:
		r.Put(pattern, handler)
	case http.MethodDelete:
		r.Delete(pattern, handler)
	case http.MethodPost:
		r.Post(pattern, handler)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestCategoryListEmpty(t *testing.T) {
	h := categoriesHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var categories []models.Category
	if err := json.NewDecoder(w.Body).Decode(&categories); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(categories) != 0 {
		t.Errorf("expected empty list, got %d", len(categories))
	}
}

func TestCategoryCreate(t *testing.T) {
	h := categoriesHandler(t)

	req := createCategoryRequest(t, models.CreateCategoryRequest{
		Name: "Infrastructure",
		Icon: "server",
	})
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var cat models.Category
	if err := json.NewDecoder(w.Body).Decode(&cat); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if cat.Name != "Infrastructure" {
		t.Errorf("expected name 'Infrastructure', got '%s'", cat.Name)
	}
	if cat.Icon != "server" {
		t.Errorf("expected icon 'server', got '%s'", cat.Icon)
	}
	if cat.ID == 0 {
		t.Error("expected non-zero ID")
	}
	if cat.SortOrder != 0 {
		t.Errorf("expected sort_order 0, got %d", cat.SortOrder)
	}
	if cat.Slug != "infrastructure" {
		t.Errorf("expected slug 'infrastructure', got '%s'", cat.Slug)
	}
}

func TestCategoryCreateMissingName(t *testing.T) {
	h := categoriesHandler(t)

	req := createCategoryRequest(t, models.CreateCategoryRequest{
		Name: "",
		Icon: "server",
	})
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCategoryCreateMissingIcon(t *testing.T) {
	h := categoriesHandler(t)

	req := createCategoryRequest(t, models.CreateCategoryRequest{
		Name: "Test",
		Icon: "",
	})
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCategoryCreateInvalidJSON(t *testing.T) {
	h := categoriesHandler(t)

	req := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	h.Create(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCategoryGet(t *testing.T) {
	h := categoriesHandler(t)

	// Create a category first
	cat, err := h.DB.CreateCategory("Media", "film")
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/categories/%d", cat.ID), nil)
	w := serveWithChi(http.MethodGet, "/api/categories/{id}", h.Get, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var got models.Category
	json.NewDecoder(w.Body).Decode(&got)
	if got.Name != "Media" {
		t.Errorf("expected 'Media', got '%s'", got.Name)
	}
	if got.Slug != "media" {
		t.Errorf("expected slug 'media', got '%s'", got.Slug)
	}
}

func TestCategoryGetNotFound(t *testing.T) {
	h := categoriesHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/categories/999", nil)
	w := serveWithChi(http.MethodGet, "/api/categories/{id}", h.Get, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCategoryGetInvalidID(t *testing.T) {
	h := categoriesHandler(t)

	req := httptest.NewRequest(http.MethodGet, "/api/categories/abc", nil)
	w := serveWithChi(http.MethodGet, "/api/categories/{id}", h.Get, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCategoryUpdate(t *testing.T) {
	h := categoriesHandler(t)

	cat, err := h.DB.CreateCategory("Old Name", "folder")
	if err != nil {
		t.Fatalf("failed to create: %v", err)
	}

	body, _ := json.Marshal(models.UpdateCategoryRequest{
		Name: "New Name",
		Icon: "star",
	})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/categories/%d", cat.ID), bytes.NewReader(body))
	w := serveWithChi(http.MethodPut, "/api/categories/{id}", h.Update, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.Category
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Name != "New Name" {
		t.Errorf("expected 'New Name', got '%s'", updated.Name)
	}
	if updated.Icon != "star" {
		t.Errorf("expected 'star', got '%s'", updated.Icon)
	}
}

func TestCategoryUpdateNotFound(t *testing.T) {
	h := categoriesHandler(t)

	body, _ := json.Marshal(models.UpdateCategoryRequest{
		Name: "Test",
		Icon: "test",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/categories/999", bytes.NewReader(body))
	w := serveWithChi(http.MethodPut, "/api/categories/{id}", h.Update, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCategoryUpdateMissingName(t *testing.T) {
	h := categoriesHandler(t)

	cat, _ := h.DB.CreateCategory("Test", "folder")
	body, _ := json.Marshal(models.UpdateCategoryRequest{Name: "", Icon: "star"})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/categories/%d", cat.ID), bytes.NewReader(body))
	w := serveWithChi(http.MethodPut, "/api/categories/{id}", h.Update, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCategoryUpdateMissingIcon(t *testing.T) {
	h := categoriesHandler(t)

	cat, _ := h.DB.CreateCategory("Test", "folder")
	body, _ := json.Marshal(models.UpdateCategoryRequest{Name: "Test", Icon: ""})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/categories/%d", cat.ID), bytes.NewReader(body))
	w := serveWithChi(http.MethodPut, "/api/categories/{id}", h.Update, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCategoryDelete(t *testing.T) {
	h := categoriesHandler(t)

	cat, _ := h.DB.CreateCategory("ToDelete", "trash")

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/categories/%d", cat.ID), nil)
	w := serveWithChi(http.MethodDelete, "/api/categories/{id}", h.Delete, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify it's gone
	got, err := h.DB.GetCategory(cat.ID)
	if err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if got != nil {
		t.Error("category should be deleted")
	}
}

func TestCategoryDeleteNotFound(t *testing.T) {
	h := categoriesHandler(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/categories/999", nil)
	w := serveWithChi(http.MethodDelete, "/api/categories/{id}", h.Delete, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCategoryReorder(t *testing.T) {
	h := categoriesHandler(t)

	cat1, _ := h.DB.CreateCategory("First", "one")
	cat2, _ := h.DB.CreateCategory("Second", "two")
	cat3, _ := h.DB.CreateCategory("Third", "three")

	// Reverse order
	body, _ := json.Marshal(models.ReorderRequest{
		IDs: []int64{cat3.ID, cat1.ID, cat2.ID},
	})
	req := httptest.NewRequest(http.MethodPut, "/api/categories/reorder", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Reorder(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var categories []models.Category
	json.NewDecoder(w.Body).Decode(&categories)

	if len(categories) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(categories))
	}
	// After reorder, first should be cat3 (sort_order=0)
	if categories[0].ID != cat3.ID {
		t.Errorf("expected first category ID %d, got %d", cat3.ID, categories[0].ID)
	}
	if categories[1].ID != cat1.ID {
		t.Errorf("expected second category ID %d, got %d", cat1.ID, categories[1].ID)
	}
	if categories[2].ID != cat2.ID {
		t.Errorf("expected third category ID %d, got %d", cat2.ID, categories[2].ID)
	}
}

func TestCategoryReorderEmptyIDs(t *testing.T) {
	h := categoriesHandler(t)

	body, _ := json.Marshal(models.ReorderRequest{IDs: []int64{}})
	req := httptest.NewRequest(http.MethodPut, "/api/categories/reorder", bytes.NewReader(body))
	w := httptest.NewRecorder()
	h.Reorder(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCategoryListAfterCreate(t *testing.T) {
	h := categoriesHandler(t)

	h.DB.CreateCategory("Alpha", "a")
	h.DB.CreateCategory("Beta", "b")

	req := httptest.NewRequest(http.MethodGet, "/api/categories", nil)
	w := httptest.NewRecorder()
	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var categories []models.Category
	json.NewDecoder(w.Body).Decode(&categories)

	if len(categories) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(categories))
	}
	if categories[0].Name != "Alpha" {
		t.Errorf("expected first 'Alpha', got '%s'", categories[0].Name)
	}
	if categories[1].Name != "Beta" {
		t.Errorf("expected second 'Beta', got '%s'", categories[1].Name)
	}
}

func TestCategorySortOrderAutoIncrement(t *testing.T) {
	h := categoriesHandler(t)

	cat1, _ := h.DB.CreateCategory("First", "one")
	cat2, _ := h.DB.CreateCategory("Second", "two")
	cat3, _ := h.DB.CreateCategory("Third", "three")

	if cat1.SortOrder != 0 {
		t.Errorf("expected sort_order 0, got %d", cat1.SortOrder)
	}
	if cat2.SortOrder != 1 {
		t.Errorf("expected sort_order 1, got %d", cat2.SortOrder)
	}
	if cat3.SortOrder != 2 {
		t.Errorf("expected sort_order 2, got %d", cat3.SortOrder)
	}
}

func TestCategorySlugStableOnRename(t *testing.T) {
	h := categoriesHandler(t)

	cat, _ := h.DB.CreateCategory("Old Name", "folder")
	originalSlug := cat.Slug

	body, _ := json.Marshal(models.UpdateCategoryRequest{
		Name: "New Name",
		Icon: "star",
	})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/categories/%d", cat.ID), bytes.NewReader(body))
	w := serveWithChi(http.MethodPut, "/api/categories/{id}", h.Update, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.Category
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Slug != originalSlug {
		t.Errorf("slug should not change on rename: expected '%s', got '%s'", originalSlug, updated.Slug)
	}
}

func TestCategorySlugCollision(t *testing.T) {
	h := categoriesHandler(t)

	cat1, _ := h.DB.CreateCategory("Test", "folder")
	cat2, _ := h.DB.CreateCategory("Test", "folder")

	if cat1.Slug != "test" {
		t.Errorf("expected first slug 'test', got '%s'", cat1.Slug)
	}
	if cat2.Slug != "test-2" {
		t.Errorf("expected second slug 'test-2', got '%s'", cat2.Slug)
	}
}

func TestCategorySlugFrenchAccents(t *testing.T) {
	h := categoriesHandler(t)

	cat, _ := h.DB.CreateCategory("Réseau Sécurité", "shield")
	if cat.Slug != "reseau-securite" {
		t.Errorf("expected slug 'reseau-securite', got '%s'", cat.Slug)
	}
}
