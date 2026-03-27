package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	t.Parallel()
	// Load from a non-existent file should return defaults
	cfg, err := Load("/tmp/nidus_test_nonexistent.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 3777 {
		t.Errorf("expected default port 3777, got %d", cfg.Server.Port)
	}
	if cfg.Server.BaseURL != "http://localhost:3777" {
		t.Errorf("expected default base_url, got %s", cfg.Server.BaseURL)
	}
	if cfg.Database.Path != "./data/nidus.db" {
		t.Errorf("expected default db path, got %s", cfg.Database.Path)
	}
}

func TestLoadFromYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := []byte(`server:
  port: 9090
  base_url: "http://example.com"
database:
  path: "/tmp/test.db"
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Server.Port)
	}
	if cfg.Server.BaseURL != "http://example.com" {
		t.Errorf("expected base_url http://example.com, got %s", cfg.Server.BaseURL)
	}
	if cfg.Database.Path != "/tmp/test.db" {
		t.Errorf("expected db path /tmp/test.db, got %s", cfg.Database.Path)
	}
}

func TestLoadPartialYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	// Only override port, rest should stay default
	content := []byte(`server:
  port: 3000
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 3000 {
		t.Errorf("expected port 3000, got %d", cfg.Server.Port)
	}
	if cfg.Database.Path != "./data/nidus.db" {
		t.Errorf("expected default db path, got %s", cfg.Database.Path)
	}
}

func TestEnvOverrides(t *testing.T) {
	t.Setenv("NIDUS_PORT", "4000")
	t.Setenv("NIDUS_BASE_URL", "http://env.local")
	t.Setenv("NIDUS_DB_PATH", "/env/path.db")

	cfg, err := Load("/tmp/nidus_test_nonexistent.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 4000 {
		t.Errorf("expected port 4000 from env, got %d", cfg.Server.Port)
	}
	if cfg.Server.BaseURL != "http://env.local" {
		t.Errorf("expected base_url from env, got %s", cfg.Server.BaseURL)
	}
	if cfg.Database.Path != "/env/path.db" {
		t.Errorf("expected db path from env, got %s", cfg.Database.Path)
	}
}

func TestEnvOverridesYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := []byte(`server:
  port: 9090
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Env should override YAML value
	t.Setenv("NIDUS_PORT", "5555")

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 5555 {
		t.Errorf("expected env port 5555 to override yaml 9090, got %d", cfg.Server.Port)
	}
}

func TestValidateInvalidPort(t *testing.T) {
	tests := []struct {
		name string
		port string
	}{
		{"zero", "0"},
		{"negative", "-1"},
		{"too high", "70000"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NIDUS_PORT", tt.port)
			_, err := Load("/tmp/nidus_test_nonexistent.yaml")
			if err == nil {
				t.Error("expected validation error for invalid port")
			}
		})
	}
}

func TestValidateEmptyDBPath(t *testing.T) {
	t.Setenv("NIDUS_DB_PATH", "")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	content := []byte(`database:
  path: ""
`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected validation error for empty db path")
	}
}

func TestInvalidYAML(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	content := []byte(`server: [invalid yaml structure`)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestInvalidEnvPortIgnored(t *testing.T) {
	// Non-numeric port env var should be ignored, keeping default
	t.Setenv("NIDUS_PORT", "notanumber")

	cfg, err := Load("/tmp/nidus_test_nonexistent.yaml")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Server.Port != 3777 {
		t.Errorf("expected default port 3777 when env is invalid, got %d", cfg.Server.Port)
	}
}
