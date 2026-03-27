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
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(10000)&_txlock=immediate")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Limit concurrent connections to avoid SQLite locking issues
	db.SetMaxOpenConns(1)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// Enable WAL mode
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode=WAL").Scan(&journalMode); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting journal mode: %w", err)
	}
	if journalMode != "wal" {
		db.Close()
		return nil, fmt.Errorf("expected WAL journal mode, got %s", journalMode)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}
	var fkEnabled int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled); err != nil {
		db.Close()
		return nil, fmt.Errorf("checking foreign keys: %w", err)
	}
	if fkEnabled != 1 {
		db.Close()
		return nil, fmt.Errorf("foreign keys not enabled")
	}

	// Performance: synchronous=NORMAL is safe with WAL mode
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting synchronous mode: %w", err)
	}

	// 64MB page cache (default ~2MB)
	if _, err := db.Exec("PRAGMA cache_size = -64000"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting cache size: %w", err)
	}

	// Temp tables in memory
	if _, err := db.Exec("PRAGMA temp_store = MEMORY"); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting temp store: %w", err)
	}

	return &DB{db}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}
