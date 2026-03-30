package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/security"
	"github.com/tdebuilt/nidus/internal/services/calendar"
)

// CalendarHandler handles calendar-related HTTP requests.
type CalendarHandler struct {
	Cache *cache.Cache
}

// GetEvents godoc
// @Summary Get upcoming calendar events from iCal feeds
// @Description Fetches and merges events from comma-separated iCal URLs
// @Tags calendar
// @Produce json
// @Param urls query string true "Comma-separated iCal feed URLs"
// @Param days query int false "Number of days to look ahead (default 14, max 90)"
// @Success 200 {object} calendar.CalendarData
// @Failure 400 {object} map[string]string
// @Failure 502 {object} map[string]string
// @Router /calendar [get]
// @Security BearerAuth
func (h *CalendarHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
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

	days := 14
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}

	// Cache key based on URLs and days
	cacheKey := "calendar:" + urlsParam + ":" + strconv.Itoa(days)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client := calendar.NewClient(nil)
	data, err := client.FetchEvents(r.Context(), urls, days)
	if err != nil {
		slog.Error("calendar: failed to fetch events", "error", err)
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "failed to fetch calendar: " + err.Error()})
		return
	}

	h.Cache.Set(cacheKey, data)
	writeJSON(w, http.StatusOK, data)
}
