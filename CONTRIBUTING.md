# Contributing to Nidus

Thanks for your interest in contributing! Whether it's a bug fix, new widget, translation, or documentation improvement — all contributions are welcome.

## Getting started

### Prerequisites

- **Node.js 20+** and **npm** — for the Svelte frontend
- **Go 1.24+** — for the backend (or use Docker)
- **Docker** (optional) — for containerized builds and testing

### Local development

```bash
# Clone the repo
git clone https://github.com/tdebuilt/Nidus-Dashboard.git
cd Nidus-Dashboard

# Install frontend dependencies
cd web && npm ci && cd ..

# Run in dev mode (Go backend + Svelte dev server)
make dev

# Open http://localhost:5173 (Svelte dev server with HMR)
# API runs on http://localhost:3777
```

### Running tests

```bash
# All tests
make test

# Frontend only
cd web && npm test -- --run

# Backend only (requires Go installed, or use Docker)
go test ./...
# or via Docker:
docker compose exec nidus go test ./...
```

### Building

```bash
# Full production build (frontend + Go binary)
make build

# Docker
docker compose up --build -d
```

## Project structure

```
cmd/nidus/              # Go entry point
internal/
  config/               # YAML + env config loading
  database/             # SQLite connection, migrations, queries
  crypto/               # AES-256-GCM encrypt/decrypt
  middleware/            # JWT auth, rate limiting, CORS
  handlers/             # HTTP handlers (REST API)
  models/               # Go structs (User, Category, Widget, Service...)
  services/             # External API clients
    portainer/           #   Docker via Portainer
    proxmox/             #   Proxmox VE
    homeassistant/       #   Home Assistant
    adguard/             #   AdGuard Home
    jdownloader/         #   MyJDownloader
    transmission/        #   Transmission RPC
  cache/                # In-memory cache with TTL
  server/               # HTTP server setup and routing
web/
  src/
    lib/
      api/              # API client
      components/       # Svelte components (organized by widget type)
      config/           # Widget config forms
      i18n/             # Translation files (fr.json, en.json)
      stores/           # Svelte stores (auth, theme, edit mode...)
      widgetRegistry.ts # Central widget registration
    pages/              # Page components (Dashboard, Login, Setup, Settings)
  public/               # Static assets (icons, manifest, service worker)
  static/               # Build output (embedded in Go binary)
data/                   # Runtime data (SQLite DB) — not committed
```

## How to add a new widget

Thanks to the widget registry, adding a new widget requires changes in only 3 places:

### 1. Backend service (Go)

Create a new package in `internal/services/<yourwidget>/`:

```
internal/services/yourwidget/
  client.go     # API client
  types.go      # Request/response types
```

Create a handler in `internal/handlers/yourwidget.go` and register routes in `internal/server/server.go`.

Add your service to the registry in `internal/handlers/services.go`:

```go
var ServiceRegistry = map[string]ServiceDefinition{
    // ... existing services
    "yourwidget": {CachePrefix: "yw:", TestPath: "/api/status"},
}
```

### 2. Frontend component (Svelte)

Create your widget components in `web/src/lib/components/yourwidget/`:

```
web/src/lib/components/yourwidget/
  YourWidget.svelte        # Main widget component
  YourWidgetConfig.svelte  # Config form (optional)
```

Your widget component receives a `config` prop (JSON string):

```svelte
<script lang="ts">
  interface Props {
    config?: string
  }
  let { config = '{}' }: Props = $props()
</script>
```

### 3. Register in the widget registry

In `web/src/lib/widgetRegistry.ts`:

```typescript
import YourWidget from './components/yourwidget/YourWidget.svelte'
import YourWidgetConfig from './components/yourwidget/YourWidgetConfig.svelte' // optional
import { SomeIcon } from 'lucide-svelte'

register({
  type: 'yourwidget',
  label: 'Your Widget',
  icon: SomeIcon,
  component: YourWidget,
  configComponent: YourWidgetConfig, // omit if no config needed
  serviceType: 'yourwidget',        // must match ServiceRegistry key
})
```

Don't forget to add translations in `web/src/lib/i18n/fr.json` and `en.json`.

## How to add a translation

Adding a new language requires only 2 steps:

1. Copy `web/src/lib/i18n/fr.json` to `web/src/lib/i18n/<code>.json` (e.g. `de.json`)
2. Add your locale metadata in `web/src/lib/i18n/locales.ts`:
   ```ts
   de: { label: 'Deutsch', flag: '🇩🇪' },
   ```

That's it — the i18n system auto-discovers JSON files via `import.meta.glob` and registers them automatically.

### Translation file format

Translation files are JSON with **one level of nesting** — sections containing key-value pairs:

```json
{
  "section": {
    "key": "Translated text",
    "keyWithParam": "Hello {name}, you have {count} items"
  }
}
```

**Rules:**
- **Keys** are in English, dot-separated when used in code: `$t('section.key')`
- **Values** are the translated strings in the target language
- **Placeholders** use `{paramName}` syntax — these are replaced at runtime
- **Keep all keys identical** to `fr.json` (the reference file) — don't add or remove keys
- Run `npm run i18n:validate` to verify your file matches the reference

### Fallback chain

When a key is missing in the current locale, the system falls back:
1. Selected language → 2. English → 3. French → 4. Raw key

### Browser detection

On first visit (no stored preference), Nidus detects the browser language and selects the best matching locale automatically.

## Code conventions

- **Code language**: English (variable names, functions, comments)
- **UI language**: French by default, English supported — all user-facing strings go through i18n
- **Backend**: Go standard formatting (`go fmt`), packages in `internal/`
- **Frontend**: Svelte 5 (runes: `$state`, `$derived`, `$effect`), Tailwind CSS with CSS variables
- **Tests**: `*_test.go` for Go, `*.test.ts` for Svelte (vitest + testing-library)

## Commit messages

Use [Conventional Commits](https://www.conventionalcommits.org/) format:

```
feat: add weather widget
fix: correct CPU calculation for paused containers
docs: update installation instructions
refactor: extract cache logic into shared helper
test: add WidgetGrid drag-and-drop tests
```

## Pull request process

1. **Fork** the repo and create a branch from `main`
2. **Implement** your changes following the conventions above
3. **Write tests** for new functionality
4. **Run the full test suite** — make sure everything passes
5. **Open a PR** with a clear title and description
6. A maintainer will review and merge

## Reporting bugs

Open an [issue](https://github.com/tdebuilt/Nidus-Dashboard/issues) with:
- Steps to reproduce
- Expected vs actual behavior
- Browser / OS / Nidus version
- Screenshots if applicable

## Feature requests

Open an [issue](https://github.com/tdebuilt/Nidus-Dashboard/issues) describing:
- What you'd like to see
- Why it would be useful
- Any implementation ideas you have
