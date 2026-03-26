.PHONY: build dev test clean lint frontend backend desktop-dev desktop-build docker test-e2e test-e2e-headed lint-go lint-go-fix setup

TRIPLE := $(shell rustc -vV 2>/dev/null | grep host | cut -d' ' -f2 || echo "x86_64-unknown-linux-gnu")
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Build production binary (Svelte → embed → Go binary)
build: frontend backend

frontend:
	cd web && npm ci && npm run build

backend:
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o nidus ./cmd/nidus/

# Development: run Go backend + Svelte dev server
dev:
	@echo "Starting backend on :3777 and frontend dev server..."
	@trap 'kill 0' EXIT; \
		go run ./cmd/nidus/ & \
		cd web && npm run dev & \
		wait

# Desktop: build Go sidecar + run Tauri dev
desktop-dev: frontend
	CGO_ENABLED=0 go build -o desktop/src-tauri/binaries/nidus-$(TRIPLE) ./cmd/nidus/
	cd desktop/src-tauri && cargo tauri dev

# Desktop: production Tauri build
desktop-build: frontend
	CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=$(VERSION)" -o desktop/src-tauri/binaries/nidus-$(TRIPLE) ./cmd/nidus/
	cd desktop/src-tauri && cargo tauri build

# Docker
docker:
	docker compose up --build -d

# Run all tests
test: test-backend test-frontend

test-backend:
	go test ./...

test-frontend:
	cd web && npm test -- --run

# Lint
lint: lint-go lint-frontend

lint-go:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.11.4 golangci-lint run

lint-go-fix:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.11.4 golangci-lint run --fix

lint-frontend:
	@cd web && npm run lint 2>/dev/null || true

# E2E tests (requires app running or binary built)
test-e2e: frontend backend
	cd web && npx playwright test --project=auth-setup --project=e2e

test-e2e-headed: frontend backend
	cd web && npx playwright test --project=auth-setup --project=e2e --headed

# Setup dev environment (git hooks)
setup:
	git config core.hooksPath .githooks
	@echo "Git hooks configured from .githooks/"

# Clean build artifacts
clean:
	rm -f nidus
	rm -rf web/static/assets web/static/index.html web/static/.vite
	rm -f desktop/src-tauri/binaries/nidus-*
