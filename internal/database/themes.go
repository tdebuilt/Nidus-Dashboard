package database

import (
	"context"
	"database/sql"

	"github.com/tdebuilt/nidus/internal/models"
)

// ListCustomThemes returns all custom themes ordered by creation date.
func (db *DB) ListCustomThemes(ctx context.Context) ([]models.CustomTheme, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name, theme_json, created_at, updated_at FROM custom_themes ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var themes []models.CustomTheme
	for rows.Next() {
		var t models.CustomTheme
		if err := rows.Scan(&t.ID, &t.Name, &t.ThemeJSON, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		themes = append(themes, t)
	}
	if themes == nil {
		themes = []models.CustomTheme{}
	}
	return themes, rows.Err()
}

// GetCustomTheme returns a custom theme by ID.
func (db *DB) GetCustomTheme(ctx context.Context, id int64) (*models.CustomTheme, error) {
	var t models.CustomTheme
	err := db.QueryRowContext(ctx,
		"SELECT id, name, theme_json, created_at, updated_at FROM custom_themes WHERE id = ?", id,
	).Scan(&t.ID, &t.Name, &t.ThemeJSON, &t.CreatedAt, &t.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CountCustomThemes returns the number of custom themes.
func (db *DB) CountCustomThemes(ctx context.Context) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM custom_themes").Scan(&count)
	return count, err
}

// CreateCustomTheme creates a new custom theme and returns its ID.
func (db *DB) CreateCustomTheme(ctx context.Context, name, themeJSON string) (int64, error) {
	result, err := db.ExecContext(ctx,
		"INSERT INTO custom_themes (name, theme_json) VALUES (?, ?)",
		name, themeJSON,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateCustomTheme updates a custom theme's name and JSON.
func (db *DB) UpdateCustomTheme(ctx context.Context, id int64, name, themeJSON string) error {
	_, err := db.ExecContext(ctx,
		"UPDATE custom_themes SET name = ?, theme_json = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		name, themeJSON, id,
	)
	return err
}

// DeleteCustomTheme deletes a custom theme by ID.
func (db *DB) DeleteCustomTheme(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, "DELETE FROM custom_themes WHERE id = ?", id)
	return err
}
