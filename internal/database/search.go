package database

import (
	"context"
	"fmt"
)

// SearchWidgetResult holds a widget search result with its category info.
type SearchWidgetResult struct {
	ID           int64
	Title        string
	CategoryID   int64
	CategoryName string
}

// SearchCategoryResult holds a category search result.
type SearchCategoryResult struct {
	ID   int64
	Name string
}

// SearchWidgets searches widgets whose title matches the query (case-insensitive).
func (db *DB) SearchWidgets(ctx context.Context, query string) ([]SearchWidgetResult, error) {
	pattern := "%" + query + "%"
	rows, err := db.QueryContext(ctx,
		`SELECT w.id, w.title, w.category_id, c.name
		 FROM widgets w
		 JOIN categories c ON w.category_id = c.id
		 WHERE w.title LIKE ? COLLATE NOCASE
		 ORDER BY w.title`,
		pattern,
	)
	if err != nil {
		return nil, fmt.Errorf("searching widgets: %w", err)
	}
	defer rows.Close()

	var results []SearchWidgetResult
	for rows.Next() {
		var r SearchWidgetResult
		if err := rows.Scan(&r.ID, &r.Title, &r.CategoryID, &r.CategoryName); err != nil {
			return nil, fmt.Errorf("scanning widget result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

// SearchCategories searches categories whose name matches the query (case-insensitive).
func (db *DB) SearchCategories(ctx context.Context, query string) ([]SearchCategoryResult, error) {
	pattern := "%" + query + "%"
	rows, err := db.QueryContext(ctx,
		"SELECT id, name FROM categories WHERE name LIKE ? COLLATE NOCASE ORDER BY name",
		pattern,
	)
	if err != nil {
		return nil, fmt.Errorf("searching categories: %w", err)
	}
	defer rows.Close()

	var results []SearchCategoryResult
	for rows.Next() {
		var r SearchCategoryResult
		if err := rows.Scan(&r.ID, &r.Name); err != nil {
			return nil, fmt.Errorf("scanning category result: %w", err)
		}
		results = append(results, r)
	}
	return results, rows.Err()
}
