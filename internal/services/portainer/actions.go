package portainer

import (
	"fmt"
)

// ContainerAction performs an action on a container (start/stop/restart/kill).
func (c *Client) ContainerAction(envID int, containerID, action string) error {
	path := fmt.Sprintf("/api/endpoints/%d/docker/containers/%s/%s", envID, containerID, action)
	return c.post(path, nil, nil)
}

// StartContainer starts a stopped container.
func (c *Client) StartContainer(envID int, containerID string) error {
	return c.ContainerAction(envID, containerID, "start")
}

// StopContainer stops a running container.
func (c *Client) StopContainer(envID int, containerID string) error {
	return c.ContainerAction(envID, containerID, "stop")
}

// RestartContainer restarts a container.
func (c *Client) RestartContainer(envID int, containerID string) error {
	return c.ContainerAction(envID, containerID, "restart")
}

// RecreateContainer pulls the latest image and recreates the container via Portainer.
// This uses the Portainer-specific recreate endpoint.
func (c *Client) RecreateContainer(envID int, containerID string, pullImage bool) error {
	path := fmt.Sprintf("/api/docker/%d/containers/%s/recreate", envID, containerID)
	body := map[string]bool{"PullImage": pullImage}
	return c.post(path, body, nil)
}

// GetStackFile retrieves the docker-compose file content of a stack.
func (c *Client) GetStackFile(stackID int) (string, error) {
	var result struct {
		StackFileContent string `json:"StackFileContent"`
	}
	path := fmt.Sprintf("/api/stacks/%d/file", stackID)
	if err := c.get(path, &result); err != nil {
		return "", fmt.Errorf("getting stack file: %w", err)
	}
	return result.StackFileContent, nil
}

// UpdateStack redeploys a stack with its current compose file, optionally pulling images.
func (c *Client) UpdateStack(stackID, envID int, pullImage bool) error {
	composeContent, err := c.GetStackFile(stackID)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/stacks/%d?endpointId=%d", stackID, envID)
	body := map[string]any{
		"stackFileContent": composeContent,
		"pullImage":        pullImage,
	}
	return c.put(path, body, nil)
}

// StackAction performs a start/stop action on an entire stack.
func (c *Client) StackAction(stackID int, envID int, action string) error {
	path := fmt.Sprintf("/api/stacks/%d/%s", stackID, action)
	if action == "start" || action == "stop" {
		path = fmt.Sprintf("/api/stacks/%d/%s?endpointId=%d", stackID, action, envID)
	}
	return c.post(path, nil, nil)
}

// StartStack starts all containers in a stack.
func (c *Client) StartStack(stackID, envID int) error {
	return c.StackAction(stackID, envID, "start")
}

// StopStack stops all containers in a stack.
func (c *Client) StopStack(stackID, envID int) error {
	return c.StackAction(stackID, envID, "stop")
}
