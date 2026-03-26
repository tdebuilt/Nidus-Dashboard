package database

import (
	"path/filepath"
	"testing"
)

func TestOpenAndClose(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Verify we can query
	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if result != 1 {
		t.Errorf("expected 1, got %d", result)
	}

	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}
}

func TestWALMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode").Scan(&journalMode); err != nil {
		t.Fatalf("failed to check journal mode: %v", err)
	}
	if journalMode != "wal" {
		t.Errorf("expected WAL journal mode, got %s", journalMode)
	}
}

func TestForeignKeysEnabled(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	var fkEnabled int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled); err != nil {
		t.Fatalf("failed to check foreign keys: %v", err)
	}
	if fkEnabled != 1 {
		t.Errorf("expected foreign keys enabled (1), got %d", fkEnabled)
	}
}

func TestForeignKeysEnforced(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create parent and child tables
	_, err = db.Exec(`
		CREATE TABLE parent (id INTEGER PRIMARY KEY);
		CREATE TABLE child (id INTEGER PRIMARY KEY, parent_id INTEGER REFERENCES parent(id));
	`)
	if err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	// Insert into child with non-existent parent should fail
	_, err = db.Exec("INSERT INTO child (id, parent_id) VALUES (1, 999)")
	if err == nil {
		t.Error("expected foreign key violation error, got nil")
	}
}

func TestCloseIsClean(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("failed to close database: %v", err)
	}

	// After close, queries should fail
	var result int
	err = db.QueryRow("SELECT 1").Scan(&result)
	if err == nil {
		t.Error("expected error after close, got nil")
	}
}

func TestCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nested", "test.db")

	db, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open database in nested dir: %v", err)
	}
	defer db.Close()

	var result int
	if err := db.QueryRow("SELECT 1").Scan(&result); err != nil {
		t.Fatalf("failed to query: %v", err)
	}
}

func TestReopenExistingDB(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.db")

	// First open: create table and insert data
	db1, err := Open(path)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	_, err = db1.Exec("CREATE TABLE test (val TEXT)")
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	_, err = db1.Exec("INSERT INTO test (val) VALUES ('hello')")
	if err != nil {
		t.Fatalf("failed to insert: %v", err)
	}
	db1.Close()

	// Second open: data should persist
	db2, err := Open(path)
	if err != nil {
		t.Fatalf("failed to reopen database: %v", err)
	}
	defer db2.Close()

	var val string
	if err := db2.QueryRow("SELECT val FROM test").Scan(&val); err != nil {
		t.Fatalf("failed to query: %v", err)
	}
	if val != "hello" {
		t.Errorf("expected 'hello', got '%s'", val)
	}
}
