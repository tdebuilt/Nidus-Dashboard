package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/portainer"
)

// extractHost extracts the hostname/IP from a URL.
func extractHost(rawURL string) string {
	if strings.HasPrefix(rawURL, "unix://") {
		return ""
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Hostname()
}

// resolveToIP resolves a hostname to an IP address. Returns the hostname as-is if already an IP.
func resolveToIP(hostname string) string {
	if hostname == "" {
		return ""
	}
	if ip := net.ParseIP(hostname); ip != nil {
		return hostname
	}
	ips, err := net.LookupHost(hostname)
	if err != nil || len(ips) == 0 {
		return hostname
	}
	return ips[0]
}

// DockerHandler handles Docker-related HTTP requests via Portainer.
type DockerHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *DockerHandler) getPortainerClient(ctx context.Context) (*portainer.Client, error) {
	svc, err := h.DB.GetServiceByType(ctx, "portainer")
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	client := portainer.NewClient(svc.URL, nil)

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
			Username string `json:"username"`
			Password string `json:"password"`
			Token    string `json:"token"`
		}
		if err := json.Unmarshal([]byte(creds), &authData); err != nil {
			// Not JSON — treat the raw string as an API token
			client.SetToken(creds)
			return client, nil
		}
		if authData.Token != "" {
			client.SetToken(authData.Token)
		} else if authData.Username != "" {
			if err := client.Authenticate(authData.Username, authData.Password); err != nil {
				return nil, err
			}
		}
	}

	return client, nil
}

// ListEnvironments godoc
// @Summary List Docker environments from Portainer
// @Tags docker
// @Produce json
// @Success 200 {array} portainer.EnvironmentInfo
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /docker/environments [get]
// @Security BearerAuth
func (h *DockerHandler) ListEnvironments(w http.ResponseWriter, r *http.Request) {
	if val, ok := h.Cache.Get("docker:environments"); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getPortainerClient(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to Portainer"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Portainer not configured"})
		return
	}

	envs, err := client.ListEnvironments()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch environments"})
		return
	}

	// Get Portainer server IP as fallback for unix socket endpoints
	portainerHost := ""
	svc, _ := h.DB.GetServiceByType(r.Context(), "portainer")
	if svc != nil {
		portainerHost = resolveToIP(extractHost(svc.URL))
	}

	result := make([]portainer.EnvironmentInfo, 0, len(envs))
	for _, e := range envs {
		status := "up"
		if e.Status != 1 {
			status = "down"
		}
		// Priority: PublicURL > endpoint URL > Portainer server IP
		host := ""
		if e.PublicURL != "" {
			host = extractHost(e.PublicURL)
			if host == "" {
				host = e.PublicURL // PublicURL might be just an IP without scheme
			}
		}
		if host == "" {
			host = extractHost(e.URL)
		}
		if host == "" {
			host = portainerHost
		}
		result = append(result, portainer.EnvironmentInfo{
			ID:     e.ID,
			Name:   e.Name,
			Status: status,
			Host:   host,
		})
	}

	h.Cache.Set("docker:environments", result)
	writeJSON(w, http.StatusOK, result)
}

// ListContainers godoc
// @Summary List containers for a Docker environment
// @Tags docker
// @Produce json
// @Param envId path int true "Environment ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /docker/environments/{envId}/containers [get]
// @Security BearerAuth
func (h *DockerHandler) ListContainers(w http.ResponseWriter, r *http.Request) {
	envID, err := strconv.Atoi(chi.URLParam(r, "envId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid environment ID"})
		return
	}

	cacheKey := "docker:containers:" + strconv.Itoa(envID)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getPortainerClient(r.Context())
	if err != nil {
		log.Printf("docker: getPortainerClient error: %v", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}
	if client == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Portainer not configured"})
		return
	}

	log.Printf("docker: fetching containers for env %d (token prefix: %.10s...)", envID, client.GetTokenPrefix())
	containers, err := client.ListContainers(envID)
	if err != nil {
		log.Printf("docker: ListContainers error: %v", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch containers"})
		return
	}

	stacks, err := client.ListStacks(envID)
	if err != nil {
		stacks = nil
	}

	grouped, standalone := portainer.GroupContainers(containers, envID)
	grouped = portainer.MergeWithPortainerStacks(grouped, stacks)

	result := map[string]any{
		"stacks":     grouped,
		"standalone": standalone,
	}

	h.Cache.Set(cacheKey, result)
	writeJSON(w, http.StatusOK, result)
}
