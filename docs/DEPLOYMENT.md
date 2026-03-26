# Deployment Guide

Nidus can be deployed in three ways: Docker (recommended for servers), standalone binary, or desktop application.

## Docker (Server)

The recommended way to run Nidus on a server.

```bash
docker compose up -d
```

The default `docker-compose.yml` exposes port 3777 and persists data in `./data/`.

### Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `NIDUS_PORT` | `3777` | Server port |
| `NIDUS_DB_PATH` | `/data/nidus.db` | Database file path |
| `NIDUS_BASE_URL` | `http://localhost:3777` | Public URL |

You can also mount a `config.yaml` file:

```yaml
server:
  port: 3777
  base_url: https://nidus.example.com
database:
  path: /data/nidus.db
```

### Using the GitHub Container Registry image

```bash
docker pull ghcr.io/tdebuilt/nidus-dashboard:latest
docker run -d -p 3777:3777 -v ./data:/data ghcr.io/tdebuilt/nidus-dashboard:latest
```

## Standalone Binary

Download the binary for your platform from the [GitHub Releases](https://github.com/tdebuilt/Nidus-Dashboard/releases) page.

| Platform | Binary |
|----------|--------|
| Linux x86_64 | `nidus-x86_64-unknown-linux-gnu` |
| macOS Intel | `nidus-x86_64-apple-darwin` |
| macOS Apple Silicon | `nidus-aarch64-apple-darwin` |
| Windows x64 | `nidus-x86_64-pc-windows-msvc.exe` |

```bash
chmod +x nidus-x86_64-unknown-linux-gnu
./nidus-x86_64-unknown-linux-gnu
```

The server starts on `http://localhost:3777`. Data is stored in `./data/nidus.db` by default.

## Desktop Application (Tauri)

Download the installer for your platform from the [GitHub Releases](https://github.com/tdebuilt/Nidus-Dashboard/releases) page.

| Platform | Format |
|----------|--------|
| Linux | `.deb`, `.AppImage` |
| macOS | `.dmg` |
| Windows | `.msi`, `.exe` |

### How it works

The desktop app bundles the Nidus server and runs it locally. The app:

- Starts the server on `127.0.0.1` (localhost only, not network-accessible)
- Automatically finds a free port if 3777 is in use
- Stores data in your user config directory:
  - Linux: `~/.config/nidus/`
  - macOS: `~/Library/Application Support/nidus/`
  - Windows: `%APPDATA%\nidus\`

### Desktop vs Server

| Feature | Desktop | Server |
|---------|---------|--------|
| Network access | Localhost only | Configurable |
| Data location | User config dir | `./data/` or Docker volume |
| Port | Auto-detected | Configurable |
| Local tools (DDEV, etc.) | Available | Not available |

### Platform-specific notes

**macOS**: If Gatekeeper blocks the app, run:
```bash
xattr -d com.apple.quarantine /Applications/Nidus.app
```

**Windows**: If SmartScreen shows a warning, click "More info" then "Run anyway". This happens because the binary is not code-signed.

## CI/CD

Releases are fully automated via GitHub Actions.

### Automatic release workflow

1. Tag a commit: `git tag v1.0.0 && git push --tags`
2. GitHub Actions automatically:
   - Builds Go binaries for all 4 platforms
   - Builds Tauri desktop installers for Linux, macOS, Windows
   - Builds and pushes the Docker image to `ghcr.io`
   - Creates a GitHub Release with all artifacts attached

### Test workflow

Every push to any branch and every pull request triggers:
- Go tests (`go vet` + `go test ./...`)
- Frontend tests (Vitest)
- Full build check (frontend + backend)

## Building from Source

### Prerequisites

- Go 1.24+
- Node.js 20+
- Rust (for desktop builds only)

### Server build

```bash
make build          # Build production binary
make dev            # Development mode (Go + Svelte HMR)
make docker         # Build and run via Docker
```

### Desktop build

```bash
make desktop-dev    # Development mode with Tauri
make desktop-build  # Production Tauri build
```
