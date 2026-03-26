package proxmox

import (
	"strings"
	"testing"
)

func TestStartVM(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	taskID, err := client.StartVM("pve1", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(taskID, "start") {
		t.Fatalf("expected task ID containing 'start', got '%s'", taskID)
	}
}

func TestStopVM(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	taskID, err := client.StopVM("pve1", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(taskID, "stop") {
		t.Fatalf("expected task ID containing 'stop', got '%s'", taskID)
	}
}

func TestShutdownVM(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	taskID, err := client.ShutdownVM("pve1", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(taskID, "shutdown") {
		t.Fatalf("expected task ID containing 'shutdown', got '%s'", taskID)
	}
}

func TestRebootVM(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	taskID, err := client.RebootVM("pve1", 100)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(taskID, "reboot") {
		t.Fatalf("expected task ID containing 'reboot', got '%s'", taskID)
	}
}

func TestVMActionUnauthorized(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	// No auth

	_, err := client.StartVM("pve1", 100)
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("expected 'unauthorized' in error, got: %v", err)
	}
}

func TestVMActionNetworkError(t *testing.T) {
	client := NewClient("http://localhost:1", nil)
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	_, err := client.StartVM("pve1", 100)
	if err == nil {
		t.Fatal("expected network error")
	}
}

func TestLXCActions(t *testing.T) {
	// LXC endpoints use the same mock pattern — they share the same handler
	// pattern as VMs, so we test the method routing works correctly.
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	// LXC actions aren't registered in mock (only qemu/100),
	// so these should return 404 errors.
	_, err := client.StartLXC("pve1", 200)
	if err == nil {
		t.Fatal("expected error for unregistered LXC route")
	}
}

func TestAllVMActions(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	actions := []struct {
		name string
		fn   func(string, int) (string, error)
	}{
		{"start", client.StartVM},
		{"stop", client.StopVM},
		{"shutdown", client.ShutdownVM},
		{"reboot", client.RebootVM},
	}

	for _, a := range actions {
		taskID, err := a.fn("pve1", 100)
		if err != nil {
			t.Fatalf("action %s: unexpected error: %v", a.name, err)
		}
		if taskID == "" {
			t.Fatalf("action %s: expected non-empty task ID", a.name)
		}
	}
}
