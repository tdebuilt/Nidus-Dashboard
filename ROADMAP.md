# Roadmap

## Planned

### Third-party Plugin System
- [ ] Define plugin format (manifest.json, Svelte component, optional Go handler)
- [ ] Plugin loader from a `plugins/` directory
- [ ] Auto-register plugins with the widget registry
- [ ] Security sandbox (iframe or Web Components)
- [ ] Developer documentation for creating plugins
- [ ] Starter plugin template

## Ideas (Not Yet Planned)

- **Mobile app** — React Native or Flutter companion app
- **Multi-dashboard** — multiple dashboards per user with different layouts
- **Shared dashboards** — public read-only dashboards via link
- **Widget marketplace** — community-contributed widgets
- **Backup scheduling** — automated periodic config exports
- **LDAP/SSO authentication** — integrate with existing identity providers
- **Notification rules engine** — advanced conditions with templating

## Completed

### Infrastructure
- [x] E2E tests with Playwright (setup wizard, auth, CRUD, drag/resize, services)
- [x] CI/CD with GitHub Actions (Go lint, frontend lint, unit tests, E2E tests)
- [x] Docker image build (amd64)
- [x] Tauri desktop app (Linux, macOS, Windows)
- [x] OpenAPI/Swagger documentation
- [x] Incoming webhooks with HMAC validation
- [x] Pre-commit hooks (golangci-lint + eslint)

### Features
- [x] 20 widget types (Docker, Proxmox, Home Assistant, Uptime Kuma, Plex/Jellyfin, *arr stack, Pi-hole, Weather, Calendar, RSS, System, Bookmarks, Reolink, Finance, AdGuard, JDownloader, Transmission, qBittorrent, Grafana, SSH Terminal)
- [x] Multi-user with roles (admin, editor, viewer)
- [x] 4 built-in themes with accent color picker and custom CSS
- [x] 11 languages with RTL support
- [x] Kiosk mode with auto-rotation
- [x] Keyboard shortcuts
- [x] Push notifications (Gotify, Ntfy, Apprise)
- [x] YAML config import/export
- [x] Free-placement drag & drop grid
- [x] Responsive layout (mobile, tablet, desktop, TV)
- [x] Embedded go2rtc for camera streaming
- [x] Grafana widget (embed panels via iframe, multi-panel grid, dashboard/panel picker)
- [x] Lazy-loaded widget components (code-splitting)
- [x] Responsive column config per widget (desktop/tablet/mobile)
- [x] Add service modal dialog with type selection grid and alphabetical sort
