package testutil

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/tdebuilt/nidus/internal/database"
)

// SetupTestDB creates a temporary SQLite database with all migrations applied.
// The database is automatically cleaned up when the test finishes.
func SetupTestDB(t *testing.T) *database.DB {
	t.Helper()
	ctx := context.Background()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
