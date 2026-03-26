package database

import (
	"path/filepath"
	"testing"
)

func openTestDB(t *testing.T) *DB {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")
	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	return db
}

func TestMigrateCreatesAllTables(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	if err := db.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	expectedTables := []string{"migrations", "users", "settings", "categories", "widgets", "services", "app_links"}
	for _, table := range expectedTables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %q not found: %v", table, err)
		}
	}
}

func TestMigrateIsIdempotent(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()

	// Run migration twice — second run should be a no-op
	if err := db.Migrate(); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}

	// Verify migration was only recorded once
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM migrations WHERE version = 1").Scan(&count); err != nil {
		t.Fatalf("failed to count migrations: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 migration record, got %d", count)
	}
}

func TestMigrateUsersTable(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Insert a user
	_, err := db.Exec("INSERT INTO users (username, password_hash) VALUES ('admin', '$2a$10$hash')")
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Verify unique constraint on username
	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES ('admin', '$2a$10$other')")
	if err == nil {
		t.Error("expected unique constraint violation for duplicate username")
	}

	// Verify defaults
	var totpEnabled int
	if err := db.QueryRow("SELECT totp_enabled FROM users WHERE username='admin'").Scan(&totpEnabled); err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if totpEnabled != 0 {
		t.Errorf("expected totp_enabled=0, got %d", totpEnabled)
	}
}

func TestMigrateSettingsTable(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Create user first (FK)
	_, err := db.Exec("INSERT INTO users (username, password_hash) VALUES ('admin', '$2a$10$hash')")
	if err != nil {
		t.Fatalf("failed to insert user: %v", err)
	}

	// Insert setting
	_, err = db.Exec("INSERT INTO settings (user_id, key, value) VALUES (1, 'theme', 'dark')")
	if err != nil {
		t.Fatalf("failed to insert setting: %v", err)
	}

	// Verify unique constraint on (user_id, key)
	_, err = db.Exec("INSERT INTO settings (user_id, key, value) VALUES (1, 'theme', 'light')")
	if err == nil {
		t.Error("expected unique constraint violation for duplicate (user_id, key)")
	}
}

func TestMigrateSettingsFKConstraint(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Insert setting with non-existent user_id should fail
	_, err := db.Exec("INSERT INTO settings (user_id, key, value) VALUES (999, 'theme', 'dark')")
	if err == nil {
		t.Error("expected foreign key violation for non-existent user_id")
	}
}

func TestMigrateCategoriesTable(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	_, err := db.Exec("INSERT INTO categories (name, icon, sort_order) VALUES ('Infrastructure', 'server', 0)")
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}

	var icon string
	if err := db.QueryRow("SELECT icon FROM categories WHERE name='Infrastructure'").Scan(&icon); err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if icon != "server" {
		t.Errorf("expected icon='server', got %q", icon)
	}
}

func TestMigrateWidgetsFKConstraint(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Widget with non-existent category should fail
	_, err := db.Exec("INSERT INTO widgets (category_id, type, title) VALUES (999, 'docker', 'Test')")
	if err == nil {
		t.Error("expected foreign key violation for non-existent category_id")
	}

	// Create category, then widget should succeed
	_, err = db.Exec("INSERT INTO categories (name) VALUES ('Test')")
	if err != nil {
		t.Fatalf("failed to insert category: %v", err)
	}
	_, err = db.Exec("INSERT INTO widgets (category_id, type, title) VALUES (1, 'docker', 'My Container')")
	if err != nil {
		t.Fatalf("failed to insert widget: %v", err)
	}
}

func TestMigrateWidgetCascadeDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Create category + widget
	db.Exec("INSERT INTO categories (name) VALUES ('ToDelete')")
	db.Exec("INSERT INTO widgets (category_id, type, title) VALUES (1, 'docker', 'Widget1')")

	// Delete category — widget should cascade
	_, err := db.Exec("DELETE FROM categories WHERE id = 1")
	if err != nil {
		t.Fatalf("failed to delete category: %v", err)
	}

	var count int
	db.QueryRow("SELECT COUNT(*) FROM widgets").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 widgets after cascade delete, got %d", count)
	}
}

func TestMigrateServicesTable(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	_, err := db.Exec("INSERT INTO services (type, name, url) VALUES ('portainer', 'Portainer', 'https://portainer.example.com')")
	if err != nil {
		t.Fatalf("failed to insert service: %v", err)
	}

	// Unique constraint on type
	_, err = db.Exec("INSERT INTO services (type, name, url) VALUES ('portainer', 'Portainer2', 'https://other.com')")
	if err == nil {
		t.Error("expected unique constraint violation for duplicate service type")
	}
}

func TestMigrateAppLinksTable(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	_, err := db.Exec("INSERT INTO app_links (name, url, icon) VALUES ('Grafana', 'https://grafana.example.com', 'activity')")
	if err != nil {
		t.Fatalf("failed to insert app link: %v", err)
	}

	var icon string
	db.QueryRow("SELECT icon FROM app_links WHERE name='Grafana'").Scan(&icon)
	if icon != "activity" {
		t.Errorf("expected icon='activity', got %q", icon)
	}
}

func TestMigrateSettingsCascadeDelete(t *testing.T) {
	db := openTestDB(t)
	defer db.Close()
	if err := db.Migrate(); err != nil {
		t.Fatalf("migration failed: %v", err)
	}

	// Create user + settings
	db.Exec("INSERT INTO users (username, password_hash) VALUES ('admin', '$2a$10$hash')")
	db.Exec("INSERT INTO settings (user_id, key, value) VALUES (1, 'theme', 'dark')")
	db.Exec("INSERT INTO settings (user_id, key, value) VALUES (1, 'language', 'fr')")

	// Delete user — settings should cascade
	db.Exec("DELETE FROM users WHERE id = 1")

	var count int
	db.QueryRow("SELECT COUNT(*) FROM settings").Scan(&count)
	if count != 0 {
		t.Errorf("expected 0 settings after cascade delete, got %d", count)
	}
}
