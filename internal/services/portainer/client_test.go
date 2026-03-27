package portainer

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockPortainerServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req AuthRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Username == "admin" && req.Password == "secret" {
			json.NewEncoder(w).Encode(AuthResponse{JWT: "test-jwt-token"})
			return
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})

	mux.HandleFunc("/api/endpoints", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-jwt-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		envs := []Environment{
			{ID: 1, Name: "local", Type: 1, URL: "unix:///var/run/docker.sock", Status: 1},
			{ID: 2, Name: "remote", Type: 1, URL: "tcp://192.168.1.10:2375", Status: 1},
		}
		json.NewEncoder(w).Encode(envs)
	})

	mux.HandleFunc("/api/stacks", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-jwt-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		stacks := []Stack{
			{ID: 1, Name: "web-stack", Type: 2, EndpointID: 1, Status: 1},
			{ID: 2, Name: "monitoring", Type: 2, EndpointID: 1, Status: 1},
		}
		json.NewEncoder(w).Encode(stacks)
	})

	mux.HandleFunc("/api/endpoints/1/docker/containers/json", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-jwt-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		containers := []Container{
			{
				ID:    "abc123",
				Names: []string{"/web-app"},
				Image: "nginx:latest",
				State: "running",
				Status: "Up 2 hours",
				Created: 1700000000,
				Ports: []Port{
					{IP: "0.0.0.0", PrivatePort: 80, PublicPort: 8080, Type: "tcp"},
				},
				Labels: ContainerLabel{
					"com.docker.compose.project": "web-stack",
					"com.docker.compose.service": "web",
				},
			},
			{
				ID:    "def456",
				Names: []string{"/db"},
				Image: "postgres:16",
				State: "running",
				Status: "Up 2 hours (healthy)",
				Created: 1700000000,
				Ports: []Port{
					{IP: "0.0.0.0", PrivatePort: 5432, PublicPort: 5432, Type: "tcp"},
				},
				Labels: ContainerLabel{
					"com.docker.compose.project": "web-stack",
					"com.docker.compose.service": "db",
				},
			},
			{
				ID:    "ghi789",
				Names: []string{"/standalone-redis"},
				Image: "redis:7",
				State: "running",
				Status: "Up 1 hour",
				Created: 1700000000,
				Labels: ContainerLabel{},
			},
		}
		json.NewEncoder(w).Encode(containers)
	})

	mux.HandleFunc("/api/endpoints/1/docker/containers/abc123/json", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-jwt-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		details := ContainerDetails{
			ID:   "abc123",
			Name: "/web-app",
			State: &ContainerState{
				Status:  "running",
				Running: true,
				Health:  &HealthState{Status: "healthy"},
			},
		}
		json.NewEncoder(w).Encode(details)
	})

	return httptest.NewServer(mux)
}

func TestAuthenticate(t *testing.T) {
	t.Parallel()
	srv := mockPortainerServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())

	err := client.Authenticate(context.Background(), "admin", "secret")
	if err != nil {
		t.Fatalf("expected successful auth, got: %v", err)
	}
	if client.token != "test-jwt-token" {
		t.Fatalf("expected token 'test-jwt-token', got '%s'", client.token)
	}
}

func TestAuthenticateInvalidCredentials(t *testing.T) {
	t.Parallel()
	srv := mockPortainerServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())

	err := client.Authenticate(context.Background(), "admin", "wrong")
	if err == nil {
		t.Fatal("expected auth error for invalid credentials")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected 401 in error, got: %v", err)
	}
}

func TestAuthenticateNetworkError(t *testing.T) {
	t.Parallel()
	client := NewClient("http://localhost:1", nil)

	err := client.Authenticate(context.Background(), "admin", "secret")
	if err == nil {
		t.Fatal("expected network error")
	}
}

func TestListEnvironments(t *testing.T) {
	t.Parallel()
	srv := mockPortainerServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-jwt-token")

	envs, err := client.ListEnvironments(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(envs) != 2 {
		t.Fatalf("expected 2 environments, got %d", len(envs))
	}
	if envs[0].Name != "local" {
		t.Fatalf("expected first env name 'local', got '%s'", envs[0].Name)
	}
	if envs[1].Name != "remote" {
		t.Fatalf("expected second env name 'remote', got '%s'", envs[1].Name)
	}
}

func TestListEnvironmentsUnauthorized(t *testing.T) {
	t.Parallel()
	srv := mockPortainerServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	// No token set

	_, err := client.ListEnvironments(context.Background())
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("expected 'unauthorized' in error, got: %v", err)
	}
}

func TestListStacks(t *testing.T) {
	t.Parallel()
	srv := mockPortainerServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-jwt-token")

	stacks, err := client.ListStacks(context.Background(), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stacks) != 2 {
		t.Fatalf("expected 2 stacks, got %d", len(stacks))
	}
	if stacks[0].Name != "web-stack" {
		t.Fatalf("expected first stack 'web-stack', got '%s'", stacks[0].Name)
	}
}

func TestListContainers(t *testing.T) {
	t.Parallel()
	srv := mockPortainerServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-jwt-token")

	containers, err := client.ListContainers(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(containers) != 3 {
		t.Fatalf("expected 3 containers, got %d", len(containers))
	}
	if containers[0].ID != "abc123" {
		t.Fatalf("expected first container ID 'abc123', got '%s'", containers[0].ID)
	}
	if containers[0].State != "running" {
		t.Fatalf("expected state 'running', got '%s'", containers[0].State)
	}
	if len(containers[0].Ports) != 1 {
		t.Fatalf("expected 1 port, got %d", len(containers[0].Ports))
	}
	if containers[0].Ports[0].PublicPort != 8080 {
		t.Fatalf("expected public port 8080, got %d", containers[0].Ports[0].PublicPort)
	}
}

func TestListContainersUnauthorized(t *testing.T) {
	t.Parallel()
	srv := mockPortainerServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())

	_, err := client.ListContainers(context.Background(), 1)
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
}

func TestInspectContainer(t *testing.T) {
	t.Parallel()
	srv := mockPortainerServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-jwt-token")

	details, err := client.InspectContainer(context.Background(), 1, "abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details.ID != "abc123" {
		t.Fatalf("expected ID 'abc123', got '%s'", details.ID)
	}
	if details.State == nil {
		t.Fatal("expected non-nil state")
	}
	if !details.State.Running {
		t.Fatal("expected container to be running")
	}
	if details.State.Health == nil {
		t.Fatal("expected non-nil health")
	}
	if details.State.Health.Status != "healthy" {
		t.Fatalf("expected health 'healthy', got '%s'", details.State.Health.Status)
	}
}

func TestSetToken(t *testing.T) {
	t.Parallel()
	client := NewClient("http://example.com", nil)
	client.SetToken("my-api-key")

	if client.token != "my-api-key" {
		t.Fatalf("expected token 'my-api-key', got '%s'", client.token)
	}
}

func TestFullAuthFlow(t *testing.T) {
	t.Parallel()
	srv := mockPortainerServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())

	ctx := context.Background()

	// Authenticate
	if err := client.Authenticate(ctx, "admin", "secret"); err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	// List environments
	envs, err := client.ListEnvironments(ctx)
	if err != nil {
		t.Fatalf("list envs failed: %v", err)
	}
	if len(envs) == 0 {
		t.Fatal("expected at least 1 environment")
	}

	// List stacks
	stacks, err := client.ListStacks(ctx, 0)
	if err != nil {
		t.Fatalf("list stacks failed: %v", err)
	}
	if len(stacks) == 0 {
		t.Fatal("expected at least 1 stack")
	}

	// List containers
	containers, err := client.ListContainers(ctx, envs[0].ID)
	if err != nil {
		t.Fatalf("list containers failed: %v", err)
	}
	if len(containers) == 0 {
		t.Fatal("expected at least 1 container")
	}
}

func TestTrailingSlashInBaseURL(t *testing.T) {
	t.Parallel()
	srv := mockPortainerServer(t)
	defer srv.Close()

	client := NewClient(srv.URL+"/", srv.Client())
	client.SetToken("test-jwt-token")

	envs, err := client.ListEnvironments(context.Background())
	if err != nil {
		t.Fatalf("unexpected error with trailing slash: %v", err)
	}
	if len(envs) != 2 {
		t.Fatalf("expected 2 envs, got %d", len(envs))
	}
}
