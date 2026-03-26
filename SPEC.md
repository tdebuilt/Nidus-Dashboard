# Nidus — Specification

## Overview

Nidus is a lightweight, self-hosted dashboard for monitoring and managing Docker containers, Proxmox VMs/LXCs, and self-hosted services. It runs on low-spec servers (2GB RAM) and provides both visibility and control from a single interface. Even with embedded camera streaming (go2rtc), the whole stack stays under 70 MB of RAM.

**Target user:** Homelab admins managing Proxmox, Docker, and various self-hosted services.

---

## Architecture

### Stack

| Component | Technology | Notes |
|---|---|---|
| Backend | Go (Chi router) | Single binary, ~15-25MB RAM |
| Frontend | Svelte 5 | Compiled to static, embedded in Go binary |
| Database | SQLite (WAL mode) | Single file, zero config |
| Streaming | go2rtc (embedded) | RTSP → WebRTC/MSE for cameras |
| Desktop | Tauri | Native app for Linux, macOS, Windows |
| Deployment | Docker Compose | Multi-arch (amd64 + arm64) |

### No Agent

Nidus does **not** deploy agents on remote machines. It connects to:
- **Portainer API** for Docker management (not Docker API directly)
- **Proxmox API** for VM/LXC management
- **Service APIs** directly (Home Assistant, AdGuard, Uptime Kuma, Plex/Jellyfin, etc.)

### High-Level Architecture

```
┌─────────────────────────────────────────────────┐
│                   Nidus                          │
│                                                  │
│  ┌──────────┐  ┌──────────┐  ┌───────────────┐  │
│  │ Svelte   │  │ Go API   │  │ SQLite        │  │
│  │ Frontend │◄─┤ Backend  ├──┤ (config, auth, │  │
│  │ (static) │  │ (REST+WS)│  │  layout, i18n)│  │
│  └──────────┘  └────┬─────┘  └───────────────┘  │
│                     │                            │
│  ┌──────────────────┴──────────────────────┐     │
│  │ go2rtc (embedded subprocess)            │     │
│  │ RTSP → WebRTC/MSE streaming             │     │
│  └─────────────────────────────────────────┘     │
└─────────────────────┼────────────────────────────┘
                      │
        ┌─────────────┼─────────────────┐
        │             │                 │
   ┌────▼────┐  ┌─────▼─────┐  ┌───────▼───────┐
   │Portainer│  │ Proxmox   │  │ Services      │
   │ API     │  │ API       │  │ (HA, AdGuard, │
   │(CE + EE)│  │           │  │  UptimeKuma…) │
   └─────────┘  └───────────┘  └───────────────┘
```

---

## Authentication

### Multi-User with Roles

- **Admin** — full access (settings, services, users, dashboard)
- **Editor** — modify dashboard (widgets, categories) but not settings/services
- **Viewer** — read-only (edit mode hidden)
- User invitation system (link or code)
- Username + password (bcrypt hashed)
- Optional TOTP 2FA (Google Authenticator, Authy, etc.)
- JWT auth (HTTP-only cookie, 24-hour expiration)
- Rate limiting on login endpoint (configurable, brute force protection)

### Setup Wizard (First Launch)

On first launch, Nidus detects no admin account exists and presents a wizard:

1. **Create admin account** — username, password
2. **Create first category** — name, icon

Service configuration is done afterward in Settings.

---

## Dashboard & Layout

### Sidebar (Left)

- Always visible on desktop
- Collapsible burger menu on mobile
- Lists categories (user-created)
- Each category has a name and icon (1700+ Lucide icons)
- Settings, help, about, and logout links
- App version display in footer

### Main Area

- Free-placement widget grid with CSS grid positioning
- Widgets are **resizable** (width and height via drag handles)
- **Drag & drop** repositioning with pointer events
- Auto-compaction (no vertical gaps)
- Collision resolution on drop
- Responsive: 12 columns desktop, 6 tablet, 1 mobile
- Layout saved per category in SQLite
- **Global search bar** — find any category or widget

### Categories

- Created, renamed, reordered, deleted via inline tab UI
- Slug-based URLs (`/dashboard/{slug}`)
- Each category contains a set of widgets
- Tab drag reordering in edit mode

### Edit Mode

- Toggle in header (pencil icon)
- When locked: drag, resize, rename, delete, add buttons are hidden
- Prevents accidental changes

---

## Widgets

### Docker (via Portainer API)

- Stack grouping (stack containers grouped, standalone shown separately)
- Stack actions: start all, stop all, restart all, update (redeploy)
- Container actions: start, stop, restart
- CPU and RAM usage per container
- Docker health check status
- Update check (image version comparison)
- Quick link to open services (mapped ports)

### Proxmox

- VMs and LXCs with status, CPU, RAM, uptime
- Actions: start, stop, shutdown, reboot
- Multi-node support

### Home Assistant

- Any entity type: lights, switches, sensors, climate, cameras, covers, locks, buttons, scenes, scripts
- Real-time updates via WebSocket
- Entity picker with search and domain filtering
- Configurable layout and entity size

### Uptime Kuma

- Monitor list with status (up/down/pending/maintenance)
- Latency display
- 24h uptime percentage

### Plex / Jellyfin

- Active sessions with poster art
- Progress tracking
- Configurable server type (Plex or Jellyfin)

### *arr Stack (Sonarr / Radarr / Lidarr / Prowlarr)

- Calendar view (upcoming releases)
- Download queue with progress
- Library stats

### Pi-hole

- DNS query stats (total, blocked, percentage)
- Toggle filtering on/off

### AdGuard

- DNS query stats, blocked count, average response time
- Toggle filtering on/off
- Filter rules count

### Weather (OpenWeatherMap)

- Current conditions: temperature, humidity, wind, icon
- 5-day forecast
- Configurable city or GPS coordinates, units (°C/°F)

### Calendar (iCal)

- Parse iCal/CalDAV feeds
- Event list with date grouping (today, tomorrow, upcoming)
- Multiple feed URLs

### RSS / Atom

- Article list with title, date, source
- Multiple feed URLs
- Configurable max articles

### System Stats

- CPU, RAM, disk usage gauges
- Uptime, hostname, core count

### App Links (Bookmarks)

- Grid layout with favicon auto-fetch
- Group support with tags
- Health check (HTTP ping with status indicator)
- Sortable (manual, name, status)

### Reolink Cameras

- Snapshot grid with configurable refresh
- Live streaming via go2rtc (WebRTC/MSE)
- Camera discovery
- Fullscreen view (double-click)

### Finance (Yahoo Finance)

- Stock/ETF/crypto quotes
- Price, change, change percent
- Volume, day high/low
- Market state indicator

### Grafana

- Embed Grafana panels via iframe
- Multi-panel grid layout
- Dashboard and panel picker from Grafana API
- Configurable Grafana URL and API token

### JDownloader

- Download queue with progress
- Add links
- Start/pause queue
- Cleanup finished downloads

### Transmission

- Torrent list with progress and speed
- Add torrents (URL/magnet or file upload)
- Start/stop individual or all

---

## Internationalization (i18n)

- **11 languages:** French, English, Spanish, German, Portuguese, Italian, Dutch, Russian, Chinese (Simplified), Japanese, Arabic
- **RTL support** for Arabic
- Dynamic locale auto-discovery
- Fallback chain: selected language → English → French → raw key
- Browser language auto-detection on first visit
- Language selectable in settings
- Translation template available in `docs/i18n-template.json`

---

## Themes & Customization

- **4 built-in themes:** Dark, Light, Nord, Dracula
- Visual preview cards in theme selector
- **Accent color picker** with auto-generated hover/active variants
- **Custom CSS injection** (textarea in settings, stored in DB)
- Theme preference saved per user

---

## Notifications

- **Gotify** integration
- **Ntfy** integration
- **Apprise** integration (multi-provider)
- Configurable notification rules (what to notify, thresholds)
- Event types: container down, service unreachable

---

## Webhooks

- Incoming webhooks (`POST /api/webhooks/{id}`)
- HMAC signature validation
- Configurable actions: send notification, refresh widget, invalidate cache
- Management UI in settings

---

## Kiosk Mode

- Route `/kiosk` or `?kiosk=true`
- Header, sidebar, and edit buttons hidden
- Automatic rotation between categories (configurable interval)
- Fullscreen display

---

## Keyboard Shortcuts

- `E` — toggle edit mode
- `1-9` — switch to category by number
- `/` — focus search bar
- `?` — show shortcuts help
- `Escape` — close dialogs
- Can be disabled in settings

---

## UI Design

### Design Language

- Clean, modern, minimal
- Card-based widgets with subtle borders/shadows
- Lucide icon set (1700+ icons)
- Smooth transitions and animations
- Color-coded status indicators (green = running, red = stopped, yellow = warning)
- Toast notifications (in-app, no sounds, always silent)

### PWA (Progressive Web App)

- Installable on mobile devices
- Network-first caching strategy
- App icon + splash screen

### Responsive Breakpoints

- **TV** (>1920px): larger base font, more columns
- **Desktop** (>1024px): sidebar always visible, 12-column grid
- **Tablet** (768-1024px): sidebar collapsible, 6-column grid
- **Mobile** (<768px): burger menu, single column

---

## Configuration

### config.yaml (Bootstrap)

Minimal bootstrap config. Most configuration happens via UI and is stored in SQLite.

```yaml
server:
  port: 3777

database:
  path: "./data/nidus.db"
```

### SQLite Database

Stores:
- User accounts (hashed passwords, 2FA secrets, roles, token versions)
- Connected services (encrypted credentials — AES-256-GCM)
- Categories (name, icon, slug, order)
- Widget layout per category (position, size, config)
- User preferences (theme, accent color, language, custom CSS)
- Webhooks (name, secret, actions)
- Notification providers and rules

### YAML Config Import/Export

- Full dashboard export as YAML (services, categories, widgets, settings)
- Import YAML to restore or bootstrap a new instance
- Auto-import from `config.yaml` on first startup if DB is empty

### Encrypted Export/Import

- Export full config as encrypted JSON (AES-256-GCM)
- Password-protected backup and restore

---

## Deployment

### Docker Compose

```yaml
services:
  nidus:
    image: ghcr.io/tdebuilt/nidus-dashboard:latest
    ports:
      - "3777:3777"
      - "1984:1984"  # go2rtc streaming
    volumes:
      - ./data:/data
    restart: unless-stopped
```

### Standalone Binary

Single binary with embedded frontend. Download from GitHub Releases.

### Desktop App (Tauri)

Native installers for Linux (.deb, .AppImage), macOS (.dmg), Windows (.msi, .exe).

### Data Persistence

Single `./data` volume containing:
- `nidus.db` (SQLite database)
- `config.yaml` (bootstrap config, optional)
- `go2rtc.yaml` (auto-generated streaming config)

---

## API

- RESTful JSON API on `/api/*`
- WebSocket on `/api/ws` for real-time broadcasts
- OpenAPI/Swagger documentation served at `/api/docs`
- JWT authentication (cookie or Bearer header)

---

## Data Refresh Strategy

- **Polling** — configurable interval (default 30s)
- **WebSocket** for Home Assistant (real-time entity updates)
- **WebSocket broadcast** for dashboard events (widget updates, notifications)
- Backend caches external API responses (30s TTL)
