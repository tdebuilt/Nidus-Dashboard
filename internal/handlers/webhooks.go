package handlers

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/services/notifications"
	nidusws "github.com/tdebuilt/nidus/internal/websocket"
)

// WebhooksHandler handles webhook HTTP requests.
type WebhooksHandler struct {
	DB     *database.DB
	Hub    *nidusws.Hub
	Cache  *cache.Cache
	Sender *notifications.Sender
}

func generateSecret() (string, error) {
	b := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func validateSignature(body []byte, secret, signature string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) == 1
}

// List godoc
// @Summary List all webhooks
// @Tags webhooks
// @Produce json
// @Success 200 {array} models.WebhookResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /webhooks [get]
// @Security BearerAuth
func (h *WebhooksHandler) List(w http.ResponseWriter, r *http.Request) {
	webhooks, err := h.DB.ListWebhooks()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list webhooks"})
		return
	}
	writeJSON(w, http.StatusOK, webhooks)
}

// Create godoc
// @Summary Create a new webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Param body body models.CreateWebhookRequest true "Webhook data"
// @Success 201 {object} models.CreateWebhookResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /webhooks [post]
// @Security BearerAuth
func (h *WebhooksHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "name is required"})
		return
	}

	secret, err := generateSecret()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate secret"})
		return
	}

	webhook, err := h.DB.CreateWebhook(req.Name, secret)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create webhook"})
		return
	}

	writeJSON(w, http.StatusCreated, models.CreateWebhookResponse{
		ID:     webhook.ID,
		Name:   webhook.Name,
		Secret: secret,
		URL:    fmt.Sprintf("/api/webhooks/%d", webhook.ID),
	})
}

// Update godoc
// @Summary Update a webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path int true "Webhook ID"
// @Param body body models.UpdateWebhookRequest true "Fields to update"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /webhooks/{id} [put]
// @Security BearerAuth
func (h *WebhooksHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid webhook ID"})
		return
	}

	var req models.UpdateWebhookRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if err := h.DB.UpdateWebhook(id, req); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update webhook"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Delete godoc
// @Summary Delete a webhook
// @Tags webhooks
// @Param id path int true "Webhook ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /webhooks/{id} [delete]
// @Security BearerAuth
func (h *WebhooksHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid webhook ID"})
		return
	}
	if err := h.DB.DeleteWebhook(id); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to delete webhook"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// ListActions godoc
// @Summary List actions for a webhook
// @Tags webhooks
// @Produce json
// @Param id path int true "Webhook ID"
// @Success 200 {array} models.WebhookAction
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /webhooks/{id}/actions [get]
// @Security BearerAuth
func (h *WebhooksHandler) ListActions(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid webhook ID"})
		return
	}
	actions, err := h.DB.ListWebhookActions(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list actions"})
		return
	}
	writeJSON(w, http.StatusOK, actions)
}

// CreateAction godoc
// @Summary Add an action to a webhook
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path int true "Webhook ID"
// @Param body body models.CreateWebhookActionRequest true "Action data"
// @Success 201 {object} models.WebhookAction
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /webhooks/{id}/actions [post]
// @Security BearerAuth
func (h *WebhooksHandler) CreateAction(w http.ResponseWriter, r *http.Request) {
	webhookID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid webhook ID"})
		return
	}

	var req models.CreateWebhookActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}
	if !models.ValidActionTypes[req.ActionType] {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid action type"})
		return
	}

	action, err := h.DB.CreateWebhookAction(webhookID, req.ActionType, req.Config)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create action"})
		return
	}
	writeJSON(w, http.StatusCreated, action)
}

// DeleteAction godoc
// @Summary Delete a webhook action
// @Tags webhooks
// @Param id path int true "Webhook ID"
// @Param actionId path int true "Action ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /webhooks/{id}/actions/{actionId} [delete]
// @Security BearerAuth
func (h *WebhooksHandler) DeleteAction(w http.ResponseWriter, r *http.Request) {
	actionID, err := strconv.ParseInt(chi.URLParam(r, "actionId"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid action ID"})
		return
	}
	if err := h.DB.DeleteWebhookAction(actionID); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to delete action"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Receive godoc
// @Summary Receive an incoming webhook event
// @Description Public endpoint (no JWT). Validates HMAC-SHA256 signature and triggers configured actions.
// @Tags webhooks
// @Accept json
// @Produce json
// @Param id path int true "Webhook ID"
// @Param X-Webhook-Signature header string true "HMAC-SHA256 signature of the request body"
// @Success 200 {object} map[string]string
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 403 {object} models.ErrorResponse
// @Failure 404 {object} models.ErrorResponse
// @Router /webhooks/{id} [post]
func (h *WebhooksHandler) Receive(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid webhook ID"})
		return
	}

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "failed to read body"})
		return
	}

	signature := r.Header.Get("X-Webhook-Signature")
	if signature == "" {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "missing signature"})
		return
	}

	webhook, err := h.DB.GetWebhook(id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "webhook not found"})
		return
	}

	if !validateSignature(body, webhook.Secret, signature) {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "invalid signature"})
		return
	}

	if !webhook.Enabled {
		writeJSON(w, http.StatusForbidden, models.ErrorResponse{Error: "webhook disabled"})
		return
	}

	actions, err := h.DB.ListWebhookActions(id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to load actions"})
		return
	}

	for _, action := range actions {
		h.executeAction(action, body)
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *WebhooksHandler) executeAction(action models.WebhookAction, body []byte) {
	switch action.ActionType {
	case "notify":
		h.executeNotify(action.Config)
	case "refresh_widget":
		h.executeRefreshWidget(action.Config)
	case "invalidate_cache":
		h.executeInvalidateCache(action.Config)
	}
}

func (h *WebhooksHandler) executeNotify(config string) {
	var cfg struct {
		ProviderID int64  `json:"provider_id"`
		Title      string `json:"title"`
		Message    string `json:"message"`
	}
	if err := json.Unmarshal([]byte(config), &cfg); err != nil {
		return
	}
	if cfg.ProviderID == 0 || h.Sender == nil {
		return
	}

	provider, err := h.DB.GetNotificationProvider(cfg.ProviderID)
	if err != nil || provider == nil {
		return
	}

	title := cfg.Title
	if title == "" {
		title = "Webhook"
	}
	message := cfg.Message
	if message == "" {
		message = "Webhook event received"
	}

	h.Sender.Send(provider.Type, provider.URL, provider.Token, provider.Config, title, message)
}

func (h *WebhooksHandler) executeRefreshWidget(config string) {
	var cfg struct {
		WidgetID int64 `json:"widget_id"`
	}
	if err := json.Unmarshal([]byte(config), &cfg); err != nil || cfg.WidgetID == 0 {
		return
	}

	if h.Hub != nil {
		h.Hub.BroadcastType("widget_refresh", map[string]int64{"widget_id": cfg.WidgetID})
	}
}

func (h *WebhooksHandler) executeInvalidateCache(config string) {
	var cfg struct {
		Prefix string `json:"prefix"`
	}
	if err := json.Unmarshal([]byte(config), &cfg); err != nil || cfg.Prefix == "" {
		return
	}

	if h.Cache != nil {
		h.Cache.InvalidatePrefix(cfg.Prefix)
	}
}
