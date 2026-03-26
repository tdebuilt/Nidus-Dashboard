package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/go2rtc"
	"github.com/tdebuilt/nidus/internal/services/reolink"
)

// ReolinkHandler handles Reolink camera HTTP requests.
type ReolinkHandler struct {
	DB         *database.DB
	Cache      *cache.Cache
	Go2RTC     *go2rtc.Manager
	mu         sync.Mutex
	clients    map[string]*reolink.Client // keyed by camera ID
}

func (h *ReolinkHandler) getClient(cam *reolink.Camera) *reolink.Client {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients == nil {
		h.clients = make(map[string]*reolink.Client)
	}
	if c, ok := h.clients[cam.ID]; ok {
		return c
	}
	c := reolink.NewClient(cam.IP, cam.Username, cam.Password, cam.Channel, nil)
	h.clients[cam.ID] = c
	return c
}

func (h *ReolinkHandler) getReolinkConfig() (*reolink.CameraConfig, error) {
	widgets, err := h.DB.GetAllWidgets()
	if err != nil {
		return nil, err
	}

	config := &reolink.CameraConfig{}
	seen := make(map[string]bool)
	for _, w := range widgets {
		if w.Type != "reolink" || w.Config == "" {
			continue
		}
		var wc reolink.CameraConfig
		if err := json.Unmarshal([]byte(w.Config), &wc); err != nil {
			continue
		}
		if wc.Go2RTCURL != "" && config.Go2RTCURL == "" {
			config.Go2RTCURL = wc.Go2RTCURL
		}
		for _, cam := range wc.Cameras {
			key := fmt.Sprintf("%s:%d", cam.IP, cam.Channel)
			if seen[key] {
				continue
			}
			seen[key] = true
			config.Cameras = append(config.Cameras, cam)
		}
	}

	if len(config.Cameras) == 0 {
		return nil, nil
	}
	return config, nil
}

func (h *ReolinkHandler) findCamera(id string) (*reolink.Camera, *reolink.CameraConfig, error) {
	config, err := h.getReolinkConfig()
	if err != nil || config == nil {
		return nil, config, err
	}
	for _, entry := range config.Cameras {
		cam := entry.ToCamera()
		if cam.ID == id {
			return &cam, config, nil
		}
	}
	return nil, config, nil
}

// ListCameras godoc
// @Summary List configured Reolink cameras
// @Tags reolink
// @Produce json
// @Success 200 {array} reolink.CameraResponse
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /reolink/cameras [get]
// @Security BearerAuth
func (h *ReolinkHandler) ListCameras(w http.ResponseWriter, r *http.Request) {
	config, err := h.getReolinkConfig()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to load config"})
		return
	}
	if config == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "Reolink not configured"})
		return
	}

	var cameras []reolink.CameraResponse
	for _, entry := range config.Cameras {
		cameras = append(cameras, entry.ToCamera().ToResponse())
	}
	if cameras == nil {
		cameras = []reolink.CameraResponse{}
	}
	writeJSON(w, http.StatusOK, cameras)
}

// GetSnapshot godoc
// @Summary Get camera snapshot image
// @Tags reolink
// @Produce image/jpeg
// @Param id path string true "Camera ID"
// @Success 200 {file} binary "JPEG image"
// @Failure 404 {object} models.ErrorResponse
// @Failure 502 {object} models.ErrorResponse
// @Router /reolink/cameras/{id}/snapshot [get]
// @Security BearerAuth
func (h *ReolinkHandler) GetSnapshot(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cam, config, err := h.findCamera(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to load config"})
		return
	}
	if cam == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "camera not found"})
		return
	}

	if cam.Source == "homeassistant" {
		h.proxyHASnapshot(w, r, cam.EntityID)
		return
	}

	// Direct Reolink snapshot
	client := h.getClient(cam)
	_ = config

	data, contentType, err := client.GetSnapshot()
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "snapshot failed: " + err.Error()})
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "no-cache")
	w.Write(data)
}

func (h *ReolinkHandler) proxyHASnapshot(w http.ResponseWriter, r *http.Request, entityID string) {
	// Reuse HA handler logic
	svc, err := h.DB.GetServiceByType("homeassistant")
	if err != nil || svc == nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "Home Assistant not configured"})
		return
	}

	encKey, _ := h.DB.GetSystemSetting("encryption_key")
	if encKey == "" {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not found"})
		return
	}
	creds, err := crypto.Decrypt(svc.Credentials, encKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to decrypt credentials"})
		return
	}
	var authData struct {
		Token string `json:"token"`
	}
	json.Unmarshal([]byte(creds), &authData)

	url := fmt.Sprintf("%s/api/camera_proxy/%s", svc.URL, entityID)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Authorization", "Bearer "+authData.Token)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "HA request failed"})
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(resp.StatusCode)
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
		}
		if err != nil {
			break
		}
	}
}

// GetStreamURL godoc
// @Summary Get stream URLs for a camera
// @Tags reolink
// @Produce json
// @Param id path string true "Camera ID"
// @Success 200 {object} reolink.StreamInfo
// @Failure 404 {object} models.ErrorResponse
// @Router /reolink/cameras/{id}/stream [get]
// @Security BearerAuth
func (h *ReolinkHandler) GetStreamURL(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	cam, _, err := h.findCamera(id)
	if err != nil || cam == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "camera not found"})
		return
	}

	info := reolink.StreamInfo{
		Snapshot: fmt.Sprintf("/api/reolink/cameras/%s/snapshot", id),
	}

	if cam.Source == "direct" {
		client := h.getClient(cam)
		info.RTSP = client.GetRTSPURL("main")
	}

	if h.Go2RTC != nil && h.Go2RTC.IsRunning() {
		info.Go2RTC = fmt.Sprintf("/api/go2rtc/ws?src=%s", cam.Name)
	}

	writeJSON(w, http.StatusOK, info)
}

// Discover godoc
// @Summary Discover cameras on the local network via ONVIF
// @Tags reolink
// @Produce json
// @Success 200 {array} reolink.DiscoveredCamera
// @Failure 500 {object} models.ErrorResponse
// @Router /reolink/discover [get]
// @Security BearerAuth
func (h *ReolinkHandler) Discover(w http.ResponseWriter, r *http.Request) {
	cameras, err := reolink.DiscoverCameras(3 * time.Second)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "discovery failed: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, cameras)
}
