package homeassistant

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func mockHAServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/states", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-ha-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		entities := []Entity{
			{
				EntityID: "light.living_room",
				State:    "on",
				Attributes: map[string]interface{}{
					"friendly_name": "Living Room Light",
					"brightness":    float64(255),
					"icon":          "mdi:lightbulb",
				},
				LastChanged: "2026-03-17T10:00:00Z",
			},
			{
				EntityID: "sensor.temperature",
				State:    "21.5",
				Attributes: map[string]interface{}{
					"friendly_name":       "Temperature",
					"unit_of_measurement": "°C",
					"icon":                "mdi:thermometer",
				},
				LastChanged: "2026-03-17T10:05:00Z",
			},
			{
				EntityID: "switch.bedroom_fan",
				State:    "off",
				Attributes: map[string]interface{}{
					"friendly_name": "Bedroom Fan",
				},
				LastChanged: "2026-03-17T09:00:00Z",
			},
		}
		json.NewEncoder(w).Encode(entities)
	})

	mux.HandleFunc("/api/states/light.living_room", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-ha-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		entity := Entity{
			EntityID: "light.living_room",
			State:    "on",
			Attributes: map[string]interface{}{
				"friendly_name": "Living Room Light",
				"brightness":    float64(255),
			},
			LastChanged: "2026-03-17T10:00:00Z",
		}
		json.NewEncoder(w).Encode(entity)
	})

	mux.HandleFunc("/api/services/light/turn_off", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-ha-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req ServiceCallRequest
		json.NewDecoder(r.Body).Decode(&req)
		resp := ServiceCallResponse{
			{EntityID: req.EntityID, State: "off"},
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/api/camera_proxy/camera.front_door", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-ha-token" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "image/jpeg")
		w.Write([]byte("fake-image-data"))
	})

	return httptest.NewServer(mux)
}

func TestListStates(t *testing.T) {
	srv := mockHAServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-ha-token")

	entities, err := client.ListStates()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entities) != 3 {
		t.Fatalf("expected 3 entities, got %d", len(entities))
	}
	if entities[0].EntityID != "light.living_room" {
		t.Fatalf("expected 'light.living_room', got '%s'", entities[0].EntityID)
	}
	if entities[0].State != "on" {
		t.Fatalf("expected state 'on', got '%s'", entities[0].State)
	}
}

func TestListStatesUnauthorized(t *testing.T) {
	srv := mockHAServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())

	_, err := client.ListStates()
	if err == nil {
		t.Fatal("expected unauthorized error")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Fatalf("expected 'unauthorized' in error, got: %v", err)
	}
}

func TestGetState(t *testing.T) {
	srv := mockHAServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-ha-token")

	entity, err := client.GetState("light.living_room")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if entity.EntityID != "light.living_room" {
		t.Fatalf("expected 'light.living_room', got '%s'", entity.EntityID)
	}
	if entity.State != "on" {
		t.Fatalf("expected 'on', got '%s'", entity.State)
	}
}

func TestCallService(t *testing.T) {
	srv := mockHAServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-ha-token")

	resp, err := client.CallService("light", "turn_off", ServiceCallRequest{
		EntityID: "light.living_room",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp) != 1 {
		t.Fatalf("expected 1 entity in response, got %d", len(resp))
	}
	if resp[0].State != "off" {
		t.Fatalf("expected state 'off', got '%s'", resp[0].State)
	}
}

func TestGetCameraSnapshot(t *testing.T) {
	srv := mockHAServer(t)
	defer srv.Close()

	client := NewClient(srv.URL, srv.Client())
	client.SetToken("test-ha-token")

	data, contentType, err := client.GetCameraSnapshot("camera.front_door")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "fake-image-data" {
		t.Fatalf("expected 'fake-image-data', got '%s'", string(data))
	}
	if contentType != "image/jpeg" {
		t.Fatalf("expected 'image/jpeg', got '%s'", contentType)
	}
}

func TestToEntityInfo(t *testing.T) {
	entity := Entity{
		EntityID: "sensor.temperature",
		State:    "21.5",
		Attributes: map[string]interface{}{
			"friendly_name":       "Temperature",
			"unit_of_measurement": "°C",
			"icon":                "mdi:thermometer",
		},
		LastChanged: "2026-03-17T10:00:00Z",
	}

	info := ToEntityInfo(entity)

	if info.Domain != "sensor" {
		t.Fatalf("expected domain 'sensor', got '%s'", info.Domain)
	}
	if info.Name != "Temperature" {
		t.Fatalf("expected name 'Temperature', got '%s'", info.Name)
	}
	if info.UnitOfMeasure != "°C" {
		t.Fatalf("expected unit '°C', got '%s'", info.UnitOfMeasure)
	}
	if info.Icon != "mdi:thermometer" {
		t.Fatalf("expected icon 'mdi:thermometer', got '%s'", info.Icon)
	}
}

func TestTrailingSlash(t *testing.T) {
	srv := mockHAServer(t)
	defer srv.Close()

	client := NewClient(srv.URL+"/", srv.Client())
	client.SetToken("test-ha-token")

	entities, err := client.ListStates()
	if err != nil {
		t.Fatalf("unexpected error with trailing slash: %v", err)
	}
	if len(entities) != 3 {
		t.Fatalf("expected 3 entities, got %d", len(entities))
	}
}

func TestNetworkError(t *testing.T) {
	client := NewClient("http://localhost:1", nil)
	client.SetToken("test-ha-token")

	_, err := client.ListStates()
	if err == nil {
		t.Fatal("expected network error")
	}
}
