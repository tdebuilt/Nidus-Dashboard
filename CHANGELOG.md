# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- Grafana widget — embed panels from Grafana dashboards with multi-panel grid
- Add service dialog — modal with type selection grid (replaces inline panel)
- Alphabetical sort toggle (A→Z / Z→A) for services list and add service dialog
- Responsive column config per widget (desktop/tablet/mobile)

### Changed
- Services settings page uses flat alphabetical list instead of category grouping
- Widget components are lazy-loaded with dynamic imports (code-splitting)
- Pre-commit hook now runs Go and frontend tests in addition to linting

## [0.1.0] - Initial Release

### Core
- Go backend (Chi router) with SQLite database and embedded Svelte frontend
- JWT authentication with optional TOTP 2FA
- Setup wizard for first-launch configuration
- Dashboard with sidebar navigation, category tabs, and widget grid
- Multi-user support with admin, editor, and viewer roles
- Encrypted config export/import (AES-256-GCM)
- YAML configuration with startup auto-import

### Widgets
- **Docker** — stacks and containers via Portainer API (CE + EE), start/stop/restart/update, CPU/RAM stats
- **Proxmox** — VMs and LXCs monitoring, status, metrics, start/stop
- **Home Assistant** — any entity with real-time actions via WebSocket
- **AdGuard** — DNS query stats, toggle filtering
- **JDownloader** — add links, manage queue via MyJDownloader API
- **Transmission** — add torrents, pause/resume, progress monitoring
- **Uptime Kuma** — status page, monitor cards
- **Plex/Jellyfin** — active sessions, posters, progress tracking
- **Pi-hole** — DNS stats, toggle filtering
- **Sonarr/Radarr/Lidarr/Prowlarr** — calendar, queue, status
- **Weather** — OpenWeatherMap current conditions and forecast
- **Calendar** — iCal feed parser with event display
- **RSS/Atom** — subscribe and display articles
- **System stats** — CPU, RAM, disk usage monitoring
- **App Links** — bookmarks with health checks, favicons, and group support
- **Reolink cameras** — snapshots and live streaming via go2rtc
- **Finance** — stock quotes via Yahoo Finance

### UI/UX
- Free-placement drag & drop widget grid with auto-compaction
- 4 built-in themes (dark, light, Nord, Dracula) with accent color picker
- Custom CSS injection
- Kiosk mode with auto-rotation
- Keyboard shortcuts
- Responsive layout (mobile, tablet, desktop, TV)
- 11 languages (fr, en, es, de, pt, it, nl, ru, zh, ja, ar) with RTL support

### Infrastructure
- Push notifications via Gotify, Ntfy, and Apprise
- Incoming webhooks with HMAC validation
- OpenAPI/Swagger documentation
- Tauri desktop app (Linux, macOS, Windows)
- Docker multi-arch image (amd64 + arm64)
- CI/CD with GitHub Actions (tests, lint, E2E, Docker, desktop builds)
- Embedded go2rtc for camera streaming
