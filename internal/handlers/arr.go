package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/services/arr"
)

// ArrHandler handles *arr stack HTTP requests.
type ArrHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

var arrServiceTypes = []string{"sonarr", "radarr", "lidarr", "prowlarr"}

func (h *ArrHandler) getArrClient(ctx context.Context, serviceType string) (*arr.Client, error) {
	svc, err := h.DB.GetServiceByType(ctx, serviceType)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	apiKey := ""
	if svc.Credentials != "" {
		encKey, err := h.DB.GetSystemSetting(ctx, "encryption_key")
		if err != nil || encKey == "" {
			return nil, err
		}
		creds, err := crypto.Decrypt(svc.Credentials, encKey)
		if err != nil {
			return nil, err
		}
		var authData struct {
			APIKey string `json:"api_key"`
		}
		if err := json.Unmarshal([]byte(creds), &authData); err == nil {
			apiKey = authData.APIKey
		}
	}

	cfg, ok := arr.Configs[serviceType]
	if !ok {
		return nil, nil
	}

	return arr.NewClient(svc.URL, apiKey, cfg.APIVersion, nil), nil
}

func (h *ArrHandler) fetchServiceOverview(ctx context.Context, serviceType string) arr.ArrOverview {
	cfg := arr.Configs[serviceType]
	overview := arr.ArrOverview{
		Type:  serviceType,
		Label: cfg.Label,
	}

	client, err := h.getArrClient(ctx, serviceType)
	if err != nil {
		overview.Error = "connection error"
		return overview
	}
	if client == nil {
		overview.Error = "not_configured"
		return overview
	}

	status, err := client.GetSystemStatus(ctx)
	if err == nil {
		overview.Status = status
	}

	queue, err := client.GetQueue(ctx, 20)
	if err == nil {
		overview.QueueCount = queue.TotalRecords
		overview.QueueItems = queue.Records
	}

	if cfg.CalendarPath != "" {
		now := time.Now()
		end := now.AddDate(0, 0, 7)
		cal, err := client.GetCalendar(ctx, now, end)
		if err == nil {
			overview.CalendarItems = cal
		}
	}

	count, err := client.GetLibraryCount(ctx, cfg.LibraryPath)
	if err == nil {
		overview.LibraryCount = count
	}

	return overview
}

// GetOverview godoc
// @Summary Get overview of all configured *arr services
// @Description Fetches status, queue, calendar, and library counts from Sonarr, Radarr, Lidarr, and Prowlarr
// @Tags arr
// @Produce json
// @Success 200 {array} arr.ArrOverview
// @Router /arr/overview [get]
// @Security BearerAuth
func (h *ArrHandler) GetOverview(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.Cache.Get("arr:overview"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	var wg sync.WaitGroup
	results := make([]arr.ArrOverview, len(arrServiceTypes))

	for i, svcType := range arrServiceTypes {
		wg.Add(1)
		go func(idx int, st string) {
			defer wg.Done()
			results[idx] = h.fetchServiceOverview(r.Context(), st)
		}(i, svcType)
	}
	wg.Wait()

	// Filter to only configured services
	var configured []arr.ArrOverview
	for _, ov := range results {
		if ov.Error != "not_configured" {
			configured = append(configured, ov)
		}
	}

	if configured == nil {
		configured = []arr.ArrOverview{}
	}

	h.Cache.Set("arr:overview", configured)
	writeJSON(w, http.StatusOK, configured)
}
