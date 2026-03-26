package handlers

import (
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

func (h *DockerHandler) getPortainerClient() (*portainer.Client, error) {
	svc, err := h.DB.GetServiceByType("portainer")
	if err != nil {
		return nil, err
	}
	if svc == nil {
		return nil, nil
	}

	client := portainer.NewClient(svc.URL, nil)

	if svc.Credentials != "" {
		encKey, err := h.DB.GetSystemSetting("encryption_key")
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

	client, err := h.getPortainerClient()
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
	svc, _ := h.DB.GetServiceByType("portainer")
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

	client, err := h.getPortainerClient()
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

// ContainerAction godoc
// @Summary Perform an action on a container (start, stop, restart)
// @Tags docker
// @Produce json
// @Param envId path int true "Environment ID"
// @Param containerId path string true "Container ID"
// @Param action path string true "Action to perform" Enums(start, stop, restart)
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /docker/environments/{envId}/containers/{containerId}/{action} [post]
// @Security BearerAuth
func (h *DockerHandler) ContainerAction(w http.ResponseWriter, r *http.Request) {
	envID, err := strconv.Atoi(chi.URLParam(r, "envId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid environment ID"})
		return
	}

	containerID := chi.URLParam(r, "containerId")
	action := chi.URLParam(r, "action")

	validActions := map[string]bool{
		"start": true, "stop": true, "restart": true,
	}
	if !validActions[action] {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid action"})
		return
	}

	client, err := h.getPortainerClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}

	// Workaround: Portainer's Docker proxy sends a request body on POST /start,
	// which Docker API v1.24+ rejects. Use "restart" instead — it works on both
	// running and stopped containers.
	effectiveAction := action
	if action == "start" {
		effectiveAction = "restart"
	}

	log.Printf("docker: container action %s on %s (envID: %d)", action, containerID[:12], envID)
	if err := client.ContainerAction(envID, containerID, effectiveAction); err != nil {
		log.Printf("docker: container action error: %v", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "action failed: " + err.Error()})
		return
	}

	h.Cache.InvalidatePrefix("docker:")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// RecreateContainer godoc
// @Summary Recreate a container (pull image and redeploy)
// @Tags docker
// @Accept json
// @Produce json
// @Param envId path int true "Environment ID"
// @Param containerId path string true "Container ID"
// @Param body body object false "Options" SchemaExample({"pull_image": true})
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /docker/environments/{envId}/containers/{containerId}/recreate [post]
// @Security BearerAuth
func (h *DockerHandler) RecreateContainer(w http.ResponseWriter, r *http.Request) {
	envID, err := strconv.Atoi(chi.URLParam(r, "envId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid environment ID"})
		return
	}

	containerID := chi.URLParam(r, "containerId")

	var body struct {
		PullImage bool `json:"pull_image"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	client, err := h.getPortainerClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}

	log.Printf("docker: recreate container %s (envID: %d, pull: %v)", containerID[:12], envID, body.PullImage)

	// Recreate in background — pull+restart can take minutes
	go func() {
		if err := client.RecreateContainer(envID, containerID, body.PullImage); err != nil {
			log.Printf("docker: recreate error: %v", err)
		} else {
			log.Printf("docker: recreate success: %s", containerID[:12])
		}
		h.Cache.InvalidatePrefix("docker:")
	}()

	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// StackAction godoc
// @Summary Perform an action on a stack (start or stop all containers)
// @Tags docker
// @Accept json
// @Produce json
// @Param stackId path int true "Stack ID"
// @Param action path string true "Action to perform" Enums(start, stop)
// @Param body body object false "Stack action options" SchemaExample({"env_id": 1, "stack_name": "mystack", "container_ids": ["abc123"]})
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /docker/stacks/{stackId}/{action} [post]
// @Security BearerAuth
//
// Instead of using Portainer's stack start/stop (which does compose down/up),
// we start/stop all containers in the stack individually to preserve them.
func (h *DockerHandler) StackAction(w http.ResponseWriter, r *http.Request) {
	_ = chi.URLParam(r, "stackId") // not used directly anymore

	action := chi.URLParam(r, "action")
	if action != "start" && action != "stop" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid action"})
		return
	}

	var body struct {
		EnvID        int      `json:"env_id"`
		StackName    string   `json:"stack_name"`
		ContainerIDs []string `json:"container_ids"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	client, err := h.getPortainerClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}

	containerIDs := body.ContainerIDs

	// If no container IDs provided, fetch them by listing and filtering by stack name
	if len(containerIDs) == 0 && body.StackName != "" {
		containers, err := client.ListContainers(body.EnvID)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to list containers"})
			return
		}
		for _, c := range containers {
			if c.Labels["com.docker.compose.project"] == body.StackName {
				containerIDs = append(containerIDs, c.ID)
			}
		}
	}

	if len(containerIDs) == 0 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "no containers found for stack"})
		return
	}

	log.Printf("docker: stack %s — %s %d containers (envID: %d)", body.StackName, action, len(containerIDs), body.EnvID)

	// Same workaround as ContainerAction: Portainer's Docker proxy sends a
	// request body on POST /start, which Docker API v1.24+ rejects.
	effectiveAction := action
	if action == "start" {
		effectiveAction = "restart"
	}

	var errors []string
	for _, cid := range containerIDs {
		if err := client.ContainerAction(body.EnvID, cid, effectiveAction); err != nil {
			log.Printf("docker: container %s %s error: %v", cid[:12], action, err)
			errors = append(errors, err.Error())
		}
	}

	h.Cache.InvalidatePrefix("docker:")

	if len(errors) > 0 {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":       "partial",
			"errors_count": len(errors),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// UpdateStack godoc
// @Summary Update a stack (pull images and redeploy)
// @Tags docker
// @Accept json
// @Produce json
// @Param stackId path int true "Stack ID"
// @Param body body object false "Update options" SchemaExample({"env_id": 1, "pull_image": true})
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /docker/stacks/{stackId}/update [post]
// @Security BearerAuth
//
// Redeploys the stack via Portainer (pull images + compose up).
func (h *DockerHandler) UpdateStack(w http.ResponseWriter, r *http.Request) {
	stackID, err := strconv.Atoi(chi.URLParam(r, "stackId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid stack ID"})
		return
	}

	var body struct {
		EnvID     int  `json:"env_id"`
		PullImage bool `json:"pull_image"`
	}
	json.NewDecoder(r.Body).Decode(&body)

	client, err := h.getPortainerClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}

	log.Printf("docker: update stack %d (envID: %d, pull: %v)", stackID, body.EnvID, body.PullImage)

	// Run in background — pull + redeploy can take minutes
	go func() {
		if err := client.UpdateStack(stackID, body.EnvID, body.PullImage); err != nil {
			log.Printf("docker: stack update error: %v", err)
		} else {
			log.Printf("docker: stack update success: %d", stackID)
		}
		h.Cache.InvalidatePrefix("docker:")
	}()

	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}

// ContainerStatsAll godoc
// @Summary Get CPU and memory stats for all running containers
// @Tags docker
// @Produce json
// @Param envId path int true "Environment ID"
// @Success 200 {array} portainer.ContainerStats
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /docker/environments/{envId}/stats [get]
// @Security BearerAuth
func (h *DockerHandler) ContainerStatsAll(w http.ResponseWriter, r *http.Request) {
	envID, err := strconv.Atoi(chi.URLParam(r, "envId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid environment ID"})
		return
	}

	cacheKey := "docker:stats:" + strconv.Itoa(envID)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getPortainerClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}

	containers, err := client.ListContainers(envID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to list containers"})
		return
	}

	results := make([]portainer.ContainerStats, 0)
	for _, c := range containers {
		if c.State != "running" {
			continue
		}
		stats, err := client.CalculateContainerStats(envID, c.ID)
		if err != nil {
			log.Printf("docker: stats error for %s: %v", c.ID[:12], err)
			continue
		}
		results = append(results, *stats)
	}

	h.Cache.Set(cacheKey, results)
	writeJSON(w, http.StatusOK, results)
}

// CheckUpdates godoc
// @Summary Check for container image updates
// @Description Compares local image digests with remote registry digests
// @Tags docker
// @Produce json
// @Param envId path int true "Environment ID"
// @Success 200 {array} object
// @Failure 400 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /docker/environments/{envId}/updates [get]
// @Security BearerAuth
func (h *DockerHandler) CheckUpdates(w http.ResponseWriter, r *http.Request) {
	envID, err := strconv.Atoi(chi.URLParam(r, "envId"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid environment ID"})
		return
	}

	cacheKey := "docker:updates:" + strconv.Itoa(envID)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getPortainerClient()
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}

	containers, err := client.ListContainers(envID)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to list containers"})
		return
	}

	type UpdateInfo struct {
		ContainerID string `json:"container_id"`
		Image       string `json:"image"`
		HasUpdate   bool   `json:"has_update"`
	}

	imageChecked := make(map[string]bool)
	results := make([]UpdateInfo, 0)

	for _, c := range containers {
		imageName := c.Image
		if checked, done := imageChecked[imageName]; done {
			results = append(results, UpdateInfo{
				ContainerID: c.ID,
				Image:       imageName,
				HasUpdate:   checked,
			})
			continue
		}

		hasUpdate := false

		localImg, err := client.InspectImage(envID, c.ImageID)
		if err != nil {
			imageChecked[imageName] = false
			results = append(results, UpdateInfo{ContainerID: c.ID, Image: imageName})
			continue
		}

		remoteInfo, err := client.GetDistribution(envID, imageName)
		if err != nil {
			imageChecked[imageName] = false
			results = append(results, UpdateInfo{ContainerID: c.ID, Image: imageName})
			continue
		}

		remoteDigest := remoteInfo.Descriptor.Digest
		if remoteDigest != "" {
			hasUpdate = true
			for _, rd := range localImg.RepoDigests {
				parts := strings.SplitN(rd, "@", 2)
				if len(parts) == 2 && parts[1] == remoteDigest {
					hasUpdate = false
					break
				}
			}
		}

		imageChecked[imageName] = hasUpdate
		results = append(results, UpdateInfo{
			ContainerID: c.ID,
			Image:       imageName,
			HasUpdate:   hasUpdate,
		})
	}

	h.Cache.Set(cacheKey, results)
	writeJSON(w, http.StatusOK, results)
}
