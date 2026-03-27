package models

import "time"

// Webhook represents an incoming webhook endpoint.
type Webhook struct {
	ID        int64     `json:"id" example:"1"`
	Name      string    `json:"name" example:"GitHub Deploy"`
	Secret    string    `json:"-"`
	Enabled   bool      `json:"enabled" example:"true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// WebhookResponse is the response for webhook listings (no secret).
type WebhookResponse struct {
	ID        int64     `json:"id" example:"1"`
	Name      string    `json:"name" example:"GitHub Deploy"`
	HasSecret bool      `json:"has_secret" example:"true"`
	Enabled   bool      `json:"enabled" example:"true"`
	URL       string    `json:"url" example:"/api/webhooks/1"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateWebhookResponse is returned on creation with the plaintext secret.
type CreateWebhookResponse struct {
	ID     int64  `json:"id" example:"1"`
	Name   string `json:"name" example:"GitHub Deploy"`
	Secret string `json:"secret" example:"a1b2c3d4..."`
	URL    string `json:"url" example:"/api/webhooks/1"`
}

// CreateWebhookRequest is the payload for POST /api/webhooks.
type CreateWebhookRequest struct {
	Name string `json:"name" example:"GitHub Deploy"`
}

// UpdateWebhookRequest is the payload for PUT /api/webhooks/{id}.
type UpdateWebhookRequest struct {
	Name    *string `json:"name,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
}

// WebhookAction represents an action triggered by a webhook.
type WebhookAction struct {
	ID         int64  `json:"id" example:"1"`
	WebhookID  int64  `json:"webhook_id" example:"1"`
	ActionType string `json:"action_type" example:"notify"`
	Config     string `json:"config" example:"{}"`
}

// CreateWebhookActionRequest is the payload for POST /api/webhooks/{id}/actions.
type CreateWebhookActionRequest struct {
	ActionType string `json:"action_type" example:"notify"`
	Config     string `json:"config" example:"{}"`
}

// ValidActionTypes contains all valid webhook action types.
var ValidActionTypes = map[string]bool{
	"notify":           true,
	"refresh_widget":   true,
	"invalidate_cache": true,
}
