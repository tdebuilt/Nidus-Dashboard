package portainer

import (
	"testing"
)

func makeContainer(id, name, image, state, status string, labels ContainerLabel) Container {
	return Container{
		ID:      id,
		Names:   []string{"/" + name},
		Image:   image,
		State:   state,
		Status:  status,
		Created: 1700000000,
		Labels:  labels,
	}
}

func TestGroupContainersWithStack(t *testing.T) {
	t.Parallel()
	containers := []Container{
		makeContainer("c1", "web", "nginx", "running", "Up 1h", ContainerLabel{
			"com.docker.compose.project": "my-stack",
			"com.docker.compose.service": "web",
		}),
		makeContainer("c2", "db", "postgres", "running", "Up 1h (healthy)", ContainerLabel{
			"com.docker.compose.project": "my-stack",
			"com.docker.compose.service": "db",
		}),
	}

	stacks, standalone := GroupContainers(containers, 1)

	if len(standalone) != 0 {
		t.Fatalf("expected 0 standalone, got %d", len(standalone))
	}
	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if stacks[0].Name != "my-stack" {
		t.Fatalf("expected stack name 'my-stack', got '%s'", stacks[0].Name)
	}
	if len(stacks[0].Containers) != 2 {
		t.Fatalf("expected 2 containers in stack, got %d", len(stacks[0].Containers))
	}
	if stacks[0].Status != "running" {
		t.Fatalf("expected stack status 'running', got '%s'", stacks[0].Status)
	}
}

func TestGroupContainersStandalone(t *testing.T) {
	t.Parallel()
	containers := []Container{
		makeContainer("c1", "redis", "redis:7", "running", "Up 1h", ContainerLabel{}),
		makeContainer("c2", "mongo", "mongo:6", "running", "Up 1h", nil),
	}

	stacks, standalone := GroupContainers(containers, 1)

	if len(stacks) != 0 {
		t.Fatalf("expected 0 stacks, got %d", len(stacks))
	}
	if len(standalone) != 2 {
		t.Fatalf("expected 2 standalone, got %d", len(standalone))
	}
	if standalone[0].Name != "mongo" {
		t.Fatalf("expected name 'mongo', got '%s'", standalone[0].Name)
	}
	if standalone[1].Name != "redis" {
		t.Fatalf("expected name 'redis', got '%s'", standalone[1].Name)
	}
}

func TestGroupContainersMixed(t *testing.T) {
	t.Parallel()
	containers := []Container{
		makeContainer("c1", "web", "nginx", "running", "Up 1h", ContainerLabel{
			"com.docker.compose.project": "app",
		}),
		makeContainer("c2", "db", "postgres", "running", "Up 1h", ContainerLabel{
			"com.docker.compose.project": "app",
		}),
		makeContainer("c3", "redis", "redis:7", "running", "Up 1h", ContainerLabel{}),
	}

	stacks, standalone := GroupContainers(containers, 1)

	if len(stacks) != 1 {
		t.Fatalf("expected 1 stack, got %d", len(stacks))
	}
	if len(standalone) != 1 {
		t.Fatalf("expected 1 standalone, got %d", len(standalone))
	}
	if stacks[0].Name != "app" {
		t.Fatalf("expected stack 'app', got '%s'", stacks[0].Name)
	}
	if standalone[0].Name != "redis" {
		t.Fatalf("expected standalone 'redis', got '%s'", standalone[0].Name)
	}
}

func TestGroupContainersMultipleStacks(t *testing.T) {
	t.Parallel()
	containers := []Container{
		makeContainer("c1", "web", "nginx", "running", "Up 1h", ContainerLabel{
			"com.docker.compose.project": "frontend",
		}),
		makeContainer("c2", "api", "node", "running", "Up 1h", ContainerLabel{
			"com.docker.compose.project": "backend",
		}),
		makeContainer("c3", "db", "postgres", "running", "Up 1h", ContainerLabel{
			"com.docker.compose.project": "backend",
		}),
	}

	stacks, _ := GroupContainers(containers, 1)

	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(stacks))
	}

	stackNames := make(map[string]int)
	for _, s := range stacks {
		stackNames[s.Name] = len(s.Containers)
	}

	if stackNames["frontend"] != 1 {
		t.Fatalf("expected 1 container in frontend, got %d", stackNames["frontend"])
	}
	if stackNames["backend"] != 2 {
		t.Fatalf("expected 2 containers in backend, got %d", stackNames["backend"])
	}
}

func TestStackStatusPartial(t *testing.T) {
	t.Parallel()
	containers := []Container{
		makeContainer("c1", "web", "nginx", "running", "Up 1h", ContainerLabel{
			"com.docker.compose.project": "app",
		}),
		makeContainer("c2", "db", "postgres", "exited", "Exited (0) 5m ago", ContainerLabel{
			"com.docker.compose.project": "app",
		}),
	}

	stacks, _ := GroupContainers(containers, 1)

	if stacks[0].Status != "partial" {
		t.Fatalf("expected 'partial', got '%s'", stacks[0].Status)
	}
}

func TestStackStatusStopped(t *testing.T) {
	t.Parallel()
	containers := []Container{
		makeContainer("c1", "web", "nginx", "exited", "Exited (0)", ContainerLabel{
			"com.docker.compose.project": "app",
		}),
		makeContainer("c2", "db", "postgres", "exited", "Exited (0)", ContainerLabel{
			"com.docker.compose.project": "app",
		}),
	}

	stacks, _ := GroupContainers(containers, 1)

	if stacks[0].Status != "stopped" {
		t.Fatalf("expected 'stopped', got '%s'", stacks[0].Status)
	}
}

func TestHealthExtraction(t *testing.T) {
	t.Parallel()
	tests := []struct {
		status   string
		expected string
	}{
		{"Up 1h (healthy)", "healthy"},
		{"Up 1h (unhealthy)", "unhealthy"},
		{"Up 30s (starting)", "starting"},
		{"Up 1h", ""},
		{"Exited (0)", ""},
	}

	for _, tc := range tests {
		c := makeContainer("c1", "test", "img", "running", tc.status, nil)
		health := extractHealth(c)
		if health != tc.expected {
			t.Errorf("status '%s': expected health '%s', got '%s'", tc.status, tc.expected, health)
		}
	}
}

func TestContainerInfoNameTrimsSlash(t *testing.T) {
	t.Parallel()
	c := makeContainer("c1", "myapp", "img", "running", "Up 1h", nil)
	info := toContainerInfo(c, 1)

	if info.Name != "myapp" {
		t.Fatalf("expected 'myapp', got '%s'", info.Name)
	}
	if info.EnvID != 1 {
		t.Fatalf("expected envID 1, got %d", info.EnvID)
	}
}

func TestMergeWithPortainerStacks(t *testing.T) {
	t.Parallel()
	grouped := []StackInfo{
		{Name: "web-stack", EnvID: 1},
		{Name: "monitoring", EnvID: 1},
		{Name: "unknown", EnvID: 1},
	}

	portainerStacks := []Stack{
		{ID: 10, Name: "web-stack", EndpointID: 1},
		{ID: 20, Name: "monitoring", EndpointID: 1},
	}

	result := MergeWithPortainerStacks(grouped, portainerStacks)

	if result[0].ID != 10 {
		t.Fatalf("expected ID 10 for web-stack, got %d", result[0].ID)
	}
	if result[1].ID != 20 {
		t.Fatalf("expected ID 20 for monitoring, got %d", result[1].ID)
	}
	if result[2].ID != 0 {
		t.Fatalf("expected ID 0 for unknown stack, got %d", result[2].ID)
	}
}

func TestGroupMultiEnv(t *testing.T) {
	t.Parallel()
	envContainers := map[int][]Container{
		1: {
			makeContainer("c1", "web", "nginx", "running", "Up 1h", ContainerLabel{
				"com.docker.compose.project": "app",
			}),
			makeContainer("c2", "redis", "redis", "running", "Up 1h", ContainerLabel{}),
		},
		2: {
			makeContainer("c3", "api", "node", "running", "Up 1h", ContainerLabel{
				"com.docker.compose.project": "backend",
			}),
		},
	}

	stacks, standalone := GroupMultiEnv(envContainers)

	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks across envs, got %d", len(stacks))
	}
	if len(standalone) != 1 {
		t.Fatalf("expected 1 standalone, got %d", len(standalone))
	}

	// Verify env IDs are preserved
	envIDs := make(map[int]bool)
	for _, s := range stacks {
		envIDs[s.EnvID] = true
	}
	if !envIDs[1] || !envIDs[2] {
		t.Fatal("expected stacks from both environments")
	}
}

func TestGroupContainersEmpty(t *testing.T) {
	t.Parallel()
	stacks, standalone := GroupContainers(nil, 1)

	if stacks != nil {
		t.Fatalf("expected nil stacks, got %v", stacks)
	}
	if standalone != nil {
		t.Fatalf("expected nil standalone, got %v", standalone)
	}
}
