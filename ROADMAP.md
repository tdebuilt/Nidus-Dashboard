# Roadmap

## Planned

### Third-party Plugin System
- [ ] Define plugin format (manifest.json, Svelte component, optional Go handler)
- [ ] Plugin loader from a `plugins/` directory
- [ ] Auto-register plugins with the widget registry
- [ ] Security sandbox (iframe or Web Components)
- [ ] Developer documentation for creating plugins
- [ ] Starter plugin template

### Grafana Widget
- [ ] Backend service `internal/services/grafana/` (API client)
- [ ] Fetch dashboards list, panels, and snapshots
- [ ] Embed panels via iframe or render snapshots
- [ ] Handler `internal/handlers/grafana.go` + routes
- [ ] Add `grafana` to ServiceRegistry
- [ ] Frontend component `web/src/lib/components/grafana/`
- [ ] Widget config: Grafana URL, API key, dashboard/panel selection
- [ ] Register in `widgetRegistry.ts`
- [ ] Add translations (all locales)
- [ ] Tests backend + frontend


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
- [x] Docker multi-arch build (amd64 + arm64)
- [x] Tauri desktop app (Linux, macOS, Windows)
- [x] OpenAPI/Swagger documentation
- [x] Incoming webhooks with HMAC validation
- [x] Pre-commit hooks (golangci-lint + eslint)

### Features
- [x] 17 widget types (Docker, Proxmox, Home Assistant, Uptime Kuma, Plex/Jellyfin, *arr stack, Pi-hole, Weather, Calendar, RSS, System, Bookmarks, Reolink, Finance, AdGuard, JDownloader, Transmission)
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
