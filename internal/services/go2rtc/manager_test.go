package go2rtc

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/tdebuilt/nidus/internal/database"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")
	db, err := database.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestNewManager(t *testing.T) {
	db := setupTestDB(t)
	dataDir := t.TempDir()

	m := NewManager(db, dataDir)

	if m.db != db {
		t.Error("expected db to be set")
	}
	if m.configPath != filepath.Join(dataDir, "go2rtc.yaml") {
		t.Errorf("expected config path %s, got %s", filepath.Join(dataDir, "go2rtc.yaml"), m.configPath)
	}
}

func TestIsAvailable_NoBinary(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, t.TempDir())
	// Override binary path to empty (simulates missing binary)
	m.binaryPath = ""

	if m.IsAvailable() {
		t.Error("expected IsAvailable() to be false when binary not found")
	}
}

func TestIsAvailable_WithBinary(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, t.TempDir())
	m.binaryPath = "/usr/local/bin/go2rtc"

	if !m.IsAvailable() {
		t.Error("expected IsAvailable() to be true when binary path is set")
	}
}

func TestIsRunning_Default(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, t.TempDir())

	if m.IsRunning() {
		t.Error("expected IsRunning() to be false by default")
	}
}

func TestURL(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, t.TempDir())

	expected := "http://localhost:1984"
	if m.URL() != expected {
		t.Errorf("expected URL %q, got %q", expected, m.URL())
	}
}

func TestStatus_NotRunning(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, t.TempDir())
	m.binaryPath = ""

	status := m.Status()

	if status.Available {
		t.Error("expected Available=false")
	}
	if status.Running {
		t.Error("expected Running=false")
	}
	if status.Uptime != "" {
		t.Errorf("expected empty Uptime, got %q", status.Uptime)
	}
	if status.CameraCount != 0 {
		t.Errorf("expected CameraCount=0, got %d", status.CameraCount)
	}
}

func TestStatus_Available(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, t.TempDir())
	m.binaryPath = "/usr/local/bin/go2rtc"

	status := m.Status()

	if !status.Available {
		t.Error("expected Available=true")
	}
	if status.Running {
		t.Error("expected Running=false")
	}
}

func TestStart_NoBinary(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, t.TempDir())
	m.binaryPath = ""

	err := m.Start()
	if err == nil {
		t.Fatal("expected error when binary not found")
	}
	if err.Error() != "go2rtc binary not found" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestStart_AlreadyRunning(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, t.TempDir())
	m.binaryPath = "/some/path"
	m.running = true

	err := m.Start()
	if err != nil {
		t.Errorf("expected nil when already running, got %v", err)
	}
}

func TestGenerateConfig_Empty(t *testing.T) {
	db := setupTestDB(t)
	dataDir := t.TempDir()
	m := NewManager(db, dataDir)

	err := m.generateConfig()
	if err != nil {
		t.Fatalf("generateConfig: %v", err)
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	content := string(data)
	if len(content) == 0 {
		t.Fatal("expected non-empty config file")
	}

	// Should contain the API listen address
	if !containsString(content, ":1984") {
		t.Error("expected config to contain :1984 listen address")
	}
}

func TestGenerateConfig_WithCameras(t *testing.T) {
	db := setupTestDB(t)
	dataDir := t.TempDir()
	m := NewManager(db, dataDir)

	// Insert a reolink widget with camera config
	cat, err := db.CreateCategory("test", "test")
	catID := cat.ID
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	config := `{"cameras":[{"name":"FrontDoor","ip":"192.168.1.100","username":"admin","password":"pass","channel":0,"source":"direct"}]}`
	_, err = db.CreateWidget(catID, "reolink", "Cameras", config, 0, 0, 6, 4)
	if err != nil {
		t.Fatalf("create widget: %v", err)
	}

	err = m.generateConfig()
	if err != nil {
		t.Fatalf("generateConfig: %v", err)
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	content := string(data)
	if !containsString(content, "FrontDoor") {
		t.Error("expected config to contain camera name 'FrontDoor'")
	}
	if !containsString(content, "rtsp://") {
		t.Error("expected config to contain RTSP URL")
	}
}

func TestGenerateConfig_SkipsHomeAssistant(t *testing.T) {
	db := setupTestDB(t)
	dataDir := t.TempDir()
	m := NewManager(db, dataDir)

	cat, err := db.CreateCategory("test", "test")
	catID := cat.ID
	if err != nil {
		t.Fatalf("create category: %v", err)
	}
	config := `{"cameras":[{"name":"HACam","ip":"","entity_id":"camera.front","channel":0,"source":"homeassistant"}]}`
	_, err = db.CreateWidget(catID, "reolink", "Cameras", config, 0, 0, 6, 4)
	if err != nil {
		t.Fatalf("create widget: %v", err)
	}

	err = m.generateConfig()
	if err != nil {
		t.Fatalf("generateConfig: %v", err)
	}

	data, err := os.ReadFile(m.configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}

	content := string(data)
	if containsString(content, "HACam") {
		t.Error("expected HA cameras to be skipped in go2rtc config")
	}
}

func TestGetCameras_DeduplicatesbyIPAndChannel(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, t.TempDir())

	cat, err := db.CreateCategory("test", "test")
	catID := cat.ID
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	// Two widgets with the same camera
	config := `{"cameras":[{"name":"Cam1","ip":"192.168.1.100","username":"admin","password":"pass","channel":0}]}`
	_, _ = db.CreateWidget(catID, "reolink", "Widget1", config, 0, 0, 6, 4)
	_, _ = db.CreateWidget(catID, "reolink", "Widget2", config, 0, 0, 6, 4)

	cameras, err := m.getCameras()
	if err != nil {
		t.Fatalf("getCameras: %v", err)
	}

	if len(cameras) != 1 {
		t.Errorf("expected 1 camera (deduplicated), got %d", len(cameras))
	}
}

func TestGetCameras_IgnoresNonReolinkWidgets(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, t.TempDir())

	cat, err := db.CreateCategory("test", "test")
	catID := cat.ID
	if err != nil {
		t.Fatalf("create category: %v", err)
	}

	_, _ = db.CreateWidget(catID, "docker", "Docker", `{}`, 0, 0, 6, 4)
	_, _ = db.CreateWidget(catID, "proxmox", "Proxmox", `{}`, 0, 0, 6, 4)

	cameras, err := m.getCameras()
	if err != nil {
		t.Fatalf("getCameras: %v", err)
	}

	if len(cameras) != 0 {
		t.Errorf("expected 0 cameras, got %d", len(cameras))
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		d        time.Duration
		expected string
	}{
		{"seconds", 45 * time.Second, "45s"},
		{"minutes", 5*time.Minute + 30*time.Second, "5m 30s"},
		{"hours", 2*time.Hour + 15*time.Minute, "2h 15m"},
		{"zero", 0, "0s"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.d)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, result, tt.expected)
			}
		})
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
