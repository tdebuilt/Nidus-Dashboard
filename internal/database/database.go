package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps a sql.DB connection with SQLite-specific configuration.
type DB struct {
	*sql.DB
}

// Open creates or opens a SQLite database at the given path.
// It enables WAL mode and foreign keys.
func Open(path string) (*DB, error) {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(10000)&_txlock=immediate")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	if err := configurePragmas(db); err != nil {
		db.Close()
		return nil, err
	}

	return &DB{db}, nil
}

// configurePragmas enables WAL mode, foreign keys, and performance settings.
func configurePragmas(db *sql.DB) error {
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode=WAL").Scan(&journalMode); err != nil {
		return fmt.Errorf("setting journal mode: %w", err)
	}
	if journalMode != "wal" {
		return fmt.Errorf("expected WAL journal mode, got %s", journalMode)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("enabling foreign keys: %w", err)
	}
	var fkEnabled int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled); err != nil {
		return fmt.Errorf("checking foreign keys: %w", err)
	}
	if fkEnabled != 1 {
		return fmt.Errorf("foreign keys not enabled")
	}

	pragmas := []struct{ sql, desc string }{
		{"PRAGMA synchronous = NORMAL", "setting synchronous mode"},
		{"PRAGMA cache_size = -64000", "setting cache size"},
		{"PRAGMA temp_store = MEMORY", "setting temp store"},
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p.sql); err != nil {
			return fmt.Errorf("%s: %w", p.desc, err)
		}
	}
	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}
