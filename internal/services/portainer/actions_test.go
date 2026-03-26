package portainer

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockActionsServer(t *testing.T) (*httptest.Server, *actionLog) {
	t.Helper()
	log := &actionLog{}
	mux := http.NewServeMux()

	// Container actions
	for _, action := range []string{"start", "stop", "restart", "kill"} {
		action := action
		mux.HandleFunc("/api/endpoints/1/docker/containers/abc123/"+action, func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") != "Bearer test-token" {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			log.add("container:" + action)
			w.WriteHeader(http.StatusNoContent)
		})
	}

	// Container recreate (Portainer-specific)
	mux.HandleFunc("/api/docker/1/containers/abc123/recreate", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var body map[string]bool
		json.NewDecoder(r.Body).Decode(&body)
		if body["PullImage"] {
			log.add("container:recreate:pull")
		} else {
			log.add("container:recreate:nopull")
		}
		w.WriteHeader(http.StatusOK)
	})

	// Stack actions
	mux.HandleFunc("/api/stacks/10/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		log.add("stack:start")
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/stacks/10/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		log.add("stack:stop")
		w.WriteHeader(http.StatusOK)
	})

	// Error scenario
	mux.HandleFunc("/api/endpoints/1/docker/containers/notfound/start", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"No such container"}`, http.StatusNotFound)
	})

	return httptest.NewServer(mux), log
}

type actionLog struct {
	actions []string
}

func (l *actionLog) add(a string) {
	l.actions = append(l.actions, a)
}

func TestStartContainer(t *testing.T) {
	srv, log := mockActionsServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-token")

	if err := client.StartContainer(1, "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(log.actions) != 1 || log.actions[0] != "container:start" {
		t.Fatalf("expected [container:start], got %v", log.actions)
	}
}

func TestStopContainer(t *testing.T) {
	srv, log := mockActionsServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-token")

	if err := client.StopContainer(1, "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.actions[0] != "container:stop" {
		t.Fatalf("expected container:stop, got %s", log.actions[0])
	}
}

func TestRestartContainer(t *testing.T) {
	srv, log := mockActionsServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-token")

	if err := client.RestartContainer(1, "abc123"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.actions[0] != "container:restart" {
		t.Fatalf("expected container:restart, got %s", log.actions[0])
	}
}

func TestRecreateContainerWithPull(t *testing.T) {
	srv, log := mockActionsServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-token")

	if err := client.RecreateContainer(1, "abc123", true); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.actions[0] != "container:recreate:pull" {
		t.Fatalf("expected container:recreate:pull, got %s", log.actions[0])
	}
}

func TestRecreateContainerWithoutPull(t *testing.T) {
	srv, log := mockActionsServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-token")

	if err := client.RecreateContainer(1, "abc123", false); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.actions[0] != "container:recreate:nopull" {
		t.Fatalf("expected container:recreate:nopull, got %s", log.actions[0])
	}
}

func TestStartStack(t *testing.T) {
	srv, log := mockActionsServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-token")

	if err := client.StartStack(10, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.actions[0] != "stack:start" {
		t.Fatalf("expected stack:start, got %s", log.actions[0])
	}
}

func TestStopStack(t *testing.T) {
	srv, log := mockActionsServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-token")

	if err := client.StopStack(10, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if log.actions[0] != "stack:stop" {
		t.Fatalf("expected stack:stop, got %s", log.actions[0])
	}
}

func TestContainerActionUnauthorized(t *testing.T) {
	srv, _ := mockActionsServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	// No token

	err := client.StartContainer(1, "abc123")
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("expected 'unauthorized', got: %v", err)
	}
}

func TestContainerActionNotFound(t *testing.T) {
	srv, _ := mockActionsServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-token")

	err := client.StartContainer(1, "notfound")
	if err == nil {
		t.Fatal("expected error for non-existent container")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected 404 in error, got: %v", err)
	}
}

func TestContainerActionNetworkError(t *testing.T) {
	client := NewClient("http://localhost:1", nil)
	client.SetToken("test-token")

	err := client.StartContainer(1, "abc123")
	if err == nil {
		t.Fatal("expected network error")
	}
}
