package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/handlers"
)

func registerViewerServiceRoutes(r chi.Router, srv *Server, db *database.DB) {
	dockerHandler := &handlers.DockerHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/docker/environments", dockerHandler.ListEnvironments)
	r.Get("/docker/environments/{envId}/containers", dockerHandler.ListContainers)
	r.Get("/docker/environments/{envId}/stats", dockerHandler.ContainerStatsAll)
	r.Get("/docker/environments/{envId}/updates", dockerHandler.CheckUpdates)

	proxmoxHandler := &handlers.ProxmoxHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/proxmox/nodes", proxmoxHandler.ListNodes)
	r.Get("/proxmox/vms", proxmoxHandler.ListVMs)

	haHandler := &handlers.HomeAssistantHandler{DB: db, Cache: srv.ServiceCache, Hub: srv.WSHub}
	r.Get("/homeassistant/entities", haHandler.ListEntities)
	r.Get("/homeassistant/entities/{entityId}", haHandler.GetEntity)
	r.Get("/homeassistant/camera/{entityId}/snapshot", haHandler.CameraSnapshot)

	adguardHandler := &handlers.AdGuardHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/adguard/stats", adguardHandler.GetStats)

	grafanaHandler := &handlers.GrafanaHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/grafana/dashboards", grafanaHandler.ListDashboards)
	r.Get("/grafana/dashboards/{uid}/panels", grafanaHandler.GetDashboardPanels)
	r.Get("/grafana/embed-url", grafanaHandler.GetEmbedURL)

	jdHandler := &handlers.JDownloaderHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/jdownloader/queue", jdHandler.GetQueue)

	txHandler := &handlers.TransmissionHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/transmission/torrents", txHandler.ListTorrents)

	kumaHandler := &handlers.UptimeKumaHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/uptimekuma/monitors/{slug}", kumaHandler.GetMonitors)

	mediaHandler := &handlers.MediaServerHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/mediaserver/{type}/sessions", mediaHandler.GetSessions)
	r.Get("/mediaserver/{type}/libraries", mediaHandler.GetLibraries)
	r.Get("/mediaserver/{type}/proxy", mediaHandler.ProxyImage)

	weatherHandler := &handlers.WeatherHandler{Cache: srv.ServiceCache}
	r.Get("/weather", weatherHandler.GetWeather)

	calendarHandler := &handlers.CalendarHandler{Cache: srv.ServiceCache}
	r.Get("/calendar", calendarHandler.GetEvents)

	rssHandler := &handlers.RSSHandler{Cache: srv.ServiceCache}
	r.Get("/rss", rssHandler.GetFeed)

	financeHandler := &handlers.FinanceHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/finance/quotes", financeHandler.GetQuotes)
	r.Get("/finance/search", financeHandler.SearchSymbol)
	r.Get("/finance/symbol-count", financeHandler.GetSymbolCount)

	systemHandler := &handlers.SystemHandler{Cache: srv.ServiceCache}
	r.Get("/system", systemHandler.GetStats)

	appLinkHandler := &handlers.AppLinkHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/applinks/health", appLinkHandler.HealthCheck)
	r.Get("/applinks/favicon", appLinkHandler.Favicon)

	piholeHandler := &handlers.PiholeHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/pihole/stats", piholeHandler.GetStats)

	arrHandler := &handlers.ArrHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/arr/overview", arrHandler.GetOverview)
	r.Get("/arr/{type}/library", arrHandler.GetLibrary)
	r.Get("/arr/sonarr/episodes/{seriesId}", arrHandler.GetEpisodes)
	r.Get("/arr/{type}/qualityprofiles", arrHandler.GetQualityProfiles)
	r.Get("/arr/{type}/rootfolders", arrHandler.GetRootFolders)
	r.Get("/arr/{type}/lookup", arrHandler.LookupMedia)
	r.Get("/arr/prowlarr/indexers", arrHandler.GetProwlarrIndexers)

	reolinkHandler := &handlers.ReolinkHandler{DB: db, Cache: srv.ServiceCache, Go2RTC: srv.Go2RTCManager}
	r.Get("/reolink/cameras", reolinkHandler.ListCameras)
	r.Get("/reolink/cameras/{id}/snapshot", reolinkHandler.GetSnapshot)
	r.Get("/reolink/cameras/{id}/stream", reolinkHandler.GetStreamURL)
	r.Get("/reolink/discover", reolinkHandler.Discover)

	qbtHandler := &handlers.QBittorrentHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/qbittorrent/torrents", qbtHandler.ListTorrents)
	r.Get("/qbittorrent/categories", qbtHandler.GetCategories)

	go2rtcHandler := &handlers.Go2RTCHandler{Manager: srv.Go2RTCManager}
	r.Get("/go2rtc/status", go2rtcHandler.Status)
	r.Get("/go2rtc/ws", go2rtcHandler.ProxyWS)

	terminalHandler := &handlers.TerminalHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/terminal/ws", terminalHandler.HandleWS)
}
