package handlers

import (
	"fmt"

	"github.com/tdebuilt/nidus/internal/models"
)

// convertToYAMLConfig converts the internal export format to the YAML-friendly nested format.
func convertToYAMLConfig(cfg *models.EncryptedExport) models.YAMLConfig {
	// Build a map from category ID to its widgets
	widgetsByCategory := make(map[int64][]models.YAMLWidget)
	for _, w := range cfg.Widgets {
		yw := models.YAMLWidget{
			Type:      w.Type,
			Title:     w.Title,
			Config:    w.Config,
			Collapsed: w.Collapsed,
			PosX:      w.PosX,
			PosY:      w.PosY,
			Width:     w.Width,
			Height:    w.Height,
		}
		widgetsByCategory[w.CategoryID] = append(widgetsByCategory[w.CategoryID], yw)
	}

	categories := make([]models.YAMLCategory, 0, len(cfg.Categories))
	for _, c := range cfg.Categories {
		yc := models.YAMLCategory{
			Name:      c.Name,
			Slug:      c.Slug,
			Icon:      c.Icon,
			SortOrder: c.SortOrder,
			Widgets:   widgetsByCategory[c.ID],
		}
		if yc.Widgets == nil {
			yc.Widgets = []models.YAMLWidget{}
		}
		categories = append(categories, yc)
	}

	services := make([]models.YAMLService, 0, len(cfg.Services))
	for _, s := range cfg.Services {
		services = append(services, models.YAMLService(s))
	}

	return models.YAMLConfig{
		Version: 2,
		Settings: models.YAMLSettings{
			Theme:           cfg.Settings.Theme,
			Language:        cfg.Settings.Language,
			RefreshInterval: cfg.Settings.RefreshInterval,
			AccentColor:     cfg.Settings.AccentColor,
			CustomCSS:       cfg.Settings.CustomCSS,
		},
		Categories: categories,
		Services:   services,
	}
}

// convertFromYAMLConfig converts YAML format (nested widgets) to internal format (flat widgets with category IDs).
// Categories get temporary IDs that ImportConfigFull will remap.
func convertFromYAMLConfig(yamlCfg *models.YAMLConfig) models.EncryptedExport {
	categories := make([]models.Category, 0, len(yamlCfg.Categories))
	var widgets []models.Widget

	for i, yc := range yamlCfg.Categories {
		catID := int64(i + 1) // temporary ID
		categories = append(categories, models.Category{
			ID:        catID,
			Name:      yc.Name,
			Slug:      yc.Slug,
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

	return models.EncryptedExport{
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
}

// validateYAMLConfig validates a YAML configuration before import.
func validateYAMLConfig(cfg *models.YAMLConfig) error {
	if err := validateYAMLCategories(cfg.Categories); err != nil {
		return err
	}
	return validateYAMLServices(cfg.Services)
}

// validateYAMLCategories validates categories and their nested widgets in YAML format.
func validateYAMLCategories(categories []models.YAMLCategory) error {
	if len(categories) > MaxCategories {
		return fmt.Errorf("too many categories: %d (max %d)", len(categories), MaxCategories)
	}

	totalWidgets := 0
	for i, c := range categories {
		if c.Name == "" {
			return fmt.Errorf("category at index %d: name is required", i)
		}
		if len(c.Name) > MaxNameLength {
			return fmt.Errorf("category at index %d: name too long (%d chars, max %d)", i, len(c.Name), MaxNameLength)
		}
		totalWidgets += len(c.Widgets)
		if totalWidgets > MaxWidgets {
			return fmt.Errorf("too many widgets: %d (max %d)", totalWidgets, MaxWidgets)
		}
		for j, w := range c.Widgets {
			if err := validateYAMLWidget(c.Name, j, w); err != nil {
				return err
			}
		}
	}
	return nil
}

// validateYAMLWidget validates a single widget within a YAML category.
func validateYAMLWidget(catName string, idx int, w models.YAMLWidget) error {
	if w.Title == "" {
		return fmt.Errorf("category '%s', widget at index %d: title is required", catName, idx)
	}
	if len(w.Title) > MaxNameLength {
		return fmt.Errorf("category '%s', widget at index %d: title too long (%d chars, max %d)", catName, idx, len(w.Title), MaxNameLength)
	}
	if w.Type == "" {
		return fmt.Errorf("category '%s', widget at index %d: type is required", catName, idx)
	}
	if w.Width < 1 || w.Width > 24 {
		return fmt.Errorf("category '%s', widget at index %d: width must be between 1 and 24", catName, idx)
	}
	if w.Height < 0 || w.Height > MaxWidgetHeight {
		return fmt.Errorf("category '%s', widget at index %d: height must be between 0 and %d", catName, idx, MaxWidgetHeight)
	}
	if w.PosX < 0 || w.PosX >= 24 {
		return fmt.Errorf("category '%s', widget at index %d: pos_x must be between 0 and 23", catName, idx)
	}
	if w.PosY < 0 {
		return fmt.Errorf("category '%s', widget at index %d: pos_y must be >= 0", catName, idx)
	}
	return nil
}

// validateYAMLServices validates service entries in YAML format.
func validateYAMLServices(services []models.YAMLService) error {
	if len(services) > MaxServicesImport {
		return fmt.Errorf("too many services: %d (max %d)", len(services), MaxServicesImport)
	}
	for i, s := range services {
		if s.Type == "" {
			return fmt.Errorf("service at index %d: type is required", i)
		}
		if !ValidServiceTypes[s.Type] {
			return fmt.Errorf("service at index %d: invalid type '%s'", i, s.Type)
		}
		if len(s.Name) > MaxNameLength {
			return fmt.Errorf("service at index %d: name too long (%d chars, max %d)", i, len(s.Name), MaxNameLength)
		}
		if len(s.URL) > MaxURLLength {
			return fmt.Errorf("service at index %d: URL too long (%d chars, max %d)", i, len(s.URL), MaxURLLength)
		}
		reg, ok := ServiceRegistry[s.Type]
		if ok && reg.NeedsURL && s.URL == "" {
			return fmt.Errorf("service at index %d: url is required", i)
		}
	}
	return nil
}
