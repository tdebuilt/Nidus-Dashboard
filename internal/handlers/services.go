package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
)

// ServiceDefinition holds metadata for a supported service type.
type ServiceDefinition struct {
	CachePrefix string `json:"-"`
	TestPath    string `json:"-"`
	DisplayName string `json:"display_name"`
	AuthType    string `json:"auth_type"` // "token", "userpass", "apikey", "password", "jdownloader", "dual", "none"
	NeedsURL    bool   `json:"needs_url"`
}

// ServiceRegistry maps service type → definition.
var ServiceRegistry = map[string]ServiceDefinition{
	"portainer":     {CachePrefix: "docker:", TestPath: "/api/status", DisplayName: "Portainer", AuthType: "dual", NeedsURL: true},
	"proxmox":       {CachePrefix: "proxmox:", TestPath: "/api2/json/version", DisplayName: "Proxmox", AuthType: "dual", NeedsURL: true},
	"homeassistant": {CachePrefix: "ha:", TestPath: "/api/", DisplayName: "Home Assistant", AuthType: "token", NeedsURL: true},
	"adguard":       {CachePrefix: "adguard:", TestPath: "/control/status", DisplayName: "AdGuard Home", AuthType: "userpass", NeedsURL: true},
	"pihole":        {CachePrefix: "pihole:", TestPath: "/api/auth", DisplayName: "Pi-hole", AuthType: "password", NeedsURL: true},
	"jdownloader":   {CachePrefix: "jd:", TestPath: "", DisplayName: "JDownloader", AuthType: "jdownloader", NeedsURL: false},
	"transmission":  {CachePrefix: "tx:", TestPath: "/transmission/rpc", DisplayName: "Transmission", AuthType: "userpass", NeedsURL: true},
	"uptimekuma":    {CachePrefix: "kuma:", TestPath: "/api/entry-page", DisplayName: "Uptime Kuma", AuthType: "none", NeedsURL: true},
	"plex":          {CachePrefix: "plex:", TestPath: "/identity", DisplayName: "Plex", AuthType: "token", NeedsURL: true},
	"jellyfin":      {CachePrefix: "jellyfin:", TestPath: "/System/Info/Public", DisplayName: "Jellyfin", AuthType: "token", NeedsURL: true},
	"sonarr":        {CachePrefix: "arr:", TestPath: "/api/v3/system/status", DisplayName: "Sonarr", AuthType: "apikey", NeedsURL: true},
	"radarr":        {CachePrefix: "arr:", TestPath: "/api/v3/system/status", DisplayName: "Radarr", AuthType: "apikey", NeedsURL: true},
	"lidarr":        {CachePrefix: "arr:", TestPath: "/api/v1/system/status", DisplayName: "Lidarr", AuthType: "apikey", NeedsURL: true},
	"prowlarr":      {CachePrefix: "arr:", TestPath: "/api/v1/system/status", DisplayName: "Prowlarr", AuthType: "apikey", NeedsURL: true},
	"grafana":       {CachePrefix: "grafana:", TestPath: "/api/health", DisplayName: "Grafana", AuthType: "token", NeedsURL: true},
	"reolink":       {CachePrefix: "reolink:", TestPath: "", DisplayName: "Reolink Cameras", AuthType: "none", NeedsURL: false},
	"qbittorrent":   {CachePrefix: "qbt:", TestPath: "/api/v2/app/version", DisplayName: "qBittorrent", AuthType: "userpass", NeedsURL: true},
}

// ServiceTypeInfo is the API response for a service type definition.
type ServiceTypeInfo struct {
	Type        string `json:"type"`
	DisplayName string `json:"display_name"`
	AuthType    string `json:"auth_type"`
	NeedsURL    bool   `json:"needs_url"`
}

// ValidServiceTypes defines the allowed service types (derived from registry).
var ValidServiceTypes = func() map[string]bool {
	m := make(map[string]bool, len(ServiceRegistry))
	for k := range ServiceRegistry {
		m[k] = true
	}
	return m
}()

const (
	serviceTestTimeout = 5 * time.Second
	batchStatusTimeout = 10 * time.Second
)

// ServicesHandler handles service configuration HTTP requests.
type ServicesHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *ServicesHandler) getEncryptionKey(ctx context.Context) (string, error) {
	return h.DB.GetSystemSetting(ctx, "encryption_key")
}

// List godoc
// @Summary List all configured services
// @Tags services
// @Produce json
// @Success 200 {array} models.ServiceResponse "List of services"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /services [get]
// @Security BearerAuth
func (h *ServicesHandler) List(w http.ResponseWriter, r *http.Request) {
	// If ?types=true, return available service types instead of configured services
	if r.URL.Query().Get("types") == "true" {
		types := make([]ServiceTypeInfo, 0, len(ServiceRegistry))
		for k, def := range ServiceRegistry {
			types = append(types, ServiceTypeInfo{
				Type:        k,
				DisplayName: def.DisplayName,
				AuthType:    def.AuthType,
				NeedsURL:    def.NeedsURL,
			})
		}
		writeJSON(w, http.StatusOK, types)
		return
	}

	services, err := h.DB.GetServices(r.Context())
	if err != nil {
		slog.Error("services: failed to list services", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list services"})
		return
	}

	responses := make([]models.ServiceResponse, 0, len(services))
	for _, s := range services {
		responses = append(responses, toServiceResponse(&s))
	}

	writeJSON(w, http.StatusOK, responses)
}

// Update godoc
// @Summary Create or update a service configuration
// @Tags services
// @Accept json
// @Produce json
// @Param type path string true "Service type"
// @Param request body models.UpdateServiceRequest true "Service configuration"
// @Success 200 {object} models.ServiceResponse "Updated service"
// @Failure 400 {object} models.ErrorResponse "Invalid service type or request body"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /services/{type} [put]
// @Security BearerAuth
func (h *ServicesHandler) Update(w http.ResponseWriter, r *http.Request) {
	serviceType := chi.URLParam(r, "type")
	if !ValidServiceTypes[serviceType] {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid service type"})
		return
	}

	var req models.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := validateServiceRequest(req, serviceType); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	encryptedCreds, err := h.encryptCredentials(r.Context(), req.Credentials, serviceType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: sanitizeError(err)})
		return
	}

	svc, err := h.DB.UpsertService(r.Context(), serviceType, req.Name, req.URL, encryptedCreds, req.Config, enabled)
	if err != nil {
		slog.Error("services: failed to save service", "service_type", serviceType, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save service"})
		return
	}

	h.invalidateServiceCache(serviceType)
	slog.Info("services: service updated", "service_type", serviceType, "name", req.Name)
	writeJSON(w, http.StatusOK, toServiceResponse(svc))
}

func validateServiceRequest(req models.UpdateServiceRequest, serviceType string) error {
	if req.Name == "" {
		return fmt.Errorf("name is required")
	}
	if def := ServiceRegistry[serviceType]; req.URL == "" && def.NeedsURL {
		return fmt.Errorf("url is required")
	}
	return nil
}

func toServiceResponse(svc *models.Service) models.ServiceResponse {
	return models.ServiceResponse{
		ID:        svc.ID,
		Type:      svc.Type,
		Name:      svc.Name,
		URL:       svc.URL,
		Enabled:   svc.Enabled,
		Config:    svc.Config,
		HasCreds:  svc.Credentials != "",
		CreatedAt: svc.CreatedAt,
		UpdatedAt: svc.UpdatedAt,
	}
}

// Delete godoc
// @Summary Delete a service configuration
// @Tags services
// @Produce json
// @Param type path string true "Service type"
// @Success 200 {object} object "Service deleted confirmation"
// @Failure 400 {object} models.ErrorResponse "Invalid service type"
// @Failure 404 {object} models.ErrorResponse "Service not found"
// @Router /services/{type} [delete]
// @Security BearerAuth
func (h *ServicesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	serviceType := chi.URLParam(r, "type")
	if !ValidServiceTypes[serviceType] {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid service type"})
		return
	}

	if err := h.DB.DeleteService(r.Context(), serviceType); err != nil {
		slog.Warn("services: service not found for deletion", "service_type", serviceType, "error", err)
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "service not found"})
		return
	}

	// Invalidate cached data for this service
	h.invalidateServiceCache(serviceType)

	slog.Info("services: service deleted", "service_type", serviceType)
	writeJSON(w, http.StatusOK, map[string]string{"message": "service deleted"})
}

// Test godoc
// @Summary Test connectivity to a configured service
// @Tags services
// @Produce json
// @Param type path string true "Service type"
// @Success 200 {object} models.TestServiceResponse "Test result"
// @Failure 400 {object} models.ErrorResponse "Invalid service type or URL not configured"
// @Failure 404 {object} models.ErrorResponse "Service not configured"
// @Failure 500 {object} models.ErrorResponse "Database error"
// @Router /services/{type}/test [post]
// @Security BearerAuth
func (h *ServicesHandler) Test(w http.ResponseWriter, r *http.Request) {
	serviceType := chi.URLParam(r, "type")
	if !ValidServiceTypes[serviceType] {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid service type"})
		return
	}

	svc, err := h.DB.GetServiceByType(r.Context(), serviceType)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		slog.Error("services: failed to get service for test", "service_type", serviceType, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if svc == nil {
		slog.Warn("services: service not configured for test", "service_type", serviceType)
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "service not configured"})
		return
	}
	if svc.URL == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "service URL not configured"})
		return
	}

	writeJSON(w, http.StatusOK, testServiceConnectivity(svc.URL, serviceType))
}

// testServiceConnectivity probes a service URL and returns the result.
func testServiceConnectivity(baseURL, serviceType string) models.TestServiceResponse {
	testURL := baseURL
	if def, ok := ServiceRegistry[serviceType]; ok && def.TestPath != "" {
		testURL += def.TestPath
	}
	client := &http.Client{Timeout: serviceTestTimeout}
	resp, err := client.Get(testURL)
	if err != nil {
		return models.TestServiceResponse{Success: false, Message: "connection failed: " + sanitizeError(err)}
	}
	defer resp.Body.Close()
	// 401/403 means reachable (auth issue is separate); 409 is normal for Transmission
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		return models.TestServiceResponse{Success: true, Message: fmt.Sprintf("connected successfully (HTTP %d)", resp.StatusCode)}
	}
	return models.TestServiceResponse{Success: false, Message: fmt.Sprintf("server returned HTTP %d", resp.StatusCode)}
}

// BatchStatus returns connectivity status for all configured services.
func (h *ServicesHandler) BatchStatus(w http.ResponseWriter, r *http.Request) {
	services, err := h.DB.GetServices(r.Context())
	if err != nil {
		slog.Error("services: failed to list services for batch status", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list services"})
		return
	}

	statuses := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	ctx, cancel := context.WithTimeout(r.Context(), batchStatusTimeout)
	defer cancel()

	client := &http.Client{Timeout: serviceTestTimeout}

	for _, svc := range services {
		if svc.URL == "" {
			continue
		}
		wg.Add(1)
		go func(svcType, svcURL string) {
			defer wg.Done()
			reachable := checkSingleService(ctx, client, svcType, svcURL)
			mu.Lock()
			statuses[svcType] = reachable
			mu.Unlock()
		}(svc.Type, svc.URL)
	}

	wg.Wait()
	writeJSON(w, http.StatusOK, map[string]interface{}{"statuses": statuses})
}

// checkSingleService probes a service URL and returns whether it is reachable.
func checkSingleService(ctx context.Context, client *http.Client, svcType, svcURL string) bool {
	testURL := svcURL
	if def, ok := ServiceRegistry[svcType]; ok && def.TestPath != "" {
		testURL += def.TestPath
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, testURL, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 500
}

// encryptCredentials encrypts credentials if provided, returning the encrypted string.
func (h *ServicesHandler) encryptCredentials(ctx context.Context, creds, serviceType string) (string, error) {
	if creds == "" {
		return "", nil
	}
	encKey, err := h.getEncryptionKey(ctx)
	if err != nil || encKey == "" {
		slog.Error("services: encryption key not configured", "service_type", serviceType)
		return "", fmt.Errorf("encryption key not configured")
	}
	encrypted, err := crypto.Encrypt(creds, encKey)
	if err != nil {
		slog.Error("services: failed to encrypt credentials", "service_type", serviceType, "error", err)
		return "", fmt.Errorf("failed to encrypt credentials")
	}
	return encrypted, nil
}

// invalidateServiceCache clears cached data for a given service type.
func (h *ServicesHandler) invalidateServiceCache(serviceType string) {
	if h.Cache == nil {
		return
	}
	if def, ok := ServiceRegistry[serviceType]; ok && def.CachePrefix != "" {
		h.Cache.InvalidatePrefix(def.CachePrefix)
	}
	if serviceType == "grafana" {
		h.Cache.Invalidate("csp:frame-src")
	}
}
