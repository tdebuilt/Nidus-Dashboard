package main

import (
	"context"
	"flag"
	"fmt"
	"log"
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

	db, err := database.Open(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Migrate(context.Background()); err != nil {
		db.Close()
		log.Fatalf("Failed to run migrations: %v", err)
	}
	defer db.Close()

	// Auto-import dashboard config from config.yaml on first startup
	if cfg.HasDashboardConfig() && db.IsEmpty(context.Background()) {
		encKey, err := db.GetSystemSetting(context.Background(), "encryption_key")
		if err == nil && encKey != "" {
			dashCfg := cfg.DashboardConfig()
			if importErr := db.ImportYAMLConfig(context.Background(), dashCfg, encKey); importErr != nil {
				log.Printf("Warning: failed to auto-import dashboard config from config.yaml: %v", importErr)
			} else {
				log.Println("Dashboard configuration imported from config.yaml")
			}
		}
	}

	// Initialize go2rtc manager if binary is available
	if _, lookErr := exec.LookPath("go2rtc"); lookErr == nil {
		dataDir := filepath.Dir(cfg.Database.Path)
		srv.Go2RTCManager = go2rtc.NewManager(db, dataDir)
		log.Println("go2rtc binary found, streaming manager available")

		// Auto-start go2rtc if cameras are configured
		if err := srv.Go2RTCManager.Start(); err != nil {
			log.Printf("go2rtc auto-start failed: %v", err)
		}
	}

	r := server.NewFromEmbed(srv, cfg, db, web.StaticFS, "static")

	if err := server.Run(srv, cfg, r); err != nil {
		log.Printf("Server error: %v", err)
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
