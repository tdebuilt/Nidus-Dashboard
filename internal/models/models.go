package models

// ErrorResponse is a standard error response.
type ErrorResponse struct {
	Error string `json:"error" example:"not found"`
}
