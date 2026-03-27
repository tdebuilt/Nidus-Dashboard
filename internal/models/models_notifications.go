package models

import "time"

// NotificationProvider represents a configured notification provider.
type NotificationProvider struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`      // gotify, ntfy, apprise
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Token     string    `json:"token,omitempty"`
	Enabled   bool      `json:"enabled"`
	Config    string    `json:"config"`    // JSON extra config (e.g. topic for ntfy)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NotificationProviderResponse hides sensitive token.
type NotificationProviderResponse struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	HasToken  bool      `json:"has_token"`
	Enabled   bool      `json:"enabled"`
	Config    string    `json:"config"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateNotificationProviderRequest is the payload for creating a notification provider.
type CreateNotificationProviderRequest struct {
	Type   string `json:"type"`
	Name   string `json:"name"`
	URL    string `json:"url"`
	Token  string `json:"token"`
	Config string `json:"config"`
}

// UpdateNotificationProviderRequest is the payload for updating a notification provider.
type UpdateNotificationProviderRequest struct {
	Name    *string `json:"name,omitempty"`
	URL     *string `json:"url,omitempty"`
	Token   *string `json:"token,omitempty"`
	Enabled *bool   `json:"enabled,omitempty"`
	Config  *string `json:"config,omitempty"`
}

// NotificationRule represents a notification rule for an event type.
type NotificationRule struct {
	ID         int64     `json:"id"`
	EventType  string    `json:"event_type"`  // container_down, service_unreachable
	ProviderID int64     `json:"provider_id"`
	Enabled    bool      `json:"enabled"`
	Config     string    `json:"config"`      // JSON thresholds/extra config
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateNotificationRuleRequest is the payload for creating a notification rule.
type CreateNotificationRuleRequest struct {
	EventType  string `json:"event_type"`
	ProviderID int64  `json:"provider_id"`
	Config     string `json:"config"`
}

// UpdateNotificationRuleRequest is the payload for updating a notification rule.
type UpdateNotificationRuleRequest struct {
	Enabled *bool   `json:"enabled,omitempty"`
	Config  *string `json:"config,omitempty"`
}

// TestNotificationRequest is the payload for POST /api/notifications/test.
type TestNotificationRequest struct {
	ProviderID int64 `json:"provider_id"`
}

// ValidProviderTypes contains all valid notification provider types.
var ValidProviderTypes = map[string]bool{
	"gotify":  true,
	"ntfy":    true,
	"apprise": true,
}

// ValidEventTypes contains all valid notification event types.
var ValidEventTypes = map[string]bool{
	"container_down":      true,
	"service_unreachable": true,
}
