package database

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func setupServiceDB(t *testing.T) *DB {
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

func TestCreateService(t *testing.T) {
	t.Parallel()
	db := setupServiceDB(t)
	ctx := context.Background()

	svc, err := db.UpsertService(ctx, "portainer", "Portainer", "https://portainer.local:9443", "enc-token", "{}", true)
	if err != nil {
		t.Fatalf("UpsertService: %v", err)
	}
	if svc == nil {
		t.Fatal("expected service, got nil")
	}
	if svc.ID < 1 {
		t.Errorf("expected positive ID, got %d", svc.ID)
	}
	if svc.Type != "portainer" {
		t.Errorf("expected type 'portainer', got %q", svc.Type)
	}
	if svc.Name != "Portainer" {
		t.Errorf("expected name 'Portainer', got %q", svc.Name)
	}
	if svc.URL != "https://portainer.local:9443" {
		t.Errorf("expected URL 'https://portainer.local:9443', got %q", svc.URL)
	}
	if !svc.Enabled {
		t.Error("expected service to be enabled")
	}
}

func TestCreateService_DefaultConfig(t *testing.T) {
	t.Parallel()
	db := setupServiceDB(t)
	ctx := context.Background()

	// Empty config should default to "{}"
	svc, err := db.UpsertService(ctx, "proxmox", "Proxmox", "https://pve.local:8006", "token", "", true)
	if err != nil {
		t.Fatalf("UpsertService: %v", err)
	}
	if svc.Config != "{}" {
		t.Errorf("expected config '{}', got %q", svc.Config)
	}
}

func TestGetServiceByType(t *testing.T) {
	t.Parallel()
	db := setupServiceDB(t)
	ctx := context.Background()

	if _, err := db.UpsertService(ctx, "homeassistant", "Home Assistant", "https://ha.local:8123", "token123", "{}", true); err != nil {
		t.Fatalf("UpsertService: %v", err)
	}

	svc, err := db.GetServiceByType(ctx, "homeassistant")
	if err != nil {
		t.Fatalf("GetServiceByType: %v", err)
	}
	if svc == nil {
		t.Fatal("expected service, got nil")
	}
	if svc.Name != "Home Assistant" {
		t.Errorf("expected name 'Home Assistant', got %q", svc.Name)
	}
	if svc.Credentials != "token123" {
		t.Errorf("expected credentials 'token123', got %q", svc.Credentials)
	}
}

func TestGetServiceByType_NotFound(t *testing.T) {
	t.Parallel()
	db := setupServiceDB(t)
	ctx := context.Background()

	svc, err := db.GetServiceByType(ctx, "nonexistent")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetServiceByType: expected ErrNotFound, got %v", err)
	}
	if svc != nil {
		t.Errorf("expected nil for nonexistent service, got %+v", svc)
	}
}

func TestListServices(t *testing.T) {
	t.Parallel()
	db := setupServiceDB(t)
	ctx := context.Background()

	if _, err := db.UpsertService(ctx, "portainer", "Portainer", "https://p.local", "t1", "{}", true); err != nil {
		t.Fatalf("UpsertService: %v", err)
	}
	if _, err := db.UpsertService(ctx, "adguard", "AdGuard", "https://a.local", "t2", "{}", false); err != nil {
		t.Fatalf("UpsertService: %v", err)
	}

	services, err := db.GetServices(ctx)
	if err != nil {
		t.Fatalf("GetServices: %v", err)
	}
	if len(services) != 2 {
		t.Fatalf("expected 2 services, got %d", len(services))
	}

	// Ordered by type alphabetically
	if services[0].Type != "adguard" {
		t.Errorf("expected first service type 'adguard', got %q", services[0].Type)
	}
	if services[1].Type != "portainer" {
		t.Errorf("expected second service type 'portainer', got %q", services[1].Type)
	}
}

func TestUpdateService(t *testing.T) {
	t.Parallel()
	db := setupServiceDB(t)
	ctx := context.Background()

	if _, err := db.UpsertService(ctx, "portainer", "Old Name", "https://old.local", "cred", "{}", true); err != nil {
		t.Fatalf("UpsertService: %v", err)
	}

	// Upsert with same type updates the existing record
	updated, err := db.UpsertService(ctx, "portainer", "New Name", "https://new.local", "newcred", `{"key":"val"}`, false)
	if err != nil {
		t.Fatalf("UpsertService update: %v", err)
	}
	if updated.Name != "New Name" {
		t.Errorf("expected name 'New Name', got %q", updated.Name)
	}
	if updated.URL != "https://new.local" {
		t.Errorf("expected URL 'https://new.local', got %q", updated.URL)
	}
	if updated.Enabled {
		t.Error("expected service to be disabled after update")
	}

	// Verify only one service exists (upsert, not duplicate)
	services, err := db.GetServices(ctx)
	if err != nil {
		t.Fatalf("GetServices: %v", err)
	}
	if len(services) != 1 {
		t.Errorf("expected 1 service after upsert, got %d", len(services))
	}
}

func TestUpdateService_EmptyCredentials(t *testing.T) {
	t.Parallel()
	db := setupServiceDB(t)
	ctx := context.Background()

	if _, err := db.UpsertService(ctx, "portainer", "Portainer", "https://p.local", "secret-token", "{}", true); err != nil {
		t.Fatalf("UpsertService: %v", err)
	}

	// Update with empty credentials should preserve existing credentials
	updated, err := db.UpsertService(ctx, "portainer", "Portainer", "https://p.local", "", "{}", true)
	if err != nil {
		t.Fatalf("UpsertService: %v", err)
	}
	if updated.Credentials != "secret-token" {
		t.Errorf("expected preserved credentials 'secret-token', got %q", updated.Credentials)
	}
}

func TestDeleteService(t *testing.T) {
	t.Parallel()
	db := setupServiceDB(t)
	ctx := context.Background()

	if _, err := db.UpsertService(ctx, "portainer", "Portainer", "https://p.local", "t", "{}", true); err != nil {
		t.Fatalf("UpsertService: %v", err)
	}

	if err := db.DeleteService(ctx, "portainer"); err != nil {
		t.Fatalf("DeleteService: %v", err)
	}

	svc, err := db.GetServiceByType(ctx, "portainer")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetServiceByType after delete: expected ErrNotFound, got %v", err)
	}
	if svc != nil {
		t.Error("expected nil after delete")
	}
}

func TestDeleteService_NotFound(t *testing.T) {
	t.Parallel()
	db := setupServiceDB(t)
	ctx := context.Background()

	err := db.DeleteService(ctx, "nonexistent")
	if err == nil {
		t.Error("expected error deleting nonexistent service")
	}
}
