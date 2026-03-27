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
	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/config"
	"github.com/tdebuilt/nidus/internal/database"
	nidusmw "github.com/tdebuilt/nidus/internal/middleware"
	"github.com/tdebuilt/nidus/internal/services/go2rtc"
	nidusws "github.com/tdebuilt/nidus/internal/websocket"
)

const (
	serverReadTimeout    = 15 * time.Second
	serverWriteTimeout   = 15 * time.Second
	serverIdleTimeout    = 60 * time.Second
	shutdownTimeout      = 10 * time.Second
	defaultCacheTTL      = 30 * time.Second
	cacheCleanupInterval = time.Minute
	authRateLimitWindow  = 15 * time.Minute
)

// Server holds all shared dependencies for the Nidus server.
type Server struct {
	AppVersion      string
	Go2RTCManager   *go2rtc.Manager
	ServiceCache    *cache.Cache
	AuthRateLimiter *nidusmw.RateLimiter
	WSHub           *nidusws.Hub
	StaticFiles     fs.FS
}

// NewServer creates a Server with default values for cache, hub, and rate limiter.
func NewServer(version string) *Server {
	return &Server{
		AppVersion:      version,
		ServiceCache:    cache.New(defaultCacheTTL, cacheCleanupInterval),
		AuthRateLimiter: newAuthRateLimiter(),
		WSHub:           nidusws.NewHub(),
	}
}

func newAuthRateLimiter() *nidusmw.RateLimiter {
	limit := 5
	if v := os.Getenv("NIDUS_AUTH_RATE_LIMIT"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	return nidusmw.NewRateLimiter(limit, authRateLimitWindow)
}

// New creates a chi router with all routes configured.
func New(srv *Server, cfg config.Config, db *database.DB) *chi.Mux {
	r := chi.NewRouter()
	registerMiddleware(r, srv, db)
	go srv.WSHub.Run()
	registerSwagger(r)
	r.Route("/api", func(r chi.Router) {
		registerAPIRoutes(r, srv, db)
	})
	serveStaticFiles(r, srv)
	return r
}

// NewFromEmbed creates a router using an embed.FS for static files.
func NewFromEmbed(srv *Server, cfg config.Config, db *database.DB, staticFS embed.FS, subDir string) *chi.Mux {
	sub, err := fs.Sub(staticFS, subDir)
	if err != nil {
		log.Fatalf("failed to get sub filesystem: %v", err)
	}
	srv.StaticFiles = sub
	return New(srv, cfg, db)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func versionHandler(srv *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, dockerErr := os.Stat("/.dockerenv")
		isDocker := dockerErr == nil
		fmt.Fprintf(w, `{"version":%q,"is_docker":%t}`, srv.AppVersion, isDocker)
	}
}

// Run starts the HTTP server and blocks until a shutdown signal is received.
func Run(srv *Server, cfg config.Config, r http.Handler) error {
	return RunWithContext(context.Background(), srv, cfg, r)
}

// RunWithContext starts the HTTP server and blocks until the context is cancelled
// or a shutdown signal is received. Used by tests to stop the server without SIGTERM.
func RunWithContext(ctx context.Context, srv *Server, cfg config.Config, r http.Handler) error {
	httpSrv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.BindAddress, cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
		IdleTimeout:  serverIdleTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("Nidus server starting on :%d", cfg.Server.Port)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := httpSrv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	stopServices(srv)
	log.Println("Server stopped gracefully")
	return nil
}

func stopServices(srv *Server) {
	if srv.Go2RTCManager != nil {
		srv.Go2RTCManager.Stop()
	}
	srv.WSHub.Stop()
	srv.ServiceCache.Stop()
	srv.AuthRateLimiter.Stop()
}
