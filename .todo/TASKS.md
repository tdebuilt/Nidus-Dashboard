# Tasks

## Next Up

### 1. Pagination for Transmission & JDownloader widgets
- Add a reusable `Pager` component (prev/next, page indicator)
- Transmission: paginate the torrent list (default 10 items per page, configurable in widget config)
- JDownloader: paginate the package list (same approach)
- Keep the existing `overflow-y-auto` scroll within each page for consistency
- Add "items per page" option in widget config dialog (5, 10, 20, 50)
- No backend changes needed (lists are already small enough client-side)

### 2. Search/filter for Transmission & JDownloader
- Add a compact search input above the list (filter by name)
- Debounced input (300ms) to avoid excessive re-renders
- Search should work alongside pagination (filter first, then paginate)
- Show "no results" message when filter matches nothing

### 3. Docker container logs viewer
- Add a "Logs" action button on ContainerCard (alongside start/stop/restart)
- Open a modal/drawer with tailing logs (last 100 lines)
- Backend: new endpoint `GET /api/services/{id}/docker/containers/{containerId}/logs`
- Use Portainer API for log retrieval
- Auto-scroll to bottom, monospace font, basic ANSI color support

### 4. Per-widget refresh button
- Add a manual refresh icon button in WidgetCardHeader (visible in view mode)
- Triggers immediate data fetch (bypasses polling interval)
- Brief spin animation on the icon during fetch
- Works for all widget types via a generic callback pattern

### 5. Widget detail modal for Transmission & JDownloader
- Click on a torrent/package item to open a detail dialog
- Show: full name, progress bar, size, download/upload speed, ETA, peers, files list
- Transmission: show individual file progress within a torrent
- JDownloader: show individual links within a package

### 6. Docker pause/unpause action
- Add pause/unpause button in ContainerCard (toggle based on container state)
- Backend: new endpoint `POST /api/services/{id}/docker/containers/{containerId}/pause`
- Visual indicator for paused state (distinct from stopped)

### 7. Notification delivery history
- Store notification delivery attempts in SQLite (provider, event, status, timestamp)
- New tab or section in NotificationsTab showing recent deliveries
- Show success/failure status with error messages
- Auto-cleanup of entries older than 30 days

## Backlog

### Sorting options for list widgets
- Add sort controls to Transmission (by name, progress, size, speed, ETA)
- Add sort controls to JDownloader (by name, progress, size)
- Persist sort preference in widget config

### Per-widget polling interval
- Allow overriding the global polling interval per widget in config
- Useful for high-frequency widgets (transmission: 5s) vs low-frequency (weather: 5min)
- UI: optional field in widget edit dialog

### Dark mode auto-detection
- Detect system preference via `prefers-color-scheme` media query
- Add "Auto" option in appearance settings (alongside manual theme selection)
- Switch theme automatically when system preference changes

### Bulk container actions
- Select multiple containers in Docker widget
- Apply start/stop/restart to selection
- Confirmation dialog before bulk actions
