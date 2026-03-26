# Nidus — Instructions pour Claude

## Projet

Nidus est un dashboard self-hosted pour gérer Docker (via Portainer), Proxmox, et des services (HA, AdGuard, JDownloader, Transmission). Stack : Go (Chi) + Svelte + SQLite.

## Fichiers clés

- `planning/SPEC.md` — Spécification complète du projet
- `planning/PLAN.md` — Plan d'implémentation (phases 1-4)
- `planning/ROADMAP_TASKS.md` — Roadmap détaillée (phases 2-6) avec suivi
- `planning/TESTING.md` — Checklist de tests manuels
- `README.md` — Description du projet

## Avancement

Suivi dans `planning/ROADMAP_TASKS.md`. Utiliser `/next-task` pour implémenter la prochaine tâche pending.

## Conventions

- **Langue du code** : anglais (noms de variables, fonctions, commentaires)
- **Langue de l'UI** : français par défaut, anglais supporté (i18n)
- **Backend Go** : packages dans `internal/`, point d'entrée dans `cmd/nidus/`
- **Frontend Svelte** : dans `web/`, build vers `web/static/` (embedded dans le binaire Go)
- **Base de données** : SQLite dans `./data/nidus.db`
- **Credentials** : chiffrés AES-256-GCM en base, jamais en clair

## Structure cible

```
cmd/nidus/main.go
internal/
  config/       # Chargement config.yaml + env vars
  database/     # Connexion SQLite, migrations, queries
  crypto/       # AES-256-GCM encrypt/decrypt
  middleware/   # Auth JWT, rate limiting, CORS
  handlers/     # HTTP handlers (auth, categories, widgets, services, settings)
  models/       # Structs Go (User, Category, Widget, Service, etc.)
  services/     # Clients API externes (portainer, proxmox, homeassistant, adguard, jdownloader, transmission)
web/
  src/          # Code source Svelte
  static/       # Build output (embedded)
data/           # Volume Docker (nidus.db, config.yaml)
```

## Build

- **Go et les outils backend ne sont pas installés localement** — tout passe par Docker
- **Build complet** : `docker compose up --build -d` (méthode principale)
- **Frontend seul** : `cd web && npm run build` (Node/npm sont disponibles localement)

## Tests

- **Backend Go** : `docker compose exec nidus go test ./...` (via Docker)
- **Frontend Svelte** : `vitest` après chaque composant
- **Lint** : `go vet ./...` (via Docker) + `eslint` (local)
- Les fichiers de test Go suivent la convention `*_test.go` dans le même package
- Les tests Svelte sont dans `web/src/**/*.test.ts`

## Git

- **Ne jamais mentionner "Claude"** dans les messages de commit, noms de branches, ou métadonnées git
- **Ne jamais s'ajouter comme co-author** (pas de `Co-authored-by`)
- Les commits doivent avoir l'air écrits par un développeur humain

## Commandes

- `docker compose up --build -d` — **Build et lance l'app via Docker** (méthode principale)
- `make dev` — Lance le backend Go + frontend Svelte en dev (hors Docker)
- `make build` — Build production (Svelte → embed → Go binary, via Docker)
- `docker compose up` — Lance via Docker
