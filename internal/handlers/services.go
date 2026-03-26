package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
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
	"uptimekuma":    {CachePrefix: "kuma:", TestPath: "/api/status-page/heartbeat/default", DisplayName: "Uptime Kuma", AuthType: "none", NeedsURL: true},
	"plex":          {CachePrefix: "plex:", TestPath: "/identity", DisplayName: "Plex", AuthType: "token", NeedsURL: true},
	"jellyfin":      {CachePrefix: "jellyfin:", TestPath: "/System/Info/Public", DisplayName: "Jellyfin", AuthType: "token", NeedsURL: true},
	"sonarr":        {CachePrefix: "arr:", TestPath: "/api/v3/system/status", DisplayName: "Sonarr", AuthType: "apikey", NeedsURL: true},
	"radarr":        {CachePrefix: "arr:", TestPath: "/api/v3/system/status", DisplayName: "Radarr", AuthType: "apikey", NeedsURL: true},
	"lidarr":        {CachePrefix: "arr:", TestPath: "/api/v1/system/status", DisplayName: "Lidarr", AuthType: "apikey", NeedsURL: true},
	"prowlarr":      {CachePrefix: "arr:", TestPath: "/api/v1/system/status", DisplayName: "Prowlarr", AuthType: "apikey", NeedsURL: true},
	"reolink":       {CachePrefix: "reolink:", TestPath: "", DisplayName: "Reolink Cameras", AuthType: "none", NeedsURL: false},
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

// ServicesHandler handles service configuration HTTP requests.
type ServicesHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *ServicesHandler) getEncryptionKey() (string, error) {
	return h.DB.GetSystemSetting("encryption_key")
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

	services, err := h.DB.GetServices()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list services"})
		return
	}

	responses := make([]models.ServiceResponse, 0, len(services))
	for _, s := range services {
		responses = append(responses, models.ServiceResponse{
			ID:        s.ID,
			Type:      s.Type,
			Name:      s.Name,
			URL:       s.URL,
			Enabled:   s.Enabled,
			Config:    s.Config,
			HasCreds:  s.Credentials != "",
			CreatedAt: s.CreatedAt,
			UpdatedAt: s.UpdatedAt,
		})
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

	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "name is required"})
		return
	}
	def := ServiceRegistry[serviceType]
	if req.URL == "" && def.NeedsURL {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "url is required"})
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	// Encrypt credentials if provided
	encryptedCreds := ""
	if req.Credentials != "" {
		encKey, err := h.getEncryptionKey()
		if err != nil || encKey == "" {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
			return
		}
		encryptedCreds, err = crypto.Encrypt(req.Credentials, encKey)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to encrypt credentials"})
			return
		}
	}

	svc, err := h.DB.UpsertService(serviceType, req.Name, req.URL, encryptedCreds, req.Config, enabled)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to save service"})
		return
	}

	// Invalidate cached data for this service
	h.invalidateServiceCache(serviceType)

	writeJSON(w, http.StatusOK, models.ServiceResponse{
		ID:        svc.ID,
		Type:      svc.Type,
		Name:      svc.Name,
		URL:       svc.URL,
		Enabled:   svc.Enabled,
		Config:    svc.Config,
		HasCreds:  svc.Credentials != "",
		CreatedAt: svc.CreatedAt,
		UpdatedAt: svc.UpdatedAt,
	})
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

	if err := h.DB.DeleteService(serviceType); err != nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "service not found"})
		return
	}

	// Invalidate cached data for this service
	h.invalidateServiceCache(serviceType)

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

	svc, err := h.DB.GetServiceByType(serviceType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if svc == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "service not configured"})
		return
	}

	if svc.URL == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "service URL not configured"})
		return
	}

	// Real connectivity test — HTTP GET to the service URL
	testClient := &http.Client{Timeout: 5 * time.Second}
	testURL := svc.URL
	if def, ok := ServiceRegistry[serviceType]; ok && def.TestPath != "" {
		testURL += def.TestPath
	}

	resp, err := testClient.Get(testURL)
	if err != nil {
		writeJSON(w, http.StatusOK, models.TestServiceResponse{
			Success: false,
			Message: "connection failed: " + sanitizeError(err),
		})
		return
	}
	resp.Body.Close()

	// 401/403 means the server is reachable (auth issue is separate)
	// 409 is normal for Transmission (session-id)
	if resp.StatusCode >= 200 && resp.StatusCode < 500 {
		writeJSON(w, http.StatusOK, models.TestServiceResponse{
			Success: true,
			Message: fmt.Sprintf("connected successfully (HTTP %d)", resp.StatusCode),
		})
	} else {
		writeJSON(w, http.StatusOK, models.TestServiceResponse{
			Success: false,
			Message: fmt.Sprintf("server returned HTTP %d", resp.StatusCode),
		})
	}
}

// invalidateServiceCache clears cached data for a given service type.
func (h *ServicesHandler) invalidateServiceCache(serviceType string) {
	if h.Cache == nil {
		return
	}
	if def, ok := ServiceRegistry[serviceType]; ok && def.CachePrefix != "" {
		h.Cache.InvalidatePrefix(def.CachePrefix)
	}
}
