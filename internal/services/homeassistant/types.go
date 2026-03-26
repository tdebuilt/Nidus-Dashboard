package homeassistant

// Entity represents a Home Assistant entity state from the REST API.
type Entity struct {
	EntityID    string                 `json:"entity_id"`
	State       string                 `json:"state"`
	Attributes  map[string]interface{} `json:"attributes"`
	LastChanged string                 `json:"last_changed"`
	LastUpdated string                 `json:"last_updated"`
}

// EntityInfo represents a simplified entity for the Nidus frontend.
type EntityInfo struct {
	EntityID     string                 `json:"entity_id"`
	Domain       string                 `json:"domain"`
	Name         string                 `json:"name"`
	State        string                 `json:"state"`
	Attributes   map[string]interface{} `json:"attributes"`
	Icon         string                 `json:"icon,omitempty"`
	UnitOfMeasure string               `json:"unit_of_measurement,omitempty"`
	LastChanged  string                 `json:"last_changed"`
}

// ServiceCallRequest represents a request to call a Home Assistant service.
type ServiceCallRequest struct {
	EntityID string                 `json:"entity_id"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

// ServiceCallResponse represents a HA service call response.
type ServiceCallResponse []Entity

// WSAuthMessage is the HA WebSocket auth message.
type WSAuthMessage struct {
	Type        string `json:"type"`
	AccessToken string `json:"access_token,omitempty"`
}

// WSMessage represents a generic HA WebSocket message.
type WSMessage struct {
	ID      int             `json:"id,omitempty"`
	Type    string          `json:"type"`
	Success *bool           `json:"success,omitempty"`
	Event   *WSEvent        `json:"event,omitempty"`
}

// WSSubscribeMessage is sent to subscribe to HA events.
type WSSubscribeMessage struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	EventType string `json:"event_type,omitempty"`
}

// WSEvent represents a HA event from WebSocket.
type WSEvent struct {
	EventType string                 `json:"event_type"`
	Data      map[string]interface{} `json:"data"`
}
