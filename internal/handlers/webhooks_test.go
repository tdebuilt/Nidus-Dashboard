package handlers

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/tdebuilt/nidus/internal/cache"
	"github.com/tdebuilt/nidus/internal/models"
)

func setupWebhookHandler(t *testing.T) *WebhooksHandler {
	t.Helper()
	db := setupTestDB(t)
	return &WebhooksHandler{
		DB:    db,
		Cache: cache.New(30*time.Second, time.Minute),
	}
}

func signPayload(body []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}



func TestCreateWebhook(t *testing.T) {
	h := setupWebhookHandler(t)

	body, _ := json.Marshal(models.CreateWebhookRequest{Name: "Test Webhook"})
	req := httptest.NewRequest(http.MethodPost, "/api/webhooks", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.CreateWebhookResponse
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Name != "Test Webhook" {
		t.Fatalf("expected name 'Test Webhook', got '%s'", resp.Name)
	}
	if resp.Secret == "" {
		t.Fatal("expected secret to be returned on creation")
	}
	if resp.ID == 0 {
		t.Fatal("expected non-zero ID")
	}
}

func TestListWebhooks(t *testing.T) {
	h := setupWebhookHandler(t)

	// Create a webhook first
	h.DB.CreateWebhook("Test", "secret123")

	req := httptest.NewRequest(http.MethodGet, "/api/webhooks", nil)
	w := httptest.NewRecorder()

	h.List(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var webhooks []models.WebhookResponse
	json.NewDecoder(w.Body).Decode(&webhooks)

	if len(webhooks) != 1 {
		t.Fatalf("expected 1 webhook, got %d", len(webhooks))
	}
	if webhooks[0].Name != "Test" {
		t.Fatalf("expected name 'Test', got '%s'", webhooks[0].Name)
	}
	if !webhooks[0].HasSecret {
		t.Fatal("expected has_secret to be true")
	}
}

func TestReceiveWebhookValid(t *testing.T) {
	h := setupWebhookHandler(t)

	// Create webhook
	wh, err := h.DB.CreateWebhook("Test", "mysecret")
	if err != nil {
		t.Fatalf("failed to create webhook: %v", err)
	}

	payload := []byte(`{"event":"deploy","status":"success"}`)
	signature := signPayload(payload, "mysecret")

	r := chi.NewRouter()
	r.Post("/api/webhooks/{id}", h.Receive)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/"+itoa(wh.ID), bytes.NewReader(payload))
	req.Header.Set("X-Webhook-Signature", signature)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestReceiveInvalidSignature(t *testing.T) {
	h := setupWebhookHandler(t)

	wh, _ := h.DB.CreateWebhook("Test", "mysecret")

	payload := []byte(`{"event":"deploy"}`)

	r := chi.NewRouter()
	r.Post("/api/webhooks/{id}", h.Receive)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/"+itoa(wh.ID), bytes.NewReader(payload))
	req.Header.Set("X-Webhook-Signature", "invalidsignature")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestReceiveDisabledWebhook(t *testing.T) {
	h := setupWebhookHandler(t)

	wh, _ := h.DB.CreateWebhook("Test", "mysecret")
	enabled := false
	h.DB.UpdateWebhook(wh.ID, models.UpdateWebhookRequest{Enabled: &enabled})

	payload := []byte(`{"event":"deploy"}`)
	signature := signPayload(payload, "mysecret")

	r := chi.NewRouter()
	r.Post("/api/webhooks/{id}", h.Receive)

	req := httptest.NewRequest(http.MethodPost, "/api/webhooks/"+itoa(wh.ID), bytes.NewReader(payload))
	req.Header.Set("X-Webhook-Signature", signature)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestWebhookActions(t *testing.T) {
	h := setupWebhookHandler(t)

	wh, _ := h.DB.CreateWebhook("Test", "secret")

	// Create action
	action, err := h.DB.CreateWebhookAction(wh.ID, "invalidate_cache", `{"prefix":"docker:"}`)
	if err != nil {
		t.Fatalf("failed to create action: %v", err)
	}
	if action.ActionType != "invalidate_cache" {
		t.Fatalf("expected action_type 'invalidate_cache', got '%s'", action.ActionType)
	}

	// List actions
	actions, err := h.DB.ListWebhookActions(wh.ID)
	if err != nil {
		t.Fatalf("failed to list actions: %v", err)
	}
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}

	// Delete action
	if err := h.DB.DeleteWebhookAction(action.ID); err != nil {
		t.Fatalf("failed to delete action: %v", err)
	}

	actions, _ = h.DB.ListWebhookActions(wh.ID)
	if len(actions) != 0 {
		t.Fatalf("expected 0 actions, got %d", len(actions))
	}
}

func TestDeleteWebhookCascade(t *testing.T) {
	h := setupWebhookHandler(t)

	wh, _ := h.DB.CreateWebhook("Test", "secret")
	h.DB.CreateWebhookAction(wh.ID, "notify", `{"provider_id":1}`)
	h.DB.CreateWebhookAction(wh.ID, "invalidate_cache", `{"prefix":"all:"}`)

	// Delete webhook — should cascade to actions
	if err := h.DB.DeleteWebhook(wh.ID); err != nil {
		t.Fatalf("failed to delete webhook: %v", err)
	}

	actions, _ := h.DB.ListWebhookActions(wh.ID)
	if len(actions) != 0 {
		t.Fatalf("expected cascade delete of actions, got %d", len(actions))
	}
}

