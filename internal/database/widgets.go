package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tdebuilt/nidus/internal/models"
)

// CreateWidget inserts a new widget for a category.
func (db *DB) CreateWidget(ctx context.Context, categoryID int64, wType, title, config string, posX, posY, width, height int) (*models.Widget, error) {
	if config == "" {
		config = "{}"
	}
	if width < 1 {
		width = 1
	}
	if height < 0 {
		height = 0
	}
	result, err := db.ExecContext(ctx,
		"INSERT INTO widgets (category_id, type, title, config, pos_x, pos_y, width, height) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		categoryID, wType, title, config, posX, posY, width, height,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting widget: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetWidget(ctx, id)
}

// GetWidget retrieves a widget by ID, or nil if not found.
func (db *DB) GetWidget(ctx context.Context, id int64) (*models.Widget, error) {
	w := &models.Widget{}
	var collapsed int
	err := db.QueryRowContext(ctx,
		"SELECT id, category_id, type, title, config, collapsed, pos_x, pos_y, width, height, created_at, updated_at FROM widgets WHERE id = ?", id,
	).Scan(&w.ID, &w.CategoryID, &w.Type, &w.Title, &w.Config, &collapsed, &w.PosX, &w.PosY, &w.Width, &w.Height, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying widget: %w", err)
	}
	w.Collapsed = collapsed == 1
	return w, nil
}

// GetWidgetsByCategory retrieves all widgets for a category.
func (db *DB) GetWidgetsByCategory(ctx context.Context, categoryID int64) ([]models.Widget, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT id, category_id, type, title, config, collapsed, pos_x, pos_y, width, height, created_at, updated_at FROM widgets WHERE category_id = ? ORDER BY pos_y, pos_x",
		categoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying widgets: %w", err)
	}
	defer rows.Close()

	var widgets []models.Widget
	for rows.Next() {
		var w models.Widget
		var collapsed int
		if err := rows.Scan(&w.ID, &w.CategoryID, &w.Type, &w.Title, &w.Config, &collapsed, &w.PosX, &w.PosY, &w.Width, &w.Height, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning widget: %w", err)
		}
		w.Collapsed = collapsed == 1
		widgets = append(widgets, w)
	}
	return widgets, rows.Err()
}

// GetAllWidgets retrieves all widgets.
func (db *DB) GetAllWidgets(ctx context.Context) ([]models.Widget, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT id, category_id, type, title, config, collapsed, pos_x, pos_y, width, height, created_at, updated_at FROM widgets ORDER BY category_id, pos_y, pos_x",
	)
	if err != nil {
		return nil, fmt.Errorf("querying all widgets: %w", err)
	}
	defer rows.Close()

	var widgets []models.Widget
	for rows.Next() {
		var w models.Widget
		var collapsed int
		if err := rows.Scan(&w.ID, &w.CategoryID, &w.Type, &w.Title, &w.Config, &collapsed, &w.PosX, &w.PosY, &w.Width, &w.Height, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning widget: %w", err)
		}
		w.Collapsed = collapsed == 1
		widgets = append(widgets, w)
	}
	return widgets, rows.Err()
}

// UpdateWidget updates a widget's type, title, and config.
func (db *DB) UpdateWidget(ctx context.Context, id int64, wType, title, config string) (*models.Widget, error) {
	result, err := db.ExecContext(ctx,
		"UPDATE widgets SET type = ?, title = ?, config = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		wType, title, config, id,
	)
	if err != nil {
		return nil, fmt.Errorf("updating widget: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return db.GetWidget(ctx, id)
}

// SetWidgetCollapsed updates the collapsed state of a widget.
func (db *DB) SetWidgetCollapsed(ctx context.Context, id int64, collapsed bool) (*models.Widget, error) {
	val := 0
	if collapsed {
		val = 1
	}
	result, err := db.ExecContext(ctx,
		"UPDATE widgets SET collapsed = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		val, id,
	)
	if err != nil {
		return nil, fmt.Errorf("updating widget collapsed: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return db.GetWidget(ctx, id)
}

// DeleteWidget deletes a widget by ID. Returns true if deleted.
func (db *DB) DeleteWidget(ctx context.Context, id int64) (bool, error) {
	result, err := db.ExecContext(ctx, "DELETE FROM widgets WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("deleting widget: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// SaveWidgetLayout updates positions and sizes for multiple widgets in a transaction.
func (db *DB) SaveWidgetLayout(ctx context.Context, layouts []models.WidgetLayout) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	for _, l := range layouts {
		if _, err := tx.ExecContext(ctx,
			"UPDATE widgets SET pos_x = ?, pos_y = ?, width = ?, height = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			l.PosX, l.PosY, l.Width, l.Height, l.ID,
		); err != nil {
			return fmt.Errorf("updating widget layout: %w", err)
		}
	}
	return tx.Commit()
}
