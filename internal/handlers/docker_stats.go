package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

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

	client, err := h.getPortainerClient(r.Context())
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

	client, err := h.getPortainerClient(r.Context())
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
