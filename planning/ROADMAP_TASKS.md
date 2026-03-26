# Nidus — Roadmap détaillée

> Légende : `[x]` = finish, `[ ]` = pending
>
> Les tâches sont dans l'ordre logique d'exécution. Dire "tâche suivante" = première case `[ ]` non cochée.

---

## Phase 1 — Prérequis open source (voir OPEN_SOURCE_TASKS.md)

Détail complet dans [OPEN_SOURCE_TASKS.md](./OPEN_SOURCE_TASKS.md) (48 tâches).

---

## Phase 2 — Internationalisation & langues supplémentaires

### 2.1 Refactoring i18n pour support communautaire
- [x] Rendre le type `Locale` dynamique (ne plus hardcoder `'fr' | 'en'`)
- [x] Charger les fichiers de langue dynamiquement (import dynamique ou fetch)
- [x] Ajouter un mécanisme de détection automatique de la langue du navigateur
- [x] Ajouter un sélecteur de langue dans les settings (dropdown avec drapeau + nom natif)
- [x] Ajouter un fallback chain : langue choisie → anglais → français → clé brute
- [x] Créer un script de validation : vérifier que toutes les clés de `fr.json` existent dans les autres langues
- [x] Documenter le format des fichiers de traduction (structure JSON, placeholders `{param}`)

### 2.2 Nouvelles langues
- [x] Espagnol (`es.json`)
- [x] Allemand (`de.json`)
- [x] Portugais (`pt.json`)
- [x] Italien (`it.json`)
- [x] Néerlandais (`nl.json`)
- [x] Russe (`ru.json`)
- [x] Chinois simplifié (`zh.json`)
- [x] Japonais (`ja.json`)
- [x] Arabe (`ar.json`) + support RTL

### 2.3 Documentation i18n
- [x] Créer `docs/TRANSLATING.md` (guide pour les traducteurs)
- [x] Ajouter un template de fichier de langue vide (toutes les clés, valeurs à remplir)
- [x] Badge "help wanted: translations" dans le README

---

## Phase 3 — Thèmes & personnalisation

### 3.1 Système de thèmes
- [x] Extraire toutes les CSS variables dans un fichier de thème central
- [x] Créer une structure de thème (JSON/YAML : nom, auteur, couleurs)
- [x] Implémenter un thème loader qui applique les variables CSS
- [x] Créer 3-4 thèmes built-in (dark, light, nord, dracula)
- [x] Ajouter un sélecteur de thème dans les settings
- [x] Sauvegarder le thème choisi en DB (par utilisateur)
- [x] Prévisualisation live dans le sélecteur de thème

### 3.2 Couleurs d'accent
- [x] Ajouter un color picker dans les settings
- [x] Appliquer la couleur d'accent sur `--color-primary` et dérivées
- [x] Générer automatiquement hover/active à partir de la couleur choisie
- [x] Sauvegarder en DB par utilisateur

### 3.3 CSS custom
- [x] Ajouter un champ "CSS personnalisé" dans les settings (textarea)
- [x] Injecter le CSS custom dans le `<head>` au chargement
- [x] Sauvegarder en DB
- [x] Ajouter un avertissement sur les risques (XSS si multi-utilisateur)

---

## Phase 4 — Nouveaux widgets

### 4.1 Widget Uptime Kuma
- [x] Créer le service backend `internal/services/uptimekuma/` (client API)
- [x] Ajouter les types (Monitor, Heartbeat, Status)
- [x] Implémenter les endpoints : liste monitors, statut, uptime %
- [x] Créer le handler `internal/handlers/uptimekuma.go`
- [x] Enregistrer les routes dans `server.go`
- [x] Ajouter `uptimekuma` dans `ServiceRegistry` (services.go)
- [x] Créer le composant frontend `web/src/lib/components/uptimekuma/`
- [x] Composant principal : UptimeKumaWidget.svelte (liste des monitors avec statut)
- [x] Composant MonitorCard.svelte (nom, statut, uptime %, latence)
- [x] Enregistrer dans `widgetRegistry.ts`
- [x] Ajouter les traductions fr/en
- [x] Tests backend + frontend

### 4.2 Widget Plex / Jellyfin
- [x] Créer le service backend `internal/services/mediaserver/` (client générique)
- [x] Support Plex API (en cours de lecture, bibliothèque)
- [x] Support Jellyfin API (en cours de lecture, bibliothèque)
- [x] Créer le handler `internal/handlers/mediaserver.go`
- [x] Enregistrer les routes dans `server.go`
- [x] Ajouter dans `ServiceRegistry`
- [x] Créer le composant frontend `web/src/lib/components/mediaserver/`
- [x] Widget : sessions actives, poster, progression lecture
- [x] Config : choix Plex ou Jellyfin
- [x] Enregistrer dans `widgetRegistry.ts`
- [x] Ajouter les traductions fr/en
- [x] Tests backend + frontend

### 4.3 Widget *arr stack (Sonarr / Radarr / Lidarr / Prowlarr)
- [x] Créer le service backend `internal/services/arr/` (client générique *arr)
- [x] Support Sonarr API (séries, calendrier, queue)
- [x] Support Radarr API (films, calendrier, queue)
- [x] Support Lidarr API (musique)
- [x] Support Prowlarr API (indexers)
- [x] Créer le handler `internal/handlers/arr.go`
- [x] Enregistrer les routes dans `server.go`
- [x] Ajouter dans `ServiceRegistry`
- [x] Créer le composant frontend `web/src/lib/components/arr/`
- [x] Widget : prochaines sorties, queue de téléchargement, statut
- [x] Config : choix du type *arr + URL
- [x] Enregistrer dans `widgetRegistry.ts`
- [x] Ajouter les traductions fr/en
- [x] Tests backend + frontend

### 4.4 Widget Pi-hole
- [x] Créer le service backend `internal/services/pihole/` (client API)
- [x] Ajouter les types (Stats, TopDomains, QueryLog)
- [x] Implémenter : stats DNS, toggle filtrage, top domaines
- [x] Créer le handler `internal/handlers/pihole.go`
- [x] Enregistrer les routes + ServiceRegistry
- [x] Créer le composant frontend `web/src/lib/components/pihole/`
- [x] Widget : requêtes totales, bloquées, %, toggle on/off
- [x] Enregistrer dans `widgetRegistry.ts`
- [x] Ajouter les traductions fr/en
- [x] Tests backend + frontend

### 4.5 Widget Météo
- [x] Créer le service backend `internal/services/weather/` (client OpenWeatherMap)
- [x] Ajouter les types (CurrentWeather, Forecast)
- [x] Implémenter : météo actuelle, prévisions 5 jours
- [x] Créer le handler `internal/handlers/weather.go`
- [x] Enregistrer les routes
- [x] Créer le composant frontend `web/src/lib/components/weather/`
- [x] Widget : température, icône météo, humidité, vent, prévisions
- [x] Config : ville ou coordonnées GPS, unités (°C/°F)
- [x] Enregistrer dans `widgetRegistry.ts`
- [x] Ajouter les traductions fr/en
- [x] Tests backend + frontend

### 4.6 Widget Calendrier
- [x] Créer le service backend `internal/services/calendar/` (parser iCal/CalDAV)
- [x] Ajouter les types (Event, Calendar)
- [x] Implémenter : fetch URL iCal, parse événements, prochains RDV
- [x] Créer le handler `internal/handlers/calendar.go`
- [x] Enregistrer les routes
- [x] Créer le composant frontend `web/src/lib/components/calendar/`
- [x] Widget : vue liste des prochains événements, vue mini-calendrier
- [x] Config : URL(s) iCal, nombre de jours à afficher
- [x] Enregistrer dans `widgetRegistry.ts`
- [x] Ajouter les traductions fr/en
- [x] Tests backend + frontend

### 4.7 Widget Flux RSS
- [x] Créer le service backend `internal/services/rss/` (parser RSS/Atom)
- [x] Ajouter les types (Feed, FeedItem)
- [x] Implémenter : fetch feed, parse articles, cache
- [x] Créer le handler `internal/handlers/rss.go`
- [x] Enregistrer les routes
- [x] Créer le composant frontend `web/src/lib/components/rss/`
- [x] Widget : liste des articles récents avec titre, date, source
- [x] Config : URL(s) du feed, nombre d'articles à afficher
- [x] Enregistrer dans `widgetRegistry.ts`
- [x] Ajouter les traductions fr/en
- [x] Tests backend + frontend

### 4.8 Widget Système (machine hôte)
- [x] Créer le service backend `internal/services/system/` (stats OS)
- [x] Implémenter : CPU %, RAM %, disques, uptime, hostname
- [x] Option : stats via agent local OU via Proxmox node API
- [x] Créer le handler `internal/handlers/system.go`
- [x] Enregistrer les routes
- [x] Créer le composant frontend `web/src/lib/components/system/`
- [x] Widget : jauges CPU/RAM/disque, uptime, température (si dispo)
- [x] Enregistrer dans `widgetRegistry.ts`
- [x] Ajouter les traductions fr/en
- [x] Tests backend + frontend

### 4.9 Widget Bookmarks améliorés
- [x] Étendre le modèle AppLink en DB (groupes, tags, icône auto)
- [x] Ajouter un endpoint pour fetch automatique de favicon depuis l'URL
- [x] Créer le composant frontend amélioré
- [x] Widget : grille de liens avec favicon, groupés par tags
- [x] Config : gestion des groupes et tags
- [x] Enregistrer dans `widgetRegistry.ts`
- [x] Ajouter les traductions fr/en
- [x] Tests backend + frontend

### 4.10 Widget Reolink (caméras RTSP)
- [x] Créer le service backend `internal/services/reolink/` (découverte + client API)
- [x] Ajouter les types (Camera, StreamInfo, Snapshot)
- [x] Implémenter : découverte caméras, snapshots, URLs RTSP
- [x] Créer le handler `internal/handlers/reolink.go`
- [x] Enregistrer les routes dans `server.go`
- [x] Ajouter `reolink` dans `ServiceRegistry` (services.go)
- [x] Créer le composant frontend `web/src/lib/components/reolink/`
- [x] Widget : grille de caméras avec snapshot et stream live (WebRTC/MSE via go2rtc)
- [x] Composant config `ReolinkConfig.svelte` (sélection caméras, taille, rafraîchissement)
- [x] Enregistrer dans `widgetRegistry.ts`
- [x] Ajouter les traductions fr/en
- [x] Tests backend + frontend

### 4.11 go2rtc embarqué (streaming caméras)
- [x] Télécharger les binaires go2rtc par plateforme durant le build (linux/amd64, darwin/amd64, darwin/arm64, windows/amd64)
- [x] Inclure go2rtc dans l'image Docker (Dockerfile)
- [x] Inclure go2rtc dans les archives de release (binaires natifs)
- [x] Gestion du sous-processus : démarrage auto de go2rtc au lancement de Nidus
- [x] Génération automatique de `go2rtc.yaml` depuis les configs des widgets Reolink (URLs RTSP)
- [x] Monitoring du sous-processus : restart on crash, graceful shutdown
- [x] Auto-configuration : `go2rtc_url` = `http://localhost:1984` par défaut
- [x] Régénération de `go2rtc.yaml` et reload go2rtc quand la config caméra change
- [x] Option pour utiliser un go2rtc externe (override `go2rtc_url`)
- [x] Tests backend

---

## Phase 5 — Fonctionnalités UX

### 5.1 Multi-utilisateur & rôles
- [x] Ajouter un modèle `Role` en DB (admin, editor, viewer)
- [x] Modifier le modèle User pour inclure le rôle
- [x] Ajouter un middleware de vérification de rôle
- [x] Admin : tout accès
- [x] Editor : modifier dashboard (widgets, catégories) mais pas les settings/services
- [x] Viewer : consultation seule (edit mode masqué)
- [x] Page de gestion des utilisateurs dans les settings (admin only)
- [x] Invitation d'utilisateur (lien ou code)
- [x] Tests backend pour chaque rôle
- [x] Ajouter les traductions fr/en

### 5.2 Drag & drop grille libre
- [x] Remplacer la grille swap actuelle par un placement libre (type CSS grid-row)
- [x] Permettre le chevauchement contrôlé ou auto-push des widgets
- [x] Snap to grid avec preview de placement
- [x] Auto-compaction verticale (éviter les trous)
- [x] Tests frontend

### 5.3 Configuration YAML
- [x] Définir un schéma YAML pour la config complète (services, catégories, widgets, layout)
- [x] Endpoint d'export en YAML (`GET /api/config/yaml`)
- [x] Endpoint d'import YAML (`POST /api/config/yaml`)
- [x] Détection de fichier `config.yaml` au démarrage (alternative à la DB)
- [x] Documentation du format YAML
- [x] Tests backend

### 5.4 Notifications push
- [x] Créer un système de notification interne (événements : container down, service unreachable)
- [x] Intégration Gotify (envoi de notifications)
- [x] Intégration Ntfy (envoi de notifications)
- [x] Intégration Apprise (multi-provider)
- [x] Config dans les settings : provider, URL, token
- [x] Règles de notification configurables (quoi notifier, seuils)
- [x] Ajouter les traductions fr/en
- [x] Tests backend

### 5.5 Mode kiosk
- [x] Route `/kiosk` ou paramètre `?kiosk=true`
- [x] Masquer header, sidebar, boutons d'édition
- [x] Rotation automatique entre catégories (configurable)
- [x] Plein écran automatique
- [x] Ajouter les traductions fr/en

### 5.6 Raccourcis clavier
- [x] Définir les raccourcis (E = edit mode, 1-9 = catégories, / = recherche, ? = aide)
- [x] Implémenter un gestionnaire de raccourcis global
- [x] Afficher un modal d'aide raccourcis (touche ?)
- [x] Option pour désactiver dans les settings
- [x] Ajouter les traductions fr/en

### 5.7 Responsive amélioré
- [x] Optimiser le layout tablette (breakpoint intermédiaire)
- [x] Layout TV/grand écran (plus de colonnes, police plus grande)
- [x] Tester et corriger les widgets sur petits écrans
- [x] Touch : améliorer le drag & resize sur tactile
- [x] Tests visuels

---

## Phase 6 — Technique

### 6.1 API publique documentée
- [x] Ajouter les annotations OpenAPI/Swagger sur chaque handler Go
- [x] Générer la spec OpenAPI automatiquement (swag ou oapi-codegen)
- [x] Servir Swagger UI sur `/api/docs`
- [x] Documenter l'authentification API (JWT)
- [x] Exemples d'utilisation avec curl

### 6.2 Webhooks entrants
- [x] Créer un endpoint `POST /api/webhooks/{id}` pour recevoir des events
- [x] Modèle Webhook en DB (nom, secret, actions)
- [x] Actions configurables (notification, refresh widget, exécuter action)
- [x] Validation HMAC des webhooks
- [x] Page de gestion des webhooks dans les settings
- [x] Ajouter les traductions fr/en
- [x] Tests backend

### 6.3 Tests E2E
- [x] Installer et configurer Playwright
- [x] Test E2E : setup wizard complet
- [x] Test E2E : login / logout
- [x] Test E2E : créer catégorie, ajouter widget, modifier, supprimer
- [x] Test E2E : drag & drop, resize
- [x] Test E2E : configurer un service
- [x] Intégrer dans le CI GitHub Actions

### 6.4 CI/CD améliorations
- [x] Ajouter le linting Go (golangci-lint) dans le CI
- [x] Ajouter le linting frontend (eslint) dans le CI
- [x] Ajouter la vérification de types Svelte (svelte-check) dans le CI
- [x] Build multi-arch Docker (amd64 + arm64)
- [x] Release notes automatiques depuis les commits conventionnels

### 6.5 Plugins tiers
- [ ] Définir le format de plugin (manifest.json, composant Svelte, handler Go optionnel)
- [ ] Mécanisme de chargement de plugins depuis un dossier `plugins/`
- [ ] Intégration avec le widget registry (auto-register)
- [ ] Sandbox de sécurité (iframe ou Web Components)
- [ ] Documentation développeur pour créer un plugin
- [ ] Template de plugin starter

---

## Résumé global

| Phase | Section | Total | Finish | Pending |
|-------|---------|-------|--------|---------|
| 1 | Prérequis open source | 48 | 48 | 0 |
| 2 | Internationalisation | 19 | 19 | 0 |
| 3 | Thèmes & personnalisation | 15 | 15 | 0 |
| 4 | Nouveaux widgets | 121 | 121 | 0 |
| 5 | Fonctionnalités UX | 44 | 44 | 0 |
| 6 | Technique | 30 | 24 | 6 |
| **Total** | | **262** | **256** | **6** |
