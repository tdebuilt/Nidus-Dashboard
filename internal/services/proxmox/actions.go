package proxmox

import (
	"fmt"
	"net/url"
)

// VMAction performs a status action on a VM or LXC (start/stop/shutdown/reboot).
func (c *Client) VMAction(node, vmType string, vmid int, action string) (string, error) {
	path := fmt.Sprintf("/api2/json/nodes/%s/%s/%d/status/%s", node, vmType, vmid, action)

	var taskID string
	if err := c.post(path, url.Values{}, &taskID); err != nil {
		return "", fmt.Errorf("%s %s/%d: %w", action, vmType, vmid, err)
	}
	return taskID, nil
}

// StartVM starts a VM.
func (c *Client) StartVM(node string, vmid int) (string, error) {
	return c.VMAction(node, "qemu", vmid, "start")
}

// StopVM force-stops a VM.
func (c *Client) StopVM(node string, vmid int) (string, error) {
	return c.VMAction(node, "qemu", vmid, "stop")
}

// ShutdownVM gracefully shuts down a VM.
func (c *Client) ShutdownVM(node string, vmid int) (string, error) {
	return c.VMAction(node, "qemu", vmid, "shutdown")
}

// RebootVM reboots a VM.
func (c *Client) RebootVM(node string, vmid int) (string, error) {
	return c.VMAction(node, "qemu", vmid, "reboot")
}

// StartLXC starts an LXC container.
func (c *Client) StartLXC(node string, vmid int) (string, error) {
	return c.VMAction(node, "lxc", vmid, "start")
}

// StopLXC force-stops an LXC container.
func (c *Client) StopLXC(node string, vmid int) (string, error) {
	return c.VMAction(node, "lxc", vmid, "stop")
}

// ShutdownLXC gracefully shuts down an LXC container.
func (c *Client) ShutdownLXC(node string, vmid int) (string, error) {
	return c.VMAction(node, "lxc", vmid, "shutdown")
}

// RebootLXC reboots an LXC container.
func (c *Client) RebootLXC(node string, vmid int) (string, error) {
	return c.VMAction(node, "lxc", vmid, "reboot")
}
