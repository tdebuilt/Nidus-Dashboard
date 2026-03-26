package server

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
	_ "github.com/tdebuilt/nidus/docs"
	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/config"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/handlers"
	nidusmw "github.com/tdebuilt/nidus/internal/middleware"
	"github.com/tdebuilt/nidus/internal/services/go2rtc"
	"github.com/tdebuilt/nidus/internal/services/notifications"
	nidusws "github.com/tdebuilt/nidus/internal/websocket"
)

// AppVersion is the application version, set from main.go.
var AppVersion = "dev"

// Go2RTCManager is the embedded go2rtc process manager (nil if not available).
var Go2RTCManager *go2rtc.Manager

// ServiceCache is the shared cache for external service responses.
var ServiceCache = cache.New(30*time.Second, time.Minute)

// AuthRateLimiter is the rate limiter for auth endpoints.
// Exported for testing purposes. Limit is raised via NIDUS_AUTH_RATE_LIMIT env var.
var AuthRateLimiter = newAuthRateLimiter()

func newAuthRateLimiter() *nidusmw.RateLimiter {
	limit := 5
	if v := os.Getenv("NIDUS_AUTH_RATE_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	return nidusmw.NewRateLimiter(limit, 15*time.Minute)
}

// WSHub is the shared WebSocket hub for real-time broadcasts.
var WSHub = nidusws.NewHub()

// StaticFiles holds the embedded frontend files.
// It is set from main.go via the web package embed.
var StaticFiles fs.FS

// New creates a chi router with all routes configured.
func New(cfg config.Config, db *database.DB) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	// Security headers
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-Frame-Options", "DENY")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("X-XSS-Protection", "0")
			w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; connect-src 'self' ws: wss:; media-src 'self' blob:; font-src 'self'")
			if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
				w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains")
			}
			next.ServeHTTP(w, r)
		})
	})

	// Start WebSocket hub
	go WSHub.Run()

	// Swagger UI (disabled by default in production)
	if os.Getenv("NIDUS_ENABLE_DOCS") == "true" {
		r.Get("/api/docs/*", httpSwagger.Handler(
			httpSwagger.URL("/api/docs/doc.json"),
		))
	}

	// API routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/health", healthHandler)
		r.Get("/version", versionHandler)

		if db != nil {
			// WebSocket endpoint (auth handled manually inside handler)
			wsHandler := &handlers.WebSocketHandler{DB: db, Hub: WSHub}
			r.Get("/ws", wsHandler.HandleWS)

			// Public webhook receive (no auth — HMAC validated inside handler)
			webhookHandler := &handlers.WebhooksHandler{DB: db, Hub: WSHub, Cache: ServiceCache, Sender: notifications.NewSender()}
			r.Post("/webhooks/{id}", webhookHandler.Receive)

			authHandler := &handlers.AuthHandler{DB: db}

			// Public auth routes (no auth required)
			r.Get("/auth/status", authHandler.Status)

			usersHandler := &handlers.UsersHandler{DB: db}

			// Public auth routes with rate limiting
			r.Group(func(r chi.Router) {
				r.Use(AuthRateLimiter.Limit)
				r.Post("/auth/setup", authHandler.Setup)
				r.Post("/auth/login", authHandler.Login)
				r.Post("/auth/register", usersHandler.Register)
				r.Post("/auth/reset-password", usersHandler.ResetPassword)
			})

			// Protected routes — all authenticated users (viewer+)
			r.Group(func(r chi.Router) {
				r.Use(nidusmw.Auth(db))
				r.Post("/auth/logout", authHandler.Logout)
				r.Post("/auth/totp/generate", authHandler.TOTPGenerate)
				r.Post("/auth/totp/enable", authHandler.TOTPEnable)
				r.Delete("/auth/totp", authHandler.TOTPDisable)
				r.Put("/auth/account", authHandler.UpdateAccount)

				// Read-only endpoints — all roles can access
				catHandler := &handlers.CategoriesHandler{DB: db}
				r.Get("/categories", catHandler.List)
				r.Get("/categories/{id}", catHandler.Get)

				widgetHandler := &handlers.WidgetsHandler{DB: db}
				if Go2RTCManager != nil {
					widgetHandler.OnReolinkChange = func() { Go2RTCManager.Reload() }
				}
				r.Get("/categories/{id}/widgets", widgetHandler.ListByCategory)

				svcHandler := &handlers.ServicesHandler{DB: db, Cache: ServiceCache}
				r.Get("/services", svcHandler.List)
				r.Get("/services/status", svcHandler.BatchStatus)

				settingsHandler := &handlers.SettingsHandler{DB: db}
				r.Get("/settings", settingsHandler.Get)

				prefsHandler := &handlers.UserPreferencesHandler{DB: db}
				r.Get("/preferences", prefsHandler.GetPreferences)
				r.Put("/preferences", prefsHandler.UpdatePreferences)

				themeHandler := &handlers.ThemesHandler{DB: db}
				r.Get("/themes", themeHandler.List)

				searchHandler := &handlers.SearchHandler{DB: db, Cache: ServiceCache}
				r.Get("/search", searchHandler.Search)

				dockerHandler := &handlers.DockerHandler{DB: db, Cache: ServiceCache}
				r.Get("/docker/environments", dockerHandler.ListEnvironments)
				r.Get("/docker/environments/{envId}/containers", dockerHandler.ListContainers)
				r.Get("/docker/environments/{envId}/stats", dockerHandler.ContainerStatsAll)
				r.Get("/docker/environments/{envId}/updates", dockerHandler.CheckUpdates)

				proxmoxHandler := &handlers.ProxmoxHandler{DB: db, Cache: ServiceCache}
				r.Get("/proxmox/nodes", proxmoxHandler.ListNodes)
				r.Get("/proxmox/vms", proxmoxHandler.ListVMs)

				haHandler := &handlers.HomeAssistantHandler{DB: db, Cache: ServiceCache, Hub: WSHub}
				r.Get("/homeassistant/entities", haHandler.ListEntities)
				r.Get("/homeassistant/entities/{entityId}", haHandler.GetEntity)
				r.Get("/homeassistant/camera/{entityId}/snapshot", haHandler.CameraSnapshot)

				adguardHandler := &handlers.AdGuardHandler{DB: db, Cache: ServiceCache}
				r.Get("/adguard/stats", adguardHandler.GetStats)

				jdHandler := &handlers.JDownloaderHandler{DB: db, Cache: ServiceCache}
				r.Get("/jdownloader/queue", jdHandler.GetQueue)

				txHandler := &handlers.TransmissionHandler{DB: db, Cache: ServiceCache}
				r.Get("/transmission/torrents", txHandler.ListTorrents)

				kumaHandler := &handlers.UptimeKumaHandler{DB: db, Cache: ServiceCache}
				r.Get("/uptimekuma/monitors/{slug}", kumaHandler.GetMonitors)

				mediaHandler := &handlers.MediaServerHandler{DB: db, Cache: ServiceCache}
				r.Get("/mediaserver/{type}/sessions", mediaHandler.GetSessions)
				r.Get("/mediaserver/{type}/libraries", mediaHandler.GetLibraries)
				r.Get("/mediaserver/{type}/proxy", mediaHandler.ProxyImage)

				weatherHandler := &handlers.WeatherHandler{Cache: ServiceCache}
				r.Get("/weather", weatherHandler.GetWeather)

				calendarHandler := &handlers.CalendarHandler{Cache: ServiceCache}
				r.Get("/calendar", calendarHandler.GetEvents)

				rssHandler := &handlers.RSSHandler{Cache: ServiceCache}
				r.Get("/rss", rssHandler.GetFeed)

				financeHandler := &handlers.FinanceHandler{DB: db, Cache: ServiceCache}
				r.Get("/finance/quotes", financeHandler.GetQuotes)
				r.Get("/finance/search", financeHandler.SearchSymbol)
				r.Get("/finance/symbol-count", financeHandler.GetSymbolCount)

				systemHandler := &handlers.SystemHandler{Cache: ServiceCache}
				r.Get("/system", systemHandler.GetStats)

				appLinkHandler := &handlers.AppLinkHandler{DB: db, Cache: ServiceCache}
				r.Get("/applinks/health", appLinkHandler.HealthCheck)
				r.Get("/applinks/favicon", appLinkHandler.Favicon)

				piholeHandler := &handlers.PiholeHandler{DB: db, Cache: ServiceCache}
				r.Get("/pihole/stats", piholeHandler.GetStats)

				arrHandler := &handlers.ArrHandler{DB: db, Cache: ServiceCache}
				r.Get("/arr/overview", arrHandler.GetOverview)

				reolinkHandler := &handlers.ReolinkHandler{DB: db, Cache: ServiceCache, Go2RTC: Go2RTCManager}
				r.Get("/reolink/cameras", reolinkHandler.ListCameras)
				r.Get("/reolink/cameras/{id}/snapshot", reolinkHandler.GetSnapshot)
				r.Get("/reolink/cameras/{id}/stream", reolinkHandler.GetStreamURL)
				r.Get("/reolink/discover", reolinkHandler.Discover)

				go2rtcHandler := &handlers.Go2RTCHandler{Manager: Go2RTCManager}
				r.Get("/go2rtc/status", go2rtcHandler.Status)
				r.Get("/go2rtc/ws", go2rtcHandler.ProxyWS)

				// Editor routes — editor+ can modify dashboard content
				r.Group(func(r chi.Router) {
					r.Use(nidusmw.RequireRole("editor"))

					r.Post("/categories", catHandler.Create)
					r.Put("/categories/reorder", catHandler.Reorder)
					r.Put("/categories/{id}", catHandler.Update)
					r.Delete("/categories/{id}", catHandler.Delete)

					r.Post("/categories/{id}/widgets", widgetHandler.Create)
					r.Put("/widgets/layout", widgetHandler.SaveLayout)
					r.Put("/widgets/{id}", widgetHandler.Update)
					r.Patch("/widgets/{id}/toggle-collapse", widgetHandler.ToggleCollapse)
					r.Delete("/widgets/{id}", widgetHandler.Delete)

					// Service interactions (not config, just usage)
					r.Post("/docker/environments/{envId}/containers/{containerId}/{action}", dockerHandler.ContainerAction)
					r.Post("/docker/environments/{envId}/containers/{containerId}/recreate", dockerHandler.RecreateContainer)
					r.Post("/docker/stacks/{stackId}/update", dockerHandler.UpdateStack)
					r.Post("/docker/stacks/{stackId}/{action}", dockerHandler.StackAction)
					r.Post("/proxmox/vms/{node}/{vmType}/{vmid}/{action}", proxmoxHandler.VMAction)
					r.Post("/homeassistant/services/{domain}/{service}", haHandler.CallService)
					r.Post("/adguard/filtering/toggle", adguardHandler.ToggleFiltering)
					r.Post("/pihole/blocking", piholeHandler.ToggleBlocking)
					r.Post("/jdownloader/links", jdHandler.AddLinks)
					r.Post("/jdownloader/queue/start", jdHandler.StartQueue)
					r.Post("/jdownloader/queue/pause", jdHandler.PauseQueue)
					r.Post("/jdownloader/queue/cleanup", jdHandler.CleanupFinished)
					r.Post("/transmission/torrents", txHandler.AddTorrent)
					r.Post("/transmission/torrents/{id}/start", txHandler.StartTorrent)
					r.Post("/transmission/torrents/{id}/stop", txHandler.StopTorrent)
					r.Post("/transmission/torrents/start-all", txHandler.StartAllTorrents)
					r.Post("/transmission/torrents/stop-all", txHandler.StopAllTorrents)
					r.Post("/transmission/torrents/cleanup", txHandler.CleanupCompleted)
				})

				// Admin routes — admin only
				r.Group(func(r chi.Router) {
					r.Use(nidusmw.RequireRole("admin"))

					r.Put("/settings", settingsHandler.Update)

					r.Post("/themes", themeHandler.Create)
					r.Put("/themes/{id}", themeHandler.Update)
					r.Delete("/themes/{id}", themeHandler.Delete)

					r.Put("/services/{type}", svcHandler.Update)
					r.Delete("/services/{type}", svcHandler.Delete)
					r.Post("/services/{type}/test", svcHandler.Test)

					cfgHandler := &handlers.ConfigHandler{DB: db}
					r.Post("/config/export", cfgHandler.Export)
					r.Post("/config/import", cfgHandler.Import)
					r.Get("/config/yaml", cfgHandler.ExportYAML)
					r.Post("/config/yaml", cfgHandler.ImportYAML)

					r.Get("/users", usersHandler.List)
					r.Put("/users/{id}/role", usersHandler.UpdateRole)
					r.Delete("/users/{id}", usersHandler.Delete)
					r.Post("/users/{id}/reset", usersHandler.CreateReset)
					r.Get("/invites", usersHandler.ListInvites)
					r.Post("/invites", usersHandler.CreateInvite)
					r.Delete("/invites/{id}", usersHandler.DeleteInvite)

					r.Get("/webhooks", webhookHandler.List)
					r.Post("/webhooks", webhookHandler.Create)
					r.Put("/webhooks/{id}", webhookHandler.Update)
					r.Delete("/webhooks/{id}", webhookHandler.Delete)
					r.Get("/webhooks/{id}/actions", webhookHandler.ListActions)
					r.Post("/webhooks/{id}/actions", webhookHandler.CreateAction)
					r.Delete("/webhooks/{id}/actions/{actionId}", webhookHandler.DeleteAction)

					notifHandler := &handlers.NotificationsHandler{DB: db, Sender: notifications.NewSender()}
					r.Get("/notifications/providers", notifHandler.ListProviders)
					r.Post("/notifications/providers", notifHandler.CreateProvider)
					r.Put("/notifications/providers/{id}", notifHandler.UpdateProvider)
					r.Delete("/notifications/providers/{id}", notifHandler.DeleteProvider)
					r.Post("/notifications/test", notifHandler.TestProvider)
					r.Get("/notifications/rules", notifHandler.ListRules)
					r.Post("/notifications/rules", notifHandler.CreateRule)
					r.Put("/notifications/rules/{id}", notifHandler.UpdateRule)
					r.Delete("/notifications/rules/{id}", notifHandler.DeleteRule)

					r.Post("/go2rtc/start", go2rtcHandler.Start)
					r.Post("/go2rtc/stop", go2rtcHandler.Stop)
					r.Post("/go2rtc/restart", go2rtcHandler.Restart)
				})
			})
		}
	})

	// Serve static files (SPA)
	if StaticFiles != nil {
		fileServer := http.FileServer(http.FS(StaticFiles))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file directly
			path := r.URL.Path
			if path == "/" {
				path = "/index.html"
			}
			// Check if file exists
			f, err := StaticFiles.Open(path[1:]) // strip leading /
			if err != nil {
				// SPA fallback: serve index.html for unknown routes
				r.URL.Path = "/"
				fileServer.ServeHTTP(w, r)
				return
			}
			f.Close()
			fileServer.ServeHTTP(w, r)
		})
	}

	return r
}

// NewFromEmbed creates a router using an embed.FS for static files.
func NewFromEmbed(cfg config.Config, db *database.DB, staticFS embed.FS, subDir string) *chi.Mux {
	sub, err := fs.Sub(staticFS, subDir)
	if err != nil {
		log.Fatalf("failed to get sub filesystem: %v", err)
	}
	StaticFiles = sub
	return New(cfg, db)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, dockerErr := os.Stat("/.dockerenv")
	isDocker := dockerErr == nil
	fmt.Fprintf(w, `{"version":%q,"is_docker":%t}`, AppVersion, isDocker)
}

// Run starts the HTTP server and blocks until a shutdown signal is received.
// It performs a graceful shutdown with a 10-second timeout.
func Run(cfg config.Config, r http.Handler) error {
	return RunWithContext(context.Background(), cfg, r)
}

// RunWithContext starts the HTTP server and blocks until the context is cancelled
// or a shutdown signal is received. Used by tests to stop the server without SIGTERM.
func RunWithContext(ctx context.Context, cfg config.Config, r http.Handler) error {
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.BindAddress, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Channel to capture server errors
	errCh := make(chan error, 1)
	go func() {
		log.Printf("Nidus server starting on :%d", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// Wait for interrupt signal or context cancellation
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		log.Printf("Received signal %s, shutting down...", sig)
	case <-ctx.Done():
		log.Println("Context cancelled, shutting down...")
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	}
	signal.Stop(quit)

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	if Go2RTCManager != nil {
		Go2RTCManager.Stop()
	}

	log.Println("Server stopped gracefully")
	return nil
}
