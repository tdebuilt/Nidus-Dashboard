package handlers

import (
	"log/slog"
	"net/http"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/services/system"
)

// SystemHandler handles system stats HTTP requests.
type SystemHandler struct {
	Cache *cache.Cache
}

// GetStats godoc
// @Summary Get host system statistics
// @Description Returns CPU, RAM, disk, uptime, and hostname stats of the host machine
// @Tags system
// @Produce json
// @Success 200 {object} system.SystemStats
// @Failure 500 {object} map[string]string
// @Router /system [get]
// @Security BearerAuth
func (h *SystemHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	cacheKey := "system:stats"
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	stats, err := system.GetStats()
	if err != nil {
		slog.Error("system: failed to read system stats", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to read system stats: " + sanitizeError(err)})
		return
	}

	h.Cache.Set(cacheKey, stats)
	writeJSON(w, http.StatusOK, stats)
}
