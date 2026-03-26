package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/tdebuilt/nidus/internal/models"
)

func settingsRouter(t *testing.T) *chi.Mux {
	t.Helper()
	db := setupTestDB(t)
	h := &SettingsHandler{DB: db}
	r := chi.NewRouter()
	r.Get("/api/settings", h.Get)
	r.Put("/api/settings", h.Update)
	return r
}

func TestSettingsGetDefaults(t *testing.T) {
	r := settingsRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var settings models.Settings
	if err := json.NewDecoder(w.Body).Decode(&settings); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}

	if settings.Theme != "dark" {
		t.Errorf("expected default theme 'dark', got '%s'", settings.Theme)
	}
	if settings.Language != "fr" {
		t.Errorf("expected default language 'fr', got '%s'", settings.Language)
	}
	if settings.RefreshInterval != 30 {
		t.Errorf("expected default refresh_interval 30, got %d", settings.RefreshInterval)
	}
}

func TestSettingsUpdateTheme(t *testing.T) {
	r := settingsRouter(t)

	theme := "light"
	body, _ := json.Marshal(models.UpdateSettingsRequest{Theme: &theme})
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var settings models.Settings
	json.NewDecoder(w.Body).Decode(&settings)
	if settings.Theme != "light" {
		t.Errorf("expected theme 'light', got '%s'", settings.Theme)
	}
	// Other settings should remain defaults
	if settings.Language != "fr" {
		t.Errorf("expected language 'fr', got '%s'", settings.Language)
	}
	if settings.RefreshInterval != 30 {
		t.Errorf("expected refresh_interval 30, got %d", settings.RefreshInterval)
	}
}

func TestSettingsUpdateLanguage(t *testing.T) {
	r := settingsRouter(t)

	lang := "en"
	body, _ := json.Marshal(models.UpdateSettingsRequest{Language: &lang})
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var settings models.Settings
	json.NewDecoder(w.Body).Decode(&settings)
	if settings.Language != "en" {
		t.Errorf("expected language 'en', got '%s'", settings.Language)
	}
}

func TestSettingsUpdateRefreshInterval(t *testing.T) {
	r := settingsRouter(t)

	interval := 60
	body, _ := json.Marshal(models.UpdateSettingsRequest{RefreshInterval: &interval})
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var settings models.Settings
	json.NewDecoder(w.Body).Decode(&settings)
	if settings.RefreshInterval != 60 {
		t.Errorf("expected refresh_interval 60, got %d", settings.RefreshInterval)
	}
}

func TestSettingsUpdateMultiple(t *testing.T) {
	r := settingsRouter(t)

	theme := "light"
	lang := "en"
	interval := 10
	body, _ := json.Marshal(models.UpdateSettingsRequest{
		Theme:           &theme,
		Language:        &lang,
		RefreshInterval: &interval,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var settings models.Settings
	json.NewDecoder(w.Body).Decode(&settings)
	if settings.Theme != "light" {
		t.Errorf("expected theme 'light', got '%s'", settings.Theme)
	}
	if settings.Language != "en" {
		t.Errorf("expected language 'en', got '%s'", settings.Language)
	}
	if settings.RefreshInterval != 10 {
		t.Errorf("expected refresh_interval 10, got %d", settings.RefreshInterval)
	}
}

func TestSettingsRefreshIntervalTooLow(t *testing.T) {
	r := settingsRouter(t)

	interval := 2
	body, _ := json.Marshal(models.UpdateSettingsRequest{RefreshInterval: &interval})
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSettingsRefreshIntervalTooHigh(t *testing.T) {
	r := settingsRouter(t)

	interval := 500
	body, _ := json.Marshal(models.UpdateSettingsRequest{RefreshInterval: &interval})
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSettingsInvalidJSON(t *testing.T) {
	r := settingsRouter(t)

	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSettingsPersistence(t *testing.T) {
	r := settingsRouter(t)

	// Update
	theme := "light"
	lang := "en"
	interval := 15
	body, _ := json.Marshal(models.UpdateSettingsRequest{
		Theme:           &theme,
		Language:        &lang,
		RefreshInterval: &interval,
	})
	putReq := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	putW := httptest.NewRecorder()
	r.ServeHTTP(putW, putReq)

	if putW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", putW.Code)
	}

	// Read back with GET
	getReq := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	getW := httptest.NewRecorder()
	r.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", getW.Code)
	}

	var settings models.Settings
	json.NewDecoder(getW.Body).Decode(&settings)
	if settings.Theme != "light" {
		t.Errorf("expected theme 'light', got '%s'", settings.Theme)
	}
	if settings.Language != "en" {
		t.Errorf("expected language 'en', got '%s'", settings.Language)
	}
	if settings.RefreshInterval != 15 {
		t.Errorf("expected refresh_interval 15, got %d", settings.RefreshInterval)
	}
}

func TestSettingsPartialUpdatePreservesOthers(t *testing.T) {
	r := settingsRouter(t)

	// First update all
	theme := "light"
	lang := "en"
	interval := 60
	body, _ := json.Marshal(models.UpdateSettingsRequest{
		Theme:           &theme,
		Language:        &lang,
		RefreshInterval: &interval,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Now update only theme
	theme2 := "dark"
	body2, _ := json.Marshal(models.UpdateSettingsRequest{Theme: &theme2})
	req2 := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body2))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w2.Code)
	}

	var settings models.Settings
	json.NewDecoder(w2.Body).Decode(&settings)
	if settings.Theme != "dark" {
		t.Errorf("expected theme 'dark', got '%s'", settings.Theme)
	}
	if settings.Language != "en" {
		t.Errorf("expected language 'en' preserved, got '%s'", settings.Language)
	}
	if settings.RefreshInterval != 60 {
		t.Errorf("expected refresh_interval 60 preserved, got %d", settings.RefreshInterval)
	}
}

func TestSettingsRefreshIntervalBoundary(t *testing.T) {
	r := settingsRouter(t)

	// Min boundary (5) should work
	minInterval := 5
	body, _ := json.Marshal(models.UpdateSettingsRequest{RefreshInterval: &minInterval})
	req := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 for min boundary, got %d", w.Code)
	}

	// Max boundary (300) should work
	maxInterval := 300
	body2, _ := json.Marshal(models.UpdateSettingsRequest{RefreshInterval: &maxInterval})
	req2 := httptest.NewRequest(http.MethodPut, "/api/settings", bytes.NewReader(body2))
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected 200 for max boundary, got %d", w2.Code)
	}
}
