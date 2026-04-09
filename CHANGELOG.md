# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Fixed
- **Terminal widget shows real SSH error** — replaced the opaque "WebSocket failed" message with structured error frames classifying authentication, network, timeout, host resolution, decryption-key mismatch and handshake failures, rendered in red inside xterm

## [0.3.0-beta.7] - 2026-04-08

### Added
- Responsive columns for Uptime Kuma widget

### Fixed
- **Reolink live stream stuck on frozen snapshot** — Svelte 5 reactive tracking created unintended dependencies in the initialization effect, destroying the MSE player before it could connect
- Camera snapshots showing stale images (blob URL + forced src update)
- Service worker caching snapshot and stream endpoints
- WebSocket connections rejected when accessed through a reverse proxy
- Fullscreen icon not visible on mobile
- Config import failing on unknown service types instead of skipping them

### Infrastructure
- Release workflow auto-creates git tag and GitHub release

## [0.3.0-beta.1] - 2026-04-08

### Added
- **SSH Terminal** — interactive SSH terminal widget with xterm.js, command snippets, and mobile toolbar
- SECURITY.md with vulnerability reporting guidelines
- Improved .dockerignore for smaller build context

### Infrastructure
- Pre-release Docker tags no longer update `latest`
- Desktop builds skipped for pre-release versions

## [0.1.0] - 2026-03-26

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
- **Grafana** — embed panels from Grafana dashboards with multi-panel grid

### UI/UX
- Free-placement drag & drop widget grid with auto-compaction
- 4 built-in themes (dark, light, Nord, Dracula) with accent color picker
- Custom CSS injection
- Kiosk mode with auto-rotation
- Keyboard shortcuts
- Responsive layout (mobile, tablet, desktop, TV)
- Responsive column config per widget (desktop/tablet/mobile)
- 11 languages (fr, en, es, de, pt, it, nl, ru, zh, ja, ar) with RTL support
- Add service dialog — modal with type selection grid
- Alphabetical sort toggle (A→Z / Z→A) for services list
- Widget components lazy-loaded with dynamic imports (code-splitting)

### Infrastructure
- Push notifications via Gotify, Ntfy, and Apprise
- Incoming webhooks with HMAC validation
- OpenAPI/Swagger documentation
- Tauri desktop app (Linux, macOS, Windows)
- Docker image (amd64)
- CI/CD with GitHub Actions (tests, lint, E2E, Docker, desktop builds)
- Embedded go2rtc for camera streaming
- Pre-commit hook with Go lint, frontend lint, and tests
