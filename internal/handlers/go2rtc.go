package handlers

import (
	"io"
	"log"
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
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "start failed: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, h.Manager.Status())
}

// Stop stops the go2rtc subprocess.
func (h *Go2RTCHandler) Stop(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "go2rtc not available"})
		return
	}
	h.Manager.Stop()
	writeJSON(w, http.StatusOK, h.Manager.Status())
}

// Restart restarts the go2rtc subprocess.
func (h *Go2RTCHandler) Restart(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "go2rtc not available"})
		return
	}
	if err := h.Manager.Restart(); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "restart failed: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, h.Manager.Status())
}

// ProxyWS proxies a WebSocket connection to the embedded go2rtc instance.
func (h *Go2RTCHandler) ProxyWS(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil || !h.Manager.IsRunning() {
		http.Error(w, "go2rtc not running", http.StatusBadGateway)
		return
	}

	// Build target URL: ws://localhost:1984/api/ws?src=...
	target, _ := url.Parse(strings.Replace(h.Manager.URL(), "http", "ws", 1) + "/api/ws")
	target.RawQuery = r.URL.RawQuery

	// Connect to go2rtc
	backendConn, _, err := websocket.DefaultDialer.Dial(target.String(), nil)
	if err != nil {
		log.Printf("[go2rtc-proxy] backend dial error: %v", err)
		http.Error(w, "cannot connect to go2rtc", http.StatusBadGateway)
		return
	}
	defer backendConn.Close()

	// Upgrade client connection
	clientConn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("[go2rtc-proxy] client upgrade error: %v", err)
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

// ProxyHTTP proxies an HTTP request to the embedded go2rtc API.
func (h *Go2RTCHandler) ProxyHTTP(w http.ResponseWriter, r *http.Request) {
	if h.Manager == nil || !h.Manager.IsRunning() {
		http.Error(w, "go2rtc not running", http.StatusBadGateway)
		return
	}

	target := h.Manager.URL() + "/api" + strings.TrimPrefix(r.URL.Path, "/api/go2rtc")
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(r.Context(), r.Method, target, r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Header = r.Header.Clone()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "go2rtc unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

