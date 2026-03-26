package proxmox

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockProxmoxServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Auth - ticket
	mux.HandleFunc("/api2/json/access/ticket", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()
		if r.FormValue("username") == "root@pam" && r.FormValue("password") == "secret" {
			json.NewEncoder(w).Encode(APIResponse[TicketResponse]{
				Data: TicketResponse{
					Ticket:              "PVE:root@pam:12345::abcdef",
					CSRFPreventionToken: "csrf-token-123",
					Username:            "root@pam",
				},
			})
			return
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})

	// List nodes
	mux.HandleFunc("/api2/json/nodes", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		nodes := []Node{
			{Node: "pve1", Status: "online", CPU: 0.25, MaxCPU: 8, Mem: 4294967296, MaxMem: 17179869184, Uptime: 86400},
			{Node: "pve2", Status: "online", CPU: 0.10, MaxCPU: 4, Mem: 2147483648, MaxMem: 8589934592, Uptime: 43200},
		}
		json.NewEncoder(w).Encode(APIResponse[[]Node]{Data: nodes})
	})

	// List VMs for pve1
	mux.HandleFunc("/api2/json/nodes/pve1/qemu", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		vms := []VM{
			{VMID: 100, Name: "ubuntu-server", Status: "running", CPU: 0.15, CPUs: 4, Mem: 2147483648, MaxMem: 4294967296, Uptime: 3600},
			{VMID: 101, Name: "windows-11", Status: "stopped", CPU: 0, CPUs: 2, Mem: 0, MaxMem: 8589934592, Uptime: 0},
		}
		json.NewEncoder(w).Encode(APIResponse[[]VM]{Data: vms})
	})

	// List LXCs for pve1
	mux.HandleFunc("/api2/json/nodes/pve1/lxc", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		lxcs := []VM{
			{VMID: 200, Name: "nginx-proxy", Status: "running", CPU: 0.05, CPUs: 2, Mem: 536870912, MaxMem: 1073741824, Uptime: 7200},
		}
		json.NewEncoder(w).Encode(APIResponse[[]VM]{Data: lxcs})
	})

	// List VMs for pve2
	mux.HandleFunc("/api2/json/nodes/pve2/qemu", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		vms := []VM{
			{VMID: 300, Name: "docker-host", Status: "running", CPU: 0.40, CPUs: 8, Mem: 8589934592, MaxMem: 17179869184, Uptime: 172800},
		}
		json.NewEncoder(w).Encode(APIResponse[[]VM]{Data: vms})
	})

	// List LXCs for pve2
	mux.HandleFunc("/api2/json/nodes/pve2/lxc", func(w http.ResponseWriter, r *http.Request) {
		if !checkAuth(r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(APIResponse[[]VM]{Data: []VM{}})
	})

	// VM actions
	for _, action := range []string{"start", "stop", "shutdown", "reboot"} {
		action := action
		mux.HandleFunc("/api2/json/nodes/pve1/qemu/100/status/"+action, func(w http.ResponseWriter, r *http.Request) {
			if !checkAuth(r) {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			json.NewEncoder(w).Encode(APIResponse[string]{Data: "UPID:pve1:00001234:abcdef:" + action})
		})
	}

	return httptest.NewServer(mux)
}

func checkAuth(r *http.Request) bool {
	// Check API token
	if auth := r.Header.Get("Authorization"); auth != "" {
		return strings.HasPrefix(auth, "PVEAPIToken=")
	}
	// Check ticket cookie
	cookie, err := r.Cookie("PVEAuthCookie")
	return err == nil && cookie.Value != ""
}

func TestAuthenticate(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	if err := client.Authenticate("root@pam", "secret"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.ticket == "" {
		t.Fatal("expected ticket to be set")
	}
	if client.csrfToken == "" {
		t.Fatal("expected CSRF token to be set")
	}
}

func TestAuthenticateInvalidCredentials(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	err := client.Authenticate("root@pam", "wrong")
	if err == nil {
		t.Fatal("expected error for invalid credentials")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401 in error, got: %v", err)
	}
}

func TestAuthenticateNetworkError(t *testing.T) {
	client := NewClient("http://localhost:1", nil)
	err := client.Authenticate("root@pam", "secret")
	if err == nil {
		t.Fatal("expected network error")
	}
}

func TestListNodes(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	nodes, err := client.ListNodes()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
	if nodes[0].Node != "pve1" {
		t.Fatalf("expected node 'pve1', got '%s'", nodes[0].Node)
	}
	if nodes[0].CPU != 0.25 {
		t.Fatalf("expected CPU 0.25, got %f", nodes[0].CPU)
	}
	if nodes[0].MaxCPU != 8 {
		t.Fatalf("expected MaxCPU 8, got %d", nodes[0].MaxCPU)
	}
}

func TestListNodesUnauthorized(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	_, err := client.ListNodes()
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestListVMs(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	vms, err := client.ListVMs("pve1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(vms) != 2 {
		t.Fatalf("expected 2 VMs, got %d", len(vms))
	}
	if vms[0].Name != "ubuntu-server" {
		t.Fatalf("expected 'ubuntu-server', got '%s'", vms[0].Name)
	}
	if vms[0].Type != "qemu" {
		t.Fatalf("expected type 'qemu', got '%s'", vms[0].Type)
	}
	if vms[0].Node != "pve1" {
		t.Fatalf("expected node 'pve1', got '%s'", vms[0].Node)
	}
}

func TestListLXCs(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	lxcs, err := client.ListLXCs("pve1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(lxcs) != 1 {
		t.Fatalf("expected 1 LXC, got %d", len(lxcs))
	}
	if lxcs[0].Type != "lxc" {
		t.Fatalf("expected type 'lxc', got '%s'", lxcs[0].Type)
	}
	if lxcs[0].Name != "nginx-proxy" {
		t.Fatalf("expected 'nginx-proxy', got '%s'", lxcs[0].Name)
	}
}

func TestListAllVMs(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	all, err := client.ListAllVMs()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// pve1: 2 VMs + 1 LXC, pve2: 1 VM + 0 LXC = 4 total
	if len(all) != 4 {
		t.Fatalf("expected 4 total VMs/LXCs, got %d", len(all))
	}

	typeCount := map[string]int{}
	for _, vm := range all {
		typeCount[vm.Type]++
	}
	if typeCount["qemu"] != 3 {
		t.Fatalf("expected 3 qemu VMs, got %d", typeCount["qemu"])
	}
	if typeCount["lxc"] != 1 {
		t.Fatalf("expected 1 LXC, got %d", typeCount["lxc"])
	}
}

func TestSetAPIToken(t *testing.T) {
	client := NewClient("http://example.com", nil)
	client.SetAPIToken("user@pam!token=uuid")

	if client.apiToken != "user@pam!token=uuid" {
		t.Fatalf("expected token set, got '%s'", client.apiToken)
	}
}

func TestFullAuthFlow(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())

	// Authenticate with password
	if err := client.Authenticate("root@pam", "secret"); err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	// List nodes
	nodes, err := client.ListNodes()
	if err != nil {
		t.Fatalf("list nodes failed: %v", err)
	}
	if len(nodes) == 0 {
		t.Fatal("expected at least 1 node")
	}

	// List VMs
	vms, err := client.ListVMs(nodes[0].Node)
	if err != nil {
		t.Fatalf("list VMs failed: %v", err)
	}
	if len(vms) == 0 {
		t.Fatal("expected at least 1 VM")
	}

	// List all
	all, err := client.ListAllVMs()
	if err != nil {
		t.Fatalf("list all failed: %v", err)
	}
	if len(all) == 0 {
		t.Fatal("expected at least 1 VM/LXC")
	}
}

func TestMetricsValues(t *testing.T) {
	srv := mockProxmoxServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetAPIToken("user@pam!mytoken=uuid-value")

	vms, err := client.ListVMs("pve1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ubuntu := vms[0]
	if ubuntu.CPU != 0.15 {
		t.Fatalf("expected CPU 0.15, got %f", ubuntu.CPU)
	}
	if ubuntu.CPUs != 4 {
		t.Fatalf("expected 4 CPUs, got %d", ubuntu.CPUs)
	}
	if ubuntu.Mem != 2147483648 {
		t.Fatalf("expected mem 2GB, got %d", ubuntu.Mem)
	}
	if ubuntu.MaxMem != 4294967296 {
		t.Fatalf("expected maxmem 4GB, got %d", ubuntu.MaxMem)
	}
	if ubuntu.Uptime != 3600 {
		t.Fatalf("expected uptime 3600, got %d", ubuntu.Uptime)
	}
}
