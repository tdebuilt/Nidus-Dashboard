package proxmox

// Node represents a Proxmox cluster node.
type Node struct {
	Node   string  `json:"node"`
	Status string  `json:"status"`
	CPU    float64 `json:"cpu"`
	MaxCPU int     `json:"maxcpu"`
	Mem    int64   `json:"mem"`
	MaxMem int64   `json:"maxmem"`
	Uptime int64   `json:"uptime"`
}

// VM represents a virtual machine or LXC container on Proxmox.
type VM struct {
	VMID    int     `json:"vmid"`
	Name    string  `json:"name"`
	Status  string  `json:"status"`
	Type    string  `json:"type"` // "qemu" or "lxc"
	Node    string  `json:"node"`
	CPU     float64 `json:"cpu"`
	CPUs    int     `json:"cpus"`
	Mem     int64   `json:"mem"`
	MaxMem  int64   `json:"maxmem"`
	Disk    int64   `json:"disk"`
	MaxDisk int64   `json:"maxdisk"`
	Uptime  int64   `json:"uptime"`
	NetIn   int64   `json:"netin"`
	NetOut  int64   `json:"netout"`
}

// APIResponse wraps the Proxmox API response format.
type APIResponse[T any] struct {
	Data T `json:"data"`
}

// TicketRequest is the payload for POST /api2/json/access/ticket.
type TicketRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// TicketResponse is the response from the ticket endpoint.
type TicketResponse struct {
	CSRFPreventionToken string `json:"CSRFPreventionToken"`
	Ticket              string `json:"ticket"`
	Username            string `json:"username"`
}

// ---- Nidus aggregated types ----

// NodeInfo is the Nidus-level representation of a node.
type NodeInfo struct {
	Name       string  `json:"name"`
	Status     string  `json:"status"`
	CPUUsage   float64 `json:"cpu_usage"`
	CPUCores   int     `json:"cpu_cores"`
	MemUsed    int64   `json:"mem_used"`
	MemTotal   int64   `json:"mem_total"`
	Uptime     int64   `json:"uptime"`
}

// VMInfo is the Nidus-level representation of a VM or LXC.
type VMInfo struct {
	VMID     int     `json:"vmid"`
	Name     string  `json:"name"`
	Node     string  `json:"node"`
	Type     string  `json:"type"`
	Status   string  `json:"status"`
	CPUUsage float64 `json:"cpu_usage"`
	CPUCores int     `json:"cpu_cores"`
	MemUsed  int64   `json:"mem_used"`
	MemTotal int64   `json:"mem_total"`
	Uptime   int64   `json:"uptime"`
}
