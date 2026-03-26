package reolink

import (
	"crypto/sha256"
	"fmt"
)

// Camera represents a configured Reolink camera.
type Camera struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"-"`
	Password string `json:"-"`
	Channel  int    `json:"channel"`
	Source   string `json:"source"`    // "direct" or "homeassistant"
	EntityID string `json:"entity_id"` // if source=homeassistant
}

// CameraResponse is the API response for a camera (no credentials).
type CameraResponse struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Channel  int    `json:"channel"`
	Source   string `json:"source"`
	EntityID string `json:"entity_id,omitempty"`
}

// CameraConfig is the JSON stored in the service credentials field.
type CameraConfig struct {
	Cameras   []CameraEntry `json:"cameras"`
	Go2RTCURL string        `json:"go2rtc_url,omitempty"`
}

// CameraEntry is a single camera in the config (with credentials).
type CameraEntry struct {
	Name     string `json:"name"`
	IP       string `json:"ip"`
	Port     int    `json:"port,omitempty"`
	Username string `json:"username"`
	Password string `json:"password"`
	Channel  int    `json:"channel"`
	Source   string `json:"source,omitempty"`    // "direct" (default) or "homeassistant"
	EntityID string `json:"entity_id,omitempty"` // for HA source
}

// DiscoveredCamera is a camera found via ONVIF discovery.
type DiscoveredCamera struct {
	IP    string `json:"ip"`
	Name  string `json:"name"`
	Model string `json:"model"`
}

// StreamInfo contains stream URLs for a camera.
type StreamInfo struct {
	RTSP     string `json:"rtsp,omitempty"`
	Go2RTC   string `json:"go2rtc,omitempty"`
	Snapshot string `json:"snapshot"`
}

// GenerateCameraID creates a stable ID from IP + channel.
func GenerateCameraID(ip string, channel int) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%d", ip, channel)))
	return fmt.Sprintf("%x", hash[:8])
}

// ToCamera converts a CameraEntry to a Camera with generated ID.
func (e CameraEntry) ToCamera() Camera {
	source := e.Source
	if source == "" {
		source = "direct"
	}
	id := GenerateCameraID(e.IP, e.Channel)
	if source == "homeassistant" {
		id = GenerateCameraID(e.EntityID, 0)
	}
	return Camera{
		ID:       id,
		Name:     e.Name,
		IP:       e.IP,
		Port:     e.Port,
		Username: e.Username,
		Password: e.Password,
		Channel:  e.Channel,
		Source:   source,
		EntityID: e.EntityID,
	}
}

// ToResponse converts a Camera to a CameraResponse (no credentials).
func (c Camera) ToResponse() CameraResponse {
	return CameraResponse{
		ID:       c.ID,
		Name:     c.Name,
		IP:       c.IP,
		Channel:  c.Channel,
		Source:   c.Source,
		EntityID: c.EntityID,
	}
}
