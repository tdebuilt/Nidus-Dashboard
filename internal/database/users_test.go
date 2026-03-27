package database

import (
	"context"
	"path/filepath"
	"testing"
)

func setupUserDB(t *testing.T) *DB {
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

func TestCreateUser(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	id, err := db.CreateUser(ctx, "alice", "hash123", "admin")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if id < 1 {
		t.Errorf("expected positive ID, got %d", id)
	}
}

func TestCreateUser_DefaultRole(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	id, err := db.CreateUser(ctx, "bob", "hash456", "")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	u, err := db.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if u.Role != "admin" {
		t.Errorf("expected default role 'admin', got %q", u.Role)
	}
}

func TestGetUserByUsername(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	_, err := db.CreateUser(ctx, "charlie", "hash789", "editor")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	u, err := db.GetUserByUsername(ctx, "charlie")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.Username != "charlie" {
		t.Errorf("expected username 'charlie', got %q", u.Username)
	}
	if u.PasswordHash != "hash789" {
		t.Errorf("expected password_hash 'hash789', got %q", u.PasswordHash)
	}
	if u.Role != "editor" {
		t.Errorf("expected role 'editor', got %q", u.Role)
	}
	if u.TOTPEnabled {
		t.Error("expected TOTP disabled by default")
	}
}

func TestGetUserByUsername_NotFound(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	u, err := db.GetUserByUsername(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if u != nil {
		t.Errorf("expected nil for nonexistent user, got %+v", u)
	}
}

func TestGetUserByID(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	id, err := db.CreateUser(ctx, "diana", "hashABC", "viewer")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	u, err := db.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if u == nil {
		t.Fatal("expected user, got nil")
	}
	if u.ID != id {
		t.Errorf("expected ID %d, got %d", id, u.ID)
	}
	if u.Username != "diana" {
		t.Errorf("expected username 'diana', got %q", u.Username)
	}
}

func TestCountUsers(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func()
		expected int
	}{
		{"zero users", func() {}, 0},
		{"one user", func() {
			db.CreateUser(ctx, "user1", "h1", "admin")
		}, 1},
		{"two users", func() {
			db.CreateUser(ctx, "user2", "h2", "admin")
		}, 2},
	}

	for _, tt := range tests {
		tt.setup()
		count, err := db.CountUsers(ctx)
		if err != nil {
			t.Fatalf("%s: CountUsers: %v", tt.name, err)
		}
		if count != tt.expected {
			t.Errorf("%s: expected %d, got %d", tt.name, tt.expected, count)
		}
	}
}

func TestListUsers(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	if _, err := db.CreateUser(ctx, "alpha", "h1", "admin"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if _, err := db.CreateUser(ctx, "beta", "h2", "viewer"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	users, err := db.ListUsers(ctx)
	if err != nil {
		t.Fatalf("ListUsers: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Username != "alpha" {
		t.Errorf("expected first user 'alpha', got %q", users[0].Username)
	}
	if users[1].Username != "beta" {
		t.Errorf("expected second user 'beta', got %q", users[1].Username)
	}
	// ListUsers should not return password hashes
	if users[0].PasswordHash != "" {
		t.Error("expected empty PasswordHash from ListUsers")
	}
}

func TestDeleteUser(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	id, err := db.CreateUser(ctx, "todelete", "h1", "admin")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := db.DeleteUser(ctx, id); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}

	u, err := db.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID after delete: %v", err)
	}
	if u != nil {
		t.Error("expected nil after delete, got user")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	err := db.DeleteUser(ctx, 9999)
	if err == nil {
		t.Error("expected error deleting nonexistent user")
	}
}

func TestUpdateUserPassword(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	id, err := db.CreateUser(ctx, "passuser", "oldhash", "admin")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := db.UpdateUserPassword(ctx, id, "newhash"); err != nil {
		t.Fatalf("UpdateUserPassword: %v", err)
	}

	u, err := db.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if u.PasswordHash != "newhash" {
		t.Errorf("expected password_hash 'newhash', got %q", u.PasswordHash)
	}
}

func TestUpdateUserRole(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	id, err := db.CreateUser(ctx, "roleuser", "h1", "viewer")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	if err := db.UpdateUserRole(ctx, id, "admin"); err != nil {
		t.Fatalf("UpdateUserRole: %v", err)
	}

	u, err := db.GetUserByID(ctx, id)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if u.Role != "admin" {
		t.Errorf("expected role 'admin', got %q", u.Role)
	}
}

func TestUpdateUserRole_NotFound(t *testing.T) {
	t.Parallel()
	db := setupUserDB(t)
	ctx := context.Background()

	err := db.UpdateUserRole(ctx, 9999, "admin")
	if err == nil {
		t.Error("expected error updating role of nonexistent user")
	}
}
