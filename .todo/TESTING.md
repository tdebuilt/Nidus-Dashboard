# Nidus — Manual Testing Checklist

> Check each box after testing. App runs on `http://localhost:3777`.
>
> This checklist covers all implemented features. Untested items from unimplemented phases are not included.

---

## 1. Setup & Auth

- [ ] Setup page shows on first launch (`/setup`)
- [ ] Create admin account (username + password >= 8 chars)
- [ ] Error if passwords don't match
- [ ] Portainer/Proxmox steps can be skipped
- [ ] First category created in wizard
- [ ] Redirect to dashboard after setup
- [ ] Logout works
- [ ] Login page shows (`/login`)
- [ ] Login with created credentials
- [ ] Error on wrong credentials
- [ ] 2FA: enable TOTP → scan QR → code validated → 2FA active
- [ ] 2FA: login requires TOTP code
- [ ] 2FA: disable TOTP

## 2. Categories & Widgets

- [ ] Create a category from sidebar
- [ ] Rename a category
- [ ] Delete a category (confirmation required)
- [ ] Reorder categories via drag-and-drop
- [ ] Add a widget to a category
- [ ] Choose widget type from registry
- [ ] Delete a widget (confirmation required)
- [ ] Drag-and-drop widgets in grid (free placement, 12 columns)
- [ ] Resize widget (bottom-right handle)
- [ ] Snap to grid with placement preview
- [ ] Auto-compaction verticale (no gaps)
- [ ] Navigate between categories via sidebar

## 3. Services Settings

- [ ] Services listed in flat alphabetical order (no categories)
- [ ] Sort toggle (A→Z / Z→A) changes service order
- [ ] "+" button visible in section header when services available to add
- [ ] Click "+" opens add service modal dialog
- [ ] Modal shows available services as clickable cards with icons
- [ ] Sort toggle in modal works (A→Z / Z→A)
- [ ] Click a service type → config form appears (URL, credentials)
- [ ] Back button returns to service type grid
- [ ] Save → service added, modal closes, list refreshes
- [ ] Empty state shows "Add a service" button that opens modal
- [ ] Edit existing service inline (click gear icon on card)
- [ ] Test service connection (lightning bolt icon)
- [ ] Delete service (overflow menu → confirmation)

## 4. Docker (Portainer)

- [ ] Configure Portainer in Settings (URL + token or username/password)
- [ ] Test Portainer connection
- [ ] Docker widget shows "not configured" if no service
- [ ] Docker widget shows stacks grouped by Compose project
- [ ] Standalone containers shown separately
- [ ] Colored status dots (green running, red stopped)
- [ ] Start/stop/restart buttons on containers
- [ ] Recreate button (pull + recreate) with confirmation
- [ ] External link to mapped container port
- [ ] Expand/collapse stacks

## 5. Proxmox

- [ ] Configure Proxmox in Settings (URL + token or username/password)
- [ ] Proxmox widget shows VMs and LXCs
- [ ] CPU/RAM/uptime info displayed
- [ ] Start/stop/shutdown/reboot buttons
- [ ] Confirmation before stop/shutdown

## 6. Home Assistant

- [ ] Configure HA in Settings (URL + long-lived access token)
- [ ] HA widget shows entities filtered by domain
- [ ] **Light**: toggle on/off + brightness slider
- [ ] **Switch**: toggle on/off
- [ ] **Sensor**: value + unit read-only
- [ ] **Climate**: current temp, adjustable target (+/-), HVAC mode
- [ ] **Camera**: snapshot refreshed every 10s
- [ ] Widget config: filter by `entities` or `domains` in JSON
- [ ] Real-time updates via WebSocket

## 7. AdGuard

- [ ] Configure AdGuard in Settings (URL + username/password)
- [ ] Widget shows total DNS queries count
- [ ] Blocked queries count and percentage
- [ ] Average response time
- [ ] Active filters and rules count
- [ ] Toggle protection on/off (confirmation to disable)

## 8. JDownloader

- [ ] Configure JDownloader in Settings (direct URL)
- [ ] Widget shows download queue
- [ ] Progress bars per package
- [ ] Global speed displayed
- [ ] Start/pause queue buttons
- [ ] "+" button opens add links dialog
- [ ] Add links (multi-line textarea) → appear in queue

## 9. Transmission

- [ ] Configure Transmission in Settings (URL + username/password)
- [ ] Widget shows torrents with progress bars
- [ ] Global download/upload speeds
- [ ] Active/total counter
- [ ] Start/stop per torrent (confirmation for stop)
- [ ] "Start all" / "Stop all" buttons
- [ ] "+" button opens add torrent dialog
- [ ] Add magnet link → torrent appears
- [ ] Ratio displayed for seeding torrents

## 10. Uptime Kuma

- [ ] Configure Uptime Kuma in Settings (URL)
- [ ] Widget shows list of monitors
- [ ] Monitor status (up/down) with colored indicator
- [ ] Uptime percentage displayed
- [ ] Latency displayed

## 11. Media Server (Plex / Jellyfin)

- [ ] Configure Plex in Settings (URL + token)
- [ ] Configure Jellyfin in Settings (URL + API key)
- [ ] Widget shows active sessions
- [ ] Poster / thumbnail displayed
- [ ] Playback progress bar
- [ ] Config: choose between Plex and Jellyfin

## 12. Weather

- [ ] Configure weather widget (city or GPS coordinates)
- [ ] Current temperature displayed
- [ ] Weather icon displayed
- [ ] Humidity and wind info
- [ ] 5-day forecast
- [ ] Config: units (°C/°F)

## 13. Calendar (iCal)

- [ ] Configure calendar widget with iCal URL
- [ ] List view shows upcoming events
- [ ] Mini-calendar view works
- [ ] Config: number of days to display

## 14. RSS Feed

- [ ] Configure RSS widget with feed URL
- [ ] Recent articles listed with title, date, source
- [ ] Config: number of articles to display

## 15. System Monitor

- [ ] System widget shows CPU usage (gauge)
- [ ] RAM usage (gauge)
- [ ] Disk usage (gauge)
- [ ] Uptime displayed
- [ ] Hostname displayed
- [ ] Temperature (if available)

## 16. Bookmarks (App Links)

- [ ] Create bookmark widget
- [ ] Links display with favicon (auto-fetched)
- [ ] Links grouped by tags
- [ ] Click link opens in new tab
- [ ] Health check: green dot if service responds

## 17. Grafana

- [ ] Configure Grafana in Settings (URL + API token)
- [ ] Test Grafana connection
- [ ] Grafana widget shows embedded panels
- [ ] Panel picker lists dashboards and panels from Grafana API
- [ ] Multi-panel grid layout displays selected panels
- [ ] Config: select dashboard and panels to embed

## 18. Internationalization (i18n)

- [ ] Switch language in Settings (dropdown with flag + native name)
- [ ] French: all labels correct
- [ ] English: all labels correct
- [ ] Spanish: all labels correct
- [ ] German: all labels correct
- [ ] Other languages: verify key labels are translated
- [ ] Browser language auto-detected on first visit
- [ ] Language preference persists after reload

## 19. Themes & Customization

- [ ] Settings → Theme selector with live preview
- [ ] Dark theme: all elements properly styled
- [ ] Light theme: all text readable, indicators visible
- [ ] Nord theme: consistent styling
- [ ] Dracula theme: consistent styling
- [ ] Theme preference persists after reload
- [ ] Accent color picker: change primary color
- [ ] Accent color applied across all UI elements
- [ ] Custom CSS textarea in Settings
- [ ] Custom CSS injected and applied

## 20. Multi-user & Roles

- [ ] Admin can manage users in Settings
- [ ] Create invitation link/code
- [ ] Register via invitation link (`/register`)
- [ ] **Admin**: full access to everything
- [ ] **Editor**: can modify dashboard (widgets, categories) but NOT settings/services
- [ ] **Viewer**: read-only (edit mode hidden)

## 21. YAML Configuration

- [ ] Settings → Export YAML: file downloaded
- [ ] Exported YAML contains categories, widgets, services, settings
- [ ] Import YAML: config restored correctly
- [ ] Import invalid YAML → descriptive error message
- [ ] After failed import, existing data intact
- [ ] Auto-import from `config.yaml` on first startup (empty DB)

## 22. Notifications

- [ ] Settings → Notifications section (admin only)
- [ ] Add Gotify provider (URL + token)
- [ ] Add Ntfy provider (URL + topic)
- [ ] Add Apprise provider (URL)
- [ ] Test notification: sends test message to provider
- [ ] Delete provider
- [ ] Add notification rule (event type + provider)
- [ ] Delete notification rule
- [ ] Event types: container down, service unreachable

## 23. Kiosk Mode

- [ ] Navigate to `/kiosk`
- [ ] No sidebar, no header, no edit buttons
- [ ] Auto-rotation between categories (default 30s)
- [ ] Custom rotation interval via `?rotate=N` in URL
- [ ] Fullscreen triggered automatically
- [ ] Hidden top bar appears on hover (category tabs, controls)
- [ ] Click category tab → switches to that category
- [ ] Rotation toggle button (pause/resume)
- [ ] Exit button returns to dashboard
- [ ] Keyboard: Escape → exit kiosk
- [ ] Keyboard: ← / → → navigate categories
- [ ] Keyboard: R → toggle rotation
- [ ] Keyboard: F → fullscreen
- [ ] Rotation indicator dots at bottom

## 24. Search

- [ ] Search bar visible in header
- [ ] Type >= 2 chars → results in dropdown
- [ ] Results show matching categories and widgets
- [ ] Click result navigates to correct category
- [ ] Escape or click outside closes dropdown

## 25. Polling & Refresh

- [ ] Settings → Refresh interval selector
- [ ] Options: 10s, 15s, 30s, 60s, 120s, 300s
- [ ] Change interval → widgets refresh at new rate
- [ ] Setting persists after reload

## 26. JSON Export/Import

- [ ] Settings → Export JSON config → file downloaded
- [ ] Import exported JSON → config restored
- [ ] Import invalid JSON → error message
- [ ] After failed import, existing data intact

## 27. PWA & Offline

- [ ] Site is installable (button in browser address bar)
- [ ] Manifest visible in DevTools > Application > Manifest
- [ ] Service worker registered
- [ ] Cut network → "Offline" banner shown
- [ ] Cached data still displayed while offline
- [ ] Restore network → banner disappears

## 28. Mobile / Responsive

- [ ] Mobile: single column grid
- [ ] Action buttons large enough for touch (>= 44px)
- [ ] Swipe from left edge → sidebar opens
- [ ] Swipe left on sidebar → closes
- [ ] No horizontal scroll
- [ ] Dialogs usable on mobile

## 29. General

- [ ] 404 page for unknown routes
- [ ] Network error page with retry button
- [ ] Toast notifications (success/error/info) visible
- [ ] Confirm dialogs work (delete, stop, etc.)
