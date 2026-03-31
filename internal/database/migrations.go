package database

import (
	"context"
	"fmt"
)

// migrate runs all database migrations in order.
// Each migration is idempotent (uses IF NOT EXISTS).
func (db *DB) Migrate(ctx context.Context) error {
	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version INTEGER NOT NULL UNIQUE,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	migrations := []struct {
		version int
		sql     string
	}{
		{1, migrationV1}, {2, migrationV2}, {3, migrationV3}, {4, migrationV4},
		{5, migrationV5}, {6, migrationV6}, {7, migrationV7}, {8, migrationV8},
		{9, migrationV9}, {10, migrationV10}, {11, migrationV11}, {12, migrationV12},
		{13, migrationV13}, {14, migrationV14},
	}

	for _, m := range migrations {
		if err := db.applyMigration(ctx, m.version, m.sql); err != nil {
			return err
		}
	}
	return nil
}

// applyMigration runs a single migration if not already applied.
func (db *DB) applyMigration(ctx context.Context, version int, sqlStr string) error {
	var count int
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM migrations WHERE version = ?", version).Scan(&count); err != nil {
		return fmt.Errorf("checking migration %d: %w", version, err)
	}
	if count > 0 {
		return nil
	}
	if _, err := db.ExecContext(ctx, sqlStr); err != nil {
		return fmt.Errorf("running migration %d: %w", version, err)
	}
	if version == 12 {
		if err := db.migrateV12Slugs(ctx); err != nil {
			return fmt.Errorf("post-migration V12 slugs: %w", err)
		}
	}
	if _, err := db.ExecContext(ctx, "INSERT INTO migrations (version) VALUES (?)", version); err != nil {
		return fmt.Errorf("recording migration %d: %w", version, err)
	}
	return nil
}

// migrateV12Slugs generates slugs for all existing categories and adds a unique index.
func (db *DB) migrateV12Slugs(ctx context.Context) error {
	rows, err := db.QueryContext(ctx, "SELECT id, name FROM categories WHERE slug IS NULL")
	if err != nil {
		return fmt.Errorf("querying categories without slug: %w", err)
	}
	defer rows.Close()

	type catRow struct {
		id   int64
		name string
	}
	var cats []catRow
	for rows.Next() {
		var c catRow
		if err := rows.Scan(&c.id, &c.name); err != nil {
			return fmt.Errorf("scanning category: %w", err)
		}
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, c := range cats {
		slug, err := db.generateUniqueSlug(ctx, GenerateSlug(c.name))
		if err != nil {
			return fmt.Errorf("generating slug for category %d: %w", c.id, err)
		}
		if _, err := db.ExecContext(ctx, "UPDATE categories SET slug = ? WHERE id = ?", slug, c.id); err != nil {
			return fmt.Errorf("updating slug for category %d: %w", c.id, err)
		}
	}

	if _, err := db.ExecContext(ctx, "CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_slug ON categories(slug)"); err != nil {
		return fmt.Errorf("creating slug unique index: %w", err)
	}
	return nil
}

const migrationV1 = `
-- Users table
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	totp_secret TEXT,
	totp_enabled INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Settings table (key-value store for user preferences)
CREATE TABLE IF NOT EXISTS settings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	key TEXT NOT NULL,
	value TEXT NOT NULL,
	UNIQUE(user_id, key)
);

-- Categories table
CREATE TABLE IF NOT EXISTS categories (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	icon TEXT NOT NULL DEFAULT 'folder',
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Widgets table
CREATE TABLE IF NOT EXISTS widgets (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	category_id INTEGER NOT NULL REFERENCES categories(id) ON DELETE CASCADE,
	type TEXT NOT NULL,
	title TEXT NOT NULL,
	config TEXT NOT NULL DEFAULT '{}',
	pos_x INTEGER NOT NULL DEFAULT 0,
	pos_y INTEGER NOT NULL DEFAULT 0,
	width INTEGER NOT NULL DEFAULT 1,
	height INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Services table (external service connections with encrypted credentials)
CREATE TABLE IF NOT EXISTS services (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	url TEXT,
	credentials TEXT,
	enabled INTEGER NOT NULL DEFAULT 1,
	config TEXT NOT NULL DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- App links table (quick-access links to self-hosted services)
CREATE TABLE IF NOT EXISTS app_links (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	url TEXT NOT NULL,
	icon TEXT NOT NULL DEFAULT 'link',
	health_check_url TEXT,
	sort_order INTEGER NOT NULL DEFAULT 0,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const migrationV2 = `
-- System settings table (JWT secret, encryption key, etc.)
CREATE TABLE IF NOT EXISTS system_settings (
	key TEXT PRIMARY KEY,
	value TEXT NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const migrationV3 = `
-- Add collapsed state to widgets
ALTER TABLE widgets ADD COLUMN collapsed INTEGER NOT NULL DEFAULT 0;
`

const migrationV4 = `
-- Reset widget heights to 0 (auto height based on content)
UPDATE widgets SET height = 0;
`

const migrationV5 = `
-- Switch from 4-column to 12-column grid: scale width and pos_x by 3
UPDATE widgets SET width = width * 3, pos_x = pos_x * 3;
`

const migrationV6 = `
-- Add role column to users table (admin, editor, viewer)
ALTER TABLE users ADD COLUMN role TEXT NOT NULL DEFAULT 'admin';

-- Invitations table for user onboarding
CREATE TABLE IF NOT EXISTS invitations (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	code TEXT NOT NULL UNIQUE,
	role TEXT NOT NULL DEFAULT 'viewer',
	created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	used_by INTEGER REFERENCES users(id) ON DELETE SET NULL,
	used_at DATETIME,
	expires_at DATETIME NOT NULL,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const migrationV7 = `
-- Notification providers (Gotify, Ntfy, Apprise)
CREATE TABLE IF NOT EXISTS notification_providers (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type TEXT NOT NULL,
	name TEXT NOT NULL,
	url TEXT NOT NULL,
	token TEXT NOT NULL DEFAULT '',
	enabled INTEGER NOT NULL DEFAULT 1,
	config TEXT NOT NULL DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Notification rules (what events trigger notifications)
CREATE TABLE IF NOT EXISTS notification_rules (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	event_type TEXT NOT NULL,
	provider_id INTEGER NOT NULL REFERENCES notification_providers(id) ON DELETE CASCADE,
	enabled INTEGER NOT NULL DEFAULT 1,
	config TEXT NOT NULL DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const migrationV8 = `
-- Incoming webhooks
CREATE TABLE IF NOT EXISTS webhooks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	secret TEXT NOT NULL,
	enabled INTEGER NOT NULL DEFAULT 1,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Webhook actions (what happens when a webhook fires)
CREATE TABLE IF NOT EXISTS webhook_actions (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	webhook_id INTEGER NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
	action_type TEXT NOT NULL,
	config TEXT NOT NULL DEFAULT '{}',
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const migrationV9 = `
-- Migrate widget grid from 12 columns to 24 columns (double pos_x and width)
-- and from 80px rows to 10px rows (x8 pos_y and height)
UPDATE widgets SET pos_x = pos_x * 2, width = width * 2, pos_y = pos_y * 8, height = CASE WHEN height > 0 THEN height * 8 ELSE 0 END;
`

const migrationV10 = `
-- V9 already ran with x2 for rows, need x4 more to reach x8 total (ROW_UNIT 80->10)
UPDATE widgets SET pos_y = pos_y * 4, height = CASE WHEN height > 0 THEN height * 4 ELSE 0 END;
`

const migrationV11 = `
-- Add token version to users for JWT invalidation on password/role changes
ALTER TABLE users ADD COLUMN token_version INTEGER NOT NULL DEFAULT 0;
`

const migrationV12 = `
-- Add slug column to categories for user-friendly URLs
ALTER TABLE categories ADD COLUMN slug TEXT;
`

const migrationV13 = `
-- Custom themes created by admins
CREATE TABLE IF NOT EXISTS custom_themes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    theme_json TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`

const migrationV14 = `
-- Password reset codes for admin-initiated account resets
CREATE TABLE IF NOT EXISTS password_resets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    code TEXT NOT NULL UNIQUE,
    created_by INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    used_at DATETIME,
    expires_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
`
