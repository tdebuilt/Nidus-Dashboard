package proxmox

import (
	"context"
	"fmt"
	"net/url"
)

// VMAction performs a status action on a VM or LXC (start/stop/shutdown/reboot).
func (c *Client) VMAction(ctx context.Context, node, vmType string, vmid int, action string) (string, error) {
	path := fmt.Sprintf("/api2/json/nodes/%s/%s/%d/status/%s", node, vmType, vmid, action)

	var taskID string
	if err := c.post(ctx, path, url.Values{}, &taskID); err != nil {
		return "", fmt.Errorf("%s %s/%d: %w", action, vmType, vmid, err)
	}
	return taskID, nil
}

// StartVM starts a VM.
func (c *Client) StartVM(ctx context.Context, node string, vmid int) (string, error) {
	return c.VMAction(ctx, node, "qemu", vmid, "start")
}

// StopVM force-stops a VM.
func (c *Client) StopVM(ctx context.Context, node string, vmid int) (string, error) {
	return c.VMAction(ctx, node, "qemu", vmid, "stop")
}

// ShutdownVM gracefully shuts down a VM.
func (c *Client) ShutdownVM(ctx context.Context, node string, vmid int) (string, error) {
	return c.VMAction(ctx, node, "qemu", vmid, "shutdown")
}

// RebootVM reboots a VM.
func (c *Client) RebootVM(ctx context.Context, node string, vmid int) (string, error) {
	return c.VMAction(ctx, node, "qemu", vmid, "reboot")
}

// StartLXC starts an LXC container.
func (c *Client) StartLXC(ctx context.Context, node string, vmid int) (string, error) {
	return c.VMAction(ctx, node, "lxc", vmid, "start")
}

// StopLXC force-stops an LXC container.
func (c *Client) StopLXC(ctx context.Context, node string, vmid int) (string, error) {
	return c.VMAction(ctx, node, "lxc", vmid, "stop")
}

// ShutdownLXC gracefully shuts down an LXC container.
func (c *Client) ShutdownLXC(ctx context.Context, node string, vmid int) (string, error) {
	return c.VMAction(ctx, node, "lxc", vmid, "shutdown")
}

// RebootLXC reboots an LXC container.
func (c *Client) RebootLXC(ctx context.Context, node string, vmid int) (string, error) {
	return c.VMAction(ctx, node, "lxc", vmid, "reboot")
}
