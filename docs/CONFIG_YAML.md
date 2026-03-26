# YAML Configuration

Nidus supports exporting and importing the full dashboard configuration as a human-readable YAML file. This can be used for:

- **Backup**: Save your configuration in a readable format
- **Version control**: Track configuration changes in Git
- **Migration**: Move your setup to a new instance
- **Bootstrapping**: Pre-configure a fresh Nidus install via `config.yaml`

## Format

```yaml
version: 2

settings:
  theme: dark           # dark, light, nord, dracula
  language: fr          # fr, en, es, de, pt, it, nl, ru, zh, ja, ar
  refresh_interval: 30  # seconds
  accent_color: ""      # hex color or empty for default
  custom_css: ""        # custom CSS injected in <head>

categories:
  - name: Infrastructure
    icon: server
    sort_order: 0
    widgets:
      - type: docker
        title: Docker Containers
        config: '{"environment_id": 1}'
        pos_x: 0
        pos_y: 0
        width: 6
        height: 3
      - type: proxmox
        title: Proxmox VMs
        config: '{}'
        pos_x: 6
        pos_y: 0
        width: 6
        height: 3

  - name: Media
    icon: play-circle
    sort_order: 1
    widgets:
      - type: transmission
        title: Torrents
        config: '{}'
        pos_x: 0
        pos_y: 0
        width: 12
        height: 4

services:
  - type: portainer
    name: Portainer
    url: "https://192.168.1.100:9443"
    credentials: '{"token": "your-api-token"}'
    enabled: true
  - type: proxmox
    name: Proxmox
    url: "https://192.168.1.50:8006"
    credentials: '{"token_id": "user@pam!tokenid", "token_secret": "xxx"}'
    enabled: true
```

## Widget types

| Type | Description |
|------|-------------|
| `docker` | Docker containers (via Portainer) |
| `proxmox` | Proxmox VMs and LXCs |
| `homeassistant` | Home Assistant entities |
| `adguard` | AdGuard Home DNS stats |
| `jdownloader` | JDownloader queue |
| `transmission` | Transmission torrents |
| `applink` | Bookmark links with health check |

## Grid system

- The grid uses **12 columns** (`pos_x`: 0-11, `width`: 1-12)
- Rows are **80px** each (`pos_y` and `height` are in row units)
- `height: 0` means auto-height (fit content)
- Widgets cannot overlap; collisions are resolved automatically

## API Endpoints

### Export (admin only)

```
GET /api/config/yaml
```

Returns the full configuration as a downloadable YAML file. Service credentials are included in plaintext.

### Import (admin only)

```
POST /api/config/yaml
Content-Type: application/x-yaml

<yaml content>
```

Replaces the current configuration with the imported one. Existing categories, widgets, and services are deleted before import.

## Auto-import at startup

If `data/config.yaml` contains dashboard sections (`settings`, `categories`, or `services`) **and** the database is empty (no categories exist), Nidus will automatically import the dashboard configuration on startup.

This allows pre-configuring Nidus by adding dashboard sections to the same `config.yaml` used for server settings:

```yaml
# Server settings (always read)
server:
  port: 3777

database:
  path: ./data/nidus.db

# Dashboard settings (auto-imported on first startup only)
settings:
  theme: dark
  language: fr

categories:
  - name: Home
    icon: home
    sort_order: 0
    widgets:
      - type: docker
        title: Containers
        config: '{}'
        pos_x: 0
        pos_y: 0
        width: 12
        height: 4

services:
  - type: portainer
    name: Portainer
    url: "https://portainer.local:9443"
    credentials: '{"token": "your-token"}'
    enabled: true
```

## Security warning

The YAML export includes service credentials (API tokens, passwords) **in plaintext**. Do not commit this file to a public repository or share it without removing sensitive values first.

For encrypted backups, use the standard export/import feature in Settings which encrypts the file with a password (AES-256-GCM).
