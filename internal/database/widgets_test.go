package database

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/tdebuilt/nidus/internal/models"
)

func setupWidgetDB(t *testing.T) (*DB, int64) {
	t.Helper()
	dir := t.TempDir()
	db, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(context.Background()); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Create a category for widget tests
	cat, err := db.CreateCategory(context.Background(), "TestCat", "folder")
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	return db, cat.ID
}

func TestCreateWidget(t *testing.T) {
	t.Parallel()
	db, catID := setupWidgetDB(t)
	ctx := context.Background()

	w, err := db.CreateWidget(ctx, catID, "docker", "My Docker", `{"env":"prod"}`, 0, 0, 6, 4)
	if err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}
	if w == nil {
		t.Fatal("expected widget, got nil")
	}
	if w.ID < 1 {
		t.Errorf("expected positive ID, got %d", w.ID)
	}
	if w.CategoryID != catID {
		t.Errorf("expected category_id %d, got %d", catID, w.CategoryID)
	}
	if w.Type != "docker" {
		t.Errorf("expected type 'docker', got %q", w.Type)
	}
	if w.Title != "My Docker" {
		t.Errorf("expected title 'My Docker', got %q", w.Title)
	}
}

func TestCreateWidget_Defaults(t *testing.T) {
	t.Parallel()
	db, catID := setupWidgetDB(t)
	ctx := context.Background()

	// Empty config defaults to "{}", width < 1 defaults to 1, height < 0 defaults to 0
	w, err := db.CreateWidget(ctx, catID, "weather", "Weather", "", 0, 0, 0, -1)
	if err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}
	if w.Config != "{}" {
		t.Errorf("expected config '{}', got %q", w.Config)
	}
	if w.Width != 1 {
		t.Errorf("expected width 1 (default), got %d", w.Width)
	}
	if w.Height != 0 {
		t.Errorf("expected height 0 (default), got %d", w.Height)
	}
}

func TestGetWidget(t *testing.T) {
	t.Parallel()
	db, catID := setupWidgetDB(t)
	ctx := context.Background()

	created, err := db.CreateWidget(ctx, catID, "proxmox", "Proxmox VMs", "{}", 2, 3, 8, 6)
	if err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}

	fetched, err := db.GetWidget(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetWidget: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected widget, got nil")
	}
	if fetched.PosX != 2 || fetched.PosY != 3 {
		t.Errorf("expected pos (2,3), got (%d,%d)", fetched.PosX, fetched.PosY)
	}
	if fetched.Width != 8 || fetched.Height != 6 {
		t.Errorf("expected size (8,6), got (%d,%d)", fetched.Width, fetched.Height)
	}
}

func TestGetWidget_NotFound(t *testing.T) {
	t.Parallel()
	db, _ := setupWidgetDB(t)
	ctx := context.Background()

	w, err := db.GetWidget(ctx, 9999)
	if err != nil {
		t.Fatalf("GetWidget: %v", err)
	}
	if w != nil {
		t.Errorf("expected nil for nonexistent widget, got %+v", w)
	}
}

func TestGetWidgetsByCategory(t *testing.T) {
	t.Parallel()
	db, catID := setupWidgetDB(t)
	ctx := context.Background()

	if _, err := db.CreateWidget(ctx, catID, "docker", "Docker 1", "{}", 0, 0, 6, 4); err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}
	if _, err := db.CreateWidget(ctx, catID, "proxmox", "Proxmox", "{}", 6, 0, 6, 4); err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}

	widgets, err := db.GetWidgetsByCategory(ctx, catID)
	if err != nil {
		t.Fatalf("GetWidgetsByCategory: %v", err)
	}
	if len(widgets) != 2 {
		t.Fatalf("expected 2 widgets, got %d", len(widgets))
	}
}

func TestGetAllWidgets(t *testing.T) {
	t.Parallel()
	db, catID := setupWidgetDB(t)
	ctx := context.Background()

	// Create a second category
	cat2, err := db.CreateCategory(ctx, "Cat2", "star")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	if _, err := db.CreateWidget(ctx, catID, "docker", "W1", "{}", 0, 0, 6, 4); err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}
	if _, err := db.CreateWidget(ctx, cat2.ID, "weather", "W2", "{}", 0, 0, 6, 4); err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}

	all, err := db.GetAllWidgets(ctx)
	if err != nil {
		t.Fatalf("GetAllWidgets: %v", err)
	}
	if len(all) != 2 {
		t.Errorf("expected 2 widgets across categories, got %d", len(all))
	}
}

func TestUpdateWidget(t *testing.T) {
	t.Parallel()
	db, catID := setupWidgetDB(t)
	ctx := context.Background()

	w, err := db.CreateWidget(ctx, catID, "docker", "Old Title", "{}", 0, 0, 6, 4)
	if err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}

	updated, err := db.UpdateWidget(ctx, w.ID, "proxmox", "New Title", `{"node":"pve"}`)
	if err != nil {
		t.Fatalf("UpdateWidget: %v", err)
	}
	if updated == nil {
		t.Fatal("expected updated widget, got nil")
	}
	if updated.Type != "proxmox" {
		t.Errorf("expected type 'proxmox', got %q", updated.Type)
	}
	if updated.Title != "New Title" {
		t.Errorf("expected title 'New Title', got %q", updated.Title)
	}
	if updated.Config != `{"node":"pve"}` {
		t.Errorf("expected config '{\"node\":\"pve\"}', got %q", updated.Config)
	}
}

func TestUpdateWidget_NotFound(t *testing.T) {
	t.Parallel()
	db, _ := setupWidgetDB(t)
	ctx := context.Background()

	updated, err := db.UpdateWidget(ctx, 9999, "docker", "Title", "{}")
	if err != nil {
		t.Fatalf("UpdateWidget: %v", err)
	}
	if updated != nil {
		t.Errorf("expected nil for nonexistent widget update, got %+v", updated)
	}
}

func TestDeleteWidget(t *testing.T) {
	t.Parallel()
	db, catID := setupWidgetDB(t)
	ctx := context.Background()

	w, err := db.CreateWidget(ctx, catID, "docker", "ToDelete", "{}", 0, 0, 6, 4)
	if err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}

	deleted, err := db.DeleteWidget(ctx, w.ID)
	if err != nil {
		t.Fatalf("DeleteWidget: %v", err)
	}
	if !deleted {
		t.Error("expected DeleteWidget to return true")
	}

	fetched, err := db.GetWidget(ctx, w.ID)
	if err != nil {
		t.Fatalf("GetWidget after delete: %v", err)
	}
	if fetched != nil {
		t.Error("expected nil after delete")
	}
}

func TestDeleteWidget_NotFound(t *testing.T) {
	t.Parallel()
	db, _ := setupWidgetDB(t)
	ctx := context.Background()

	deleted, err := db.DeleteWidget(ctx, 9999)
	if err != nil {
		t.Fatalf("DeleteWidget: %v", err)
	}
	if deleted {
		t.Error("expected DeleteWidget to return false for nonexistent widget")
	}
}

func TestSaveWidgetLayout(t *testing.T) {
	t.Parallel()
	db, catID := setupWidgetDB(t)
	ctx := context.Background()

	w1, err := db.CreateWidget(ctx, catID, "docker", "W1", "{}", 0, 0, 6, 4)
	if err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}
	w2, err := db.CreateWidget(ctx, catID, "proxmox", "W2", "{}", 6, 0, 6, 4)
	if err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}

	layouts := []models.WidgetLayout{
		{ID: w1.ID, PosX: 10, PosY: 20, Width: 12, Height: 8},
		{ID: w2.ID, PosX: 0, PosY: 0, Width: 24, Height: 16},
	}
	if err := db.SaveWidgetLayout(ctx, layouts); err != nil {
		t.Fatalf("SaveWidgetLayout: %v", err)
	}

	// Verify updated positions
	got1, err := db.GetWidget(ctx, w1.ID)
	if err != nil {
		t.Fatalf("GetWidget: %v", err)
	}
	if got1.PosX != 10 || got1.PosY != 20 || got1.Width != 12 || got1.Height != 8 {
		t.Errorf("w1 layout: expected (10,20,12,8), got (%d,%d,%d,%d)",
			got1.PosX, got1.PosY, got1.Width, got1.Height)
	}

	got2, err := db.GetWidget(ctx, w2.ID)
	if err != nil {
		t.Fatalf("GetWidget: %v", err)
	}
	if got2.PosX != 0 || got2.PosY != 0 || got2.Width != 24 || got2.Height != 16 {
		t.Errorf("w2 layout: expected (0,0,24,16), got (%d,%d,%d,%d)",
			got2.PosX, got2.PosY, got2.Width, got2.Height)
	}
}
