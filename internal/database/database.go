package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	_ "modernc.org/sqlite"

	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/models"
)

// DB wraps a sql.DB connection with SQLite-specific configuration.
type DB struct {
	*sql.DB
}

// Open creates or opens a SQLite database at the given path.
// It enables WAL mode and foreign keys.
func Open(path string) (*DB, error) {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating database directory: %w", err)
	}

	db, err := sql.Open("sqlite", path+"?_pragma=busy_timeout(10000)&_txlock=immediate")
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// Limit concurrent connections to avoid SQLite locking issues
	db.SetMaxOpenConns(1)

	// Verify connection
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("pinging database: %w", err)
	}

	// Enable WAL mode
	var journalMode string
	if err := db.QueryRow("PRAGMA journal_mode=WAL").Scan(&journalMode); err != nil {
		db.Close()
		return nil, fmt.Errorf("setting journal mode: %w", err)
	}
	if journalMode != "wal" {
		db.Close()
		return nil, fmt.Errorf("expected WAL journal mode, got %s", journalMode)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}
	var fkEnabled int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&fkEnabled); err != nil {
		db.Close()
		return nil, fmt.Errorf("checking foreign keys: %w", err)
	}
	if fkEnabled != 1 {
		db.Close()
		return nil, fmt.Errorf("foreign keys not enabled")
	}

	return &DB{db}, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	return db.DB.Close()
}

// CreateUser inserts a new user into the database with the given role.
func (db *DB) CreateUser(username, passwordHash, role string) (int64, error) {
	if role == "" {
		role = "admin"
	}
	result, err := db.Exec(
		"INSERT INTO users (username, password_hash, role) VALUES (?, ?, ?)",
		username, passwordHash, role,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting user: %w", err)
	}
	return result.LastInsertId()
}

// GetUserByUsername retrieves a user by username, or nil if not found.
func (db *DB) GetUserByUsername(username string) (*models.User, error) {
	u := &models.User{}
	var totpSecret sql.NullString
	var totpEnabled int
	err := db.QueryRow(
		"SELECT id, username, password_hash, role, totp_secret, totp_enabled, token_version, created_at, updated_at FROM users WHERE username = ?",
		username,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &totpSecret, &totpEnabled, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying user: %w", err)
	}
	if totpSecret.Valid {
		u.TOTPSecret = &totpSecret.String
	}
	u.TOTPEnabled = totpEnabled == 1
	return u, nil
}

// CountUsers returns the number of users in the database.
func (db *DB) CountUsers() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	return count, err
}

// GetSystemSetting retrieves a system setting by key, or empty string if not found.
func (db *DB) GetSystemSetting(key string) (string, error) {
	var value string
	err := db.QueryRow(
		"SELECT value FROM system_settings WHERE key = ?", key,
	).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// GetUserByID retrieves a user by ID, or nil if not found.
func (db *DB) GetUserByID(id int64) (*models.User, error) {
	u := &models.User{}
	var totpSecret sql.NullString
	var totpEnabled int
	err := db.QueryRow(
		"SELECT id, username, password_hash, role, totp_secret, totp_enabled, token_version, created_at, updated_at FROM users WHERE id = ?",
		id,
	).Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &totpSecret, &totpEnabled, &u.TokenVersion, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying user: %w", err)
	}
	if totpSecret.Valid {
		u.TOTPSecret = &totpSecret.String
	}
	u.TOTPEnabled = totpEnabled == 1
	return u, nil
}

// SetUserTOTPSecret stores the TOTP secret for a user.
func (db *DB) SetUserTOTPSecret(userID int64, secret string) error {
	_, err := db.Exec(
		"UPDATE users SET totp_secret = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		secret, userID,
	)
	return err
}

// EnableUserTOTP enables TOTP for a user.
func (db *DB) EnableUserTOTP(userID int64) error {
	_, err := db.Exec(
		"UPDATE users SET totp_enabled = 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID,
	)
	return err
}

// DisableUserTOTP disables TOTP and clears the secret for a user.
func (db *DB) DisableUserTOTP(userID int64) error {
	_, err := db.Exec(
		"UPDATE users SET totp_enabled = 0, totp_secret = NULL, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID,
	)
	return err
}

// CreateCategory inserts a new category. Sort order is set to max+1.
func (db *DB) CreateCategory(name, icon string) (*models.Category, error) {
	var maxOrder int
	db.QueryRow("SELECT COALESCE(MAX(sort_order), -1) FROM categories").Scan(&maxOrder)

	slug, err := db.generateUniqueSlug(GenerateSlug(name))
	if err != nil {
		return nil, fmt.Errorf("generating slug: %w", err)
	}

	result, err := db.Exec(
		"INSERT INTO categories (name, icon, sort_order, slug) VALUES (?, ?, ?, ?)",
		name, icon, maxOrder+1, slug,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting category: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetCategory(id)
}

// GetCategory retrieves a category by ID, or nil if not found.
func (db *DB) GetCategory(id int64) (*models.Category, error) {
	c := &models.Category{}
	err := db.QueryRow(
		"SELECT id, name, slug, icon, sort_order, created_at, updated_at FROM categories WHERE id = ?", id,
	).Scan(&c.ID, &c.Name, &c.Slug, &c.Icon, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying category: %w", err)
	}
	return c, nil
}

// GetCategories retrieves all categories ordered by sort_order.
func (db *DB) GetCategories() ([]models.Category, error) {
	rows, err := db.Query("SELECT id, name, slug, icon, sort_order, created_at, updated_at FROM categories ORDER BY sort_order")
	if err != nil {
		return nil, fmt.Errorf("querying categories: %w", err)
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Icon, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning category: %w", err)
		}
		categories = append(categories, c)
	}
	return categories, rows.Err()
}

// UpdateCategory updates a category's name and icon.
func (db *DB) UpdateCategory(id int64, name, icon string) (*models.Category, error) {
	result, err := db.Exec(
		"UPDATE categories SET name = ?, icon = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		name, icon, id,
	)
	if err != nil {
		return nil, fmt.Errorf("updating category: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return db.GetCategory(id)
}

// DeleteCategory deletes a category by ID. Returns true if deleted.
func (db *DB) DeleteCategory(id int64) (bool, error) {
	result, err := db.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("deleting category: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// ReorderCategories updates sort_order for categories based on the given ID order.
func (db *DB) ReorderCategories(ids []int64) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	for i, id := range ids {
		if _, err := tx.Exec(
			"UPDATE categories SET sort_order = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			i, id,
		); err != nil {
			return fmt.Errorf("updating sort order: %w", err)
		}
	}
	return tx.Commit()
}

// CreateWidget inserts a new widget for a category.
func (db *DB) CreateWidget(categoryID int64, wType, title, config string, posX, posY, width, height int) (*models.Widget, error) {
	if config == "" {
		config = "{}"
	}
	if width < 1 {
		width = 1
	}
	if height < 0 {
		height = 0
	}
	result, err := db.Exec(
		"INSERT INTO widgets (category_id, type, title, config, pos_x, pos_y, width, height) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		categoryID, wType, title, config, posX, posY, width, height,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting widget: %w", err)
	}
	id, _ := result.LastInsertId()
	return db.GetWidget(id)
}

// GetWidget retrieves a widget by ID, or nil if not found.
func (db *DB) GetWidget(id int64) (*models.Widget, error) {
	w := &models.Widget{}
	var collapsed int
	err := db.QueryRow(
		"SELECT id, category_id, type, title, config, collapsed, pos_x, pos_y, width, height, created_at, updated_at FROM widgets WHERE id = ?", id,
	).Scan(&w.ID, &w.CategoryID, &w.Type, &w.Title, &w.Config, &collapsed, &w.PosX, &w.PosY, &w.Width, &w.Height, &w.CreatedAt, &w.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying widget: %w", err)
	}
	w.Collapsed = collapsed == 1
	return w, nil
}

// GetWidgetsByCategory retrieves all widgets for a category.
func (db *DB) GetWidgetsByCategory(categoryID int64) ([]models.Widget, error) {
	rows, err := db.Query(
		"SELECT id, category_id, type, title, config, collapsed, pos_x, pos_y, width, height, created_at, updated_at FROM widgets WHERE category_id = ? ORDER BY pos_y, pos_x",
		categoryID,
	)
	if err != nil {
		return nil, fmt.Errorf("querying widgets: %w", err)
	}
	defer rows.Close()

	var widgets []models.Widget
	for rows.Next() {
		var w models.Widget
		var collapsed int
		if err := rows.Scan(&w.ID, &w.CategoryID, &w.Type, &w.Title, &w.Config, &collapsed, &w.PosX, &w.PosY, &w.Width, &w.Height, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning widget: %w", err)
		}
		w.Collapsed = collapsed == 1
		widgets = append(widgets, w)
	}
	return widgets, rows.Err()
}

// UpdateWidget updates a widget's type, title, and config.
func (db *DB) UpdateWidget(id int64, wType, title, config string) (*models.Widget, error) {
	result, err := db.Exec(
		"UPDATE widgets SET type = ?, title = ?, config = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		wType, title, config, id,
	)
	if err != nil {
		return nil, fmt.Errorf("updating widget: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return db.GetWidget(id)
}

// SetWidgetCollapsed updates the collapsed state of a widget.
func (db *DB) SetWidgetCollapsed(id int64, collapsed bool) (*models.Widget, error) {
	val := 0
	if collapsed {
		val = 1
	}
	result, err := db.Exec(
		"UPDATE widgets SET collapsed = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		val, id,
	)
	if err != nil {
		return nil, fmt.Errorf("updating widget collapsed: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return nil, nil
	}
	return db.GetWidget(id)
}

// DeleteWidget deletes a widget by ID. Returns true if deleted.
func (db *DB) DeleteWidget(id int64) (bool, error) {
	result, err := db.Exec("DELETE FROM widgets WHERE id = ?", id)
	if err != nil {
		return false, fmt.Errorf("deleting widget: %w", err)
	}
	rows, _ := result.RowsAffected()
	return rows > 0, nil
}

// SaveWidgetLayout updates positions and sizes for multiple widgets in a transaction.
func (db *DB) SaveWidgetLayout(layouts []models.WidgetLayout) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	for _, l := range layouts {
		if _, err := tx.Exec(
			"UPDATE widgets SET pos_x = ?, pos_y = ?, width = ?, height = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
			l.PosX, l.PosY, l.Width, l.Height, l.ID,
		); err != nil {
			return fmt.Errorf("updating widget layout: %w", err)
		}
	}
	return tx.Commit()
}

// GetServices retrieves all services.
func (db *DB) GetServices() ([]models.Service, error) {
	rows, err := db.Query("SELECT id, type, name, COALESCE(url,''), COALESCE(credentials,''), enabled, config, created_at, updated_at FROM services ORDER BY type")
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
func (db *DB) GetServiceByType(serviceType string) (*models.Service, error) {
	s := &models.Service{}
	var enabled int
	err := db.QueryRow(
		"SELECT id, type, name, COALESCE(url,''), COALESCE(credentials,''), enabled, config, created_at, updated_at FROM services WHERE type = ?",
		serviceType,
	).Scan(&s.ID, &s.Type, &s.Name, &s.URL, &s.Credentials, &enabled, &s.Config, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying service: %w", err)
	}
	s.Enabled = enabled == 1
	return s, nil
}

// UpsertService inserts or updates a service by type.
func (db *DB) UpsertService(serviceType, name, url, credentials, config string, enabled bool) (*models.Service, error) {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	if config == "" {
		config = "{}"
	}
	_, err := db.Exec(`
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
	return db.GetServiceByType(serviceType)
}

// DeleteService deletes a service by type.
func (db *DB) DeleteService(serviceType string) error {
	result, err := db.Exec("DELETE FROM services WHERE type = ?", serviceType)
	if err != nil {
		return fmt.Errorf("deleting service: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("service not found")
	}
	return nil
}

// SetSystemSetting stores a system setting.
func (db *DB) SetSystemSetting(key, value string) error {
	_, err := db.Exec(`
		INSERT INTO system_settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

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

// GetSettings retrieves all user-facing settings with defaults.
func (db *DB) GetSettings() (models.Settings, error) {
	theme, err := db.GetSystemSetting(SettingTheme)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting theme: %w", err)
	}
	if theme == "" {
		theme = DefaultTheme
	}

	lang, err := db.GetSystemSetting(SettingLanguage)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting language: %w", err)
	}
	if lang == "" {
		lang = DefaultLanguage
	}

	refreshStr, err := db.GetSystemSetting(SettingRefreshInterval)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting refresh_interval: %w", err)
	}
	if refreshStr == "" {
		refreshStr = DefaultRefreshInterval
	}
	refresh, err := strconv.Atoi(refreshStr)
	if err != nil {
		refresh, _ = strconv.Atoi(DefaultRefreshInterval)
	}

	accentColor, err := db.GetSystemSetting(SettingAccentColor)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting accent_color: %w", err)
	}

	customCSS, err := db.GetSystemSetting(SettingCustomCSS)
	if err != nil {
		return models.Settings{}, fmt.Errorf("getting custom_css: %w", err)
	}

	kbShortcutsStr, err := db.GetSystemSetting(SettingKeyboardShortcuts)
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

// GetUserPreferences returns preferences for a user, falling back to global defaults.
func (db *DB) GetUserPreferences(userID int64) (*models.UserPreferences, error) {
	globalSettings, err := db.GetSettings()
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

	rows, err := db.Query("SELECT key, value FROM settings WHERE user_id = ?", userID)
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
func (db *DB) SaveUserPreference(userID int64, key, value string) error {
	_, err := db.Exec(
		`INSERT INTO settings (user_id, key, value) VALUES (?, ?, ?)
		 ON CONFLICT(user_id, key) DO UPDATE SET value = excluded.value`,
		userID, key, value,
	)
	return err
}

// DeleteUserPreferences removes all preferences for a user (reset to defaults).
func (db *DB) DeleteUserPreferences(userID int64) error {
	_, err := db.Exec("DELETE FROM settings WHERE user_id = ?", userID)
	return err
}

// GetAllWidgets retrieves all widgets.
func (db *DB) GetAllWidgets() ([]models.Widget, error) {
	rows, err := db.Query(
		"SELECT id, category_id, type, title, config, collapsed, pos_x, pos_y, width, height, created_at, updated_at FROM widgets ORDER BY category_id, pos_y, pos_x",
	)
	if err != nil {
		return nil, fmt.Errorf("querying all widgets: %w", err)
	}
	defer rows.Close()

	var widgets []models.Widget
	for rows.Next() {
		var w models.Widget
		var collapsed int
		if err := rows.Scan(&w.ID, &w.CategoryID, &w.Type, &w.Title, &w.Config, &collapsed, &w.PosX, &w.PosY, &w.Width, &w.Height, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning widget: %w", err)
		}
		w.Collapsed = collapsed == 1
		widgets = append(widgets, w)
	}
	return widgets, rows.Err()
}

// ExportConfig exports the full configuration (categories, widgets, services, settings).
func (db *DB) ExportConfig() (*models.ConfigExport, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("exporting settings: %w", err)
	}

	categories, err := db.GetCategories()
	if err != nil {
		return nil, fmt.Errorf("exporting categories: %w", err)
	}
	if categories == nil {
		categories = []models.Category{}
	}

	widgets, err := db.GetAllWidgets()
	if err != nil {
		return nil, fmt.Errorf("exporting widgets: %w", err)
	}
	if widgets == nil {
		widgets = []models.Widget{}
	}

	rawServices, err := db.GetServices()
	if err != nil {
		return nil, fmt.Errorf("exporting services: %w", err)
	}
	services := make([]models.ServiceResponse, 0, len(rawServices))
	for _, s := range rawServices {
		services = append(services, models.ServiceResponse{
			ID:        s.ID,
			Type:      s.Type,
			Name:      s.Name,
			URL:       s.URL,
			Enabled:   s.Enabled,
			Config:    s.Config,
			HasCreds:  s.Credentials != "",
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
	}

	return &models.ConfigExport{
		Version:    1,
		Settings:   settings,
		Categories: categories,
		Widgets:    widgets,
		Services:   services,
	}, nil
}

// ExportConfigFull exports the full configuration including decrypted credentials.
func (db *DB) ExportConfigFull(encryptionKey string) (*models.EncryptedExport, error) {
	settings, err := db.GetSettings()
	if err != nil {
		return nil, fmt.Errorf("exporting settings: %w", err)
	}

	categories, err := db.GetCategories()
	if err != nil {
		return nil, fmt.Errorf("exporting categories: %w", err)
	}
	if categories == nil {
		categories = []models.Category{}
	}

	widgets, err := db.GetAllWidgets()
	if err != nil {
		return nil, fmt.Errorf("exporting widgets: %w", err)
	}
	if widgets == nil {
		widgets = []models.Widget{}
	}

	rawServices, err := db.GetServices()
	if err != nil {
		return nil, fmt.Errorf("exporting services: %w", err)
	}
	services := make([]models.ServiceExport, 0, len(rawServices))
	for _, s := range rawServices {
		creds := ""
		if s.Credentials != "" {
			decrypted, err := crypto.Decrypt(s.Credentials, encryptionKey)
			if err == nil {
				creds = decrypted
			}
		}
		services = append(services, models.ServiceExport{
			Type:        s.Type,
			Name:        s.Name,
			URL:         s.URL,
			Credentials: creds,
			Enabled:     s.Enabled,
			Config:      s.Config,
		})
	}

	return &models.EncryptedExport{
		Version:    2,
		Settings:   settings,
		Categories: categories,
		Widgets:    widgets,
		Services:   services,
	}, nil
}

// ImportConfigFull imports a full configuration including credentials.
// Credentials are re-encrypted with the system encryption key before storage.
func (db *DB) ImportConfigFull(cfg models.EncryptedExport, encryptionKey string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM widgets"); err != nil {
		return fmt.Errorf("clearing widgets: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM categories"); err != nil {
		return fmt.Errorf("clearing categories: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM services"); err != nil {
		return fmt.Errorf("clearing services: %w", err)
	}

	catIDMap := make(map[int64]int64)
	for _, c := range cfg.Categories {
		slug := c.Slug
		if slug == "" {
			slug = GenerateSlug(c.Name)
		}
		slug, err = generateUniqueSlugTx(tx, slug)
		if err != nil {
			return fmt.Errorf("generating slug for category '%s': %w", c.Name, err)
		}
		result, err := tx.Exec(
			"INSERT INTO categories (name, icon, sort_order, slug) VALUES (?, ?, ?, ?)",
			c.Name, c.Icon, c.SortOrder, slug,
		)
		if err != nil {
			return fmt.Errorf("importing category '%s': %w", c.Name, err)
		}
		newID, _ := result.LastInsertId()
		catIDMap[c.ID] = newID
	}

	for _, w := range cfg.Widgets {
		newCatID, ok := catIDMap[w.CategoryID]
		if !ok {
			continue
		}
		collapsedInt := 0
		if w.Collapsed {
			collapsedInt = 1
		}
		if _, err := tx.Exec(
			"INSERT INTO widgets (category_id, type, title, config, collapsed, pos_x, pos_y, width, height) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			newCatID, w.Type, w.Title, w.Config, collapsedInt, w.PosX, w.PosY, w.Width, w.Height,
		); err != nil {
			return fmt.Errorf("importing widget '%s': %w", w.Title, err)
		}
	}

	for _, s := range cfg.Services {
		enabledInt := 0
		if s.Enabled {
			enabledInt = 1
		}
		config := s.Config
		if config == "" {
			config = "{}"
		}
		encryptedCreds := ""
		if s.Credentials != "" {
			enc, err := crypto.Encrypt(s.Credentials, encryptionKey)
			if err != nil {
				return fmt.Errorf("encrypting credentials for '%s': %w", s.Type, err)
			}
			encryptedCreds = enc
		}
		if _, err := tx.Exec(
			"INSERT INTO services (type, name, url, credentials, enabled, config) VALUES (?, ?, ?, ?, ?, ?)",
			s.Type, s.Name, s.URL, encryptedCreds, enabledInt, config,
		); err != nil {
			return fmt.Errorf("importing service '%s': %w", s.Type, err)
		}
	}

	settingsMap := map[string]string{
		SettingTheme:           cfg.Settings.Theme,
		SettingLanguage:        cfg.Settings.Language,
		SettingRefreshInterval: strconv.Itoa(cfg.Settings.RefreshInterval),
		SettingAccentColor:     cfg.Settings.AccentColor,
		SettingCustomCSS:       cfg.Settings.CustomCSS,
	}
	for key, value := range settingsMap {
		if value == "" {
			continue
		}
		if _, err := tx.Exec(
			"INSERT INTO system_settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
			key, value,
		); err != nil {
			return fmt.Errorf("importing setting '%s': %w", key, err)
		}
	}

	return tx.Commit()
}

// IsEmpty returns true if the database has no categories (fresh install).
func (db *DB) IsEmpty() bool {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	return err != nil || count == 0
}

// ImportYAMLConfig converts a YAMLConfig to the internal format and imports it.
func (db *DB) ImportYAMLConfig(yamlCfg models.YAMLConfig, encryptionKey string) error {
	// Convert YAML categories/widgets to flat format
	categories := make([]models.Category, 0, len(yamlCfg.Categories))
	var widgets []models.Widget

	for i, yc := range yamlCfg.Categories {
		catID := int64(i + 1)
		categories = append(categories, models.Category{
			ID:        catID,
			Name:      yc.Name,
			Icon:      yc.Icon,
			SortOrder: yc.SortOrder,
		})
		for _, yw := range yc.Widgets {
			widgets = append(widgets, models.Widget{
				CategoryID: catID,
				Type:       yw.Type,
				Title:      yw.Title,
				Config:     yw.Config,
				Collapsed:  yw.Collapsed,
				PosX:       yw.PosX,
				PosY:       yw.PosY,
				Width:      yw.Width,
				Height:     yw.Height,
			})
		}
	}

	services := make([]models.ServiceExport, 0, len(yamlCfg.Services))
	for _, ys := range yamlCfg.Services {
		services = append(services, models.ServiceExport(ys))
	}

	encExport := models.EncryptedExport{
		Version: 2,
		Settings: models.Settings{
			Theme:           yamlCfg.Settings.Theme,
			Language:        yamlCfg.Settings.Language,
			RefreshInterval: yamlCfg.Settings.RefreshInterval,
			AccentColor:     yamlCfg.Settings.AccentColor,
			CustomCSS:       yamlCfg.Settings.CustomCSS,
		},
		Categories: categories,
		Widgets:    widgets,
		Services:   services,
	}

	return db.ImportConfigFull(encExport, encryptionKey)
}

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
func (db *DB) SearchWidgets(query string) ([]SearchWidgetResult, error) {
	pattern := "%" + query + "%"
	rows, err := db.Query(
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
func (db *DB) SearchCategories(query string) ([]SearchCategoryResult, error) {
	pattern := "%" + query + "%"
	rows, err := db.Query(
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

// ImportConfig imports a configuration, replacing categories, widgets, services, and settings.
// This runs in a transaction so either everything succeeds or nothing changes.
func (db *DB) ImportConfig(cfg models.ConfigExport) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing data (order matters for foreign keys)
	if _, err := tx.Exec("DELETE FROM widgets"); err != nil {
		return fmt.Errorf("clearing widgets: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM categories"); err != nil {
		return fmt.Errorf("clearing categories: %w", err)
	}
	if _, err := tx.Exec("DELETE FROM services"); err != nil {
		return fmt.Errorf("clearing services: %w", err)
	}

	// Import categories, build old ID -> new ID mapping
	catIDMap := make(map[int64]int64)
	for _, c := range cfg.Categories {
		slug := c.Slug
		if slug == "" {
			slug = GenerateSlug(c.Name)
		}
		slug, err = generateUniqueSlugTx(tx, slug)
		if err != nil {
			return fmt.Errorf("generating slug for category '%s': %w", c.Name, err)
		}
		result, err := tx.Exec(
			"INSERT INTO categories (name, icon, sort_order, slug) VALUES (?, ?, ?, ?)",
			c.Name, c.Icon, c.SortOrder, slug,
		)
		if err != nil {
			return fmt.Errorf("importing category '%s': %w", c.Name, err)
		}
		newID, _ := result.LastInsertId()
		catIDMap[c.ID] = newID
	}

	// Import widgets with remapped category IDs
	for _, w := range cfg.Widgets {
		newCatID, ok := catIDMap[w.CategoryID]
		if !ok {
			continue // skip widgets with unknown category
		}
		collapsedInt := 0
		if w.Collapsed {
			collapsedInt = 1
		}
		if _, err := tx.Exec(
			"INSERT INTO widgets (category_id, type, title, config, collapsed, pos_x, pos_y, width, height) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
			newCatID, w.Type, w.Title, w.Config, collapsedInt, w.PosX, w.PosY, w.Width, w.Height,
		); err != nil {
			return fmt.Errorf("importing widget '%s': %w", w.Title, err)
		}
	}

	// Import services (without credentials — those are not exported)
	for _, s := range cfg.Services {
		enabledInt := 0
		if s.Enabled {
			enabledInt = 1
		}
		config := s.Config
		if config == "" {
			config = "{}"
		}
		if _, err := tx.Exec(
			"INSERT INTO services (type, name, url, credentials, enabled, config) VALUES (?, ?, ?, '', ?, ?)",
			s.Type, s.Name, s.URL, enabledInt, config,
		); err != nil {
			return fmt.Errorf("importing service '%s': %w", s.Type, err)
		}
	}

	// Import settings
	settingsMap := map[string]string{
		SettingTheme:           cfg.Settings.Theme,
		SettingLanguage:        cfg.Settings.Language,
		SettingRefreshInterval: strconv.Itoa(cfg.Settings.RefreshInterval),
		SettingAccentColor:     cfg.Settings.AccentColor,
		SettingCustomCSS:       cfg.Settings.CustomCSS,
	}
	for key, value := range settingsMap {
		if value == "" {
			continue
		}
		if _, err := tx.Exec(
			"INSERT INTO system_settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
			key, value,
		); err != nil {
			return fmt.Errorf("importing setting '%s': %w", key, err)
		}
	}

	return tx.Commit()
}

// SaveSettings updates the given settings. Only non-nil fields are updated.
func (db *DB) SaveSettings(req models.UpdateSettingsRequest) error {
	if req.Theme != nil {
		if err := db.SetSystemSetting(SettingTheme, *req.Theme); err != nil {
			return fmt.Errorf("saving theme: %w", err)
		}
	}
	if req.Language != nil {
		if err := db.SetSystemSetting(SettingLanguage, *req.Language); err != nil {
			return fmt.Errorf("saving language: %w", err)
		}
	}
	if req.RefreshInterval != nil {
		if err := db.SetSystemSetting(SettingRefreshInterval, strconv.Itoa(*req.RefreshInterval)); err != nil {
			return fmt.Errorf("saving refresh_interval: %w", err)
		}
	}
	if req.AccentColor != nil {
		if err := db.SetSystemSetting(SettingAccentColor, *req.AccentColor); err != nil {
			return fmt.Errorf("saving accent_color: %w", err)
		}
	}
	if req.CustomCSS != nil {
		if err := db.SetSystemSetting(SettingCustomCSS, *req.CustomCSS); err != nil {
			return fmt.Errorf("saving custom_css: %w", err)
		}
	}
	if req.EnableKeyboardShortcuts != nil {
		val := "false"
		if *req.EnableKeyboardShortcuts {
			val = "true"
		}
		if err := db.SetSystemSetting(SettingKeyboardShortcuts, val); err != nil {
			return fmt.Errorf("saving keyboard_shortcuts: %w", err)
		}
	}
	return nil
}

// ListUsers retrieves all users (without password hashes).
func (db *DB) ListUsers() ([]models.User, error) {
	rows, err := db.Query(
		"SELECT id, username, role, totp_enabled, created_at, updated_at FROM users ORDER BY id",
	)
	if err != nil {
		return nil, fmt.Errorf("querying users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		var totpEnabled int
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &totpEnabled, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning user: %w", err)
		}
		u.TOTPEnabled = totpEnabled == 1
		users = append(users, u)
	}
	return users, rows.Err()
}

// UpdateUserRole updates a user's role.
func (db *DB) UpdateUserRole(userID int64, role string) error {
	result, err := db.Exec(
		"UPDATE users SET role = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		role, userID,
	)
	if err != nil {
		return fmt.Errorf("updating user role: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// IncrementTokenVersion bumps the token version for a user, invalidating all existing JWTs.
func (db *DB) IncrementTokenVersion(userID int64) error {
	_, err := db.Exec("UPDATE users SET token_version = token_version + 1, updated_at = CURRENT_TIMESTAMP WHERE id = ?", userID)
	return err
}

// DeleteUser deletes a user by ID. Returns an error if user not found.
func (db *DB) DeleteUser(userID int64) error {
	result, err := db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// CreateInvitation creates a new invitation code.
func (db *DB) CreateInvitation(code, role string, createdBy int64, expiresAt string) error {
	_, err := db.Exec(
		"INSERT INTO invitations (code, role, created_by, expires_at) VALUES (?, ?, ?, ?)",
		code, role, createdBy, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("creating invitation: %w", err)
	}
	return nil
}

// GetInvitationByCode retrieves a valid (unused, not expired) invitation by code.
func (db *DB) GetInvitationByCode(code string) (*models.Invitation, error) {
	inv := &models.Invitation{}
	var usedBy sql.NullInt64
	var usedAt sql.NullTime
	err := db.QueryRow(
		"SELECT id, code, role, created_by, used_by, used_at, expires_at, created_at FROM invitations WHERE code = ? AND used_by IS NULL AND expires_at > CURRENT_TIMESTAMP",
		code,
	).Scan(&inv.ID, &inv.Code, &inv.Role, &inv.CreatedBy, &usedBy, &usedAt, &inv.ExpiresAt, &inv.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying invitation: %w", err)
	}
	if usedBy.Valid {
		inv.UsedBy = &usedBy.Int64
	}
	if usedAt.Valid {
		inv.UsedAt = &usedAt.Time
	}
	return inv, nil
}

// UseInvitation marks an invitation as used by a user.
func (db *DB) UseInvitation(invID, userID int64) error {
	_, err := db.Exec(
		"UPDATE invitations SET used_by = ?, used_at = CURRENT_TIMESTAMP WHERE id = ?",
		userID, invID,
	)
	return err
}

// ListInvitations retrieves all invitations created by a user.
func (db *DB) ListInvitations() ([]models.Invitation, error) {
	rows, err := db.Query(
		"SELECT id, code, role, created_by, used_by, used_at, expires_at, created_at FROM invitations ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("querying invitations: %w", err)
	}
	defer rows.Close()

	var invitations []models.Invitation
	for rows.Next() {
		var inv models.Invitation
		var usedBy sql.NullInt64
		var usedAt sql.NullTime
		if err := rows.Scan(&inv.ID, &inv.Code, &inv.Role, &inv.CreatedBy, &usedBy, &usedAt, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning invitation: %w", err)
		}
		if usedBy.Valid {
			inv.UsedBy = &usedBy.Int64
		}
		if usedAt.Valid {
			inv.UsedAt = &usedAt.Time
		}
		invitations = append(invitations, inv)
	}
	return invitations, rows.Err()
}

// DeleteInvitation deletes an invitation by ID.
func (db *DB) DeleteInvitation(invID int64) error {
	_, err := db.Exec("DELETE FROM invitations WHERE id = ?", invID)
	return err
}

// --- Account Management ---

// UpdateUserUsername updates a user's username.
func (db *DB) UpdateUserUsername(userID int64, username string) error {
	result, err := db.Exec(
		"UPDATE users SET username = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		username, userID,
	)
	if err != nil {
		return fmt.Errorf("updating username: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// UpdateUserPassword updates a user's password hash.
func (db *DB) UpdateUserPassword(userID int64, passwordHash string) error {
	result, err := db.Exec(
		"UPDATE users SET password_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		passwordHash, userID,
	)
	if err != nil {
		return fmt.Errorf("updating password: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

// --- Password Resets ---

// CreatePasswordReset creates a new password reset code.
func (db *DB) CreatePasswordReset(userID int64, code string, createdBy int64, expiresAt string) error {
	_, err := db.Exec(
		"INSERT INTO password_resets (user_id, code, created_by, expires_at) VALUES (?, ?, ?, ?)",
		userID, code, createdBy, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("creating password reset: %w", err)
	}
	return nil
}

// GetPasswordResetByCode retrieves a valid (unused, not expired) reset code.
func (db *DB) GetPasswordResetByCode(code string) (*models.PasswordReset, error) {
	reset := &models.PasswordReset{}
	var usedAt sql.NullTime
	err := db.QueryRow(
		"SELECT id, user_id, code, created_by, used_at, expires_at, created_at FROM password_resets WHERE code = ? AND used_at IS NULL AND expires_at > CURRENT_TIMESTAMP",
		code,
	).Scan(&reset.ID, &reset.UserID, &reset.Code, &reset.CreatedBy, &usedAt, &reset.ExpiresAt, &reset.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("querying password reset: %w", err)
	}
	if usedAt.Valid {
		reset.UsedAt = &usedAt.Time
	}
	return reset, nil
}

// UsePasswordReset marks a reset code as used.
func (db *DB) UsePasswordReset(resetID int64) error {
	_, err := db.Exec(
		"UPDATE password_resets SET used_at = CURRENT_TIMESTAMP WHERE id = ?",
		resetID,
	)
	return err
}

// --- Notification Providers ---

// CreateNotificationProvider creates a new notification provider.
func (db *DB) CreateNotificationProvider(providerType, name, url, token, config string) (*models.NotificationProvider, error) {
	if config == "" {
		config = "{}"
	}
	result, err := db.Exec(
		"INSERT INTO notification_providers (type, name, url, token, enabled, config) VALUES (?, ?, ?, ?, 1, ?)",
		providerType, name, url, token, config,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return db.GetNotificationProvider(id)
}

// GetNotificationProvider returns a notification provider by ID.
func (db *DB) GetNotificationProvider(id int64) (*models.NotificationProvider, error) {
	var p models.NotificationProvider
	var enabled int
	err := db.QueryRow(
		"SELECT id, type, name, url, token, enabled, config, created_at, updated_at FROM notification_providers WHERE id = ?", id,
	).Scan(&p.ID, &p.Type, &p.Name, &p.URL, &p.Token, &enabled, &p.Config, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	p.Enabled = enabled == 1
	return &p, nil
}

// ListNotificationProviders returns all notification providers.
func (db *DB) ListNotificationProviders() ([]models.NotificationProvider, error) {
	rows, err := db.Query("SELECT id, type, name, url, token, enabled, config, created_at, updated_at FROM notification_providers ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var providers []models.NotificationProvider
	for rows.Next() {
		var p models.NotificationProvider
		var enabled int
		if err := rows.Scan(&p.ID, &p.Type, &p.Name, &p.URL, &p.Token, &enabled, &p.Config, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, err
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
func (db *DB) UpdateNotificationProvider(id int64, name, url, token *string, enabled *bool, config *string) error {
	if name != nil {
		if _, err := db.Exec("UPDATE notification_providers SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *name, id); err != nil {
			return err
		}
	}
	if url != nil {
		if _, err := db.Exec("UPDATE notification_providers SET url = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *url, id); err != nil {
			return err
		}
	}
	if token != nil {
		if _, err := db.Exec("UPDATE notification_providers SET token = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *token, id); err != nil {
			return err
		}
	}
	if enabled != nil {
		val := 0
		if *enabled {
			val = 1
		}
		if _, err := db.Exec("UPDATE notification_providers SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", val, id); err != nil {
			return err
		}
	}
	if config != nil {
		if _, err := db.Exec("UPDATE notification_providers SET config = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *config, id); err != nil {
			return err
		}
	}
	return nil
}

// DeleteNotificationProvider deletes a notification provider by ID.
func (db *DB) DeleteNotificationProvider(id int64) error {
	_, err := db.Exec("DELETE FROM notification_providers WHERE id = ?", id)
	return err
}

// --- Notification Rules ---

// CreateNotificationRule creates a new notification rule.
func (db *DB) CreateNotificationRule(eventType string, providerID int64, config string) (*models.NotificationRule, error) {
	if config == "" {
		config = "{}"
	}
	result, err := db.Exec(
		"INSERT INTO notification_rules (event_type, provider_id, enabled, config) VALUES (?, ?, 1, ?)",
		eventType, providerID, config,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return db.GetNotificationRule(id)
}

// GetNotificationRule returns a notification rule by ID.
func (db *DB) GetNotificationRule(id int64) (*models.NotificationRule, error) {
	var r models.NotificationRule
	var enabled int
	err := db.QueryRow(
		"SELECT id, event_type, provider_id, enabled, config, created_at, updated_at FROM notification_rules WHERE id = ?", id,
	).Scan(&r.ID, &r.EventType, &r.ProviderID, &enabled, &r.Config, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	r.Enabled = enabled == 1
	return &r, nil
}

// ListNotificationRules returns all notification rules.
func (db *DB) ListNotificationRules() ([]models.NotificationRule, error) {
	rows, err := db.Query("SELECT id, event_type, provider_id, enabled, config, created_at, updated_at FROM notification_rules ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var rules []models.NotificationRule
	for rows.Next() {
		var r models.NotificationRule
		var enabled int
		if err := rows.Scan(&r.ID, &r.EventType, &r.ProviderID, &enabled, &r.Config, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
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
func (db *DB) UpdateNotificationRule(id int64, enabled *bool, config *string) error {
	if enabled != nil {
		val := 0
		if *enabled {
			val = 1
		}
		if _, err := db.Exec("UPDATE notification_rules SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", val, id); err != nil {
			return err
		}
	}
	if config != nil {
		if _, err := db.Exec("UPDATE notification_rules SET config = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *config, id); err != nil {
			return err
		}
	}
	return nil
}

// DeleteNotificationRule deletes a notification rule by ID.
func (db *DB) DeleteNotificationRule(id int64) error {
	_, err := db.Exec("DELETE FROM notification_rules WHERE id = ?", id)
	return err
}

// --- Webhooks ---

// CreateWebhook creates a new webhook and returns it with the plaintext secret.
func (db *DB) CreateWebhook(name, secret string) (*models.Webhook, error) {
	result, err := db.Exec(
		"INSERT INTO webhooks (name, secret, enabled) VALUES (?, ?, 1)",
		name, secret,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return db.GetWebhook(id)
}

// GetWebhook returns a webhook by ID (includes secret for HMAC validation).
func (db *DB) GetWebhook(id int64) (*models.Webhook, error) {
	var w models.Webhook
	var enabled int
	err := db.QueryRow(
		"SELECT id, name, secret, enabled, created_at, updated_at FROM webhooks WHERE id = ?", id,
	).Scan(&w.ID, &w.Name, &w.Secret, &enabled, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		return nil, err
	}
	w.Enabled = enabled == 1
	return &w, nil
}

// ListWebhooks returns all webhooks without secrets.
func (db *DB) ListWebhooks() ([]models.WebhookResponse, error) {
	rows, err := db.Query("SELECT id, name, secret, enabled, created_at FROM webhooks ORDER BY id")
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
func (db *DB) UpdateWebhook(id int64, req models.UpdateWebhookRequest) error {
	if req.Name != nil {
		if _, err := db.Exec("UPDATE webhooks SET name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", *req.Name, id); err != nil {
			return err
		}
	}
	if req.Enabled != nil {
		val := 0
		if *req.Enabled {
			val = 1
		}
		if _, err := db.Exec("UPDATE webhooks SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", val, id); err != nil {
			return err
		}
	}
	return nil
}

// DeleteWebhook deletes a webhook by ID (cascades to actions).
func (db *DB) DeleteWebhook(id int64) error {
	_, err := db.Exec("DELETE FROM webhooks WHERE id = ?", id)
	return err
}

// --- Webhook Actions ---

// ListWebhookActions returns all actions for a webhook.
func (db *DB) ListWebhookActions(webhookID int64) ([]models.WebhookAction, error) {
	rows, err := db.Query("SELECT id, webhook_id, action_type, config FROM webhook_actions WHERE webhook_id = ? ORDER BY id", webhookID)
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
func (db *DB) CreateWebhookAction(webhookID int64, actionType, config string) (*models.WebhookAction, error) {
	if config == "" {
		config = "{}"
	}
	result, err := db.Exec(
		"INSERT INTO webhook_actions (webhook_id, action_type, config) VALUES (?, ?, ?)",
		webhookID, actionType, config,
	)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	var a models.WebhookAction
	err = db.QueryRow("SELECT id, webhook_id, action_type, config FROM webhook_actions WHERE id = ?", id).
		Scan(&a.ID, &a.WebhookID, &a.ActionType, &a.Config)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// DeleteWebhookAction deletes a webhook action by ID.
func (db *DB) DeleteWebhookAction(id int64) error {
	_, err := db.Exec("DELETE FROM webhook_actions WHERE id = ?", id)
	return err
}

// --- Custom Themes ---

// ListCustomThemes returns all custom themes ordered by creation date.
func (db *DB) ListCustomThemes() ([]models.CustomTheme, error) {
	rows, err := db.Query("SELECT id, name, theme_json, created_at, updated_at FROM custom_themes ORDER BY created_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var themes []models.CustomTheme
	for rows.Next() {
		var t models.CustomTheme
		if err := rows.Scan(&t.ID, &t.Name, &t.ThemeJSON, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		themes = append(themes, t)
	}
	if themes == nil {
		themes = []models.CustomTheme{}
	}
	return themes, rows.Err()
}

// GetCustomTheme returns a custom theme by ID.
func (db *DB) GetCustomTheme(id int64) (*models.CustomTheme, error) {
	var t models.CustomTheme
	err := db.QueryRow(
		"SELECT id, name, theme_json, created_at, updated_at FROM custom_themes WHERE id = ?", id,
	).Scan(&t.ID, &t.Name, &t.ThemeJSON, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// CountCustomThemes returns the number of custom themes.
func (db *DB) CountCustomThemes() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM custom_themes").Scan(&count)
	return count, err
}

// CreateCustomTheme creates a new custom theme and returns its ID.
func (db *DB) CreateCustomTheme(name, themeJSON string) (int64, error) {
	result, err := db.Exec(
		"INSERT INTO custom_themes (name, theme_json) VALUES (?, ?)",
		name, themeJSON,
	)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// UpdateCustomTheme updates a custom theme's name and JSON.
func (db *DB) UpdateCustomTheme(id int64, name, themeJSON string) error {
	_, err := db.Exec(
		"UPDATE custom_themes SET name = ?, theme_json = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		name, themeJSON, id,
	)
	return err
}

// DeleteCustomTheme deletes a custom theme by ID.
func (db *DB) DeleteCustomTheme(id int64) error {
	_, err := db.Exec("DELETE FROM custom_themes WHERE id = ?", id)
	return err
}
