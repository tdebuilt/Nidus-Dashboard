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

// AppVersion is the application version, set from main.go.
var AppVersion = "dev"

// Go2RTCManager is the embedded go2rtc process manager (nil if not available).
var Go2RTCManager *go2rtc.Manager

// ServiceCache is the shared cache for external service responses.
var ServiceCache = cache.New(defaultCacheTTL, cacheCleanupInterval)

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
	return nidusmw.NewRateLimiter(limit, authRateLimitWindow)
}

// WSHub is the shared WebSocket hub for real-time broadcasts.
var WSHub = nidusws.NewHub()

// StaticFiles holds the embedded frontend files.
// It is set from main.go via the web package embed.
var StaticFiles fs.FS

// New creates a chi router with all routes configured.
func New(cfg config.Config, db *database.DB) *chi.Mux {
	r := chi.NewRouter()
	registerMiddleware(r, db)
	go WSHub.Run()
	registerSwagger(r)
	r.Route("/api", func(r chi.Router) {
		registerAPIRoutes(r, db)
	})
	serveStaticFiles(r)
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
		ReadTimeout:  serverReadTimeout,
		WriteTimeout: serverWriteTimeout,
		IdleTimeout:  serverIdleTimeout,
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
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	if Go2RTCManager != nil {
		Go2RTCManager.Stop()
	}
	WSHub.Stop()
	ServiceCache.Stop()
	AuthRateLimiter.Stop()

	log.Println("Server stopped gracefully")
	return nil
}
