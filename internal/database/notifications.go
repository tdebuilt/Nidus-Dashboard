package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tdebuilt/nidus/internal/models"
)

// CreateNotificationProvider creates a new notification provider.
func (db *DB) CreateNotificationProvider(ctx context.Context, providerType, name, url, token, config string) (*models.NotificationProvider, error) {
	if config == "" {
		config = "{}"
	}
	result, err := db.ExecContext(ctx,
		"INSERT INTO notification_providers (type, name, url, token, enabled, config) VALUES (?, ?, ?, ?, 1, ?)",
		providerType, name, url, token, config,
	)
	if err != nil {
		return nil, fmt.Errorf("insert notification provider: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}
	return db.GetNotificationProvider(ctx, id)
}

// GetNotificationProvider returns a notification provider by ID.
func (db *DB) GetNotificationProvider(ctx context.Context, id int64) (*models.NotificationProvider, error) {
	var p models.NotificationProvider
	var enabled int
	err := db.QueryRowContext(ctx,
		"SELECT id, type, name, url, token, enabled, config, created_at, updated_at FROM notification_providers WHERE id = ?", id,
	).Scan(&p.ID, &p.Type, &p.Name, &p.URL, &p.Token, &enabled, &p.Config, &p.CreatedAt, &p.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get notification provider: %w", err)
	}
	p.Enabled = enabled == 1
	return &p, nil
}

// ListNotificationProviders returns all notification providers.
func (db *DB) ListNotificationProviders(ctx context.Context) ([]models.NotificationProvider, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, type, name, url, token, enabled, config, created_at, updated_at FROM notification_providers ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("list notification providers: %w", err)
	}
	defer rows.Close()
	var providers []models.NotificationProvider
	for rows.Next() {
		var p models.NotificationProvider
		var enabled int
		if err := rows.Scan(&p.ID, &p.Type, &p.Name, &p.URL, &p.Token, &enabled, &p.Config, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan notification provider: %w", err)
		}
		p.Enabled = enabled == 1
		providers = append(providers, p)
	}
	if providers == nil {
		providers = []models.NotificationProvider{}
	}
	return providers, rows.Err()
}

// UpdateNotificationProvider updates fields on a notification provider.
func (db *DB) UpdateNotificationProvider(ctx context.Context, id int64, name, url, token *string, enabled *bool, config *string) error {
	if name != nil {
		if _, err := db.ExecContext(ctx, "UPDATE notification_providers SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *name, id); err != nil {
			return fmt.Errorf("update provider name: %w", err)
		}
	}
	if url != nil {
		if _, err := db.ExecContext(ctx, "UPDATE notification_providers SET url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *url, id); err != nil {
			return fmt.Errorf("update provider url: %w", err)
		}
	}
	if token != nil {
		if _, err := db.ExecContext(ctx, "UPDATE notification_providers SET token = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *token, id); err != nil {
			return fmt.Errorf("update provider token: %w", err)
		}
	}
	if enabled != nil {
		val := 0
		if *enabled {
			val = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE notification_providers SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", val, id); err != nil {
			return fmt.Errorf("update provider enabled: %w", err)
		}
	}
	if config != nil {
		if _, err := db.ExecContext(ctx, "UPDATE notification_providers SET config = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *config, id); err != nil {
			return fmt.Errorf("update provider config: %w", err)
		}
	}
	return nil
}

// DeleteNotificationProvider deletes a notification provider by ID.
func (db *DB) DeleteNotificationProvider(ctx context.Context, id int64) error {
	if _, err := db.ExecContext(ctx, "DELETE FROM notification_providers WHERE id = ?", id); err != nil {
		return fmt.Errorf("delete notification provider: %w", err)
	}
	return nil
}

// CreateNotificationRule creates a new notification rule.
func (db *DB) CreateNotificationRule(ctx context.Context, eventType string, providerID int64, config string) (*models.NotificationRule, error) {
	if config == "" {
		config = "{}"
	}
	result, err := db.ExecContext(ctx,
		"INSERT INTO notification_rules (event_type, provider_id, enabled, config) VALUES (?, ?, 1, ?)",
		eventType, providerID, config,
	)
	if err != nil {
		return nil, fmt.Errorf("insert notification rule: %w", err)
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}
	return db.GetNotificationRule(ctx, id)
}

// GetNotificationRule returns a notification rule by ID.
func (db *DB) GetNotificationRule(ctx context.Context, id int64) (*models.NotificationRule, error) {
	var r models.NotificationRule
	var enabled int
	err := db.QueryRowContext(ctx,
		"SELECT id, event_type, provider_id, enabled, config, created_at, updated_at FROM notification_rules WHERE id = ?", id,
	).Scan(&r.ID, &r.EventType, &r.ProviderID, &enabled, &r.Config, &r.CreatedAt, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get notification rule: %w", err)
	}
	r.Enabled = enabled == 1
	return &r, nil
}

// ListNotificationRules returns all notification rules.
func (db *DB) ListNotificationRules(ctx context.Context) ([]models.NotificationRule, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, event_type, provider_id, enabled, config, created_at, updated_at FROM notification_rules ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("list notification rules: %w", err)
	}
	defer rows.Close()
	var rules []models.NotificationRule
	for rows.Next() {
		var r models.NotificationRule
		var enabled int
		if err := rows.Scan(&r.ID, &r.EventType, &r.ProviderID, &enabled, &r.Config, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan notification rule: %w", err)
		}
		r.Enabled = enabled == 1
		rules = append(rules, r)
	}
	if rules == nil {
		rules = []models.NotificationRule{}
	}
	return rules, rows.Err()
}

// UpdateNotificationRule updates fields on a notification rule.
func (db *DB) UpdateNotificationRule(ctx context.Context, id int64, enabled *bool, config *string) error {
	if enabled != nil {
		val := 0
		if *enabled {
			val = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE notification_rules SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", val, id); err != nil {
			return fmt.Errorf("update rule enabled: %w", err)
		}
	}
	if config != nil {
		if _, err := db.ExecContext(ctx, "UPDATE notification_rules SET config = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *config, id); err != nil {
			return fmt.Errorf("update rule config: %w", err)
		}
	}
	return nil
}

// DeleteNotificationRule deletes a notification rule by ID.
func (db *DB) DeleteNotificationRule(ctx context.Context, id int64) error {
	if _, err := db.ExecContext(ctx, "DELETE FROM notification_rules WHERE id = ?", id); err != nil {
		return fmt.Errorf("delete notification rule: %w", err)
	}
	return nil
}
