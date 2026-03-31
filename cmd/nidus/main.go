package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os/exec"
	"path/filepath"

	"github.com/tdebuilt/nidus/internal/config"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/server"
	"github.com/tdebuilt/nidus/internal/services/go2rtc"
	"github.com/tdebuilt/nidus/web"
)

// @title Nidus Dashboard API
// @version 1.0
// @description Self-hosted dashboard API for Docker, Proxmox, Home Assistant, and more.
// @basePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT token. Format: "Bearer {token}"

// Version is set at build time via -ldflags "-X main.Version=..."
var Version = "dev"

func main() {
	desktop := flag.Bool("desktop", false, "Run in desktop mode (Tauri sidecar)")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println("nidus " + Version)
		return
	}

	srv := server.NewServer(Version)

	cfg, err := config.Load("data/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if *desktop {
		applyDesktopMode(&cfg)
	}

	db, err := initDatabase(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	autoImportConfig(cfg, db)
	initGo2RTC(srv, cfg, db)

	r := server.NewFromEmbed(srv, cfg, db, web.StaticFS, "static")
	if err := server.Run(srv, cfg, r); err != nil {
		slog.Error("server error", "error", err)
	}
}

// initDatabase opens the SQLite database and runs migrations.
func initDatabase(cfg config.Config) (*database.DB, error) {
	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	if err := db.Migrate(context.Background()); err != nil {
		db.Close()
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return db, nil
}

// autoImportConfig imports dashboard config from config.yaml on first startup.
func autoImportConfig(cfg config.Config, db *database.DB) {
	if !cfg.HasDashboardConfig() || !db.IsEmpty(context.Background()) {
		return
	}
	encKey, err := db.GetSystemSetting(context.Background(), "encryption_key")
	if err != nil || encKey == "" {
		return
	}
	dashCfg := cfg.DashboardConfig()
	if err := db.ImportYAMLConfig(context.Background(), dashCfg, encKey); err != nil {
		slog.Warn("failed to auto-import dashboard config from config.yaml", "error", err)
	} else {
		slog.Info("dashboard configuration imported from config.yaml")
	}
}

// initGo2RTC sets up the go2rtc streaming manager if the binary is available.
func initGo2RTC(srv *server.Server, cfg config.Config, db *database.DB) {
	if _, err := exec.LookPath("go2rtc"); err != nil {
		return
	}
	dataDir := filepath.Dir(cfg.Database.Path)
	srv.Go2RTCManager = go2rtc.NewManager(db, dataDir)
	slog.Info("go2rtc binary found, streaming manager available")
	if err := srv.Go2RTCManager.Start(); err != nil {
		slog.Error("go2rtc auto-start failed", "error", err)
	}
}

func applyDesktopMode(cfg *config.Config) {
	cfg.Desktop = true
	cfg.Server.BindAddress = "127.0.0.1"

	// Use OS-specific data directory
	dataDir, err := config.DesktopDataDir()
	if err != nil {
		log.Fatalf("Failed to determine desktop data directory: %v", err)
	}
	cfg.Database.Path = filepath.Join(dataDir, "nidus.db")

	// Find a free port if default is taken
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", cfg.Server.Port))
	if err != nil {
		// Port is busy, find a free one
		ln, err = net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			log.Fatalf("Failed to find free port: %v", err)
		}
		cfg.Server.Port = ln.Addr().(*net.TCPAddr).Port
	}
	ln.Close()

	// Print port for Tauri to read
	fmt.Printf("NIDUS_PORT=%d\n", cfg.Server.Port)
}
