# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/), and this project adheres to [Semantic Versioning](https://semver.org/).

## [Unreleased]

### Added
- **Kiosk mode** — auto-rotation between categories and fullscreen display
- **Push notifications** — support for Gotify, Ntfy, and Apprise
- **Multi-user support** — roles admin, editor, and viewer with permissions
- **Weather widget** — OpenWeatherMap API integration with current conditions and forecast
- **Calendar widget** — iCal feed parser with event display
- **RSS/Atom feed widget** — subscribe and display articles from any feed
- **System stats widget** — CPU, RAM, and disk usage monitoring
- YAML configuration export/import with startup auto-import
- Free-placement drag & drop widget grid (replaces swap-based positioning)
- Enhanced bookmarks widget with group support and grid layout
- Torrent file upload support in Transmission widget

### Fixed
- Go syntax error in migrationV7 declaration

## [0.4.1] - 2026-03-22

### Added
- **Uptime Kuma widget** — status page API, monitor cards, 11-language i18n
- **Plex/Jellyfin widget** — active sessions, posters, and progress tracking
- App version display in sidebar footer and settings about section
- Global CPU and RAM usage summary bar in Docker widget
- Stack-level update via Portainer redeploy
- Update check progress indicator in Docker widget

### Fixed
- Docker widget mobile layout (hide container status text on mobile)
- Portainer start bug workaround (remapped to restart)

### Changed
- Hidden individual recreate button for containers in stacks

## [0.4.0] - 2026-03-22

### Added
- **Theme system** — Nord and Dracula built-in themes with visual preview cards
- Theme JSON structure, parser, loader, registry, and auto-apply
- Accent color picker in settings
- Custom CSS injection with textarea editor and DB persistence

### Changed
- Extracted CSS variables into central theme system
- Replaced theme dropdown with visual preview cards

## [0.3.0] - 2026-03-21

### Added
- Widget registry system for simplified widget development
- IconPicker with search and access to all 1700+ Lucide icons
- 9 language translations (es, de, pt, it, nl, ru, zh, ja, ar)
- Auto-detect browser language on first visit
- Dynamic locale auto-discovery with `import.meta.glob`
- i18n validation script to check locale key parity
- MIT license
- CONTRIBUTING.md with widget development guide
- ROADMAP.md and detailed task breakdown
- Screenshots in README
- GitHub issue/PR templates, code of conduct, and security policy

### Changed
- Refactored widget rendering from if/else chains to dynamic registry lookup
- Refactored backend ServiceRegistry to consolidate duplicated maps
- Made i18n locale system dynamic with registry

## [0.2.0] - 2026-03-21

### Added
- **Edit mode** — lock/unlock the dashboard to prevent accidental changes
  - Toggle button in the header (pencil icon)
  - When locked: drag, resize, rename, delete, collapse, and add buttons are hidden
  - Category tab management also respects edit mode
- Nidus logo in the header bar

## [0.1.4] - 2026-03-20

### Added
- Edit mode toggle for the dashboard
- Logo displayed in the header bar

## [0.1.3] - 2026-03-20

### Fixed
- JDownloader cleanup API call format (positional params instead of single object)
- Logo not visible in mobile sidebar (wrong component file)
- PWA caching preventing updates (switched to network-first for HTML, bumped cache version)

## [0.1.2] - 2026-03-20

### Added
- CPU and RAM consumption stats on Docker container and stack widgets
- Cleanup button on JDownloader widget to remove finished downloads
- Hexagonal Nidus logo throughout the app (favicon, PWA icons, sidebar, login, setup)
- PNG icons for PWA (192x192, 512x512) generated from SVG

### Changed
- Removed service configuration steps from the onboarding wizard (moved to settings)

## [0.1.1] - 2026-03-19

### Added
- 12-column responsive grid layout for widgets
- Drag-and-swap widget repositioning with pointer events
- Inline category tab management (rename, reorder, delete directly in the dashboard)
- Home Assistant widget config improvements (entity picker, camera resize)
- Mobile sidebar enhancements
- Category icon picker with DynamicIcon support

### Fixed
- Cookie Secure flag for HTTP connections (non-HTTPS setups)
- Frontend test failures after dashboard revamp

## [0.1.0] - 2026-03-18

### Added
- **Backend**: Go (Chi router) with SQLite database and embedded frontend
- **Authentication**: User/password with optional TOTP 2FA
- **Setup wizard**: Guided first-launch configuration
- **Dashboard**: Sidebar navigation, category tabs, widget grid
- **Docker widget**: View stacks and containers via Portainer API (CE + EE), start/stop/restart/update
- **Proxmox widget**: Monitor VMs and LXCs, status, metrics, start/stop
- **Home Assistant widget**: Any entity as a widget with real-time actions via WebSocket
- **AdGuard widget**: DNS query stats, toggle filtering on/off
- **JDownloader widget**: Add links, manage download queue via MyJDownloader API
- **Transmission widget**: Add torrents, pause/resume, monitor progress
- **App Links widget**: Custom bookmarks with automatic health checks and favicon detection
- **Widget system**: Resizable grid, drag-and-drop, collapse, inline rename, typed config forms
- **Dark / light theme**: Toggle in sidebar
- **Responsive design**: Desktop sidebar, mobile burger menu
- **i18n**: French and English
- **Config export/import**: Encrypted backup and restore (AES-256-GCM)
- **Help page**: Built-in widget guide
- **Tauri desktop app**: Native installers for Linux, macOS, Windows
- **CI/CD**: GitHub Actions for build, test, Docker image (GHCR), multi-platform binaries, Tauri builds
- **Deployment docs**: Docker, standalone binary, desktop app
