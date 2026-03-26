package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/models"
)

func servicesHandler(t *testing.T) *ServicesHandler {
	t.Helper()
	db := setupTestDB(t)
	// Set up encryption key (required for credential encryption)
	encKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate encryption key: %v", err)
	}
	if err := db.SetSystemSetting("encryption_key", encKey); err != nil {
		t.Fatalf("failed to set encryption key: %v", err)
	}
	return &ServicesHandler{DB: db}
}

func serviceRouter(h *ServicesHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/api/services", h.List)
	r.Put("/api/services/{type}", h.Update)
	r.Delete("/api/services/{type}", h.Delete)
	r.Post("/api/services/{type}/test", h.Test)
	return r
}

func TestServiceListEmpty(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var services []models.ServiceResponse
	json.NewDecoder(w.Body).Decode(&services)
	if len(services) != 0 {
		t.Errorf("expected empty list, got %d", len(services))
	}
}

func TestServiceSaveConfig(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	body, _ := json.Marshal(models.UpdateServiceRequest{
		Name:        "My Portainer",
		URL:         "https://portainer.local:9443",
		Credentials: `{"token": "secret123"}`,
		Config:      `{"env_id": 1}`,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.ServiceResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Type != "portainer" {
		t.Errorf("expected type 'portainer', got '%s'", resp.Type)
	}
	if resp.Name != "My Portainer" {
		t.Errorf("expected name 'My Portainer', got '%s'", resp.Name)
	}
	if resp.URL != "https://portainer.local:9443" {
		t.Errorf("expected URL 'https://portainer.local:9443', got '%s'", resp.URL)
	}
	if !resp.HasCreds {
		t.Error("expected has_credentials=true")
	}
}

func TestServiceCredentialsEncryptedInDB(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	plainCreds := `{"token": "mysecrettoken"}`
	body, _ := json.Marshal(models.UpdateServiceRequest{
		Name:        "Portainer",
		URL:         "https://portainer.local:9443",
		Credentials: plainCreds,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Read directly from DB — credentials should NOT be plaintext
	svc, err := h.DB.GetServiceByType("portainer")
	if err != nil {
		t.Fatalf("failed to get service: %v", err)
	}
	if svc.Credentials == plainCreds {
		t.Fatal("credentials stored in plaintext!")
	}
	if svc.Credentials == "" {
		t.Fatal("credentials should not be empty")
	}

	// Decrypt and verify
	encKey, _ := h.DB.GetSystemSetting("encryption_key")
	decrypted, err := crypto.Decrypt(svc.Credentials, encKey)
	if err != nil {
		t.Fatalf("failed to decrypt credentials: %v", err)
	}
	if decrypted != plainCreds {
		t.Errorf("decrypted credentials don't match: expected '%s', got '%s'", plainCreds, decrypted)
	}
}

func TestServiceGETReturnsNoSecrets(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	// Create a service with credentials
	body, _ := json.Marshal(models.UpdateServiceRequest{
		Name:        "Portainer",
		URL:         "https://portainer.local",
		Credentials: `{"token": "secret"}`,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// GET /services should not contain secrets
	listReq := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)

	if listW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", listW.Code)
	}

	respBody := listW.Body.String()
	if bytes.Contains([]byte(respBody), []byte(`"token"`)) {
		t.Error("response body should not contain credential secrets")
	}
	// Should not have a "credentials" key (only "has_credentials")
	if bytes.Contains([]byte(respBody), []byte(`"credentials"`)) {
		t.Error("response body should not contain raw 'credentials' field")
	}

	var services []models.ServiceResponse
	if err := json.NewDecoder(bytes.NewReader([]byte(respBody))).Decode(&services); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if len(services) != 1 {
		t.Fatalf("expected 1 service, got %d", len(services))
	}
	if !services[0].HasCreds {
		t.Error("expected has_credentials=true")
	}
}

func TestServiceUpdatePreservesCredentials(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	// Create with credentials
	body1, _ := json.Marshal(models.UpdateServiceRequest{
		Name:        "Portainer",
		URL:         "https://portainer.local",
		Credentials: `{"token": "original_secret"}`,
	})
	req1 := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(body1))
	w1 := httptest.NewRecorder()
	r.ServeHTTP(w1, req1)

	// Update without credentials — should keep existing
	body2, _ := json.Marshal(models.UpdateServiceRequest{
		Name: "Portainer Updated",
		URL:  "https://portainer.new",
	})
	req2 := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(body2))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}

	// Credentials should still be there
	svc, _ := h.DB.GetServiceByType("portainer")
	if svc.Credentials == "" {
		t.Error("credentials should be preserved when not provided in update")
	}

	encKey, _ := h.DB.GetSystemSetting("encryption_key")
	decrypted, _ := crypto.Decrypt(svc.Credentials, encKey)
	if decrypted != `{"token": "original_secret"}` {
		t.Errorf("expected original credentials, got '%s'", decrypted)
	}
}

func TestServiceInvalidType(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	body, _ := json.Marshal(models.UpdateServiceRequest{
		Name: "Test",
		URL:  "https://test.local",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/services/invalid_type", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestServiceUpdateMissingName(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	body, _ := json.Marshal(models.UpdateServiceRequest{
		Name: "",
		URL:  "https://test.local",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestServiceUpdateMissingURL(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	body, _ := json.Marshal(models.UpdateServiceRequest{
		Name: "Test",
		URL:  "",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestServiceUpdateInvalidJSON(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	req := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestServiceTestNotConfigured(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/api/services/portainer/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestServiceTestInvalidType(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	req := httptest.NewRequest(http.MethodPost, "/api/services/invalid/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestServiceTestConfigured(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	// Create a mock server to simulate Portainer
	mockSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"Version":"2.0"}`))
	}))
	defer mockSrv.Close()

	// Configure a service pointing to the mock server
	body, _ := json.Marshal(models.UpdateServiceRequest{
		Name: "Portainer",
		URL:  mockSrv.URL,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Test the connection
	testReq := httptest.NewRequest(http.MethodPost, "/api/services/portainer/test", nil)
	testW := httptest.NewRecorder()
	r.ServeHTTP(testW, testReq)

	if testW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", testW.Code, testW.Body.String())
	}

	var resp models.TestServiceResponse
	json.NewDecoder(testW.Body).Decode(&resp)
	if !resp.Success {
		t.Errorf("expected success=true, got message: %s", resp.Message)
	}
}

func TestServiceEnabledFlag(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	enabled := false
	body, _ := json.Marshal(models.UpdateServiceRequest{
		Name:    "Portainer",
		URL:     "https://portainer.local",
		Enabled: &enabled,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp models.ServiceResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Enabled {
		t.Error("expected enabled=false")
	}
}

func TestServiceAllValidTypes(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	for svcType := range ValidServiceTypes {
		body, _ := json.Marshal(models.UpdateServiceRequest{
			Name: svcType,
			URL:  "https://" + svcType + ".local",
		})
		req := httptest.NewRequest(http.MethodPut, "/api/services/"+svcType, bytes.NewReader(body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200 for type '%s', got %d: %s", svcType, w.Code, w.Body.String())
		}
	}

	// List should return all
	listReq := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)

	var services []models.ServiceResponse
	json.NewDecoder(listW.Body).Decode(&services)
	if len(services) != len(ValidServiceTypes) {
		t.Errorf("expected %d services, got %d", len(ValidServiceTypes), len(services))
	}
}

// TestServiceDBMethodsDirectly tests database layer for services.
func TestServiceDBMethodsDirectly(t *testing.T) {
	db := setupTestDB(t)

	// Upsert a new service
	svc, err := db.UpsertService("portainer", "Test", "https://test.local", "encrypted_creds", "{}", true)
	if err != nil {
		t.Fatalf("failed to upsert: %v", err)
	}
	if svc.Type != "portainer" {
		t.Errorf("expected type 'portainer', got '%s'", svc.Type)
	}
	if svc.Credentials != "encrypted_creds" {
		t.Errorf("expected credentials 'encrypted_creds', got '%s'", svc.Credentials)
	}

	// Upsert again (update)
	svc2, err := db.UpsertService("portainer", "Updated", "https://new.local", "", "{}", false)
	if err != nil {
		t.Fatalf("failed to upsert update: %v", err)
	}
	if svc2.Name != "Updated" {
		t.Errorf("expected name 'Updated', got '%s'", svc2.Name)
	}
	// Credentials should be preserved when empty string passed
	if svc2.Credentials != "encrypted_creds" {
		t.Errorf("credentials should be preserved, got '%s'", svc2.Credentials)
	}
	if svc2.Enabled {
		t.Error("expected enabled=false")
	}

	// Get by type
	got, err := db.GetServiceByType("portainer")
	if err != nil {
		t.Fatalf("failed to get: %v", err)
	}
	if got == nil {
		t.Fatal("expected service, got nil")
	}
	if got.Name != "Updated" {
		t.Errorf("expected 'Updated', got '%s'", got.Name)
	}

	// Get non-existent
	notFound, err := db.GetServiceByType("nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if notFound != nil {
		t.Error("expected nil for non-existent type")
	}
}

func TestServiceDelete(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	// Create a service first
	body, _ := json.Marshal(models.UpdateServiceRequest{
		Name: "Portainer",
		URL:  "https://portainer.local",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on create, got %d", w.Code)
	}

	// Delete it
	delReq := httptest.NewRequest(http.MethodDelete, "/api/services/portainer", nil)
	delW := httptest.NewRecorder()
	r.ServeHTTP(delW, delReq)
	if delW.Code != http.StatusOK {
		t.Fatalf("expected 200 on delete, got %d: %s", delW.Code, delW.Body.String())
	}

	// Verify it's gone
	listReq := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	listW := httptest.NewRecorder()
	r.ServeHTTP(listW, listReq)

	var services []models.ServiceResponse
	json.NewDecoder(listW.Body).Decode(&services)
	if len(services) != 0 {
		t.Errorf("expected 0 services after delete, got %d", len(services))
	}
}

func TestServiceDeleteNotFound(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/services/portainer", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestServiceDeleteInvalidType(t *testing.T) {
	h := servicesHandler(t)
	r := serviceRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/services/invalid_type", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}


