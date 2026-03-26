package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"gopkg.in/yaml.v3"

	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
)

// No hardcoded widget type list — widget types are validated dynamically
// by the frontend widget registry. The backend accepts any non-empty type string.

// ConfigHandler handles config export/import HTTP requests.
type ConfigHandler struct {
	DB *database.DB
}

// Export godoc
// @Summary Export full configuration encrypted with a password
// @Tags config
// @Accept json
// @Produce json
// @Param request body models.ExportRequest true "Export password"
// @Success 200 {object} object "Encrypted configuration data"
// @Failure 400 {object} models.ErrorResponse "Invalid request or missing password"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /config/export [post]
// @Security BearerAuth
func (h *ConfigHandler) Export(w http.ResponseWriter, r *http.Request) {
	var req models.ExportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid JSON"})
		return
	}
	if req.Password == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "password is required"})
		return
	}

	encKey, err := h.DB.GetSystemSetting("encryption_key")
	if err != nil || encKey == "" {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
		return
	}

	cfg, err := h.DB.ExportConfigFull(encKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to export config"})
		return
	}

	jsonData, err := json.Marshal(cfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to serialize config"})
		return
	}

	derivedKey := crypto.DeriveKey(req.Password)
	encrypted, err := crypto.Encrypt(string(jsonData), derivedKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to encrypt config"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"data": encrypted})
}

// Import godoc
// @Summary Import full configuration from encrypted payload
// @Tags config
// @Accept json
// @Produce json
// @Param request body models.ImportRequest true "Encrypted config data and password"
// @Success 200 {object} object "Import success confirmation"
// @Failure 400 {object} models.ErrorResponse "Invalid request, wrong password, or invalid data"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /config/import [post]
// @Security BearerAuth
func (h *ConfigHandler) Import(w http.ResponseWriter, r *http.Request) {
	var req models.ImportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid JSON"})
		return
	}
	if req.Password == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "password is required"})
		return
	}
	if req.Data == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "data is required"})
		return
	}

	derivedKey := crypto.DeriveKey(req.Password)
	decrypted, err := crypto.Decrypt(req.Data, derivedKey)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid password or corrupted file"})
		return
	}

	var cfg models.EncryptedExport
	if err := json.Unmarshal([]byte(decrypted), &cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid config data"})
		return
	}

	if cfg.Version != 2 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "unsupported config version"})
		return
	}

	if err := validateEncryptedImport(&cfg); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	encKey, err := h.DB.GetSystemSetting("encryption_key")
	if err != nil || encKey == "" {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
		return
	}

	if err := h.DB.ImportConfigFull(cfg, encKey); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to import config"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "config imported successfully"})
}

// validateEncryptedImport validates the encrypted import data.
func validateEncryptedImport(cfg *models.EncryptedExport) error {
	if err := validateCategories(cfg.Categories); err != nil {
		return err
	}
	if err := validateWidgets(cfg.Widgets, cfg.Categories); err != nil {
		return err
	}
	if err := validateServiceExports(cfg.Services); err != nil {
		return err
	}
	return nil
}

// validateCategories checks that all categories have valid fields.
func validateCategories(categories []models.Category) error {
	for i, c := range categories {
		if c.Name == "" {
			return fmt.Errorf("category at index %d: name is required", i)
		}
	}
	return nil
}

// validateWidgets checks that all widgets have valid fields and reference existing categories.
func validateWidgets(widgets []models.Widget, categories []models.Category) error {
	catIDs := make(map[int64]bool, len(categories))
	for _, c := range categories {
		catIDs[c.ID] = true
	}

	for i, w := range widgets {
		if w.Title == "" {
			return fmt.Errorf("widget at index %d: title is required", i)
		}
		if w.Type == "" {
			return fmt.Errorf("widget at index %d: type is required", i)
		}
		if !catIDs[w.CategoryID] {
			return fmt.Errorf("widget at index %d: references unknown category_id %d", i, w.CategoryID)
		}
		if w.Width < 1 || w.Width > 24 {
			return fmt.Errorf("widget at index %d: width must be between 1 and 24", i)
		}
		if w.Height < 0 {
			return fmt.Errorf("widget at index %d: height must be >= 0", i)
		}
	}
	return nil
}

// validateServiceExports checks that all service exports have valid fields.
func validateServiceExports(services []models.ServiceExport) error {
	for i, s := range services {
		if s.Type == "" {
			return fmt.Errorf("service at index %d: type is required", i)
		}
		if !ValidServiceTypes[s.Type] {
			return fmt.Errorf("service at index %d: invalid type '%s'", i, s.Type)
		}
		reg, ok := ServiceRegistry[s.Type]
		if ok && reg.NeedsURL && s.URL == "" {
			return fmt.Errorf("service at index %d: url is required", i)
		}
	}
	return nil
}

// ExportYAML godoc
// @Summary Export configuration as a downloadable YAML file
// @Tags config
// @Produce application/x-yaml
// @Success 200 {file} string "YAML configuration file"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /config/yaml [get]
// @Security BearerAuth
func (h *ConfigHandler) ExportYAML(w http.ResponseWriter, r *http.Request) {
	encKey, err := h.DB.GetSystemSetting("encryption_key")
	if err != nil || encKey == "" {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
		return
	}

	fullCfg, err := h.DB.ExportConfigFull(encKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to export config"})
		return
	}

	yamlCfg := convertToYAMLConfig(fullCfg)

	data, err := yaml.Marshal(yamlCfg)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to serialize YAML"})
		return
	}

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", "attachment; filename=nidus-config.yaml")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// ImportYAML godoc
// @Summary Import configuration from YAML payload
// @Tags config
// @Accept application/x-yaml
// @Produce json
// @Param request body models.YAMLConfig true "YAML configuration"
// @Success 200 {object} object "Import success confirmation"
// @Failure 400 {object} models.ErrorResponse "Invalid YAML or validation error"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /config/yaml [post]
// @Security BearerAuth
func (h *ConfigHandler) ImportYAML(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1 MB limit
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "failed to read request body"})
		return
	}

	var yamlCfg models.YAMLConfig
	if err := yaml.Unmarshal(body, &yamlCfg); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid YAML: " + err.Error()})
		return
	}

	if yamlCfg.Version != 2 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "unsupported config version (expected 2)"})
		return
	}

	if err := validateYAMLConfig(&yamlCfg); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	encKey, err := h.DB.GetSystemSetting("encryption_key")
	if err != nil || encKey == "" {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
		return
	}

	encExport := convertFromYAMLConfig(&yamlCfg)

	if err := h.DB.ImportConfigFull(encExport, encKey); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to import config"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "YAML config imported successfully"})
}

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
			Language:         cfg.Settings.Language,
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
	for i, c := range cfg.Categories {
		if c.Name == "" {
			return fmt.Errorf("category at index %d: name is required", i)
		}
		for j, w := range c.Widgets {
			if w.Title == "" {
				return fmt.Errorf("category '%s', widget at index %d: title is required", c.Name, j)
			}
			if w.Type == "" {
				return fmt.Errorf("category '%s', widget at index %d: type is required", c.Name, j)
			}
		}
	}
	for i, s := range cfg.Services {
		if s.Type == "" {
			return fmt.Errorf("service at index %d: type is required", i)
		}
		if !ValidServiceTypes[s.Type] {
			return fmt.Errorf("service at index %d: invalid type '%s'", i, s.Type)
		}
		reg, ok := ServiceRegistry[s.Type]
		if ok && reg.NeedsURL && s.URL == "" {
			return fmt.Errorf("service at index %d: url is required", i)
		}
	}
	return nil
}
