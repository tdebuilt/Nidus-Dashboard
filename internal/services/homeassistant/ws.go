package homeassistant

import (
	"encoding/json"
	"log/slog"
	"strings"
	"sync"
	"time"

	gws "github.com/gorilla/websocket"

	nidusws "github.com/tdebuilt/nidus/internal/websocket"
)

// WSClient connects to the Home Assistant WebSocket API and forwards events.
type WSClient struct {
	url            string
	token          string
	hub            *nidusws.Hub
	conn           *gws.Conn
	done           chan struct{}
	mu             sync.Mutex
	running        bool
	nextID         int
	OnStateChanged func() // called on each state_changed event
}

// NewWSClient creates a HA WebSocket client.
func NewWSClient(baseURL, token string, hub *nidusws.Hub) *WSClient {
	wsURL := strings.Replace(baseURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)
	wsURL = strings.TrimRight(wsURL, "/") + "/api/websocket"

	return &WSClient{
		url:   wsURL,
		token: token,
		hub:   hub,
		done:  make(chan struct{}),
	}
}

// Connect establishes the WebSocket connection and starts listening.
func (w *WSClient) Connect() error {
	w.mu.Lock()
	if w.running {
		w.mu.Unlock()
		return nil
	}
	w.mu.Unlock()

	conn, _, err := gws.DefaultDialer.Dial(w.url, nil)
	if err != nil {
		return err
	}

	w.mu.Lock()
	w.conn = conn
	w.running = true
	w.nextID = 1
	w.mu.Unlock()

	// Read auth_required
	var authReq WSMessage
	if err := conn.ReadJSON(&authReq); err != nil {
		conn.Close()
		return err
	}
	if authReq.Type != "auth_required" {
		conn.Close()
		return nil
	}

	// Send auth
	if err := conn.WriteJSON(WSAuthMessage{
		Type:        "auth",
		AccessToken: w.token,
	}); err != nil {
		conn.Close()
		return err
	}

	// Read auth result
	var authResult WSMessage
	if err := conn.ReadJSON(&authResult); err != nil {
		conn.Close()
		return err
	}
	if authResult.Type != "auth_ok" {
		conn.Close()
		return nil
	}

	// Subscribe to state_changed events
	w.mu.Lock()
	subID := w.nextID
	w.nextID++
	w.mu.Unlock()

	if err := conn.WriteJSON(WSSubscribeMessage{
		ID:        subID,
		Type:      "subscribe_events",
		EventType: "state_changed",
	}); err != nil {
		conn.Close()
		return err
	}

	go w.readLoop()
	return nil
}

func (w *WSClient) readLoop() {
	defer func() {
		w.mu.Lock()
		w.running = false
		if w.conn != nil {
			w.conn.Close()
		}
		w.mu.Unlock()
	}()

	for {
		select {
		case <-w.done:
			return
		default:
		}

		var raw json.RawMessage
		if err := w.conn.ReadJSON(&raw); err != nil {
			if !gws.IsUnexpectedCloseError(err, gws.CloseGoingAway, gws.CloseNormalClosure) {
				return
			}
			slog.Warn("homeassistant ws read error", "error", err)
			w.scheduleReconnect()
			return
		}

		var msg WSMessage
		if err := json.Unmarshal(raw, &msg); err != nil {
			continue
		}

		if msg.Type == "event" && msg.Event != nil && msg.Event.EventType == "state_changed" {
			w.hub.BroadcastType("ha:state_changed", msg.Event.Data)
			if w.OnStateChanged != nil {
				w.OnStateChanged()
			}
		}
	}
}

func (w *WSClient) scheduleReconnect() {
	go func() {
		time.Sleep(5 * time.Second)
		w.mu.Lock()
		w.running = false
		w.mu.Unlock()
		if err := w.Connect(); err != nil {
			slog.Error("homeassistant ws reconnect failed", "error", err)
		}
	}()
}

// Close shuts down the WebSocket connection.
func (w *WSClient) Close() {
	close(w.done)
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.conn != nil {
		w.conn.Close()
		w.conn = nil
	}
	w.running = false
}
