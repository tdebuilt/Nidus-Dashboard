package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/security"
	"github.com/tdebuilt/nidus/internal/services/rss"
)

// RSSHandler handles RSS feed HTTP requests.
type RSSHandler struct {
	Cache *cache.Cache
}

// GetFeed godoc
// @Summary Get RSS/Atom feed items
// @Description Fetches and merges items from comma-separated RSS/Atom feed URLs
// @Tags rss
// @Produce json
// @Param urls query string true "Comma-separated RSS/Atom feed URLs"
// @Param max query int false "Maximum number of items to return (default 20, max 100)"
// @Success 200 {object} rss.FeedData
// @Failure 400 {object} map[string]string
// @Failure 502 {object} map[string]string
// @Router /rss [get]
// @Security BearerAuth
func (h *RSSHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	urlsParam := r.URL.Query().Get("urls")
	if urlsParam == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing urls"})
		return
	}

	rawURLs := strings.Split(urlsParam, ",")
	var urls []string
	for _, u := range rawURLs {
		u = strings.TrimSpace(u)
		if err := security.ValidateExternalURL(u); err != nil {
			continue
		}
		urls = append(urls, u)
	}

	if len(urls) == 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "no valid urls provided"})
		return
	}

	maxItems := 20
	if m := r.URL.Query().Get("max"); m != "" {
		if parsed, err := strconv.Atoi(m); err == nil && parsed > 0 && parsed <= 100 {
			maxItems = parsed
		}
	}

	cacheKey := "rss:" + urlsParam + ":" + strconv.Itoa(maxItems)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client := rss.NewClient(nil)
	data, err := client.FetchFeeds(r.Context(), urls, maxItems)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to fetch feeds: " + err.Error()})
		return
	}

	h.Cache.Set(cacheKey, data)
	writeJSON(w, http.StatusOK, data)
}
