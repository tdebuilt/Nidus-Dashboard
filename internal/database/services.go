package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tdebuilt/nidus/internal/models"
)

// GetServices retrieves all services.
func (db *DB) GetServices(ctx context.Context) ([]models.Service, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, type, name, COALESCE(url,''), COALESCE(credentials,''), enabled, config, created_at, updated_at FROM services ORDER BY type")
	if err != nil {
		return nil, fmt.Errorf("querying services: %w", err)
	}
	defer rows.Close()

	var services []models.Service
	for rows.Next() {
		var s models.Service
		var enabled int
		if err := rows.Scan(&s.ID, &s.Type, &s.Name, &s.URL, &s.Credentials, &enabled, &s.Config, &s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning service: %w", err)
		}
		s.Enabled = enabled == 1
		services = append(services, s)
	}
	return services, rows.Err()
}

// GetServiceByType retrieves a service by type, or nil if not found.
func (db *DB) GetServiceByType(ctx context.Context, serviceType string) (*models.Service, error) {
	s := &models.Service{}
	var enabled int
	err := db.QueryRowContext(ctx,
		"SELECT id, type, name, COALESCE(url,''), COALESCE(credentials,''), enabled, config, created_at, updated_at FROM services WHERE type = ?",
		serviceType,
	).Scan(&s.ID, &s.Type, &s.Name, &s.URL, &s.Credentials, &enabled, &s.Config, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("querying service: %w", err)
	}
	s.Enabled = enabled == 1
	return s, nil
}

// UpsertService inserts or updates a service by type.
func (db *DB) UpsertService(ctx context.Context, serviceType, name, url, credentials, config string, enabled bool) (*models.Service, error) {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	if config == "" {
		config = "{}"
	}
	_, err := db.ExecContext(ctx, `
		INSERT INTO services (type, name, url, credentials, enabled, config)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(type) DO UPDATE SET
			name = excluded.name,
			url = excluded.url,
			credentials = CASE WHEN excluded.credentials = '' THEN services.credentials ELSE excluded.credentials END,
			enabled = excluded.enabled,
			config = excluded.config,
			updated_at = CURRENT_TIMESTAMP
	`, serviceType, name, url, credentials, enabledInt, config)
	if err != nil {
		return nil, fmt.Errorf("upserting service: %w", err)
	}
	return db.GetServiceByType(ctx, serviceType)
}

// DeleteService deletes a service by type.
func (db *DB) DeleteService(ctx context.Context, serviceType string) error {
	result, err := db.ExecContext(ctx, "DELETE FROM services WHERE type = ?", serviceType)
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("service not found")
	}
	return nil
}
