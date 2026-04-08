package handlers

import (
	"bytes"
	"context"
	"encoding/hex"
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
	ctx := context.Background()
	db := setupTestDB(t)

	// Set up encryption key (required for export/import)
	encKey, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("failed to generate encryption key: %v", err)
	}
	if err := db.SetSystemSetting(ctx, "encryption_key", encKey); err != nil {
		t.Fatalf("failed to set encryption key: %v", err)
	}

	h := &ConfigHandler{DB: db}
	r := chi.NewRouter()
	r.Post("/api/config/export", h.Export)
	r.Post("/api/config/import", h.Import)
	r.Get("/api/config/yaml", h.ExportYAML)
	r.Post("/api/config/yaml", h.ImportYAML)

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
type exportResult struct {
	Data string
	Salt string
	KDF  string
}

func exportEncrypted(t *testing.T, r *chi.Mux, password string) exportResult {
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
	return exportResult{Data: data, Salt: resp["salt"], KDF: resp["kdf"]}
}

// importEncrypted calls POST /api/config/import with encrypted data and password.
func importEncrypted(t *testing.T, r *chi.Mux, password string, exp exportResult) int {
	t.Helper()
	body, _ := json.Marshal(models.ImportRequest{Password: password, Data: exp.Data, Salt: exp.Salt, KDF: exp.KDF})
	req := httptest.NewRequest(http.MethodPost, "/api/config/import", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w.Code
}

func TestConfigExportRequiresPassword(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
	r, _ := configRouter(t)

	result := exportEncrypted(t, r, "testpass")

	// Data should be hex-encoded (not valid JSON)
	var raw json.RawMessage
	if json.Unmarshal([]byte(result.Data), &raw) == nil {
		t.Error("exported data should not be valid JSON (it should be encrypted hex)")
	}

	// Should include Argon2id salt and KDF marker
	if result.KDF != "argon2id" {
		t.Errorf("expected kdf 'argon2id', got %q", result.KDF)
	}
	if result.Salt == "" {
		t.Error("expected non-empty salt")
	}
}

func TestConfigEncryptedRoundTrip(t *testing.T) {
	t.Parallel()
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
	ctx := context.Background()
	cats, _ := h.DB.GetCategories(ctx)
	if len(cats) != 1 || cats[0].Name != "Infra" {
		t.Errorf("expected category 'Infra', got %v", cats)
	}

	widgets, _ := h.DB.GetAllWidgets(ctx)
	if len(widgets) != 1 || widgets[0].Title != "Containers" {
		t.Errorf("expected widget 'Containers', got %v", widgets)
	}

	settings, _ := h.DB.GetSettings(ctx)
	if settings.Theme != "light" || settings.Language != "en" || settings.RefreshInterval != 60 {
		t.Errorf("settings mismatch: %+v", settings)
	}
}

func TestConfigImportWrongPassword(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	data := exportEncrypted(t, r, "correct-password")
	code := importEncrypted(t, r, "wrong-password", data)

	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for wrong password, got %d", code)
	}
}

func TestConfigImportWithCredentials(t *testing.T) {
	t.Parallel()
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
	salt, _ := hex.DecodeString(data.Salt)
	derivedKey := crypto.DeriveKeyWithSalt(password, salt)
	decrypted, err := crypto.Decrypt(data.Data, derivedKey)
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
	ctx := context.Background()
	svc, _ := h.DB.GetServiceByType(ctx, "portainer")
	if svc == nil {
		t.Fatal("expected portainer service after import")
	}
	if svc.Credentials == "" {
		t.Fatal("expected credentials to be restored")
	}

	// Decrypt and verify
	encKey, _ := h.DB.GetSystemSetting(ctx, "encryption_key")
	plainCreds, err := crypto.Decrypt(svc.Credentials, encKey)
	if err != nil {
		t.Fatalf("failed to decrypt restored credentials: %v", err)
	}
	if plainCreds != `{"token": "my-secret-token"}` {
		t.Errorf("expected original credentials, got '%s'", plainCreds)
	}
}

func TestConfigImportMissingData(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
	r, _ := configRouter(t)

	req := httptest.NewRequest(http.MethodPost, "/api/config/import", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestConfigImportPreservesExistingOnFailure(t *testing.T) {
	t.Parallel()
	r, h := configRouter(t)

	// Create a category
	catBody, _ := json.Marshal(models.CreateCategoryRequest{Name: "Original", Icon: "star"})
	catReq := httptest.NewRequest(http.MethodPost, "/api/categories", bytes.NewReader(catBody))
	catW := httptest.NewRecorder()
	r.ServeHTTP(catW, catReq)

	// Try to import with wrong password — should not affect existing data
	code := importEncrypted(t, r, "wrong-pass", exportResult{Data: "deadbeefcafebabe"})
	if code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", code)
	}

	// Original data should still be there
	cats, _ := h.DB.GetCategories(context.Background())
	if len(cats) != 1 || cats[0].Name != "Original" {
		t.Errorf("existing data should be preserved after failed import, got %d categories", len(cats))
	}
}

func TestConfigValidationCategoryEmptyName(t *testing.T) {
	t.Parallel()
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
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty category name, got %d", code)
	}
}

func TestImportWidgetAnyTypeAccepted(t *testing.T) {
	t.Parallel()
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
		encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

		code := importEncrypted(t, r, password, exportResult{Data: encrypted})
		if code != http.StatusOK {
			t.Errorf("expected 200 for widget type '%s', got %d", wType, code)
		}
	}
}

func TestImportWidgetLargeWidth(t *testing.T) {
	t.Parallel()
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
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusOK {
		t.Errorf("expected 200 for large widget widths (12, 24), got %d", code)
	}
}

func TestImportWidgetZeroHeight(t *testing.T) {
	t.Parallel()
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
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusOK {
		t.Errorf("expected 200 for widget with height=0 (auto), got %d", code)
	}
}

func TestImportServiceWithoutURL(t *testing.T) {
	t.Parallel()
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
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusOK {
		t.Errorf("expected 200 for reolink service without URL (NeedsURL=false), got %d", code)
	}
}

func TestConfigImportSkipsUnknownServiceType(t *testing.T) {
	t.Parallel()
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
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusOK {
		t.Errorf("expected 200 (unknown services skipped), got %d", code)
	}
}

func TestConfigValidationServiceMissingURL(t *testing.T) {
	t.Parallel()
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
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for missing service URL, got %d", code)
	}
}

func TestImportTooManyCategories(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	categories := make([]models.Category, MaxCategories+1)
	for i := range categories {
		categories[i] = models.Category{ID: int64(i + 1), Name: "Cat" + strconv.Itoa(i), Icon: "star"}
	}
	cfg := models.EncryptedExport{
		Version:    2,
		Settings:   models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Categories: categories,
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for too many categories, got %d", code)
	}
}

func TestImportTooManyWidgets(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	widgets := make([]models.Widget, MaxWidgets+1)
	for i := range widgets {
		widgets[i] = models.Widget{CategoryID: 1, Type: "applink", Title: "W" + strconv.Itoa(i), Config: "{}", Width: 6, Height: 0}
	}
	cfg := models.EncryptedExport{
		Version:    2,
		Settings:   models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Categories: []models.Category{{ID: 1, Name: "Cat1", Icon: "star"}},
		Widgets:    widgets,
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for too many widgets, got %d", code)
	}
}

func TestImportTooManyServices(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	services := make([]models.ServiceExport, MaxServicesImport+1)
	for i := range services {
		services[i] = models.ServiceExport{Type: "portainer", Name: "S" + strconv.Itoa(i), URL: "http://example.com", Enabled: true, Config: "{}"}
	}
	cfg := models.EncryptedExport{
		Version:  2,
		Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Services: services,
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for too many services, got %d", code)
	}
}

func TestImportCategoryNameTooLong(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	longName := make([]byte, MaxNameLength+1)
	for i := range longName {
		longName[i] = 'a'
	}
	cfg := models.EncryptedExport{
		Version:    2,
		Settings:   models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Categories: []models.Category{{ID: 1, Name: string(longName), Icon: "star"}},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for category name too long, got %d", code)
	}
}

func TestImportWidgetTitleTooLong(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	longTitle := make([]byte, MaxNameLength+1)
	for i := range longTitle {
		longTitle[i] = 'a'
	}
	cfg := models.EncryptedExport{
		Version:    2,
		Settings:   models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Categories: []models.Category{{ID: 1, Name: "Cat1", Icon: "star"}},
		Widgets:    []models.Widget{{CategoryID: 1, Type: "applink", Title: string(longTitle), Config: "{}", Width: 6, Height: 0}},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for widget title too long, got %d", code)
	}
}

func TestImportWidgetHeightTooLarge(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	cfg := models.EncryptedExport{
		Version:    2,
		Settings:   models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Categories: []models.Category{{ID: 1, Name: "Cat1", Icon: "star"}},
		Widgets:    []models.Widget{{CategoryID: 1, Type: "applink", Title: "W1", Config: "{}", Width: 6, Height: MaxWidgetHeight + 1}},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for widget height too large, got %d", code)
	}
}

func TestImportServiceNameTooLong(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	longName := make([]byte, MaxNameLength+1)
	for i := range longName {
		longName[i] = 'a'
	}
	cfg := models.EncryptedExport{
		Version:  2,
		Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Services: []models.ServiceExport{{Type: "portainer", Name: string(longName), URL: "http://example.com", Enabled: true, Config: "{}"}},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for service name too long, got %d", code)
	}
}

func TestImportServiceURLTooLong(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	longURL := "http://" + string(make([]byte, MaxURLLength))
	cfg := models.EncryptedExport{
		Version:  2,
		Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Services: []models.ServiceExport{{Type: "portainer", Name: "Portainer", URL: longURL, Enabled: true, Config: "{}"}},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusBadRequest {
		t.Errorf("expected 400 for service URL too long, got %d", code)
	}
}

func TestYAMLRoundTrip(t *testing.T) {
	t.Parallel()
	r, h := configRouter(t)

	// Create data: category + widget + service + settings
	catBody, _ := json.Marshal(models.CreateCategoryRequest{Name: "Monitoring", Icon: "activity"})
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
		PosX: 0, PosY: 0, Width: 12, Height: 0,
	})
	widgetReq := httptest.NewRequest(http.MethodPost, "/api/categories/"+itoa(cat.ID)+"/widgets", bytes.NewReader(widgetBody))
	widgetW := httptest.NewRecorder()
	r.ServeHTTP(widgetW, widgetReq)

	svcBody, _ := json.Marshal(models.UpdateServiceRequest{
		Name: "Portainer", URL: "https://portainer.local",
		Credentials: `{"token":"yaml-test-token"}`,
	})
	svcReq := httptest.NewRequest(http.MethodPut, "/api/services/portainer", bytes.NewReader(svcBody))
	svcW := httptest.NewRecorder()
	r.ServeHTTP(svcW, svcReq)

	theme := "dark"
	lang := "fr"
	interval := 30
	settingsBody, _ := json.Marshal(models.UpdateSettingsRequest{
		Theme: &theme, Language: &lang, RefreshInterval: &interval,
	})
	settingsReq := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(settingsBody))
	settingsW := httptest.NewRecorder()
	r.ServeHTTP(settingsW, settingsReq)

	// Export YAML
	exportReq := httptest.NewRequest(http.MethodGet, "/api/config/yaml", nil)
	exportW := httptest.NewRecorder()
	r.ServeHTTP(exportW, exportReq)
	if exportW.Code != http.StatusOK {
		t.Fatalf("YAML export: expected 200, got %d: %s", exportW.Code, exportW.Body.String())
	}
	yamlData := exportW.Body.Bytes()

	// Import YAML
	importReq := httptest.NewRequest(http.MethodPost, "/api/config/yaml", bytes.NewReader(yamlData))
	importReq.Header.Set("Content-Type", "application/x-yaml")
	importW := httptest.NewRecorder()
	r.ServeHTTP(importW, importReq)
	if importW.Code != http.StatusOK {
		t.Fatalf("YAML import: expected 200, got %d: %s", importW.Code, importW.Body.String())
	}

	// Verify data after round-trip
	ctx := context.Background()
	cats, _ := h.DB.GetCategories(ctx)
	if len(cats) != 1 || cats[0].Name != "Monitoring" {
		t.Errorf("expected category 'Monitoring', got %v", cats)
	}

	widgets, _ := h.DB.GetAllWidgets(ctx)
	if len(widgets) != 1 || widgets[0].Title != "Containers" {
		t.Errorf("expected widget 'Containers', got %v", widgets)
	}

	settings, _ := h.DB.GetSettings(ctx)
	if settings.Theme != "dark" || settings.Language != "fr" || settings.RefreshInterval != 30 {
		t.Errorf("settings mismatch: %+v", settings)
	}

	svc, _ := h.DB.GetServiceByType(ctx, "portainer")
	if svc == nil {
		t.Fatal("expected portainer service after YAML import")
	}
}

func TestImportWidgetLargeHeightAccepted(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	// Simulate a widget resized to 800px (height=80 in 10px row units)
	// This commonly happens when locking auto-height on tall widgets
	cfg := models.EncryptedExport{
		Version:  2,
		Settings: models.Settings{Theme: "dark", Language: "fr", RefreshInterval: 30},
		Categories: []models.Category{
			{ID: 1, Name: "Cat1", Icon: "star"},
		},
		Widgets: []models.Widget{
			{CategoryID: 1, Type: "docker", Title: "Tall Widget", Config: "{}", Width: 12, Height: 80},
		},
	}
	jsonData, _ := json.Marshal(cfg)
	password := "test"
	encrypted, _ := crypto.Encrypt(string(jsonData), crypto.DeriveKey(password)) //nolint:staticcheck // testing legacy format

	code := importEncrypted(t, r, password, exportResult{Data: encrypted})
	if code != http.StatusOK {
		t.Errorf("expected 200 for widget with height=80 (800px), got %d", code)
	}
}

func TestYAMLImportWidgetLargeHeightAccepted(t *testing.T) {
	t.Parallel()
	r, _ := configRouter(t)

	yamlContent := `version: 2
settings:
  theme: dark
  language: fr
  refresh_interval: 30
categories:
  - name: Infra
    icon: server
    sort_order: 0
    widgets:
      - type: docker
        title: Tall Docker
        config: "{}"
        pos_x: 0
        pos_y: 0
        width: 12
        height: 100
`
	req := httptest.NewRequest(http.MethodPost, "/api/config/yaml", bytes.NewReader([]byte(yamlContent)))
	req.Header.Set("Content-Type", "application/x-yaml")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for YAML widget with height=100 (1000px), got %d: %s", w.Code, w.Body.String())
	}
}

func itoa(n int64) string {
	return strconv.FormatInt(n, 10)
}
