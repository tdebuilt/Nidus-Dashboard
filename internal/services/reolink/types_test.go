package reolink

import (
	"encoding/hex"
	"testing"
)

func TestGenerateCameraID(t *testing.T) {
	// Deterministic: same input always produces the same output
	id1 := GenerateCameraID("192.168.1.100", 0)
	id2 := GenerateCameraID("192.168.1.100", 0)
	if id1 != id2 {
		t.Errorf("expected deterministic output, got %s and %s", id1, id2)
	}

	// Different IPs produce different IDs
	id3 := GenerateCameraID("192.168.1.101", 0)
	if id1 == id3 {
		t.Errorf("expected different IDs for different IPs, both got %s", id1)
	}

	// Different channels produce different IDs
	id4 := GenerateCameraID("192.168.1.100", 1)
	if id1 == id4 {
		t.Errorf("expected different IDs for different channels, both got %s", id1)
	}
}

func TestGenerateCameraID_Format(t *testing.T) {
	id := GenerateCameraID("10.0.0.1", 0)

	// Should be a hex string of 16 characters (8 bytes)
	if len(id) != 16 {
		t.Errorf("expected 16 chars, got %d (%s)", len(id), id)
	}

	// Should be valid hex
	_, err := hex.DecodeString(id)
	if err != nil {
		t.Errorf("expected valid hex string, got error: %v", err)
	}
}

func TestCameraEntry_ToCamera_Direct(t *testing.T) {
	entry := CameraEntry{
		Name:     "Front Door",
		IP:       "192.168.1.50",
		Port:     80,
		Username: "admin",
		Password: "secret",
		Channel:  0,
		// Source is empty, should default to "direct"
	}

	cam := entry.ToCamera()

	if cam.Source != "direct" {
		t.Errorf("expected source 'direct', got %q", cam.Source)
	}

	// ID should be based on IP + channel
	expectedID := GenerateCameraID("192.168.1.50", 0)
	if cam.ID != expectedID {
		t.Errorf("expected ID %s, got %s", expectedID, cam.ID)
	}
}

func TestCameraEntry_ToCamera_HomeAssistant(t *testing.T) {
	entry := CameraEntry{
		Name:     "Back Yard",
		IP:       "192.168.1.60",
		Port:     80,
		Username: "admin",
		Password: "secret",
		Channel:  0,
		Source:   "homeassistant",
		EntityID: "camera.back_yard",
	}

	cam := entry.ToCamera()

	if cam.Source != "homeassistant" {
		t.Errorf("expected source 'homeassistant', got %q", cam.Source)
	}

	// ID should be based on entity_id + 0
	expectedID := GenerateCameraID("camera.back_yard", 0)
	if cam.ID != expectedID {
		t.Errorf("expected ID %s, got %s", expectedID, cam.ID)
	}

	// ID should NOT be based on IP + channel
	ipBasedID := GenerateCameraID("192.168.1.60", 0)
	if cam.ID == ipBasedID {
		t.Error("homeassistant camera ID should not be based on IP+channel")
	}
}

func TestCameraEntry_ToCamera_Fields(t *testing.T) {
	entry := CameraEntry{
		Name:     "Garage",
		IP:       "10.0.0.5",
		Port:     8080,
		Username: "user1",
		Password: "pass1",
		Channel:  2,
		Source:   "direct",
		EntityID: "",
	}

	cam := entry.ToCamera()

	if cam.Name != "Garage" {
		t.Errorf("Name: expected 'Garage', got %q", cam.Name)
	}
	if cam.IP != "10.0.0.5" {
		t.Errorf("IP: expected '10.0.0.5', got %q", cam.IP)
	}
	if cam.Port != 8080 {
		t.Errorf("Port: expected 8080, got %d", cam.Port)
	}
	if cam.Username != "user1" {
		t.Errorf("Username: expected 'user1', got %q", cam.Username)
	}
	if cam.Password != "pass1" {
		t.Errorf("Password: expected 'pass1', got %q", cam.Password)
	}
	if cam.Channel != 2 {
		t.Errorf("Channel: expected 2, got %d", cam.Channel)
	}
	if cam.Source != "direct" {
		t.Errorf("Source: expected 'direct', got %q", cam.Source)
	}
	if cam.ID == "" {
		t.Error("ID should not be empty")
	}
}

func TestCamera_ToResponse(t *testing.T) {
	cam := Camera{
		ID:       "abc12345def67890",
		Name:     "Front Door",
		IP:       "192.168.1.50",
		Port:     80,
		Username: "admin",
		Password: "supersecret",
		Channel:  0,
		Source:   "direct",
		EntityID: "",
	}

	resp := cam.ToResponse()

	if resp.ID != cam.ID {
		t.Errorf("ID: expected %q, got %q", cam.ID, resp.ID)
	}
	if resp.Name != cam.Name {
		t.Errorf("Name: expected %q, got %q", cam.Name, resp.Name)
	}
	if resp.IP != cam.IP {
		t.Errorf("IP: expected %q, got %q", cam.IP, resp.IP)
	}
	if resp.Channel != cam.Channel {
		t.Errorf("Channel: expected %d, got %d", cam.Channel, resp.Channel)
	}
	if resp.Source != cam.Source {
		t.Errorf("Source: expected %q, got %q", cam.Source, resp.Source)
	}
	if resp.EntityID != cam.EntityID {
		t.Errorf("EntityID: expected %q, got %q", cam.EntityID, resp.EntityID)
	}

	// CameraResponse must NOT contain credentials
	type hasUsername interface{ GetUsername() string }
	type hasPassword interface{ GetPassword() string }
	var _ interface{} = resp
	if _, ok := interface{}(resp).(hasUsername); ok {
		t.Error("CameraResponse should not expose Username")
	}
	if _, ok := interface{}(resp).(hasPassword); ok {
		t.Error("CameraResponse should not expose Password")
	}

	// Verify the struct fields do not include Username or Password
	// by checking that the response struct type only has the expected fields
	expected := CameraResponse{
		ID:       cam.ID,
		Name:     cam.Name,
		IP:       cam.IP,
		Channel:  cam.Channel,
		Source:   cam.Source,
		EntityID: cam.EntityID,
	}
	if resp != expected {
		t.Errorf("response mismatch: got %+v, expected %+v", resp, expected)
	}
}
