package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/models"
)

func configRouter(t *testing.T) (*chi.Mux, *ConfigHandler) {
	t.Helper()
	db := setupTestDB(t)

	// Set up encryption key (required for export/import)
	encKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate encryption key: %v", err)
	}
	if err := db.SetSystemSetting("encryption_key", encKey); err != nil {
		t.Fatalf("failed to set encryption key: %v", err)
	}

	h := &ConfigHandler{DB: db}
	r := chi.NewRouter()
	r.Post("/api/config/export", h.Export)
	r.Post("/api/config/import", h.Import)

	catH := &CategoriesHandler{DB: db}
	r.Post("/api/categories", catH.Create)
	r.Get("/api/categories", catH.List)

	widgetH := &WidgetsHandler{DB: db}
	r.Post("/api/categories/{id}/widgets", widgetH.Create)

	settingsH := &SettingsHandler{DB: db}
	r.Put("/api/settings", settingsH.Update)
	r.Get("/api/settings", settingsH.Get)

	svcH := &ServicesHandler{DB: db}
	r.Put("/api/services/{type}", svcH.Update)

	return r, h
}

// exportEncrypted calls POST /api/config/export with a password and returns the encrypted data.
func exportEncrypted(t *testing.T, r *chi.Mux, password string) string {
	t.Helper()
	body, _ := json.Marshal(models.ExportRequest{Password: password})
	req := httptest.NewRequest(http.MethodPost, "/api/config/export", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("export: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	data, ok := resp["data"]
	if !ok || data == "" {
		t.Fatal("export: expected non-empty 'data' field")
	}
	return data
}

// importEncrypted calls POST /api/config/import with encrypted data and password.
func importEncrypted(t *testing.T, r *chi.Mux, password, data string) int {
	t.Helper()
	body, _ := json.Marshal(models.ImportRequest{Password: password, Data: data})
	req := httptest.NewRequest(http.MethodPost, "/api/config/import", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestConfigExportRequiresPassword(t *testing.T) {
	r, _ := configRouter(t)

	body, _ := json.Marshal(models.ExportRequest{Password: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/config/export", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestConfigExportProducesEncryptedData(t *testing.T) {
	r, _ := configRouter(t)

	data := exportEncrypted(t, r, "testpass")

	// Data should be hex-encoded (not valid JSON)
	var raw json.RawMessage
	if json.Unmarshal([]byte(data), &raw) == nil {
		t.Error("exported data should not be valid JSON (it should be encrypted hex)")
	}
}

func TestConfigEncryptedRoundTrip(t *testing.T) {
	r, h := configRouter(t)

	// Create data: category + widget + settings
	catBody, _ := json.Marshal(models.CreateCategoryRequest{Name: "Infra", Icon: "server"})
	catReq := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader(catBody))
	catW := httptest.NewRecorder()
	r.ServeHTTP(catW, catReq)
	if catW.Code != http.StatusCreated {
		t.Fatalf("create category: expected 201, got %d: %s", catW.Code, catW.Body.String())
	}
	var cat models.Category
	json.NewDecoder(catW.Body).Decode(&cat)

	widgetBody, _ := json.Marshal(models.CreateWidgetRequest{
		Type: "docker", Title: "Containers", Config: `{"env":1}`,
		PosX: 0, PosY: 0, Width: 2, Height: 1,
	})
	widgetReq := httptest.NewRequest(http.MethodPost, "/api/categories/"+itoa(cat.ID)+"/widgets", bytes.NewReader(widgetBody))
	widgetW := httptest.NewRecorder()
	r.ServeHTTP(widgetW, widgetReq)

	theme := "light"
	lang := "en"
	interval := 60
	settingsBody, _ := json.Marshal(models.UpdateSettingsRequest{
		Theme: &theme, Language: &lang, RefreshInterval: &interval,
	})
	settingsReq := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(settingsBody))
	settingsW := httptest.NewRecorder()
	r.ServeHTTP(settingsW, settingsReq)

	// Export
	password := "my-secret-password"
	data := exportEncrypted(t, r, password)

	// Import (this replaces everything)
	code := importEncrypted(t, r, password, data)
	if code != http.StatusOK {
		t.Fatalf("import: expected 200, got %d", code)
	}

	// Verify data after round-trip
	cats, _ := h.DB.GetCategories()
	if len(cats) != 1 || cats[0].Name != "Infra" {
		t.Errorf("expected category 'Infra', got %v", cats)
	}

	widgets, _ := h.DB.GetAllWidgets()
	if len(widgets) != 1 || widgets[0].Title != "Containers" {
		t.Errorf("expected widget 'Containers', got %v", widgets)
	}

	settings, _ := h.DB.GetSettings()
	if settings.Theme != "light" || settings.Language != "en" || settings.RefreshInterval != 60 {
		t.Errorf("settings mismatch: %+v", settings)
	}
}

func TestConfigImportWrongPassword(t *testing.T) {
	r, _ := configRouter(t)

	data := exportEncrypted(t, r, "correct-password")
	code := importEncrypted(t, r, "wrong-password", data)

	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong password, got %d", code)
	}
}

func TestConfigImportWithCredentials(t *testing.T) {
	r, h := configRouter(t)

	// Create a service with credentials
	svcBody, _ := json.Marshal(models.UpdateServiceRequest{
		Name:        "Portainer",
		URL:         "https://portainer.local",
		Credentials: `{"token": "my-secret-token"}`,
	})
	svcReq := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(svcBody))
	svcW := httptest.NewRecorder()
	r.ServeHTTP(svcW, svcReq)
	if svcW.Code != http.StatusOK {
		t.Fatalf("create service: expected 200, got %d: %s", svcW.Code, svcW.Body.String())
	}

	// Export (includes decrypted credentials)
	password := "backup-password"
	data := exportEncrypted(t, r, password)

	// Verify the encrypted data contains the credential when decrypted
	derivedKey := crypto.DeriveKey(password)
	decrypted, err := crypto.Decrypt(data, derivedKey)
	if err != nil {
		t.Fatalf("manual decrypt failed: %v", err)
	}
	if !bytes.Contains([]byte(decrypted), []byte("my-secret-token")) {
		t.Error("decrypted export should contain plaintext credentials")
	}

	// Import into same DB (simulates restore)
	code := importEncrypted(t, r, password, data)
	if code != http.StatusOK {
		t.Fatalf("import: expected 200, got %d", code)
	}

	// Verify credentials were re-encrypted in DB
	svc, _ := h.DB.GetServiceByType("portainer")
	if svc == nil {
		t.Fatal("expected portainer service after import")
	}
	if svc.Credentials == "" {
		t.Fatal("expected credentials to be restored")
	}

	// Decrypt and verify
	encKey, _ := h.DB.GetSystemSetting("encryption_key")
	plainCreds, err := crypto.Decrypt(svc.Credentials, encKey)
	if err != nil {
		t.Fatalf("failed to decrypt restored credentials: %v", err)
	}
	if plainCreds != `{"token": "my-secret-token"}` {
		t.Errorf("expected original credentials, got '%s'", plainCreds)
	}
}

func TestConfigImportMissingData(t *testing.T) {
	r, _ := configRouter(t)

	body, _ := json.Marshal(models.ImportRequest{Password: "test", Data: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/config/import", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestConfigImportMissingPassword(t *testing.T) {
	r, _ := configRouter(t)

	body, _ := json.Marshal(models.ImportRequest{Password: "", Data: "deadbeef"})
	req := httptest.NewRequest(http.MethodPost, "/api/config/import", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestConfigImportInvalidJSON(t *testing.T) {
	r, _ := configRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/config/import", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestConfigImportPreservesExistingOnFailure(t *testing.T) {
	r, h := configRouter(t)

	// Create a category
	catBody, _ := json.Marshal(models.CreateCategoryRequest{Name: "Original", Icon: "star"})
	catReq := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader(catBody))
	catW := httptest.NewRecorder()
	r.ServeHTTP(catW, catReq)

	// Try to import with wrong password — should not affect existing data
	code := importEncrypted(t, r, "wrong-pass", "deadbeefcafebabe")
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", code)
	}

	// Original data should still be there
	cats, _ := h.DB.GetCategories()
	if len(cats) != 1 || cats[0].Name != "Original" {
		t.Errorf("existing data should be preserved after failed import, got %d categories", len(cats))
	}
}

func TestConfigValidationCategoryEmptyName(t *testing.T) {
	r, _ := configRouter(t)

	// Create valid encrypted data with empty category name
	cfg := models.EncryptedExport{
		Version:  2,
		Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Categories: []models.Category{
			{ID: 1, Name: "", Icon: "star"},
		},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password))

	code := importEncrypted(t, r, password, encrypted)
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty category name, got %d", code)
	}
}

func TestImportWidgetAnyTypeAccepted(t *testing.T) {
	r, _ := configRouter(t)

	// Widget types are not validated server-side (dynamic registry on frontend)
	// Any non-empty type string should be accepted
	types := []string{"reolink", "weather", "rss", "calendar", "system", "finance", "pihole", "mediaserver", "uptimekuma", "custom_widget"}
	for _, wType := range types {
		cfg := models.EncryptedExport{
			Version:  2,
			Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
			Categories: []models.Category{
				{ID: 1, Name: "Cat1", Icon: "star"},
			},
			Widgets: []models.Widget{
				{CategoryID: 1, Type: wType, Title: "W1", Config: "{}", Width: 6, Height: 0},
			},
		}
		jsonData, _ := json.Marshal(cfg)
		password := "test"
		encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password))

		code := importEncrypted(t, r, password, encrypted)
		if code != http.StatusOK {
			t.Errorf("expected 200 for widget type '%s', got %d", wType, code)
		}
	}
}

func TestImportWidgetLargeWidth(t *testing.T) {
	r, _ := configRouter(t)

	cfg := models.EncryptedExport{
		Version:  2,
		Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Categories: []models.Category{
			{ID: 1, Name: "Cat1", Icon: "star"},
		},
		Widgets: []models.Widget{
			{CategoryID: 1, Type: "docker", Title: "W1", Config: "{}", Width: 12, Height: 4},
			{CategoryID: 1, Type: "docker", Title: "W2", Config: "{}", Width: 24, Height: 2},
		},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password))

	code := importEncrypted(t, r, password, encrypted)
	if code != http.StatusOK {
		t.Errorf("expected 200 for large widget widths (12, 24), got %d", code)
	}
}

func TestImportWidgetZeroHeight(t *testing.T) {
	r, _ := configRouter(t)

	cfg := models.EncryptedExport{
		Version:  2,
		Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Categories: []models.Category{
			{ID: 1, Name: "Cat1", Icon: "star"},
		},
		Widgets: []models.Widget{
			{CategoryID: 1, Type: "applink", Title: "W1", Config: "{}", Width: 6, Height: 0},
		},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password))

	code := importEncrypted(t, r, password, encrypted)
	if code != http.StatusOK {
		t.Errorf("expected 200 for widget with height=0 (auto), got %d", code)
	}
}

func TestImportServiceWithoutURL(t *testing.T) {
	r, _ := configRouter(t)

	cfg := models.EncryptedExport{
		Version:  2,
		Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Services: []models.ServiceExport{
			{Type: "reolink", Name: "Cameras", URL: "", Enabled: true, Config: "{}"},
		},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password))

	code := importEncrypted(t, r, password, encrypted)
	if code != http.StatusOK {
		t.Errorf("expected 200 for reolink service without URL (NeedsURL=false), got %d", code)
	}
}

func TestConfigValidationServiceInvalidType(t *testing.T) {
	r, _ := configRouter(t)

	cfg := models.EncryptedExport{
		Version:  2,
		Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Services: []models.ServiceExport{
			{Type: "unknown", Name: "Bad", URL: "http://example.com", Enabled: true, Config: "{}"},
		},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password))

	code := importEncrypted(t, r, password, encrypted)
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid service type, got %d", code)
	}
}

func TestConfigValidationServiceMissingURL(t *testing.T) {
	r, _ := configRouter(t)

	cfg := models.EncryptedExport{
		Version:  2,
		Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Services: []models.ServiceExport{
			{Type: "portainer", Name: "Portainer", URL: "", Enabled: true, Config: "{}"},
		},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password))

	code := importEncrypted(t, r, password, encrypted)
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing service URL, got %d", code)
	}
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
