package database

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"github.com/tdebuilt/nidus/internal/models"
)

// Settings keys and defaults.
const (
	SettingTheme           = "setting_theme"
	SettingLanguage         = "setting_language"
	SettingRefreshInterval  = "setting_refresh_interval"
	SettingAccentColor      = "setting_accent_color"
	SettingCustomCSS            = "setting_custom_css"
	SettingKeyboardShortcuts    = "setting_keyboard_shortcuts"

	DefaultTheme               = "dark"
	DefaultLanguage            = "fr"
	DefaultRefreshInterval     = "30"
	DefaultAccentColor         = ""
	DefaultCustomCSS           = ""
	DefaultKeyboardShortcuts   = "true"
)

// GetSystemSetting retrieves a system setting by key, or empty string if not found.
func (db *DB) GetSystemSetting(ctx context.Context, key string) (string, error) {
	var value string
	err := db.QueryRowContext(ctx,
		"SELECT value FROM system_settings WHERE key = ?", key,
	).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetSystemSetting stores a system setting.
func (db *DB) SetSystemSetting(ctx context.Context, key, value string) error {
	_, err := db.ExecContext(ctx, `
		INSERT INTO system_settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

// GetSettings retrieves all user-facing settings with defaults.
func (db *DB) GetSettings(ctx context.Context) (models.Settings, error) {
	theme, err := db.GetSystemSetting(ctx, SettingTheme)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting theme: %w", err)
	}
	if theme == "" {
		theme = DefaultTheme
	}

	lang, err := db.GetSystemSetting(ctx, SettingLanguage)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting language: %w", err)
	}
	if lang == "" {
		lang = DefaultLanguage
	}

	refreshStr, err := db.GetSystemSetting(ctx, SettingRefreshInterval)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting refresh_interval: %w", err)
	}
	if refreshStr == "" {
		refreshStr = DefaultRefreshInterval
	}
	refresh, err := strconv.Atoi(refreshStr)
	if err != nil {
		refresh = 30
	}

	accentColor, err := db.GetSystemSetting(ctx, SettingAccentColor)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting accent_color: %w", err)
	}

	customCSS, err := db.GetSystemSetting(ctx, SettingCustomCSS)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting custom_css: %w", err)
	}

	kbShortcutsStr, err := db.GetSystemSetting(ctx, SettingKeyboardShortcuts)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting keyboard_shortcuts: %w", err)
	}
	if kbShortcutsStr == "" {
		kbShortcutsStr = DefaultKeyboardShortcuts
	}
	kbShortcuts := kbShortcutsStr == "true"

	return models.Settings{
		Theme:                   theme,
		Language:                lang,
		RefreshInterval:         refresh,
		AccentColor:             accentColor,
		CustomCSS:               customCSS,
		EnableKeyboardShortcuts: kbShortcuts,
	}, nil
}

// SaveSettings updates the given settings. Only non-nil fields are updated.
func (db *DB) SaveSettings(ctx context.Context, req models.UpdateSettingsRequest) error {
	if req.Theme != nil {
		if err := db.SetSystemSetting(ctx, SettingTheme, *req.Theme); err != nil {
			return fmt.Errorf("saving theme: %w", err)
		}
	}
	if req.Language != nil {
		if err := db.SetSystemSetting(ctx, SettingLanguage, *req.Language); err != nil {
			return fmt.Errorf("saving language: %w", err)
		}
	}
	if req.RefreshInterval != nil {
		if err := db.SetSystemSetting(ctx, SettingRefreshInterval, strconv.Itoa(*req.RefreshInterval)); err != nil {
			return fmt.Errorf("saving refresh_interval: %w", err)
		}
	}
	if req.AccentColor != nil {
		if err := db.SetSystemSetting(ctx, SettingAccentColor, *req.AccentColor); err != nil {
			return fmt.Errorf("saving accent_color: %w", err)
		}
	}
	if req.CustomCSS != nil {
		if err := db.SetSystemSetting(ctx, SettingCustomCSS, *req.CustomCSS); err != nil {
			return fmt.Errorf("saving custom_css: %w", err)
		}
	}
	if req.EnableKeyboardShortcuts != nil {
		val := "false"
		if *req.EnableKeyboardShortcuts {
			val = "true"
		}
		if err := db.SetSystemSetting(ctx, SettingKeyboardShortcuts, val); err != nil {
			return fmt.Errorf("saving keyboard_shortcuts: %w", err)
		}
	}
	return nil
}

// GetUserPreferences returns preferences for a user, falling back to global defaults.
func (db *DB) GetUserPreferences(ctx context.Context, userID int64) (*models.UserPreferences, error) {
	globalSettings, err := db.GetSettings(ctx)
	if err != nil {
		return nil, err
	}

	prefs := &models.UserPreferences{
		Theme:                   globalSettings.Theme,
		Language:                globalSettings.Language,
		RefreshInterval:         globalSettings.RefreshInterval,
		AccentColor:             globalSettings.AccentColor,
		EnableKeyboardShortcuts: globalSettings.EnableKeyboardShortcuts,
	}

	rows, err := db.QueryContext(ctx, "SELECT key, value FROM settings WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			continue
		}
		switch key {
		case SettingTheme:
			prefs.Theme = value
		case SettingLanguage:
			prefs.Language = value
		case SettingRefreshInterval:
			if v, err := strconv.Atoi(value); err == nil {
				prefs.RefreshInterval = v
			}
		case SettingAccentColor:
			prefs.AccentColor = value
		case SettingKeyboardShortcuts:
			prefs.EnableKeyboardShortcuts = value == "true"
		}
	}

	return prefs, nil
}

// SaveUserPreference saves a single preference for a user (upsert).
func (db *DB) SaveUserPreference(ctx context.Context, userID int64, key, value string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO settings (user_id, key, value) VALUES (?, ?, ?)
		 ON CONFLICT(user_id, key) DO UPDATE SET value = excluded.value`,
		userID, key, value,
	)
	return err
}

// DeleteUserPreferences removes all preferences for a user (reset to defaults).
func (db *DB) DeleteUserPreferences(ctx context.Context, userID int64) error {
	_, err := db.ExecContext(ctx, "DELETE FROM settings WHERE user_id = ?", userID)
	return err
}
