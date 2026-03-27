package models

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
	Salt     string `json:"salt,omitempty"`
	KDF      string `json:"kdf,omitempty"`
}
