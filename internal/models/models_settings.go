package models

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
