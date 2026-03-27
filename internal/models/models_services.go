package models

import "time"

// Service represents an external service connection.
type Service struct {
	ID          int64     `json:"id"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Credentials string    `json:"-"` // never expose encrypted credentials
	Enabled     bool      `json:"enabled"`
	Config      string    `json:"config"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ServiceResponse is the response for GET /api/services (no secrets).
type ServiceResponse struct {
	ID        int64     `json:"id" example:"1"`
	Type      string    `json:"type" example:"portainer"`
	Name      string    `json:"name" example:"Docker"`
	URL       string    `json:"url" example:"https://portainer.local:9443"`
	Enabled   bool      `json:"enabled" example:"true"`
	Config    string    `json:"config" example:"{}"`
	HasCreds  bool      `json:"has_credentials" example:"true"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// UpdateServiceRequest is the payload for PUT /api/services/{type}.
type UpdateServiceRequest struct {
	Name        string `json:"name"`
	URL         string `json:"url"`
	Credentials string `json:"credentials,omitempty"`
	Enabled     *bool  `json:"enabled,omitempty"`
	Config      string `json:"config,omitempty"`
}

// TestServiceResponse is the response from POST /api/services/{type}/test.
type TestServiceResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
