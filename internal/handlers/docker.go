package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/portainer"
)

// DockerHandler handles Docker-related HTTP requests via Portainer.
type DockerHandler struct {
	DB    *database.DB
	Cache *cache.Cache
}

func (h *DockerHandler) getPortainerClient(ctx context.Context) (*portainer.Client, error) {
	svc, err := h.DB.GetServiceByType(ctx, "portainer")
	if err != nil && !errors.Is(err, database.ErrNotFound) {
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
			if err := client.Authenticate(ctx, authData.Username, authData.Password); err != nil {
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
		slog.Error("docker: failed to connect to Portainer", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to connect to Portainer"})
		return
	}
	if client == nil {
		slog.Warn("docker: Portainer not configured")
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Portainer not configured"})
		return
	}

	envs, err := client.ListEnvironments(r.Context())
	if err != nil {
		slog.Error("docker: failed to fetch environments", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch environments"})
		return
	}

	// Get Portainer server IP as fallback for unix socket endpoints
	portainerHost := ""
	svc, _ := h.DB.GetServiceByType(r.Context(), "portainer")
	if svc != nil {
		portainerHost = portainer.ResolveToIP(portainer.ExtractHost(svc.URL))
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
			host = portainer.ExtractHost(e.PublicURL)
			if host == "" {
				host = e.PublicURL // PublicURL might be just an IP without scheme
			}
		}
		if host == "" {
			host = portainer.ExtractHost(e.URL)
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
		slog.Error("docker getPortainerClient failed", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}
	if client == nil {
		slog.Warn("docker: Portainer not configured for containers")
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Portainer not configured"})
		return
	}

	slog.Info("docker fetching containers", "env_id", envID, "token_prefix", client.GetTokenPrefix())
	containers, err := client.ListContainers(r.Context(), envID)
	if err != nil {
		slog.Error("docker ListContainers failed", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to fetch containers"})
		return
	}

	stacks, err := client.ListStacks(r.Context(), envID)
	if err != nil {
		slog.Warn("failed to list stacks", "env_id", envID, "error", err)
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
