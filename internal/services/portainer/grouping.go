package portainer

import (
	"sort"
	"strings"
	"time"
)

const (
	composeProjectLabel = "com.docker.compose.project"
	composeServiceLabel = "com.docker.compose.service"
)

// GroupContainers groups containers by stack name.
// Containers with a compose project label are grouped under that stack.
// Containers without a label are returned as standalone.
func GroupContainers(containers []Container, envID int) (stacks []StackInfo, standalone []ContainerInfo) {
	stackMap := make(map[string]*StackInfo)

	for _, c := range containers {
		info := toContainerInfo(c, envID)

		stackName := c.Labels[composeProjectLabel]
		if stackName == "" {
			standalone = append(standalone, info)
			continue
		}

		info.StackName = stackName
		si, ok := stackMap[stackName]
		if !ok {
			si = &StackInfo{
				Name:  stackName,
				EnvID: envID,
			}
			stackMap[stackName] = si
		}
		si.Containers = append(si.Containers, info)
	}

	for _, si := range stackMap {
		sort.Slice(si.Containers, func(i, j int) bool {
			return strings.ToLower(si.Containers[i].Name) < strings.ToLower(si.Containers[j].Name)
		})
		si.Status = computeStackStatus(si.Containers)
		stacks = append(stacks, *si)
	}

	sort.Slice(stacks, func(i, j int) bool {
		return strings.ToLower(stacks[i].Name) < strings.ToLower(stacks[j].Name)
	})
	sort.Slice(standalone, func(i, j int) bool {
		return strings.ToLower(standalone[i].Name) < strings.ToLower(standalone[j].Name)
	})

	return stacks, standalone
}

// MergeWithPortainerStacks enriches grouped stacks with Portainer stack IDs.
func MergeWithPortainerStacks(grouped []StackInfo, portainerStacks []Stack) []StackInfo {
	nameToID := make(map[string]int, len(portainerStacks))
	for _, ps := range portainerStacks {
		nameToID[ps.Name] = ps.ID
	}

	for i := range grouped {
		if id, ok := nameToID[grouped[i].Name]; ok {
			grouped[i].ID = id
		}
	}
	return grouped
}

// GroupMultiEnv groups containers across multiple environments.
func GroupMultiEnv(envContainers map[int][]Container) (stacks []StackInfo, standalone []ContainerInfo) {
	for envID, containers := range envContainers {
		s, st := GroupContainers(containers, envID)
		stacks = append(stacks, s...)
		standalone = append(standalone, st...)
	}
	return stacks, standalone
}

func toContainerInfo(c Container, envID int) ContainerInfo {
	name := ""
	if len(c.Names) > 0 {
		name = strings.TrimPrefix(c.Names[0], "/")
	}

	return ContainerInfo{
		ID:      c.ID,
		Name:    name,
		Image:   c.Image,
		ImageID: c.ImageID,
		State:   c.State,
		Status:  c.Status,
		Health:  extractHealth(c),
		Ports:   c.Ports,
		Created: time.Unix(c.Created, 0),
		EnvID:   envID,
	}
}

func extractHealth(c Container) string {
	status := c.Status
	if strings.Contains(status, "(healthy)") {
		return "healthy"
	}
	if strings.Contains(status, "(unhealthy)") {
		return "unhealthy"
	}
	if strings.Contains(status, "(starting)") {
		return "starting"
	}
	return ""
}

func computeStackStatus(containers []ContainerInfo) string {
	allRunning := true
	allStopped := true

	for _, c := range containers {
		if c.State != "running" {
			allRunning = false
		}
		if c.State != "exited" && c.State != "dead" {
			allStopped = false
		}
	}

	if allRunning {
		return "running"
	}
	if allStopped {
		return "stopped"
	}
	return "partial"
}
