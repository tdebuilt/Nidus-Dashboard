package database

import (
	"context"
	"fmt"
	"strconv"

	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/models"
)

// ExportConfig exports the full configuration (categories, widgets, services, settings).
func (db *DB) ExportConfig(ctx context.Context) (*models.ConfigExport, error) {
	settings, err := db.GetSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("exporting settings: %w", err)
	}

	categories, err := db.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("exporting categories: %w", err)
	}
	if categories == nil {
		categories = []models.Category{}
	}

	widgets, err := db.GetAllWidgets(ctx)
	if err != nil {
		return nil, fmt.Errorf("exporting widgets: %w", err)
	}
	if widgets == nil {
		widgets = []models.Widget{}
	}

	rawServices, err := db.GetServices(ctx)
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
func (db *DB) ExportConfigFull(ctx context.Context, encryptionKey string) (*models.EncryptedExport, error) {
	settings, err := db.GetSettings(ctx)
	if err != nil {
		return nil, fmt.Errorf("exporting settings: %w", err)
	}

	categories, err := db.GetCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("exporting categories: %w", err)
	}
	if categories == nil {
		categories = []models.Category{}
	}

	widgets, err := db.GetAllWidgets(ctx)
	if err != nil {
		return nil, fmt.Errorf("exporting widgets: %w", err)
	}
	if widgets == nil {
		widgets = []models.Widget{}
	}

	rawServices, err := db.GetServices(ctx)
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
func (db *DB) ImportConfigFull(ctx context.Context, cfg models.EncryptedExport, encryptionKey string) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM widgets"); err != nil {
		return fmt.Errorf("clearing widgets: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM categories"); err != nil {
		return fmt.Errorf("clearing categories: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM services"); err != nil {
		return fmt.Errorf("clearing services: %w", err)
	}

	catIDMap := make(map[int64]int64)
	for _, c := range cfg.Categories {
		slug := c.Slug
		if slug == "" {
			slug = GenerateSlug(c.Name)
		}
		slug, err = generateUniqueSlugTx(ctx, tx, slug)
		if err != nil {
			return fmt.Errorf("generating slug for category '%s': %w", c.Name, err)
		}
		result, err := tx.ExecContext(ctx,
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
		if _, err := tx.ExecContext(ctx,
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
		if _, err := tx.ExecContext(ctx,
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
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO system_settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
			key, value,
		); err != nil {
			return fmt.Errorf("importing setting '%s': %w", key, err)
		}
	}

	return tx.Commit()
}

// IsEmpty returns true if the database has no categories (fresh install).
func (db *DB) IsEmpty(ctx context.Context) bool {
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM categories").Scan(&count)
	return err != nil || count == 0
}

// ImportYAMLConfig converts a YAMLConfig to the internal format and imports it.
func (db *DB) ImportYAMLConfig(ctx context.Context, yamlCfg models.YAMLConfig, encryptionKey string) error {
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

	return db.ImportConfigFull(ctx, encExport, encryptionKey)
}

// ImportConfig imports a configuration, replacing categories, widgets, services, and settings.
// This runs in a transaction so either everything succeeds or nothing changes.
func (db *DB) ImportConfig(ctx context.Context, cfg models.ConfigExport) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback()

	// Clear existing data (order matters for foreign keys)
	if _, err := tx.ExecContext(ctx, "DELETE FROM widgets"); err != nil {
		return fmt.Errorf("clearing widgets: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM categories"); err != nil {
		return fmt.Errorf("clearing categories: %w", err)
	}
	if _, err := tx.ExecContext(ctx, "DELETE FROM services"); err != nil {
		return fmt.Errorf("clearing services: %w", err)
	}

	// Import categories, build old ID -> new ID mapping
	catIDMap := make(map[int64]int64)
	for _, c := range cfg.Categories {
		slug := c.Slug
		if slug == "" {
			slug = GenerateSlug(c.Name)
		}
		slug, err = generateUniqueSlugTx(ctx, tx, slug)
		if err != nil {
			return fmt.Errorf("generating slug for category '%s': %w", c.Name, err)
		}
		result, err := tx.ExecContext(ctx,
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
		if _, err := tx.ExecContext(ctx,
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
		if _, err := tx.ExecContext(ctx,
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
		if _, err := tx.ExecContext(ctx,
			"INSERT INTO system_settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value = excluded.value",
			key, value,
		); err != nil {
			return fmt.Errorf("importing setting '%s': %w", key, err)
		}
	}

	return tx.Commit()
}
