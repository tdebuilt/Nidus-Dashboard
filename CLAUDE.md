# Nidus — Instructions

## Project

Nidus is a self-hosted dashboard for managing Docker (via Portainer), Proxmox, and services (HA, AdGuard, Uptime Kuma, Plex/Jellyfin, etc.). Stack: Go (Chi) + Svelte 5 + SQLite.

## Key Files

- `SPEC.md` — Full project specification
- `ROADMAP.md` — Roadmap (planned, ideas, completed)
- `CHANGELOG.md` — Changelog by version
- `todo/TESTING.md` — Manual testing checklist
- `README.md` — Project description

## Progress

Tracked in `ROADMAP.md`. Use `/next-task` to implement the next pending task.

## Conventions

- **Code language**: English (variable names, functions, comments)
- **UI language**: French by default, 11 languages supported (i18n)
- **Backend Go**: packages in `internal/`, entry point in `cmd/nidus/`
- **Frontend Svelte**: in `web/`, builds to `web/static/` (embedded in Go binary)
- **Database**: SQLite in `./data/nidus.db`
- **Credentials**: AES-256-GCM encrypted in DB, never in plaintext

## Project Structure

```
cmd/nidus/main.go
internal/
  config/       # config.yaml + env vars
  database/     # SQLite connection, migrations, queries
  crypto/       # AES-256-GCM encrypt/decrypt
  middleware/   # JWT auth, rate limiting, CORS
  handlers/     # HTTP handlers (auth, categories, widgets, services, settings)
  models/       # Go structs (User, Category, Widget, Service, etc.)
  services/     # External API clients (portainer, proxmox, homeassistant, etc.)
  websocket/    # WebSocket hub for real-time broadcasts
web/
  src/          # Svelte source code
  static/       # Build output (embedded)
  e2e/          # Playwright E2E tests
data/           # Docker volume (nidus.db, config.yaml)
.githooks/      # Pre-commit hooks (lint)
```

## Build

- **Go and backend tools are NOT installed locally** — everything goes through Docker
- **Full build**: `docker compose up --build -d` (primary method)
- **Frontend only**: `cd web && npm run build` (Node/npm available locally)

## Tests

- **Go lint**: `make lint-go` (golangci-lint via Docker)
- **Frontend lint**: `cd web && npm run lint` (eslint, local)
- **Frontend tests**: `cd web && npm test -- --run` (vitest)
- **E2E tests**: `make test-e2e` (Playwright, requires app running)
- **Pre-commit hook**: runs Go lint + frontend lint automatically (`make setup` to enable)
- Go test files follow `*_test.go` convention in the same package
- Svelte tests are in `web/src/**/*.test.ts`
- E2E tests are in `web/e2e/**/*.spec.ts`

## Commands

- `docker compose up --build -d` — **Build and run the app via Docker** (primary method)
- `make lint-go` — Run golangci-lint via Docker
- `make lint` — Run Go lint + frontend lint
- `make test-e2e` — Run Playwright E2E tests
- `make setup` — Configure git hooks (run after cloning)
- `make dev` — Run Go backend + Svelte dev server (local, requires Go)
- `make build` — Production build (Svelte → embed → Go binary)
