package portainer

import (
	"context"
	"fmt"
)

// ContainerAction performs an action on a container (start/stop/restart/kill).
func (c *Client) ContainerAction(ctx context.Context, envID int, containerID, action string) error {
	path := fmt.Sprintf("/api/endpoints/%d/docker/containers/%s/%s", envID, containerID, action)
	return c.post(ctx, path, nil, nil)
}

// StartContainer starts a stopped container.
func (c *Client) StartContainer(ctx context.Context, envID int, containerID string) error {
	return c.ContainerAction(ctx, envID, containerID, "start")
}

// StopContainer stops a running container.
func (c *Client) StopContainer(ctx context.Context, envID int, containerID string) error {
	return c.ContainerAction(ctx, envID, containerID, "stop")
}

// RestartContainer restarts a container.
func (c *Client) RestartContainer(ctx context.Context, envID int, containerID string) error {
	return c.ContainerAction(ctx, envID, containerID, "restart")
}

// RecreateContainer pulls the latest image and recreates the container via Portainer.
// This uses the Portainer-specific recreate endpoint.
func (c *Client) RecreateContainer(ctx context.Context, envID int, containerID string, pullImage bool) error {
	path := fmt.Sprintf("/api/docker/%d/containers/%s/recreate", envID, containerID)
	body := map[string]bool{"PullImage": pullImage}
	return c.post(ctx, path, body, nil)
}

// GetStackFile retrieves the docker-compose file content of a stack.
func (c *Client) GetStackFile(ctx context.Context, stackID int) (string, error) {
	var result struct {
		StackFileContent string `json:"StackFileContent"`
	}
	path := fmt.Sprintf("/api/stacks/%d/file", stackID)
	if err := c.get(ctx, path, &result); err != nil {
		return "", fmt.Errorf("getting stack file: %w", err)
	}
	return result.StackFileContent, nil
}

// UpdateStack redeploys a stack with its current compose file, optionally pulling images.
func (c *Client) UpdateStack(ctx context.Context, stackID, envID int, pullImage bool) error {
	composeContent, err := c.GetStackFile(ctx, stackID)
	if err != nil {
		return err
	}

	path := fmt.Sprintf("/api/stacks/%d?endpointId=%d", stackID, envID)
	body := map[string]any{
		"stackFileContent": composeContent,
		"pullImage":        pullImage,
	}
	return c.put(ctx, path, body, nil)
}

// StackAction performs a start/stop action on an entire stack.
func (c *Client) StackAction(ctx context.Context, stackID int, envID int, action string) error {
	path := fmt.Sprintf("/api/stacks/%d/%s", stackID, action)
	if action == "start" || action == "stop" {
		path = fmt.Sprintf("/api/stacks/%d/%s?endpointId=%d", stackID, action, envID)
	}
	return c.post(ctx, path, nil, nil)
}

// StartStack starts all containers in a stack.
func (c *Client) StartStack(ctx context.Context, stackID, envID int) error {
	return c.StackAction(ctx, stackID, envID, "start")
}

// StopStack stops all containers in a stack.
func (c *Client) StopStack(ctx context.Context, stackID, envID int) error {
	return c.StackAction(ctx, stackID, envID, "stop")
}
