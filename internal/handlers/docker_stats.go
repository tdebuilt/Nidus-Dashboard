package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/portainer"
)

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
	envID, ok := parseIntIDParam(w, r, "envId")
	if !ok {
		return
	}

	cacheKey := "docker:stats:" + strconv.Itoa(envID)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getPortainerClient(r.Context())
	if err != nil || client == nil {
		slog.Error("docker_stats: Portainer not available for stats", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}

	containers, err := client.ListContainers(r.Context(), envID)
	if err != nil {
		slog.Error("docker_stats: failed to list containers for stats", "env_id", envID, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to list containers"})
		return
	}

	results := make([]portainer.ContainerStats, 0)
	for _, c := range containers {
		if c.State != "running" {
			continue
		}
		stats, err := client.CalculateContainerStats(r.Context(), envID, c.ID)
		if err != nil {
			slog.Error("docker stats failed", "container", c.ID[:12], "error", err)
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
type updateInfo struct {
	ContainerID string `json:"container_id"`
	Image       string `json:"image"`
	HasUpdate   bool   `json:"has_update"`
}

func (h *DockerHandler) CheckUpdates(w http.ResponseWriter, r *http.Request) {
	envID, ok := parseIntIDParam(w, r, "envId")
	if !ok {
		return
	}

	cacheKey := "docker:updates:" + strconv.Itoa(envID)
	if val, ok := h.Cache.Get(cacheKey); ok {
		writeJSON(w, http.StatusOK, val)
		return
	}

	client, err := h.getPortainerClient(r.Context())
	if err != nil || client == nil {
		slog.Error("docker_stats: Portainer not available for updates check", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}

	containers, err := client.ListContainers(r.Context(), envID)
	if err != nil {
		slog.Error("docker_stats: failed to list containers for updates", "env_id", envID, "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "failed to list containers"})
		return
	}

	results := h.checkContainerUpdates(r, envID, client, containers)
	h.Cache.Set(cacheKey, results)
	writeJSON(w, http.StatusOK, results)
}

// checkContainerUpdates compares local image digests with remote registry digests for each container.
func (h *DockerHandler) checkContainerUpdates(r *http.Request, envID int, client *portainer.Client, containers []portainer.Container) []updateInfo {
	imageChecked := make(map[string]bool)
	results := make([]updateInfo, 0, len(containers))

	for _, c := range containers {
		imageName := c.Image
		if checked, done := imageChecked[imageName]; done {
			results = append(results, updateInfo{ContainerID: c.ID, Image: imageName, HasUpdate: checked})
			continue
		}

		hasUpdate := h.checkImageUpdate(r, envID, client, c.ImageID, imageName)
		imageChecked[imageName] = hasUpdate
		results = append(results, updateInfo{ContainerID: c.ID, Image: imageName, HasUpdate: hasUpdate})
	}
	return results
}

// checkImageUpdate checks if a single image has a newer version in the remote registry.
func (h *DockerHandler) checkImageUpdate(r *http.Request, envID int, client *portainer.Client, imageID, imageName string) bool {
	localImg, err := client.InspectImage(r.Context(), envID, imageID)
	if err != nil {
		return false
	}
	remoteInfo, err := client.GetDistribution(r.Context(), envID, imageName)
	if err != nil {
		return false
	}
	remoteDigest := remoteInfo.Descriptor.Digest
	if remoteDigest == "" {
		return false
	}
	for _, rd := range localImg.RepoDigests {
		parts := strings.SplitN(rd, "@", 2)
		if len(parts) == 2 && parts[1] == remoteDigest {
			return false
		}
	}
	return true
}
