# Nidus — Specification

## Overview

Nidus is a lightweight, self-hosted dashboard for monitoring and managing Docker containers, Proxmox VMs/LXCs, and self-hosted services. It runs on low-spec servers (2GB RAM) and provides both visibility and control from a single interface.

**Target user:** Single admin managing a homelab with Proxmox, Docker, and various self-hosted services.

**Access:** Exposed via reverse proxy (Nginx) at `nidus.tdelab.eu`.

---

## Architecture

### Stack

| Component | Technology | Notes |
|---|---|---|
| Backend | Go (Chi router) | Single binary, ~15-25MB RAM |
| Frontend | Svelte | Compiled to static, embedded in Go binary |
| Database | SQLite | Single file, zero config |
| Deployment | Docker Compose | Single container |
| Reverse proxy | Nginx (GUI-managed) | Already in place |

### No Agent

Nidus does **not** deploy agents on remote machines. It connects to:
- **Portainer API** for Docker management (not Docker API directly)
- **Proxmox API** for VM/LXC management
- **Service APIs** (Home Assistant, AdGuard, JDownloader, Transmission) directly

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
└─────────────────────┼────────────────────────────┘
                      │
        ┌─────────────┼─────────────────┐
        │             │                 │
   ┌────▼────┐  ┌─────▼─────┐  ┌───────▼───────┐
   │Portainer│  │ Proxmox   │  │ Services      │
   │ API     │  │ API       │  │ (HA, AdGuard, │
   │(CE + EE)│  │           │  │  JDL, Trans.) │
   └─────────┘  └───────────┘  └───────────────┘
```

---

## Authentication

### Single User (Phase 1)

- One admin account created during setup wizard
- Username + password (bcrypt hashed)
- TOTP 2FA (Google Authenticator, Authy, etc.)
- JWT auth (HTTP-only cookie, **7-day expiration**)
- **Rate limiting** on login endpoint (brute force protection — exposed to internet)
- Future: multi-user with roles (admin / read-only)

### Setup Wizard (First Launch)

On first launch, Nidus detects no admin account exists and presents a wizard:

1. **Create admin account** — username, password, optional 2FA setup
2. **Connect Portainer** — URL + API key or credentials (all environments displayed)
3. **Connect Proxmox** — URL + API token
4. **Configure modules** — enable/disable, enter API keys for services
5. **Create first category** — name, icon, pick widgets to add

---

## Dashboard & Layout

### Sidebar (Left)

- Always visible on desktop
- Collapsible burger menu on mobile
- Lists categories (user-created)
- Each category has a name and icon
- Clicking a category shows its widgets in the main area
- "Settings" link at the bottom

### Main Area (Right)

- Grid of widgets, **drag-and-drop** repositioning
- Widgets are **resizable** (grid-based, like Grafana)
- Responsive: grid adapts to screen size, single column on mobile
- Layout saved per category in SQLite
- **~20 widgets max per category** (performance target)
- **Global search bar** — quickly find any container, VM, service, or widget

### Categories

- Created, renamed, reordered, deleted via UI
- Each category contains a set of widgets
- Examples: "Infrastructure", "Media", "Réseau", "Domotique"
- A widget belongs to one category

### Widgets

Widgets are the building blocks of the dashboard. Each module provides one or more widget types.

---

## Modules

### Docker (via Portainer API)

**Connection:** Portainer API (supports both CE and EE).

**Configuration:**
```yaml
portainer:
  url: "https://portainer.tdelab.eu"
  api_key: "ptr_xxxxxxxxxxxxxxxx"
```

**Display logic:**
- If a container belongs to a **stack** → show the stack (grouped)
- If a container has **no stack** → show the individual container
- Never show both a stack AND its individual containers
- **All Portainer environments** are displayed (multi-host)

**Stack widget:**
- Stack name, status (running/stopped/partial)
- Number of containers (running/total)
- Actions: **start all**, **stop all**, **restart all**, **update** (recreate with latest image)
- Expandable: click to see individual containers inside
- **Confirmation dialog** on all destructive actions

**Container widget (standalone only):**
- Container name, image, status, uptime
- **Docker health check status** (healthy/unhealthy) when available
- Actions: **start**, **stop**, **restart**, **update** (recreate with latest image)
- **Confirmation dialog** on all destructive actions
- Quick link to open the service (if port is mapped)

**Update action:**
- Calls Portainer "recreate" API (pull latest image + recreate)
- Shows progress/status
- Works for both stacks and individual containers

### Proxmox

**Connection:** Proxmox API with token auth.

**Configuration:**
```yaml
proxmox:
  url: "https://proxmox.tdelab.eu:8006"
  token_id: "nidus@pam!nidus-token"
  token_secret: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
  node: "pve"  # Single node
```

**VM/LXC widget:**
- Name, type (VM/LXC), VMID, status (running/stopped)
- CPU, RAM usage (**current values only**, no history graphs)
- Uptime
- Actions: **start**, **stop**, **shutdown**, **reboot**
- **Confirmation dialog** on all destructive actions

**Note:** Proxmox and Docker are displayed separately in the dashboard (different categories). No hierarchy shown even though Docker runs inside Proxmox VMs/LXCs.

### Home Assistant

**Connection:** Home Assistant REST API + WebSocket API for real-time.

**Configuration:**
```yaml
homeassistant:
  url: "https://ha.tdelab.eu"
  token: "eyJ0eXAiOi..."
```

**Entity widget:**
- User picks any HA entity (light, switch, sensor, camera, climate, etc.)
- Displays current state (on/off, temperature value, camera snapshot, etc.)
- Associated action depending on type:
  - Light → toggle on/off, brightness slider
  - Switch → toggle on/off
  - Sensor → display value only
  - Climate → set temperature, mode
  - Camera → show snapshot, link to stream
- Real-time updates via WebSocket
- Visual style inspired by HA tile cards

### AdGuard

**Connection:** AdGuard Home REST API.

**Configuration:**
```yaml
adguard:
  url: "https://adguard.tdelab.eu"
  username: "admin"
  password: "xxxxx"
```

**Widget:**
- DNS queries count (today)
- Blocked queries count + percentage
- Filtering status (enabled/disabled)
- Action: **toggle filtering** on/off
- Optional: top blocked domains chart

### JDownloader

**Connection:** JDownloader API (via MyJDownloader API or direct).

**Configuration:**
```yaml
jdownloader:
  # Option A: MyJDownloader cloud API
  email: "user@example.com"
  password: "xxxxx"
  device: "my-jdownloader"

  # Option B: Direct API (if accessible on LAN)
  # url: "http://192.168.1.x:3129"
```

**Widget:**
- Download queue: list of current/pending downloads
- Progress bars, speed, ETA
- Actions: **add links** (paste URLs), **start/pause queue**

### Transmission

**Connection:** Transmission RPC API.

**Configuration:**
```yaml
transmission:
  url: "http://192.168.1.x:9091/transmission/rpc"
  username: "admin"
  password: "xxxxx"
```

**Widget:**
- Torrent list: name, progress, status, speed
- Global upload/download speed
- Actions: **add torrent** (URL or magnet), **pause/resume** individual or all

### App Links

Configurable shortcut widgets linking to URLs or local applications.

**Widget:**
- Clickable card with custom **icon** (Lucide icon picker), name, and optional description
- **URL link** — opens in a new browser tab (always available)
- **Local app link** — launches a local application on the host machine (only available when Nidus runs locally, not via Docker/remote)
  - Uses OS-specific commands: `open` (macOS), `xdg-open` (Linux), `start` (Windows)
  - Accepts an app path (e.g., `/Applications/Firefox.app`, `C:\Program Files\App\app.exe`) or URI scheme (`vscode://`, `obsidian://`, `steam://`)
  - Local mode detected automatically via `--local` flag or absence of Docker environment
- **Health check** — periodic HTTP check on a configurable URL, status indicator (green = up, red = down)
- Each link has: `name`, `url`, `icon` (Lucide), `local_app_path` (optional), `health_check_url` (optional)

---

## Internationalization (i18n)

- **Default language:** French
- **Supported:** French, English
- **Extensible:** translation files (JSON) so more languages can be added
- **Stored in:** SQLite (user preference) + JSON files for translations
- Language selectable in settings

---

## UI Design

### Theme

- **Dark mode by default**
- Light mode toggle available in settings
- Preference saved per user in SQLite

### Design Language

- Clean, modern, minimal
- Card-based widgets with subtle borders/shadows
- Consistent icon set (Lucide or similar)
- Smooth transitions and animations (but not excessive)
- Color-coded status indicators (green = running, red = stopped, yellow = warning)
- **Toast notifications** (in-app, no sounds)
- **Always silent** — no audio feedback

### PWA (Progressive Web App)

- Installable on mobile devices (manifest.json, service worker)
- Offline: show last cached state with "offline" indicator
- App icon + splash screen

### Responsive Breakpoints

- **Desktop** (>1024px): sidebar always visible, multi-column grid
- **Tablet** (768-1024px): sidebar collapsible, 2-column grid
- **Mobile** (<768px): burger menu, single column, widgets stack vertically

---

## Configuration

### config.yaml (Initial / Fallback)

Minimal bootstrap config. Most configuration happens via UI and is stored in SQLite.

```yaml
server:
  port: 3777
  base_url: "https://nidus.tdelab.eu"

database:
  path: "./data/nidus.db"
```

### SQLite Database

Stores:
- User account (hashed password, 2FA secret)
- Connected services (Portainer, Proxmox, HA, etc. — **AES-256-GCM encrypted** credentials)
  - Master encryption key generated at setup, stored as env var `NIDUS_ENCRYPTION_KEY`
  - All API keys, tokens, and passwords encrypted at rest in SQLite
- Categories (name, icon, order)
- Widget layout per category (position, size, config)
- User preferences (theme, language)
- App links (name, URL, icon, local_app_path, health check URL)

### Config Export/Import

- **Export:** Download full config as JSON (categories, layout, widgets, service connections)
- **Import:** Upload JSON to restore config
- Available in Settings page

---

## Deployment

### Docker Compose

```yaml
version: "3.8"
services:
  nidus:
    image: nidus:latest
    container_name: nidus
    ports:
      - "3777:3777"
    volumes:
      - ./data:/data        # SQLite DB + config
    restart: unless-stopped
```

### Data Persistence

- Single `./data` volume containing:
  - `nidus.db` (SQLite database)
  - `config.yaml` (bootstrap config, optional)

### Nginx Reverse Proxy

```nginx
server {
    listen 443 ssl;
    server_name nidus.tdelab.eu;

    location / {
        proxy_pass http://localhost:3777;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /ws {
        proxy_pass http://localhost:3777;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

---

## Roadmap (Phased)

### Phase 1 — Foundation
- [ ] Go backend skeleton (Chi router, SQLite, config)
- [ ] Auth system (user/password + TOTP 2FA)
- [ ] Setup wizard
- [ ] Svelte frontend skeleton (sidebar, grid layout, dark theme)
- [ ] Category CRUD via UI
- [ ] Widget framework (drag-and-drop, resize)
- [ ] i18n system (FR + EN)

### Phase 2 — Docker & Proxmox
- [ ] Portainer API integration (CE + EE)
- [ ] Stack/container display logic (stack grouping)
- [ ] Container/stack actions (start/stop/restart/update)
- [ ] Proxmox API integration
- [ ] VM/LXC widgets (status, metrics, actions)

### Phase 3 — Services
- [ ] Home Assistant module (entity widgets + actions)
- [ ] AdGuard module (stats + toggle)
- [ ] JDownloader module (queue + add links)
- [ ] Transmission module (torrents + add)

### Phase 4 — Polish
- [ ] App links with health checks
- [ ] Light mode
- [ ] PWA support (manifest, service worker, installable)
- [ ] Mobile optimization
- [ ] Global search bar
- [ ] Config export/import (JSON)
- [ ] Performance tuning
- [ ] Rate limiting on login

### Future Integrations

For adding new service integrations beyond the initial scope (Portainer, Proxmox, HA, AdGuard, JDownloader, Transmission), look at [Homepage](https://gethomepage.dev/) (Next.js self-hosted dashboard) for inspiration. It has 200+ service integrations with well-documented API patterns that could serve as reference for widget design and API client implementation.

### Data Refresh Strategy

- **Simple polling** — fetch data every N seconds (configurable, default 30s)
- No WebSocket for MVP (except Home Assistant which requires it)
- Frontend polls backend API, backend caches and polls external APIs
