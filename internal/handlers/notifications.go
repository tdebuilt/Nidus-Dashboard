package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/notifications"
)

// NotificationsHandler handles notification provider and rule HTTP requests.
type NotificationsHandler struct {
	DB     *database.DB
	Sender *notifications.Sender
}

// ListProviders godoc
// @Summary List all notification providers
// @Tags notifications
// @Produce json
// @Success 200 {array} models.NotificationProviderResponse "List of notification providers"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/providers [get]
// @Security BearerAuth
func (h *NotificationsHandler) ListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.DB.ListNotificationProviders(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list providers"})
		return
	}

	// Hide tokens in response
	resp := make([]models.NotificationProviderResponse, 0, len(providers))
	for _, p := range providers {
		resp = append(resp, models.NotificationProviderResponse{
			ID:        p.ID,
			Type:      p.Type,
			Name:      p.Name,
			URL:       p.URL,
			HasToken:  p.Token != "",
			Enabled:   p.Enabled,
			Config:    p.Config,
			CreatedAt: p.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, resp)
}

// CreateProvider godoc
// @Summary Create a new notification provider
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body models.CreateNotificationProviderRequest true "Provider details"
// @Success 201 {object} models.NotificationProvider "Created provider"
// @Failure 400 {object} models.ErrorResponse "Invalid request or provider type"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/providers [post]
// @Security BearerAuth
func (h *NotificationsHandler) CreateProvider(w http.ResponseWriter, r *http.Request) {
	var req models.CreateNotificationProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid JSON"})
		return
	}

	if !models.ValidProviderTypes[req.Type] {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: fmt.Sprintf("invalid provider type '%s'", req.Type)})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "name is required"})
		return
	}
	if req.URL == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "url is required"})
		return
	}

	provider, err := h.DB.CreateNotificationProvider(r.Context(), req.Type, req.Name, req.URL, req.Token, req.Config)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create provider"})
		return
	}

	writeJSON(w, http.StatusCreated, provider)
}

// UpdateProvider godoc
// @Summary Update an existing notification provider
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Provider ID"
// @Param request body models.UpdateNotificationProviderRequest true "Updated provider fields"
// @Success 200 {object} models.NotificationProvider "Updated provider"
// @Failure 400 {object} models.ErrorResponse "Invalid provider ID or request body"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/providers/{id} [put]
// @Security BearerAuth
func (h *NotificationsHandler) UpdateProvider(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid provider id"})
		return
	}

	var req models.UpdateNotificationProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid JSON"})
		return
	}

	if err := h.DB.UpdateNotificationProvider(r.Context(), id, req.Name, req.URL, req.Token, req.Enabled, req.Config); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update provider"})
		return
	}

	provider, _ := h.DB.GetNotificationProvider(r.Context(), id)
	writeJSON(w, http.StatusOK, provider)
}

// DeleteProvider godoc
// @Summary Delete a notification provider
// @Tags notifications
// @Produce json
// @Param id path int true "Provider ID"
// @Success 200 {object} object "Deletion confirmation"
// @Failure 400 {object} models.ErrorResponse "Invalid provider ID"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/providers/{id} [delete]
// @Security BearerAuth
func (h *NotificationsHandler) DeleteProvider(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid provider id"})
		return
	}

	if err := h.DB.DeleteNotificationProvider(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to delete provider"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "provider deleted"})
}

// TestProvider godoc
// @Summary Send a test notification via a provider
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body models.TestNotificationRequest true "Provider ID to test"
// @Success 200 {object} object "Test notification sent"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 404 {object} models.ErrorResponse "Provider not found"
// @Failure 502 {object} models.ErrorResponse "Notification delivery failed"
// @Router /notifications/test [post]
// @Security BearerAuth
func (h *NotificationsHandler) TestProvider(w http.ResponseWriter, r *http.Request) {
	var req models.TestNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid JSON"})
		return
	}

	provider, err := h.DB.GetNotificationProvider(r.Context(), req.ProviderID)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "provider not found"})
		return
	}

	if err := h.Sender.Send(
		provider.Type, provider.URL, provider.Token, provider.Config,
		"Nidus — Test",
		"Notification de test depuis Nidus Dashboard",
	); err != nil {
		writeJSON(w, http.StatusBadGateway, models.ErrorResponse{Error: "notification failed: " + sanitizeError(err)})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "test notification sent"})
}

// ListRules godoc
// @Summary List all notification rules
// @Tags notifications
// @Produce json
// @Success 200 {array} models.NotificationRule "List of notification rules"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/rules [get]
// @Security BearerAuth
func (h *NotificationsHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.DB.ListNotificationRules(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list rules"})
		return
	}
	writeJSON(w, http.StatusOK, rules)
}

// CreateRule godoc
// @Summary Create a new notification rule
// @Tags notifications
// @Accept json
// @Produce json
// @Param request body models.CreateNotificationRuleRequest true "Rule details"
// @Success 201 {object} models.NotificationRule "Created rule"
// @Failure 400 {object} models.ErrorResponse "Invalid event type or provider not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/rules [post]
// @Security BearerAuth
func (h *NotificationsHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var req models.CreateNotificationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid JSON"})
		return
	}

	if !models.ValidEventTypes[req.EventType] {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: fmt.Sprintf("invalid event type '%s'", req.EventType)})
		return
	}

	// Verify provider exists
	if _, err := h.DB.GetNotificationProvider(r.Context(), req.ProviderID); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "provider not found"})
		return
	}

	rule, err := h.DB.CreateNotificationRule(r.Context(), req.EventType, req.ProviderID, req.Config)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create rule"})
		return
	}

	writeJSON(w, http.StatusCreated, rule)
}

// UpdateRule godoc
// @Summary Update an existing notification rule
// @Tags notifications
// @Accept json
// @Produce json
// @Param id path int true "Rule ID"
// @Param request body models.UpdateNotificationRuleRequest true "Updated rule fields"
// @Success 200 {object} models.NotificationRule "Updated rule"
// @Failure 400 {object} models.ErrorResponse "Invalid rule ID or request body"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/rules/{id} [put]
// @Security BearerAuth
func (h *NotificationsHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid rule id"})
		return
	}

	var req models.UpdateNotificationRuleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid JSON"})
		return
	}

	if err := h.DB.UpdateNotificationRule(r.Context(), id, req.Enabled, req.Config); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update rule"})
		return
	}

	rule, _ := h.DB.GetNotificationRule(r.Context(), id)
	writeJSON(w, http.StatusOK, rule)
}

// DeleteRule godoc
// @Summary Delete a notification rule
// @Tags notifications
// @Produce json
// @Param id path int true "Rule ID"
// @Success 200 {object} object "Deletion confirmation"
// @Failure 400 {object} models.ErrorResponse "Invalid rule ID"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/rules/{id} [delete]
// @Security BearerAuth
func (h *NotificationsHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid rule id"})
		return
	}

	if err := h.DB.DeleteNotificationRule(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to delete rule"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "rule deleted"})
}
