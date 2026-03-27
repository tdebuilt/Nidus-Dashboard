package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tdebuilt/nidus/internal/models"
)

// CreateCategory inserts a new category. Sort order is set to max+1.
func (db *DB) CreateCategory(ctx context.Context, name, icon string) (*models.Category, error) {
	var maxOrder int
	db.QueryRowContext(ctx, "SELECT COALESCE(MAX(sort_order), -1) FROM categories").Scan(&maxOrder)

	slug, err := db.generateUniqueSlug(ctx, GenerateSlug(name))
	if err != nil {
		return nil, fmt.Errorf("generating slug: %w", err)
	}

	result, err := db.ExecContext(ctx,
		"INSERT INTO categories (name, icon, sort_order, slug) VALUES (?, ?, ?, ?)",
		name, icon, maxOrder+1, slug,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting category: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetCategory(ctx, id)
}

// GetCategory retrieves a category by ID, or nil if not found.
func (db *DB) GetCategory(ctx context.Context, id int64) (*models.Category, error) {
	c := &models.Category{}
	err := db.QueryRowContext(ctx,
		"SELECT id, name, slug, icon, sort_order, created_at, updated_at FROM categories WHERE id = ?", id,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Icon, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying category: %w", err)
	}
	return c, nil
}

// GetCategories retrieves all categories ordered by sort_order.
func (db *DB) GetCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name, slug, icon, sort_order, created_at, updated_at FROM categories ORDER BY sort_order")
	if err != nil {
		return nil, fmt.Errorf("querying categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Icon, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// UpdateCategory updates a category's name and icon.
func (db *DB) UpdateCategory(ctx context.Context, id int64, name, icon string) (*models.Category, error) {
	result, err := db.ExecContext(ctx,
		"UPDATE categories SET name = ?, icon = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		name, icon, id,
	)
	if err != nil {
		return nil, fmt.Errorf("updating category: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return db.GetCategory(ctx, id)
}

// DeleteCategory deletes a category by ID. Returns true if deleted.
func (db *DB) DeleteCategory(ctx context.Context, id int64) (bool, error) {
	result, err := db.ExecContext(ctx, "DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("deleting category: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// ReorderCategories updates sort_order for categories based on the given ID order.
func (db *DB) ReorderCategories(ctx context.Context, ids []int64) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	for i, id := range ids {
		if _, err := tx.ExecContext(ctx,
			"UPDATE categories SET sort_order = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			i, id,
		); err != nil {
			return fmt.Errorf("updating sort order: %w", err)
		}
	}
	return tx.Commit()
}
