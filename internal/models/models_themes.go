package models

import "time"

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
