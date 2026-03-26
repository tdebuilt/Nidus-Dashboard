package models

import "time"

// User represents a user account.
type User struct {
	ID           int64     `json:"id" example:"1"`
	Username     string    `json:"username" example:"admin"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role" example:"admin"`
	TOTPSecret   *string   `json:"-"`
	TOTPEnabled  bool      `json:"totp_enabled" example:"false"`
	TokenVersion int64     `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Invitation represents a user invitation code.
type Invitation struct {
	ID        int64      `json:"id"`
	Code      string     `json:"code"`
	Role      string     `json:"role"`
	CreatedBy int64      `json:"created_by"`
	UsedBy    *int64     `json:"used_by"`
	UsedAt    *time.Time `json:"used_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// Role constants.
const (
	RoleAdmin  = "admin"
	RoleEditor = "editor"
	RoleViewer = "viewer"
)

// ValidRoles contains all valid role values.
var ValidRoles = []string{RoleAdmin, RoleEditor, RoleViewer}

// IsValidRole checks if a role string is valid.
func IsValidRole(role string) bool {
	for _, r := range ValidRoles {
		if r == role {
			return true
		}
	}
	return false
}

// SetupRequest is the payload for POST /api/auth/setup.
type SetupRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"strongpassword"`
}

// SetupResponse is the response from POST /api/auth/setup.
type SetupResponse struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

// LoginRequest is the payload for POST /api/auth/login.
type LoginRequest struct {
	Username string `json:"username" example:"admin"`
	Password string `json:"password" example:"strongpassword"`
	TOTPCode string `json:"totp_code,omitempty" example:"123456"`
}

// LoginResponse is the response from POST /api/auth/login.
type LoginResponse struct {
	Message string `json:"message" example:"login successful"`
	User    User   `json:"user"`
}

// TOTPGenerateResponse is the response from POST /api/auth/totp/generate.
type TOTPGenerateResponse struct {
	Secret string `json:"secret"`
	URL    string `json:"url"`
	QR     string `json:"qr"`
}

// TOTPEnableRequest is the payload for POST /api/auth/totp/enable.
type TOTPEnableRequest struct {
	Code string `json:"code"`
}

// AuthStatusResponse is the response from GET /api/auth/status.
type AuthStatusResponse struct {
	SetupCompleted bool `json:"setup_completed"`
}

// Category represents a dashboard category for organizing widgets.
type Category struct {
	ID        int64     `json:"id" example:"1"`
	Name      string    `json:"name" example:"Home"`
	Slug      string    `json:"slug" example:"home"`
	Icon      string    `json:"icon" example:"home"`
	SortOrder int       `json:"sort_order" example:"0"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCategoryRequest is the payload for POST /api/categories.
type CreateCategoryRequest struct {
	Name string `json:"name" example:"Media"`
	Icon string `json:"icon" example:"tv"`
}

// UpdateCategoryRequest is the payload for PUT /api/categories/{id}.
type UpdateCategoryRequest struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

// ReorderRequest is the payload for PUT /api/categories/reorder.
type ReorderRequest struct {
	IDs []int64 `json:"ids"`
}

// Widget represents a dashboard widget within a category.
type Widget struct {
	ID         int64     `json:"id" example:"1"`
	CategoryID int64     `json:"category_id" example:"1"`
	Type       string    `json:"type" example:"docker"`
	Title      string    `json:"title" example:"My Docker"`
	Config     string    `json:"config" example:"{}"`
	Collapsed  bool      `json:"collapsed" example:"false"`
	PosX       int       `json:"pos_x" example:"0"`
	PosY       int       `json:"pos_y" example:"0"`
	Width      int       `json:"width" example:"4"`
	Height     int       `json:"height" example:"0"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// ToggleCollapseRequest is the payload for PATCH /api/widgets/{id}/toggle-collapse.
type ToggleCollapseRequest struct {
	Collapsed bool `json:"collapsed"`
}

// CreateWidgetRequest is the payload for POST /api/categories/{id}/widgets.
type CreateWidgetRequest struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Config string `json:"config"`
	PosX   int    `json:"pos_x"`
	PosY   int    `json:"pos_y"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

// UpdateWidgetRequest is the payload for PUT /api/widgets/{id}.
type UpdateWidgetRequest struct {
	Type   string `json:"type"`
	Title  string `json:"title"`
	Config string `json:"config"`
}

// WidgetLayout represents one widget's position and size for layout saving.
type WidgetLayout struct {
	ID     int64 `json:"id"`
	PosX   int   `json:"pos_x"`
	PosY   int   `json:"pos_y"`
	Width  int   `json:"width"`
	Height int   `json:"height"`
}

// SaveLayoutRequest is the payload for PUT /api/widgets/layout.
type SaveLayoutRequest struct {
	Widgets []WidgetLayout `json:"widgets"`
}

// Service represents an external service connection.
type Service struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Credentials string    `json:"-"` // never expose encrypted credentials
	Enabled     bool      `json:"enabled"`
	Config      string    `json:"config"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ServiceResponse is the response for GET /api/services (no secrets).
type ServiceResponse struct {
	ID        int64     `json:"id" example:"1"`
	Type      string    `json:"type" example:"portainer"`
	Name      string    `json:"name" example:"Docker"`
	URL       string    `json:"url" example:"https://portainer.local:9443"`
	Enabled   bool      `json:"enabled" example:"true"`
	Config    string    `json:"config" example:"{}"`
	HasCreds  bool      `json:"has_credentials" example:"true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateServiceRequest is the payload for PUT /api/services/{type}.
type UpdateServiceRequest struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Credentials string `json:"credentials,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
	Config      string `json:"config,omitempty"`
}

// TestServiceResponse is the response from POST /api/services/{type}/test.
type TestServiceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// Settings represents the user-facing application settings.
type Settings struct {
	Theme                   string `json:"theme" example:"dark"`
	Language                string `json:"language" example:"fr"`
	RefreshInterval         int    `json:"refresh_interval" example:"30"`
	AccentColor             string `json:"accent_color" example:"#3b82f6"`
	CustomCSS               string `json:"custom_css" example:""`
	EnableKeyboardShortcuts bool   `json:"enable_keyboard_shortcuts" example:"true"`
}

// UpdateSettingsRequest is the payload for PUT /api/settings.
type UpdateSettingsRequest struct {
	Theme                   *string `json:"theme,omitempty"`
	Language                *string `json:"language,omitempty"`
	RefreshInterval         *int    `json:"refresh_interval,omitempty"`
	AccentColor             *string `json:"accent_color,omitempty"`
	CustomCSS               *string `json:"custom_css,omitempty"`
	EnableKeyboardShortcuts *bool   `json:"enable_keyboard_shortcuts,omitempty"`
}

// UserPreferences represents per-user preferences (no custom_css, that stays global).
type UserPreferences struct {
	Theme                   string `json:"theme" example:"dark"`
	Language                string `json:"language" example:"fr"`
	RefreshInterval         int    `json:"refresh_interval" example:"30"`
	AccentColor             string `json:"accent_color" example:"#3b82f6"`
	EnableKeyboardShortcuts bool   `json:"enable_keyboard_shortcuts" example:"true"`
}

// UpdateUserPreferencesRequest is the payload for PUT /api/user/preferences.
type UpdateUserPreferencesRequest struct {
	Theme                   *string `json:"theme,omitempty"`
	Language                *string `json:"language,omitempty"`
	RefreshInterval         *int    `json:"refresh_interval,omitempty"`
	AccentColor             *string `json:"accent_color,omitempty"`
	EnableKeyboardShortcuts *bool   `json:"enable_keyboard_shortcuts,omitempty"`
}

// ConfigExport represents the full exported configuration.
type ConfigExport struct {
	Version    int               `json:"version"`
	Settings   Settings          `json:"settings"`
	Categories []Category        `json:"categories"`
	Widgets    []Widget          `json:"widgets"`
	Services   []ServiceResponse `json:"services"`
}

// ServiceExport includes plaintext credentials for encrypted backup.
type ServiceExport struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	URL         string `json:"url"`
	Credentials string `json:"credentials"`
	Enabled     bool   `json:"enabled"`
	Config      string `json:"config"`
}

// EncryptedExport represents a full backup including credentials.
type EncryptedExport struct {
	Version    int             `json:"version"`
	Settings   Settings        `json:"settings"`
	Categories []Category      `json:"categories"`
	Widgets    []Widget        `json:"widgets"`
	Services   []ServiceExport `json:"services"`
}

// YAMLConfig represents the full dashboard configuration in YAML format.
// Widgets are nested under their categories for readability.
type YAMLConfig struct {
	Version    int              `yaml:"version" json:"version"`
	Settings   YAMLSettings     `yaml:"settings,omitempty" json:"settings,omitempty"`
	Categories []YAMLCategory   `yaml:"categories,omitempty" json:"categories,omitempty"`
	Services   []YAMLService    `yaml:"services,omitempty" json:"services,omitempty"`
}

// YAMLSettings holds dashboard display settings.
type YAMLSettings struct {
	Theme           string `yaml:"theme,omitempty" json:"theme,omitempty"`
	Language        string `yaml:"language,omitempty" json:"language,omitempty"`
	RefreshInterval int    `yaml:"refresh_interval,omitempty" json:"refresh_interval,omitempty"`
	AccentColor     string `yaml:"accent_color,omitempty" json:"accent_color,omitempty"`
	CustomCSS       string `yaml:"custom_css,omitempty" json:"custom_css,omitempty"`
}

// YAMLCategory represents a category with its widgets nested inside.
type YAMLCategory struct {
	Name      string       `yaml:"name" json:"name"`
	Slug      string       `yaml:"slug,omitempty" json:"slug,omitempty"`
	Icon      string       `yaml:"icon,omitempty" json:"icon,omitempty"`
	SortOrder int          `yaml:"sort_order" json:"sort_order"`
	Widgets   []YAMLWidget `yaml:"widgets,omitempty" json:"widgets,omitempty"`
}

// YAMLWidget represents a widget within a category.
type YAMLWidget struct {
	Type      string `yaml:"type" json:"type"`
	Title     string `yaml:"title" json:"title"`
	Config    string `yaml:"config,omitempty" json:"config,omitempty"`
	Collapsed bool   `yaml:"collapsed,omitempty" json:"collapsed,omitempty"`
	PosX      int    `yaml:"pos_x" json:"pos_x"`
	PosY      int    `yaml:"pos_y" json:"pos_y"`
	Width     int    `yaml:"width" json:"width"`
	Height    int    `yaml:"height" json:"height"`
}

// YAMLService represents an external service connection in YAML.
// Credentials are stored in plaintext — handle with care.
type YAMLService struct {
	Type        string `yaml:"type" json:"type"`
	Name        string `yaml:"name" json:"name"`
	URL         string `yaml:"url" json:"url"`
	Credentials string `yaml:"credentials,omitempty" json:"credentials,omitempty"`
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Config      string `yaml:"config,omitempty" json:"config,omitempty"`
}

// ExportRequest is the payload for POST /api/config/export.
type ExportRequest struct {
	Password string `json:"password"`
}

// ImportRequest is the payload for POST /api/config/import.
type ImportRequest struct {
	Password string `json:"password"`
	Data     string `json:"data"`
}

// UpdateUserRoleRequest is the payload for PUT /api/users/{id}/role.
type UpdateUserRoleRequest struct {
	Role string `json:"role"`
}

// CreateInviteRequest is the payload for POST /api/invites.
type CreateInviteRequest struct {
	Role string `json:"role"`
}

// CreateInviteResponse is the response from POST /api/invites.
type CreateInviteResponse struct {
	Code      string `json:"code"`
	Role      string `json:"role"`
	ExpiresAt string `json:"expires_at"`
}

// RegisterRequest is the payload for POST /api/auth/register (via invite).
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

// UpdateAccountRequest is the payload for PUT /api/auth/account.
type UpdateAccountRequest struct {
	CurrentPassword string  `json:"current_password"`
	Username        *string `json:"username,omitempty"`
	NewPassword     *string `json:"new_password,omitempty"`
}

// UpdateAccountResponse is the response from PUT /api/auth/account.
type UpdateAccountResponse struct {
	Message string `json:"message"`
	User    User   `json:"user"`
}

// PasswordReset represents a password reset code generated by an admin.
type PasswordReset struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	Code      string     `json:"code"`
	CreatedBy int64      `json:"created_by"`
	UsedAt    *time.Time `json:"used_at"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
}

// CreateResetResponse is the response from POST /api/users/{id}/reset.
type CreateResetResponse struct {
	Code      string `json:"code"`
	ExpiresAt string `json:"expires_at"`
}

// ResetPasswordRequest is the payload for POST /api/auth/reset-password.
type ResetPasswordRequest struct {
	Code        string `json:"code"`
	NewPassword string `json:"new_password"`
}

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

// ErrorResponse is a standard error response.
type ErrorResponse struct {
	Error string `json:"error" example:"not found"`
}

// CustomTheme represents a user-created theme stored in the database.
type CustomTheme struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	ThemeJSON string    `json:"theme_json"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateCustomThemeRequest is the payload for POST /api/themes.
type CreateCustomThemeRequest struct {
	Name      string `json:"name"`
	ThemeJSON string `json:"theme_json"`
}

// UpdateCustomThemeRequest is the payload for PUT /api/themes/{id}.
type UpdateCustomThemeRequest struct {
	Name      *string `json:"name,omitempty"`
	ThemeJSON *string `json:"theme_json,omitempty"`
}
