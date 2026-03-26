package websocket

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("expected non-nil hub")
	}
	if hub.ClientCount() != 0 {
		t.Fatalf("expected 0 clients, got %d", hub.ClientCount())
	}
}

func TestHubRegisterUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{
		hub:  hub,
		send: make(chan []byte, 256),
	}

	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", hub.ClientCount())
	}

	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Fatalf("expected 0 clients after unregister, got %d", hub.ClientCount())
	}
}

func TestHubBroadcast(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	c1 := &Client{hub: hub, send: make(chan []byte, 256)}
	c2 := &Client{hub: hub, send: make(chan []byte, 256)}

	hub.register <- c1
	hub.register <- c2
	time.Sleep(10 * time.Millisecond)

	msg := []byte(`{"type":"test","payload":"hello"}`)
	hub.broadcast <- msg
	time.Sleep(10 * time.Millisecond)

	select {
	case received := <-c1.send:
		if string(received) != string(msg) {
			t.Fatalf("client 1: expected %s, got %s", msg, received)
		}
	default:
		t.Fatal("client 1 did not receive broadcast")
	}

	select {
	case received := <-c2.send:
		if string(received) != string(msg) {
			t.Fatalf("client 2: expected %s, got %s", msg, received)
		}
	default:
		t.Fatal("client 2 did not receive broadcast")
	}
}

func TestHubBroadcastMessage(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{hub: hub, send: make(chan []byte, 256)}
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	hub.BroadcastMessage(Message{
		Type:    "update",
		Payload: json.RawMessage(`{"key":"value"}`),
	})
	time.Sleep(10 * time.Millisecond)

	select {
	case received := <-client.send:
		var msg Message
		if err := json.Unmarshal(received, &msg); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if msg.Type != "update" {
			t.Fatalf("expected type 'update', got '%s'", msg.Type)
		}
	default:
		t.Fatal("client did not receive message")
	}
}

func TestHubBroadcastType(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{hub: hub, send: make(chan []byte, 256)}
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	hub.BroadcastType("ha:state_changed", map[string]string{"entity_id": "light.living_room"})
	time.Sleep(10 * time.Millisecond)

	select {
	case received := <-client.send:
		var msg Message
		if err := json.Unmarshal(received, &msg); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if msg.Type != "ha:state_changed" {
			t.Fatalf("expected type 'ha:state_changed', got '%s'", msg.Type)
		}
		var payload map[string]string
		if err := json.Unmarshal(msg.Payload, &payload); err != nil {
			t.Fatalf("failed to unmarshal payload: %v", err)
		}
		if payload["entity_id"] != "light.living_room" {
			t.Fatalf("expected entity_id 'light.living_room', got '%s'", payload["entity_id"])
		}
	default:
		t.Fatal("client did not receive message")
	}
}

func TestHubUnregisterClosesSend(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{hub: hub, send: make(chan []byte, 256)}
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)

	// send channel should be closed
	_, ok := <-client.send
	if ok {
		t.Fatal("expected send channel to be closed")
	}
}

func TestHubMultipleUnregister(t *testing.T) {
	hub := NewHub()
	go hub.Run()

	client := &Client{hub: hub, send: make(chan []byte, 256)}
	hub.register <- client
	time.Sleep(10 * time.Millisecond)

	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)

	// Second unregister should not panic
	hub.unregister <- client
	time.Sleep(10 * time.Millisecond)

	if hub.ClientCount() != 0 {
		t.Fatalf("expected 0 clients, got %d", hub.ClientCount())
	}
}

func TestNewClient(t *testing.T) {
	hub := NewHub()
	client := NewClient(hub, nil, 42)

	if client.hub != hub {
		t.Fatal("expected hub reference")
	}
	if client.userID != 42 {
		t.Fatalf("expected userID 42, got %d", client.userID)
	}
	if client.send == nil {
		t.Fatal("expected non-nil send channel")
	}
}
