package go2rtc

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/services/reolink"
	"gopkg.in/yaml.v3"
)

// go2rtcConfig represents the go2rtc YAML configuration structure.
type go2rtcConfig struct {
	API     go2rtcAPI           `yaml:"api"`
	Streams map[string][]string `yaml:"streams,omitempty"`
}

// go2rtcAPI represents the API section of the go2rtc configuration.
type go2rtcAPI struct {
	Listen string `yaml:"listen"`
	Origin string `yaml:"origin"`
}

// StatusInfo holds the current state of the go2rtc subprocess.
type StatusInfo struct {
	Available   bool   `json:"available"`
	Running     bool   `json:"running"`
	Uptime      string `json:"uptime,omitempty"`
	CameraCount int    `json:"cameras"`
}

// Manager manages the embedded go2rtc subprocess lifecycle.
type Manager struct {
	db         *database.DB
	binaryPath string
	configPath string

	mu        sync.Mutex
	cmd       *exec.Cmd
	running   bool
	startedAt time.Time
	done      chan struct{}

	// debounce reload
	reloadTimer *time.Timer
}

// NewManager creates a new go2rtc manager.
func NewManager(db *database.DB, dataDir string) *Manager {
	binaryPath, _ := exec.LookPath("go2rtc")
	return &Manager{
		db:         db,
		binaryPath: binaryPath,
		configPath: filepath.Join(dataDir, "go2rtc.yaml"),
	}
}

// IsAvailable returns true if the go2rtc binary is found.
func (m *Manager) IsAvailable() bool {
	return m.binaryPath != ""
}

// IsRunning returns true if go2rtc is currently running.
func (m *Manager) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

// URL returns the go2rtc API base URL.
func (m *Manager) URL() string {
	return "http://localhost:1984"
}

// Status returns the current go2rtc status.
func (m *Manager) Status() StatusInfo {
	m.mu.Lock()
	defer m.mu.Unlock()

	info := StatusInfo{
		Available: m.binaryPath != "",
		Running:   m.running,
	}

	if m.running {
		info.Uptime = formatDuration(time.Since(m.startedAt))
	}

	count, _ := m.countCameras()
	info.CameraCount = count

	return info
}

// Start launches the go2rtc subprocess.
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.running {
		return nil
	}
	if m.binaryPath == "" {
		return fmt.Errorf("go2rtc binary not found")
	}

	if err := m.generateConfig(); err != nil {
		return fmt.Errorf("generate config: %w", err)
	}

	return m.startLocked()
}

// Stop gracefully stops the go2rtc subprocess.
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stopLocked()
}

// Restart stops then starts go2rtc.
func (m *Manager) Restart() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stopLocked()

	if err := m.generateConfig(); err != nil {
		return fmt.Errorf("generate config: %w", err)
	}

	return m.startLocked()
}

// Reload regenerates config and restarts go2rtc if running. Debounced by 500ms.
func (m *Manager) Reload() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	if m.reloadTimer != nil {
		m.reloadTimer.Stop()
	}
	m.reloadTimer = time.AfterFunc(500*time.Millisecond, func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		if !m.running {
			return
		}

		if err := m.generateConfig(); err != nil {
			log.Printf("[go2rtc] reload config error: %v", err)
			return
		}

		m.stopLocked()
		if err := m.startLocked(); err != nil {
			log.Printf("[go2rtc] restart after reload error: %v", err)
		} else {
			log.Printf("[go2rtc] reloaded with updated camera config")
		}
	})
}

func (m *Manager) startLocked() error {
	m.done = make(chan struct{})
	m.cmd = exec.Command(m.binaryPath, "-config", m.configPath)
	m.cmd.Stdout = log.Writer()
	m.cmd.Stderr = log.Writer()

	if err := m.cmd.Start(); err != nil {
		return fmt.Errorf("start go2rtc: %w", err)
	}

	m.running = true
	m.startedAt = time.Now()
	log.Printf("[go2rtc] started (pid %d)", m.cmd.Process.Pid)

	// Monitor goroutine: restart on unexpected exit
	done := m.done
	go func() {
		err := m.cmd.Wait()
		m.mu.Lock()
		defer m.mu.Unlock()

		select {
		case <-done:
			// Intentional stop, don't restart
			m.running = false
			return
		default:
		}

		log.Printf("[go2rtc] exited unexpectedly: %v, restarting in 2s", err)
		m.running = false

		time.Sleep(2 * time.Second)

		select {
		case <-done:
			return
		default:
		}

		if err := m.startLocked(); err != nil {
			log.Printf("[go2rtc] restart failed: %v", err)
		}
	}()

	return nil
}

func (m *Manager) stopLocked() {
	if !m.running || m.cmd == nil || m.cmd.Process == nil {
		m.running = false
		return
	}

	close(m.done)

	_ = m.cmd.Process.Signal(syscall.SIGTERM)

	exited := make(chan struct{})
	go func() {
		m.cmd.Wait()
		close(exited)
	}()

	select {
	case <-exited:
	case <-time.After(5 * time.Second):
		_ = m.cmd.Process.Kill()
		<-exited
	}

	m.running = false
	m.cmd = nil
	log.Printf("[go2rtc] stopped")

	// Wait for TCP ports to be released before restarting
	m.waitForPortFree(":1984", 10*time.Second)
}

// waitForPortFree blocks until the given address is available or timeout expires.
func (m *Manager) waitForPortFree(addr string, timeout time.Duration) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		ln, err := net.Listen("tcp", addr)
		if err == nil {
			ln.Close()
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Printf("[go2rtc] warning: port %s still in use after %v", addr, timeout)
}

func (m *Manager) generateConfig() error {
	cameras, err := m.getCameras()
	if err != nil {
		return err
	}

	streams := make(map[string][]string)
	for _, cam := range cameras {
		if cam.Source != "direct" && cam.Source != "" {
			continue
		}
		url := reolink.FormatRTSPURL(cam.Username, cam.Password, cam.IP, cam.Channel, "main")
		streams[cam.Name] = []string{url}
	}

	cfg := go2rtcConfig{
		API: go2rtcAPI{
			Listen: ":1984",
			Origin: "*",
		},
		Streams: streams,
	}

	data, err := yaml.Marshal(&cfg)
	if err != nil {
		return fmt.Errorf("marshal go2rtc config: %w", err)
	}

	return os.WriteFile(m.configPath, data, 0600)
}

func (m *Manager) getCameras() ([]reolink.CameraEntry, error) {
	widgets, err := m.db.GetAllWidgets()
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var cameras []reolink.CameraEntry

	for _, w := range widgets {
		if w.Type != "reolink" || w.Config == "" {
			continue
		}
		var wc struct {
			Cameras []reolink.CameraEntry `json:"cameras"`
		}
		if err := json.Unmarshal([]byte(w.Config), &wc); err != nil {
			continue
		}
		for _, cam := range wc.Cameras {
			key := fmt.Sprintf("%s:%d", cam.IP, cam.Channel)
			if seen[key] {
				continue
			}
			seen[key] = true
			cameras = append(cameras, cam)
		}
	}

	return cameras, nil
}

func (m *Manager) countCameras() (int, error) {
	cameras, err := m.getCameras()
	if err != nil {
		return 0, err
	}
	return len(cameras), nil
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60

	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}
