package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/models"
)

func widgetsHandler(t *testing.T) (*WidgetsHandler, *models.Category) {
	t.Helper()
	ctx := context.Background()
	db := setupTestDB(t)
	cat, err := db.CreateCategory(ctx, "Test", "folder")
	if err != nil {
		t.Fatalf("failed to create category: %v", err)
	}
	return &WidgetsHandler{DB: db}, cat
}

func widgetRouter(h *WidgetsHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/api/categories/{id}/widgets", h.ListByCategory)
	r.Post("/api/categories/{id}/widgets", h.Create)
	r.Put("/api/widgets/layout", h.SaveLayout)
	r.Put("/api/widgets/{id}", h.Update)
	r.Delete("/api/widgets/{id}", h.Delete)
	return r
}

func TestWidgetListEmpty(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/categories/%d/widgets", cat.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var widgets []models.Widget
	json.NewDecoder(w.Body).Decode(&widgets)
	if len(widgets) != 0 {
		t.Errorf("expected empty list, got %d", len(widgets))
	}
}

func TestWidgetListCategoryNotFound(t *testing.T) {
	t.Parallel()
	h, _ := widgetsHandler(t)
	r := widgetRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/categories/999/widgets", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestWidgetCreate(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)

	body, _ := json.Marshal(models.CreateWidgetRequest{
		Type:   "docker_stack",
		Title:  "My Stack",
		Config: `{"stack_id": 1}`,
		PosX:   0,
		PosY:   0,
		Width:  2,
		Height: 1,
	})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/categories/%d/widgets", cat.ID), bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var widget models.Widget
	json.NewDecoder(w.Body).Decode(&widget)
	if widget.Type != "docker_stack" {
		t.Errorf("expected type 'docker_stack', got '%s'", widget.Type)
	}
	if widget.Title != "My Stack" {
		t.Errorf("expected title 'My Stack', got '%s'", widget.Title)
	}
	if widget.CategoryID != cat.ID {
		t.Errorf("expected category_id %d, got %d", cat.ID, widget.CategoryID)
	}
	if widget.Width != 2 {
		t.Errorf("expected width 2, got %d", widget.Width)
	}
	if widget.ID == 0 {
		t.Error("expected non-zero ID")
	}
}

func TestWidgetCreateMissingType(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)

	body, _ := json.Marshal(models.CreateWidgetRequest{
		Type:  "",
		Title: "Test",
	})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/categories/%d/widgets", cat.ID), bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestWidgetCreateMissingTitle(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)

	body, _ := json.Marshal(models.CreateWidgetRequest{
		Type:  "docker_stack",
		Title: "",
	})
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/categories/%d/widgets", cat.ID), bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestWidgetCreateCategoryNotFound(t *testing.T) {
	t.Parallel()
	h, _ := widgetsHandler(t)
	r := widgetRouter(h)

	body, _ := json.Marshal(models.CreateWidgetRequest{
		Type:  "docker_stack",
		Title: "Test",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/categories/999/widgets", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestWidgetCreateInvalidJSON(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/categories/%d/widgets", cat.ID), bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestWidgetCreateDefaultConfig(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	ctx := context.Background()

	widget, err := h.DB.CreateWidget(ctx, cat.ID, "test", "Test", "", 0, 0, 0, 0)
	if err != nil {
		t.Fatalf("failed to create widget: %v", err)
	}
	if widget.Config != "{}" {
		t.Errorf("expected default config '{}', got '%s'", widget.Config)
	}
	if widget.Width != 1 {
		t.Errorf("expected default width 1, got %d", widget.Width)
	}
	if widget.Height != 0 {
		t.Errorf("expected default height 0 (auto), got %d", widget.Height)
	}
}

func TestWidgetUpdate(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)
	ctx := context.Background()

	widget, _ := h.DB.CreateWidget(ctx, cat.ID, "old_type", "Old Title", `{"old": true}`, 0, 0, 1, 1)

	body, _ := json.Marshal(models.UpdateWidgetRequest{
		Type:   "new_type",
		Title:  "New Title",
		Config: `{"new": true}`,
	})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/widgets/%d", widget.ID), bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var updated models.Widget
	json.NewDecoder(w.Body).Decode(&updated)
	if updated.Type != "new_type" {
		t.Errorf("expected type 'new_type', got '%s'", updated.Type)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected title 'New Title', got '%s'", updated.Title)
	}
	if updated.Config != `{"new": true}` {
		t.Errorf("expected config '{\"new\": true}', got '%s'", updated.Config)
	}
}

func TestWidgetUpdateNotFound(t *testing.T) {
	t.Parallel()
	h, _ := widgetsHandler(t)
	r := widgetRouter(h)

	body, _ := json.Marshal(models.UpdateWidgetRequest{
		Type:  "test",
		Title: "Test",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/widgets/999", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestWidgetUpdateMissingType(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)
	ctx := context.Background()

	widget, _ := h.DB.CreateWidget(ctx, cat.ID, "test", "Test", "{}", 0, 0, 1, 1)

	body, _ := json.Marshal(models.UpdateWidgetRequest{Type: "", Title: "Test"})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/widgets/%d", widget.ID), bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestWidgetUpdateMissingTitle(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)
	ctx := context.Background()

	widget, _ := h.DB.CreateWidget(ctx, cat.ID, "test", "Test", "{}", 0, 0, 1, 1)

	body, _ := json.Marshal(models.UpdateWidgetRequest{Type: "test", Title: ""})
	req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/widgets/%d", widget.ID), bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestWidgetDelete(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)
	ctx := context.Background()

	widget, _ := h.DB.CreateWidget(ctx, cat.ID, "test", "Test", "{}", 0, 0, 1, 1)

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/widgets/%d", widget.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify deleted
	got, _ := h.DB.GetWidget(ctx, widget.ID)
	if got != nil {
		t.Error("widget should be deleted")
	}
}

func TestWidgetDeleteNotFound(t *testing.T) {
	t.Parallel()
	h, _ := widgetsHandler(t)
	r := widgetRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/widgets/999", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestWidgetSaveLayout(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)
	ctx := context.Background()

	w1, _ := h.DB.CreateWidget(ctx, cat.ID, "test", "Widget 1", "{}", 0, 0, 1, 1)
	w2, _ := h.DB.CreateWidget(ctx, cat.ID, "test", "Widget 2", "{}", 1, 0, 1, 1)

	body, _ := json.Marshal(models.SaveLayoutRequest{
		Widgets: []models.WidgetLayout{
			{ID: w1.ID, PosX: 2, PosY: 3, Width: 4, Height: 2},
			{ID: w2.ID, PosX: 0, PosY: 0, Width: 2, Height: 1},
		},
	})
	req := httptest.NewRequest(http.MethodPut, "/api/widgets/layout", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify layout was saved
	got1, _ := h.DB.GetWidget(ctx, w1.ID)
	if got1.PosX != 2 || got1.PosY != 3 || got1.Width != 4 || got1.Height != 2 {
		t.Errorf("widget 1 layout not saved: pos(%d,%d) size(%d,%d)", got1.PosX, got1.PosY, got1.Width, got1.Height)
	}
	got2, _ := h.DB.GetWidget(ctx, w2.ID)
	if got2.PosX != 0 || got2.PosY != 0 || got2.Width != 2 || got2.Height != 1 {
		t.Errorf("widget 2 layout not saved: pos(%d,%d) size(%d,%d)", got2.PosX, got2.PosY, got2.Width, got2.Height)
	}
}

func TestWidgetSaveLayoutReload(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	ctx := context.Background()

	w1, _ := h.DB.CreateWidget(ctx, cat.ID, "test", "Widget 1", "{}", 0, 0, 1, 1)

	// Save layout
	err := h.DB.SaveWidgetLayout(ctx, []models.WidgetLayout{
		{ID: w1.ID, PosX: 5, PosY: 10, Width: 3, Height: 2},
	})
	if err != nil {
		t.Fatalf("failed to save layout: %v", err)
	}

	// Reload from DB
	got, _ := h.DB.GetWidget(ctx, w1.ID)
	if got.PosX != 5 || got.PosY != 10 || got.Width != 3 || got.Height != 2 {
		t.Errorf("layout not persisted: pos(%d,%d) size(%d,%d)", got.PosX, got.PosY, got.Width, got.Height)
	}
}

func TestWidgetSaveLayoutEmpty(t *testing.T) {
	t.Parallel()
	h, _ := widgetsHandler(t)
	r := widgetRouter(h)

	body, _ := json.Marshal(models.SaveLayoutRequest{Widgets: []models.WidgetLayout{}})
	req := httptest.NewRequest(http.MethodPut, "/api/widgets/layout", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestWidgetOrphanImpossible(t *testing.T) {
	t.Parallel()
	h, _ := widgetsHandler(t)
	ctx := context.Background()

	// Try to create widget with non-existent category
	_, err := h.DB.CreateWidget(ctx, 999, "test", "Orphan", "{}", 0, 0, 1, 1)
	if err == nil {
		t.Error("expected FK violation error, got nil")
	}
}

func TestWidgetCascadeDeleteWithCategory(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	ctx := context.Background()

	widget, _ := h.DB.CreateWidget(ctx, cat.ID, "test", "Widget", "{}", 0, 0, 1, 1)

	// Delete the category
	h.DB.DeleteCategory(ctx, cat.ID)

	// Widget should be gone (CASCADE)
	got, err := h.DB.GetWidget(ctx, widget.ID)
	if err != nil {
		t.Fatalf("error querying widget: %v", err)
	}
	if got != nil {
		t.Error("widget should be cascade-deleted with category")
	}
}

func TestWidgetListByCategory(t *testing.T) {
	t.Parallel()
	h, cat := widgetsHandler(t)
	r := widgetRouter(h)
	ctx := context.Background()

	// Create a second category
	cat2, _ := h.DB.CreateCategory(ctx, "Other", "box")

	h.DB.CreateWidget(ctx, cat.ID, "test", "Widget A", "{}", 0, 0, 1, 1)
	h.DB.CreateWidget(ctx, cat.ID, "test", "Widget B", "{}", 1, 0, 1, 1)
	h.DB.CreateWidget(ctx, cat2.ID, "test", "Widget C", "{}", 0, 0, 1, 1)

	// List widgets for first category
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/categories/%d/widgets", cat.ID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var widgets []models.Widget
	json.NewDecoder(w.Body).Decode(&widgets)
	if len(widgets) != 2 {
		t.Fatalf("expected 2 widgets for cat, got %d", len(widgets))
	}

	// List widgets for second category
	req2 := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/categories/%d/widgets", cat2.ID), nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	var widgets2 []models.Widget
	json.NewDecoder(w2.Body).Decode(&widgets2)
	if len(widgets2) != 1 {
		t.Fatalf("expected 1 widget for cat2, got %d", len(widgets2))
	}
}
