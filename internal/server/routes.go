package server

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/tdebuilt/nidus/docs"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/handlers"
	nidusmw "github.com/tdebuilt/nidus/internal/middleware"
)

type contextKey string

const cspNonceKey contextKey = "csp-nonce"

func generateNonce() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(b)
}

func getNonce(ctx context.Context) string {
	if nonce, ok := ctx.Value(cspNonceKey).(string); ok {
		return nonce
	}
	return ""
}

func registerMiddleware(r *chi.Mux, srv *Server, db *database.DB) {
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.CleanPath)

	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			nonce := generateNonce()
			ctx := context.WithValue(r.Context(), cspNonceKey, nonce)
			r = r.WithContext(ctx)

			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("X-XSS-Protection", "0")
			w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			csp := "default-src 'self'; script-src 'self' 'nonce-" + nonce + "'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' ws: wss:; media-src 'self' blob:; font-src 'self'"
			if frameSrc := srv.resolveCSPFrameSrc(r, db); frameSrc != "" {
				csp += "; frame-src 'self' " + frameSrc
			}
			w.Header().Set("Content-Security-Policy", csp)
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	})
}

func registerSwagger(r *chi.Mux) {
	if os.Getenv("NIDUS_ENABLE_DOCS") == "true" {
		r.Get("/api/docs/*", httpSwagger.Handler(
			httpSwagger.URL("/api/docs/doc.json"),
		))
	}
}

func registerAPIRoutes(r chi.Router, srv *Server, db *database.DB) {
	r.Use(func(next http.Handler) http.Handler {
		return http.MaxBytesHandler(next, 10<<20)
	})
	// Request timeout for non-streaming endpoints
	r.Use(func(next http.Handler) http.Handler {
		timeout := middleware.Timeout(10 * time.Second)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/api/ws" || strings.HasPrefix(r.URL.Path, "/api/go2rtc/") {
				next.ServeHTTP(w, r)
				return
			}
			timeout(next).ServeHTTP(w, r)
		})
	})
	r.Get("/health", healthHandler)
	r.Get("/version", versionHandler(srv))

	if db == nil {
		return
	}

	registerPublicRoutes(r, srv, db)
	r.Group(func(r chi.Router) {
		r.Use(nidusmw.Auth(db))
		registerViewerRoutes(r, srv, db)
		r.Group(func(r chi.Router) {
			r.Use(nidusmw.RequireRole("editor"))
			registerEditorRoutes(r, srv, db)
		})
		r.Group(func(r chi.Router) {
			r.Use(nidusmw.RequireRole("admin"))
			registerAdminRoutes(r, srv, db)
		})
	})
}

func registerPublicRoutes(r chi.Router, srv *Server, db *database.DB) {
	wsHandler := &handlers.WebSocketHandler{DB: db, Hub: srv.WSHub, BaseURL: srv.BaseURL}
	r.Get("/ws", wsHandler.HandleWS)

	webhookHandler := &handlers.WebhooksHandler{DB: db, Hub: srv.WSHub, Cache: srv.ServiceCache, Sender: srv.NotifSender}
	r.Post("/webhooks/{id}", webhookHandler.Receive)

	authHandler := &handlers.AuthHandler{DB: db}
	r.Get("/auth/status", authHandler.Status)

	usersHandler := &handlers.UsersHandler{DB: db}

	r.Group(func(r chi.Router) {
		r.Use(srv.AuthRateLimiter.Limit)
		r.Post("/auth/setup", authHandler.Setup)
		r.Post("/auth/login", authHandler.Login)
		r.Post("/auth/register", usersHandler.Register)
		r.Post("/auth/reset-password", usersHandler.ResetPassword)
	})
}

func registerViewerRoutes(r chi.Router, srv *Server, db *database.DB) {
	authHandler := &handlers.AuthHandler{DB: db}
	r.Post("/auth/logout", authHandler.Logout)
	r.Post("/auth/totp/generate", authHandler.TOTPGenerate)
	r.Post("/auth/totp/enable", authHandler.TOTPEnable)
	r.Delete("/auth/totp", authHandler.TOTPDisable)
	r.Put("/auth/account", authHandler.UpdateAccount)

	catHandler := &handlers.CategoriesHandler{DB: db}
	r.Get("/categories", catHandler.List)
	r.Get("/categories/{id}", catHandler.Get)

	widgetHandler := &handlers.WidgetsHandler{DB: db}
	if srv.Go2RTCManager != nil {
		widgetHandler.OnReolinkChange = func() { srv.Go2RTCManager.Reload() }
	}
	r.Get("/categories/{id}/widgets", widgetHandler.ListByCategory)

	svcHandler := &handlers.ServicesHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/services", svcHandler.List)
	r.Get("/services/status", svcHandler.BatchStatus)

	settingsHandler := &handlers.SettingsHandler{DB: db}
	r.Get("/settings", settingsHandler.Get)

	prefsHandler := &handlers.UserPreferencesHandler{DB: db}
	r.Get("/preferences", prefsHandler.GetPreferences)
	r.Put("/preferences", prefsHandler.UpdatePreferences)

	themeHandler := &handlers.ThemesHandler{DB: db}
	r.Get("/themes", themeHandler.List)

	searchHandler := &handlers.SearchHandler{DB: db, Cache: srv.ServiceCache}
	r.Get("/search", searchHandler.Search)

	registerViewerServiceRoutes(r, srv, db)
}

func registerEditorRoutes(r chi.Router, srv *Server, db *database.DB) {
	catHandler := &handlers.CategoriesHandler{DB: db}
	r.Post("/categories", catHandler.Create)
	r.Put("/categories/reorder", catHandler.Reorder)
	r.Put("/categories/{id}", catHandler.Update)
	r.Delete("/categories/{id}", catHandler.Delete)

	widgetHandler := &handlers.WidgetsHandler{DB: db}
	if srv.Go2RTCManager != nil {
		widgetHandler.OnReolinkChange = func() { srv.Go2RTCManager.Reload() }
	}
	r.Post("/categories/{id}/widgets", widgetHandler.Create)
	r.Put("/widgets/layout", widgetHandler.SaveLayout)
	r.Put("/widgets/{id}", widgetHandler.Update)
	r.Patch("/widgets/{id}/toggle-collapse", widgetHandler.ToggleCollapse)
	r.Delete("/widgets/{id}", widgetHandler.Delete)

	dockerHandler := &handlers.DockerHandler{DB: db, Cache: srv.ServiceCache, BgOps: &srv.DockerBgOps}
	r.Post("/docker/environments/{envId}/containers/{containerId}/{action}", dockerHandler.ContainerAction)
	r.Post("/docker/environments/{envId}/containers/{containerId}/recreate", dockerHandler.RecreateContainer)
	r.Post("/docker/stacks/{stackId}/update", dockerHandler.UpdateStack)
	r.Post("/docker/stacks/{stackId}/{action}", dockerHandler.StackAction)

	proxmoxHandler := &handlers.ProxmoxHandler{DB: db, Cache: srv.ServiceCache}
	r.Post("/proxmox/vms/{node}/{vmType}/{vmid}/{action}", proxmoxHandler.VMAction)

	haHandler := &handlers.HomeAssistantHandler{DB: db, Cache: srv.ServiceCache, Hub: srv.WSHub}
	r.Post("/homeassistant/services/{domain}/{service}", haHandler.CallService)

	adguardHandler := &handlers.AdGuardHandler{DB: db, Cache: srv.ServiceCache}
	r.Post("/adguard/filtering/toggle", adguardHandler.ToggleFiltering)

	piholeHandler := &handlers.PiholeHandler{DB: db, Cache: srv.ServiceCache}
	r.Post("/pihole/blocking", piholeHandler.ToggleBlocking)

	jdHandler := &handlers.JDownloaderHandler{DB: db, Cache: srv.ServiceCache}
	r.Post("/jdownloader/links", jdHandler.AddLinks)
	r.Post("/jdownloader/queue/start", jdHandler.StartQueue)
	r.Post("/jdownloader/queue/pause", jdHandler.PauseQueue)
	r.Post("/jdownloader/queue/cleanup", jdHandler.CleanupFinished)

	txHandler := &handlers.TransmissionHandler{DB: db, Cache: srv.ServiceCache}
	r.Post("/transmission/torrents", txHandler.AddTorrent)
	r.Post("/transmission/torrents/{id}/start", txHandler.StartTorrent)
	r.Post("/transmission/torrents/{id}/stop", txHandler.StopTorrent)
	r.Post("/transmission/torrents/start-all", txHandler.StartAllTorrents)
	r.Post("/transmission/torrents/stop-all", txHandler.StopAllTorrents)
	r.Post("/transmission/torrents/cleanup", txHandler.CleanupCompleted)

	arrHandler := &handlers.ArrHandler{DB: db, Cache: srv.ServiceCache}
	r.Post("/arr/{type}/add", arrHandler.AddMedia)
	r.Post("/arr/prowlarr/indexer/{id}/test", arrHandler.TestProwlarrIndexer)

	qbtHandler := &handlers.QBittorrentHandler{DB: db, Cache: srv.ServiceCache}
	r.Post("/qbittorrent/torrents", qbtHandler.AddTorrent)
	r.Post("/qbittorrent/torrents/{hash}/resume", qbtHandler.ResumeTorrent)
	r.Post("/qbittorrent/torrents/{hash}/pause", qbtHandler.PauseTorrent)
	r.Post("/qbittorrent/torrents/{hash}/delete", qbtHandler.DeleteTorrent)
	r.Post("/qbittorrent/torrents/resume-all", qbtHandler.ResumeAllTorrents)
	r.Post("/qbittorrent/torrents/pause-all", qbtHandler.PauseAllTorrents)
	r.Post("/qbittorrent/torrents/cleanup", qbtHandler.CleanupCompleted)
}

func registerAdminRoutes(r chi.Router, srv *Server, db *database.DB) {
	settingsHandler := &handlers.SettingsHandler{DB: db}
	r.Put("/settings", settingsHandler.Update)

	themeHandler := &handlers.ThemesHandler{DB: db}
	r.Post("/themes", themeHandler.Create)
	r.Put("/themes/{id}", themeHandler.Update)
	r.Delete("/themes/{id}", themeHandler.Delete)

	svcHandler := &handlers.ServicesHandler{DB: db, Cache: srv.ServiceCache}
	r.Put("/services/{type}", svcHandler.Update)
	r.Delete("/services/{type}", svcHandler.Delete)
	r.Post("/services/{type}/test", svcHandler.Test)

	cfgHandler := &handlers.ConfigHandler{DB: db}
	r.Post("/config/export", cfgHandler.Export)
	r.Post("/config/import", cfgHandler.Import)
	r.Get("/config/yaml", cfgHandler.ExportYAML)
	r.Post("/config/yaml", cfgHandler.ImportYAML)

	usersHandler := &handlers.UsersHandler{DB: db}
	r.Get("/users", usersHandler.List)
	r.Put("/users/{id}/role", usersHandler.UpdateRole)
	r.Delete("/users/{id}", usersHandler.Delete)
	r.Post("/users/{id}/reset", usersHandler.CreateReset)
	r.Get("/invites", usersHandler.ListInvites)
	r.Post("/invites", usersHandler.CreateInvite)
	r.Delete("/invites/{id}", usersHandler.DeleteInvite)

	webhookHandler := &handlers.WebhooksHandler{DB: db, Hub: srv.WSHub, Cache: srv.ServiceCache, Sender: srv.NotifSender}
	r.Get("/webhooks", webhookHandler.List)
	r.Post("/webhooks", webhookHandler.Create)
	r.Put("/webhooks/{id}", webhookHandler.Update)
	r.Delete("/webhooks/{id}", webhookHandler.Delete)
	r.Get("/webhooks/{id}/actions", webhookHandler.ListActions)
	r.Post("/webhooks/{id}/actions", webhookHandler.CreateAction)
	r.Delete("/webhooks/{id}/actions/{actionId}", webhookHandler.DeleteAction)

	notifHandler := &handlers.NotificationsHandler{DB: db, Sender: srv.NotifSender}
	r.Get("/notifications/providers", notifHandler.ListProviders)
	r.Post("/notifications/providers", notifHandler.CreateProvider)
	r.Put("/notifications/providers/{id}", notifHandler.UpdateProvider)
	r.Delete("/notifications/providers/{id}", notifHandler.DeleteProvider)
	r.Post("/notifications/test", notifHandler.TestProvider)
	r.Get("/notifications/rules", notifHandler.ListRules)
	r.Post("/notifications/rules", notifHandler.CreateRule)
	r.Put("/notifications/rules/{id}", notifHandler.UpdateRule)
	r.Delete("/notifications/rules/{id}", notifHandler.DeleteRule)

	go2rtcHandler := &handlers.Go2RTCHandler{Manager: srv.Go2RTCManager}
	r.Post("/go2rtc/start", go2rtcHandler.Start)
	r.Post("/go2rtc/stop", go2rtcHandler.Stop)
	r.Post("/go2rtc/restart", go2rtcHandler.Restart)
}

// resolveCSPFrameSrc returns a sanitized frame-src origin for CSP, checking
// the service cache first and falling back to a database lookup.
// Only scheme://host is returned to prevent CSP header injection.
func (srv *Server) resolveCSPFrameSrc(r *http.Request, db *database.DB) string {
	if frameSrc, ok := srv.ServiceCache.Get("csp:frame-src"); ok {
		if s, ok := frameSrc.(string); ok {
			return s
		}
	}
	if db != nil {
		if svc, _ := db.GetServiceByType(r.Context(), "grafana"); svc != nil && svc.URL != "" {
			sanitized := sanitizeCSPOrigin(svc.URL)
			if sanitized != "" {
				srv.ServiceCache.Set("csp:frame-src", sanitized)
				return sanitized
			}
		}
	}
	return ""
}

// sanitizeCSPOrigin extracts only scheme://host from a URL to prevent CSP injection.
func sanitizeCSPOrigin(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host
}
