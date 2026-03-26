package portainer

import "time"

// Environment represents a Portainer environment (endpoint).
type Environment struct {
	ID        int    `json:"Id"`
	Name      string `json:"Name"`
	Type      int    `json:"Type"`
	URL       string `json:"URL"`
	PublicURL string `json:"PublicURL"`
	Status    int    `json:"Status"`
}

// Stack represents a Docker stack managed by Portainer.
type Stack struct {
	ID            int    `json:"Id"`
	Name          string `json:"Name"`
	Type          int    `json:"Type"`
	EndpointID    int    `json:"EndpointId"`
	Status        int    `json:"Status"`
	CreationDate  int64  `json:"CreationDate"`
}

// ContainerState holds the running state of a container.
type ContainerState struct {
	Status  string `json:"Status"`
	Running bool   `json:"Running"`
	Paused  bool   `json:"Paused"`
	Dead    bool   `json:"Dead"`
	Health  *HealthState `json:"Health,omitempty"`
}

// HealthState represents a container's health check status.
type HealthState struct {
	Status string `json:"Status"`
}

// Port represents a container's port mapping.
type Port struct {
	IP          string `json:"IP"`
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort"`
	Type        string `json:"Type"`
}

// ContainerLabel is a key-value pair attached to a container.
type ContainerLabel map[string]string

// Container represents a Docker container from the Portainer API.
type Container struct {
	ID      string         `json:"Id"`
	Names   []string       `json:"Names"`
	Image   string         `json:"Image"`
	ImageID string         `json:"ImageID"`
	State   string         `json:"State"`
	Status  string         `json:"Status"`
	Created int64          `json:"Created"`
	Ports   []Port         `json:"Ports"`
	Labels  ContainerLabel `json:"Labels"`
}

// ImageInspect holds inspect data for an image.
type ImageInspect struct {
	ID          string   `json:"Id"`
	RepoDigests []string `json:"RepoDigests"`
}

// DistributionInfo holds remote registry distribution data.
type DistributionInfo struct {
	Descriptor struct {
		Digest string `json:"digest"`
	} `json:"Descriptor"`
}

// ContainerDetails holds detailed inspect data for a single container.
type ContainerDetails struct {
	ID    string          `json:"Id"`
	Name  string          `json:"Name"`
	State *ContainerState `json:"State"`
}

// AuthRequest is the payload for POST /api/auth.
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse is the response from POST /api/auth.
type AuthResponse struct {
	JWT string `json:"jwt"`
}

// ---- Docker stats types ----

// DockerStats represents the response from Docker's /containers/{id}/stats endpoint.
type DockerStats struct {
	CPUStats    CPUStats    `json:"cpu_stats"`
	PreCPUStats CPUStats    `json:"precpu_stats"`
	MemoryStats MemoryStats `json:"memory_stats"`
}

// CPUStats holds CPU usage data from Docker.
type CPUStats struct {
	CPUUsage    CPUUsage    `json:"cpu_usage"`
	SystemUsage uint64      `json:"system_cpu_usage"`
	OnlineCPUs  int         `json:"online_cpus"`
}

// CPUUsage holds per-container CPU usage.
type CPUUsage struct {
	TotalUsage uint64 `json:"total_usage"`
}

// MemoryStats holds memory usage data from Docker.
type MemoryStats struct {
	Usage uint64            `json:"usage"`
	Limit uint64            `json:"limit"`
	Stats map[string]uint64 `json:"stats"`
}

// ContainerStats is the Nidus-level stats for a single container.
type ContainerStats struct {
	ContainerID string  `json:"container_id"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemUsage    uint64  `json:"mem_usage"`
	MemLimit    uint64  `json:"mem_limit"`
	MemPercent  float64 `json:"mem_percent"`
}

// ---- Nidus aggregated types ----

// ContainerInfo is the Nidus-level representation of a container.
type ContainerInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Image       string    `json:"image"`
	ImageID     string    `json:"image_id"`
	State       string    `json:"state"`
	Status      string    `json:"status"`
	Health      string    `json:"health"`
	HasUpdate   bool      `json:"has_update"`
	StackName   string    `json:"stack_name"`
	Ports       []Port    `json:"ports"`
	Created     time.Time `json:"created"`
	EnvID       int       `json:"env_id"`
}

// StackInfo is the Nidus-level representation of a stack.
type StackInfo struct {
	ID         int             `json:"id"`
	Name       string          `json:"name"`
	EnvID      int             `json:"env_id"`
	Status     string          `json:"status"`
	Containers []ContainerInfo `json:"containers"`
}

// EnvironmentInfo is the Nidus-level representation of a Portainer environment.
type EnvironmentInfo struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Host   string `json:"host"`
}
