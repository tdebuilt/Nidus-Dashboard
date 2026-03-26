package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/security"
)

// AppLinkHandler handles app link health check requests.
type AppLinkHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

// HealthCheck godoc
// @Summary Check if an app link URL is reachable
// @Tags applinks
// @Produce json
// @Param url query string true "URL to check"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Router /applinks/health [get]
// @Security BearerAuth
func (h *AppLinkHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")
	if url == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "missing url parameter"})
		return
	}

	if err := security.ValidateExternalURL(url); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	cacheKey := fmt.Sprintf("applink:health:%s", url)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	result := h.checkHealth(url)
	h.Cache.Set(cacheKey, result)
	writeJSON(w, http.StatusOK, result)
}

func (h *AppLinkHandler) checkHealth(url string) map[string]any {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	start := time.Now()
	resp, err := client.Get(url)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		return map[string]any{
			"status": "down",
			"error":  err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return map[string]any{
			"status":           "up",
			"response_time_ms": elapsed,
		}
	}

	return map[string]any{
		"status": "down",
		"error":  fmt.Sprintf("HTTP %d", resp.StatusCode),
	}
}

var linkIconRe = regexp.MustCompile(`(?i)<link[^>]+rel=["'](?:shortcut )?icon["'][^>]*>`)
var hrefRe = regexp.MustCompile(`(?i)href=["']([^"']+)["']`)

// Favicon godoc
// @Summary Fetch and proxy a site's favicon
// @Tags applinks
// @Produce image/*
// @Param url query string true "Site URL to fetch favicon from"
// @Success 200 {file} binary
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /applinks/favicon [get]
// @Security BearerAuth
func (h *AppLinkHandler) Favicon(w http.ResponseWriter, r *http.Request) {
	rawURL := r.URL.Query().Get("url")
	if rawURL == "" {
		http.Error(w, "missing url parameter", http.StatusBadRequest)
		return
	}

	if err := security.ValidateExternalURL(rawURL); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("applink:favicon:%s", rawURL)
	if val, ok := h.Cache.Get(cacheKey); ok {
		cached := val.(cachedFavicon)
		w.Header().Set("Content-Type", cached.contentType)
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Write(cached.data)
		return
	}

	faviconURL, data, contentType := h.fetchFavicon(rawURL)
	if data == nil {
		w.Header().Set("Cache-Control", "no-store")
		http.Error(w, "favicon not found", http.StatusNotFound)
		return
	}

	_ = faviconURL
	h.Cache.Set(cacheKey, cachedFavicon{data: data, contentType: contentType})
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Write(data)
}

type cachedFavicon struct {
	data        []byte
	contentType string
}

func (h *AppLinkHandler) fetchFavicon(rawURL string) (string, []byte, string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", nil, ""
	}
	origin := fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)

	// Try parsing HTML for <link rel="icon"> first
	resp, err := client.Get(rawURL)
	if err == nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
		if err == nil {
			if faviconURL := extractFaviconURL(string(body), origin); faviconURL != "" {
				if data, ct := downloadURL(client, faviconURL); data != nil {
					return faviconURL, data, ct
				}
			}
		}
	}

	// Fallback: try /favicon.ico
	fallback := origin + "/favicon.ico"
	if data, ct := downloadURL(client, fallback); data != nil {
		return fallback, data, ct
	}

	return "", nil, ""
}

func extractFaviconURL(html, origin string) string {
	match := linkIconRe.FindString(html)
	if match == "" {
		return ""
	}
	hrefMatch := hrefRe.FindStringSubmatch(match)
	if len(hrefMatch) < 2 {
		return ""
	}
	href := hrefMatch[1]

	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	if strings.HasPrefix(href, "/") {
		return origin + href
	}
	if strings.HasPrefix(href, "http") {
		return href
	}
	return origin + "/" + href
}

func downloadURL(client *http.Client, u string) ([]byte, string) {
	resp, err := client.Get(u)
	if err != nil {
		return nil, ""
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ""
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024))
	if err != nil || len(data) == 0 {
		return nil, ""
	}

	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		ct = "image/x-icon"
	}
	return data, ct
}
