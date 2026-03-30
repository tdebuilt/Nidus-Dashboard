package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/tdebuilt/nidus/internal/models"
)

// CreateWebhook creates a new webhook and returns it with the plaintext secret.
func (db *DB) CreateWebhook(ctx context.Context, name, secret string) (*models.Webhook, error) {
	result, err := db.ExecContext(ctx,
		"INSERT INTO webhooks (name, secret, enabled) VALUES (?, ?, 1)",
		name, secret,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return db.GetWebhook(ctx, id)
}

// GetWebhook returns a webhook by ID (includes secret for HMAC validation).
func (db *DB) GetWebhook(ctx context.Context, id int64) (*models.Webhook, error) {
	var w models.Webhook
	var enabled int
	err := db.QueryRowContext(ctx,
		"SELECT id, name, secret, enabled, created_at, updated_at FROM webhooks WHERE id = ?", id,
	).Scan(&w.ID, &w.Name, &w.Secret, &enabled, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	w.Enabled = enabled == 1
	return &w, nil
}

// ListWebhooks returns all webhooks without secrets.
func (db *DB) ListWebhooks(ctx context.Context) ([]models.WebhookResponse, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name, secret, enabled, created_at FROM webhooks ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var webhooks []models.WebhookResponse
	for rows.Next() {
		var w models.WebhookResponse
		var secret string
		var enabled int
		if err := rows.Scan(&w.ID, &w.Name, &secret, &enabled, &w.CreatedAt); err != nil {
			return nil, err
		}
		w.Enabled = enabled == 1
		w.HasSecret = secret != ""
		w.URL = fmt.Sprintf("/api/webhooks/%d", w.ID)
		webhooks = append(webhooks, w)
	}
	if webhooks == nil {
		webhooks = []models.WebhookResponse{}
	}
	return webhooks, rows.Err()
}

// UpdateWebhook updates fields on a webhook.
func (db *DB) UpdateWebhook(ctx context.Context, id int64, req models.UpdateWebhookRequest) error {
	if req.Name != nil {
		if _, err := db.ExecContext(ctx, "UPDATE webhooks SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *req.Name, id); err != nil {
			return err
		}
	}
	if req.Enabled != nil {
		val := 0
		if *req.Enabled {
			val = 1
		}
		if _, err := db.ExecContext(ctx, "UPDATE webhooks SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", val, id); err != nil {
			return err
		}
	}
	return nil
}

// DeleteWebhook deletes a webhook by ID (cascades to actions).
func (db *DB) DeleteWebhook(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, "DELETE FROM webhooks WHERE id = ?", id)
	return err
}

// ListWebhookActions returns all actions for a webhook.
func (db *DB) ListWebhookActions(ctx context.Context, webhookID int64) ([]models.WebhookAction, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, webhook_id, action_type, config FROM webhook_actions WHERE webhook_id = ? ORDER BY id", webhookID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var actions []models.WebhookAction
	for rows.Next() {
		var a models.WebhookAction
		if err := rows.Scan(&a.ID, &a.WebhookID, &a.ActionType, &a.Config); err != nil {
			return nil, err
		}
		actions = append(actions, a)
	}
	if actions == nil {
		actions = []models.WebhookAction{}
	}
	return actions, rows.Err()
}

// CreateWebhookAction creates a new action for a webhook.
func (db *DB) CreateWebhookAction(ctx context.Context, webhookID int64, actionType, config string) (*models.WebhookAction, error) {
	if config == "" {
		config = "{}"
	}
	result, err := db.ExecContext(ctx,
		"INSERT INTO webhook_actions (webhook_id, action_type, config) VALUES (?, ?, ?)",
		webhookID, actionType, config,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	var a models.WebhookAction
	err = db.QueryRowContext(ctx, "SELECT id, webhook_id, action_type, config FROM webhook_actions WHERE id = ?", id).
		Scan(&a.ID, &a.WebhookID, &a.ActionType, &a.Config)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// DeleteWebhookAction deletes a webhook action by ID.
func (db *DB) DeleteWebhookAction(ctx context.Context, id int64) error {
	_, err := db.ExecContext(ctx, "DELETE FROM webhook_actions WHERE id = ?", id)
	return err
}
