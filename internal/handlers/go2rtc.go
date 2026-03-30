package handlers

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/go2rtc"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Go2RTCHandler handles go2rtc management HTTP requests.
type Go2RTCHandler struct {
	Manager *go2rtc.Manager
}

// Status returns the current go2rtc status.
func (h *Go2RTCHandler) Status(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil {
		writeJSON(w, http.StatusOK, go2rtc.StatusInfo{Available: false})
		return
	}
	writeJSON(w, http.StatusOK, h.Manager.Status())
}

// Start starts the go2rtc subprocess.
func (h *Go2RTCHandler) Start(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "go2rtc not available"})
		return
	}
	if err := h.Manager.Start(); err != nil {
		slog.Error("go2rtc: start failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "start failed: " + err.Error()})
		return
	}
	slog.Info("go2rtc: started")
	writeJSON(w, http.StatusOK, h.Manager.Status())
}

// Stop stops the go2rtc subprocess.
func (h *Go2RTCHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "go2rtc not available"})
		return
	}
	h.Manager.Stop()
	slog.Info("go2rtc: stopped")
	writeJSON(w, http.StatusOK, h.Manager.Status())
}

// Restart restarts the go2rtc subprocess.
func (h *Go2RTCHandler) Restart(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "go2rtc not available"})
		return
	}
	if err := h.Manager.Restart(); err != nil {
		slog.Error("go2rtc: restart failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "restart failed: " + err.Error()})
		return
	}
	slog.Info("go2rtc: restarted")
	writeJSON(w, http.StatusOK, h.Manager.Status())
}

// ProxyWS proxies a WebSocket connection to the embedded go2rtc instance.
func (h *Go2RTCHandler) ProxyWS(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil || !h.Manager.IsRunning() {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "go2rtc not running"})
		return
	}

	// Build target URL: ws://localhost:1984/api/ws?src=...
	target, _ := url.Parse(strings.Replace(h.Manager.URL(), "http", "ws", 1) + "/api/ws")
	target.RawQuery = r.URL.RawQuery

	// Connect to go2rtc
	backendConn, _, err := websocket.DefaultDialer.Dial(target.String(), nil)
	if err != nil {
		slog.Error("go2rtc-proxy backend dial failed", "error", err)
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "cannot connect to go2rtc"})
		return
	}
	defer backendConn.Close()

	// Upgrade client connection
	clientConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("go2rtc-proxy client upgrade failed", "error", err)
		return
	}
	defer clientConn.Close()

	// Bidirectional copy
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			msgType, msg, err := backendConn.ReadMessage()
			if err != nil {
				return
			}
			if err := clientConn.WriteMessage(msgType, msg); err != nil {
				return
			}
		}
	}()

	for {
		msgType, msg, err := clientConn.ReadMessage()
		if err != nil {
			break
		}
		if err := backendConn.WriteMessage(msgType, msg); err != nil {
			break
		}
	}
	<-done
}

