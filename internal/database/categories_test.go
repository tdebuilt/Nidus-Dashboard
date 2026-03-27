package database

import (
	"context"
	"path/filepath"
	"testing"
)

func setupCategoryDB(t *testing.T) *DB {
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
	return db
}

func TestCreateCategory(t *testing.T) {
	t.Parallel()
	db := setupCategoryDB(t)
	ctx := context.Background()

	cat, err := db.CreateCategory(ctx, "Infrastructure", "server")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	if cat == nil {
		t.Fatal("expected category, got nil")
	}
	if cat.ID < 1 {
		t.Errorf("expected positive ID, got %d", cat.ID)
	}
	if cat.Name != "Infrastructure" {
		t.Errorf("expected name 'Infrastructure', got %q", cat.Name)
	}
	if cat.Icon != "server" {
		t.Errorf("expected icon 'server', got %q", cat.Icon)
	}
	if cat.Slug == "" {
		t.Error("expected non-empty slug")
	}
}

func TestGetCategory(t *testing.T) {
	t.Parallel()
	db := setupCategoryDB(t)
	ctx := context.Background()

	created, err := db.CreateCategory(ctx, "Media", "tv")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	fetched, err := db.GetCategory(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetCategory: %v", err)
	}
	if fetched == nil {
		t.Fatal("expected category, got nil")
	}
	if fetched.Name != "Media" {
		t.Errorf("expected name 'Media', got %q", fetched.Name)
	}
	if fetched.Icon != "tv" {
		t.Errorf("expected icon 'tv', got %q", fetched.Icon)
	}
}

func TestGetCategory_NotFound(t *testing.T) {
	t.Parallel()
	db := setupCategoryDB(t)
	ctx := context.Background()

	cat, err := db.GetCategory(ctx, 9999)
	if err != nil {
		t.Fatalf("GetCategory: %v", err)
	}
	if cat != nil {
		t.Errorf("expected nil for nonexistent category, got %+v", cat)
	}
}

func TestGetCategories(t *testing.T) {
	t.Parallel()
	db := setupCategoryDB(t)
	ctx := context.Background()

	if _, err := db.CreateCategory(ctx, "First", "one"); err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}
	if _, err := db.CreateCategory(ctx, "Second", "two"); err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	cats, err := db.GetCategories(ctx)
	if err != nil {
		t.Fatalf("GetCategories: %v", err)
	}
	if len(cats) != 2 {
		t.Fatalf("expected 2 categories, got %d", len(cats))
	}
	// Ordered by sort_order (creation order)
	if cats[0].Name != "First" {
		t.Errorf("expected first category 'First', got %q", cats[0].Name)
	}
	if cats[1].Name != "Second" {
		t.Errorf("expected second category 'Second', got %q", cats[1].Name)
	}
}

func TestUpdateCategory(t *testing.T) {
	t.Parallel()
	db := setupCategoryDB(t)
	ctx := context.Background()

	created, err := db.CreateCategory(ctx, "Old Name", "folder")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	updated, err := db.UpdateCategory(ctx, created.ID, "New Name", "star")
	if err != nil {
		t.Fatalf("UpdateCategory: %v", err)
	}
	if updated == nil {
		t.Fatal("expected updated category, got nil")
	}
	if updated.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %q", updated.Name)
	}
	if updated.Icon != "star" {
		t.Errorf("expected icon 'star', got %q", updated.Icon)
	}
}

func TestUpdateCategory_NotFound(t *testing.T) {
	t.Parallel()
	db := setupCategoryDB(t)
	ctx := context.Background()

	updated, err := db.UpdateCategory(ctx, 9999, "Name", "icon")
	if err != nil {
		t.Fatalf("UpdateCategory: %v", err)
	}
	if updated != nil {
		t.Errorf("expected nil for nonexistent category update, got %+v", updated)
	}
}

func TestDeleteCategory(t *testing.T) {
	t.Parallel()
	db := setupCategoryDB(t)
	ctx := context.Background()

	cat, err := db.CreateCategory(ctx, "ToDelete", "trash")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	deleted, err := db.DeleteCategory(ctx, cat.ID)
	if err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}
	if !deleted {
		t.Error("expected DeleteCategory to return true")
	}

	fetched, err := db.GetCategory(ctx, cat.ID)
	if err != nil {
		t.Fatalf("GetCategory after delete: %v", err)
	}
	if fetched != nil {
		t.Error("expected nil after delete")
	}
}

func TestDeleteCategory_CascadeWidgets(t *testing.T) {
	t.Parallel()
	db := setupCategoryDB(t)
	ctx := context.Background()

	cat, err := db.CreateCategory(ctx, "WithWidgets", "box")
	if err != nil {
		t.Fatalf("CreateCategory: %v", err)
	}

	_, err = db.CreateWidget(ctx, cat.ID, "docker", "My Docker", "{}", 0, 0, 6, 4)
	if err != nil {
		t.Fatalf("CreateWidget: %v", err)
	}

	if _, err := db.DeleteCategory(ctx, cat.ID); err != nil {
		t.Fatalf("DeleteCategory: %v", err)
	}

	widgets, err := db.GetAllWidgets(ctx)
	if err != nil {
		t.Fatalf("GetAllWidgets: %v", err)
	}
	if len(widgets) != 0 {
		t.Errorf("expected 0 widgets after cascade delete, got %d", len(widgets))
	}
}
