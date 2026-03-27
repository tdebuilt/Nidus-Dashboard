package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/models"
)

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

	client, err := h.getPortainerClient(r.Context())
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

	client, err := h.getPortainerClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}

	log.Printf("docker: recreate container %s (envID: %d, pull: %v)", containerID[:12], envID, body.PullImage)

	// Recreate in background — pull+restart can take minutes
	go func() {
		timer := time.AfterFunc(5*time.Minute, func() {
			log.Printf("docker: recreate container %s still running after 5m", containerID[:12])
		})
		defer timer.Stop()
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

	client, err := h.getPortainerClient(r.Context())
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

	client, err := h.getPortainerClient(r.Context())
	if err != nil || client == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Portainer not available"})
		return
	}

	log.Printf("docker: update stack %d (envID: %d, pull: %v)", stackID, body.EnvID, body.PullImage)

	// Run in background — pull + redeploy can take minutes
	go func() {
		timer := time.AfterFunc(5*time.Minute, func() {
			log.Printf("docker: stack update %d still running after 5m", stackID)
		})
		defer timer.Stop()
		if err := client.UpdateStack(stackID, body.EnvID, body.PullImage); err != nil {
			log.Printf("docker: stack update error: %v", err)
		} else {
			log.Printf("docker: stack update success: %d", stackID)
		}
		h.Cache.InvalidatePrefix("docker:")
	}()

	writeJSON(w, http.StatusOK, map[string]string{"status": "started"})
}
