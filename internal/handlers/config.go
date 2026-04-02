package handlers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"gopkg.in/yaml.v3"

	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
)

// No hardcoded widget type list — widget types are validated dynamically
// by the frontend widget registry. The backend accepts any non-empty type string.

// Import validation limits to prevent abuse and DoS.
const (
	MaxCategories     = 100
	MaxWidgets        = 500
	MaxServicesImport = 50
	MaxNameLength     = 255
	MaxURLLength      = 2048
	MaxWidgetHeight   = 500
)

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

	encKey, err := h.DB.GetSystemSetting(r.Context(), "encryption_key")
	if err != nil || encKey == "" {
		slog.Error("config: encryption key not configured for export")
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
		return
	}

	cfg, err := h.DB.ExportConfigFull(r.Context(), encKey)
	if err != nil {
		slog.Error("config: failed to export config", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to export config"})
		return
	}

	result, err := encryptExportPayload(cfg, req.Password)
	if err != nil {
		slog.Error("config: export encryption failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	slog.Info("config: configuration exported")
	writeJSON(w, http.StatusOK, result)
}

// encryptExportPayload serializes, derives a key, and encrypts the config payload.
func encryptExportPayload(cfg *models.EncryptedExport, password string) (map[string]string, error) {
	jsonData, err := json.Marshal(cfg)
	if err != nil {
		return nil, fmt.Errorf("serializing config: %w", err)
	}
	derivedKey, salt, err := crypto.DeriveKeyArgon2(password)
	if err != nil {
		return nil, fmt.Errorf("deriving encryption key: %w", err)
	}
	encrypted, err := crypto.Encrypt(string(jsonData), derivedKey)
	if err != nil {
		return nil, fmt.Errorf("encrypting config: %w", err)
	}
	return map[string]string{
		"data": encrypted,
		"salt": hex.EncodeToString(salt),
		"kdf":  "argon2id",
	}, nil
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

	cfg, err := decryptImportPayload(&req)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: sanitizeError(err)})
		return
	}

	encKey, err := h.DB.GetSystemSetting(r.Context(), "encryption_key")
	if err != nil || encKey == "" {
		slog.Error("config: encryption key not configured for import")
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
		return
	}

	if err := h.DB.ImportConfigFull(r.Context(), *cfg, encKey); err != nil {
		slog.Error("config: failed to import config", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to import config"})
		return
	}

	slog.Info("config: configuration imported")
	writeJSON(w, http.StatusOK, map[string]string{"message": "config imported successfully"})
}

// decryptImportPayload derives the encryption key, decrypts and validates the import payload.
func decryptImportPayload(req *models.ImportRequest) (*models.EncryptedExport, error) {
	var derivedKey string
	if req.KDF == "argon2id" && req.Salt != "" {
		salt, err := hex.DecodeString(req.Salt)
		if err != nil {
			return nil, fmt.Errorf("invalid salt")
		}
		derivedKey = crypto.DeriveKeyWithSalt(req.Password, salt)
	} else {
		derivedKey = crypto.DeriveKey(req.Password) //nolint:staticcheck // backward compat for old exports
	}

	decrypted, err := crypto.Decrypt(req.Data, derivedKey)
	if err != nil {
		return nil, fmt.Errorf("invalid password or corrupted file")
	}

	var cfg models.EncryptedExport
	if err := json.Unmarshal([]byte(decrypted), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config data")
	}

	if cfg.Version != 2 {
		return nil, fmt.Errorf("unsupported config version")
	}

	if err := validateEncryptedImport(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
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
	if len(categories) > MaxCategories {
		return fmt.Errorf("too many categories: %d (max %d)", len(categories), MaxCategories)
	}
	for i, c := range categories {
		if c.Name == "" {
			return fmt.Errorf("category at index %d: name is required", i)
		}
		if len(c.Name) > MaxNameLength {
			return fmt.Errorf("category at index %d: name too long (%d chars, max %d)", i, len(c.Name), MaxNameLength)
		}
	}
	return nil
}

// validateWidgets checks that all widgets have valid fields and reference existing categories.
func validateWidgets(widgets []models.Widget, categories []models.Category) error {
	if len(widgets) > MaxWidgets {
		return fmt.Errorf("too many widgets: %d (max %d)", len(widgets), MaxWidgets)
	}

	catIDs := make(map[int64]bool, len(categories))
	for _, c := range categories {
		catIDs[c.ID] = true
	}

	for i, w := range widgets {
		if w.Title == "" {
			return fmt.Errorf("widget at index %d: title is required", i)
		}
		if len(w.Title) > MaxNameLength {
			return fmt.Errorf("widget at index %d: title too long (%d chars, max %d)", i, len(w.Title), MaxNameLength)
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
		if w.Height < 0 || w.Height > MaxWidgetHeight {
			return fmt.Errorf("widget at index %d: height must be between 0 and %d", i, MaxWidgetHeight)
		}
		if w.PosX < 0 || w.PosX >= 24 {
			return fmt.Errorf("widget at index %d: pos_x must be between 0 and 23", i)
		}
		if w.PosY < 0 {
			return fmt.Errorf("widget at index %d: pos_y must be >= 0", i)
		}
	}
	return nil
}

// validateServiceExports checks that all service exports have valid fields.
func validateServiceExports(services []models.ServiceExport) error {
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

// ExportYAML godoc
// @Summary Export configuration as a downloadable YAML file
// @Tags config
// @Produce application/x-yaml
// @Success 200 {file} string "YAML configuration file"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /config/yaml [get]
// @Security BearerAuth
func (h *ConfigHandler) ExportYAML(w http.ResponseWriter, r *http.Request) {
	encKey, err := h.DB.GetSystemSetting(r.Context(), "encryption_key")
	if err != nil || encKey == "" {
		slog.Error("config: encryption key not configured for YAML export")
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
		return
	}

	fullCfg, err := h.DB.ExportConfigFull(r.Context(), encKey)
	if err != nil {
		slog.Error("config: failed to export config for YAML", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to export config"})
		return
	}

	yamlCfg := convertToYAMLConfig(fullCfg)

	data, err := yaml.Marshal(yamlCfg)
	if err != nil {
		slog.Error("config: failed to serialize YAML", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to serialize YAML"})
		return
	}

	slog.Info("config: YAML configuration exported")

	w.Header().Set("Content-Type", "application/x-yaml")
	w.Header().Set("Content-Disposition", "attachment; filename=nidus-config.yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
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
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid YAML: " + sanitizeError(err)})
		return
	}

	if yamlCfg.Version != 2 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "unsupported config version (expected 2)"})
		return
	}

	if err := validateYAMLConfig(&yamlCfg); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: sanitizeError(err)})
		return
	}

	encKey, err := h.DB.GetSystemSetting(r.Context(), "encryption_key")
	if err != nil || encKey == "" {
		slog.Error("config: encryption key not configured for YAML import")
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
		return
	}

	encExport := convertFromYAMLConfig(&yamlCfg)

	if err := h.DB.ImportConfigFull(r.Context(), encExport, encKey); err != nil {
		slog.Error("config: failed to import YAML config", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to import config"})
		return
	}

	slog.Info("config: YAML configuration imported")
	writeJSON(w, http.StatusOK, map[string]string{"message": "YAML config imported successfully"})
}
