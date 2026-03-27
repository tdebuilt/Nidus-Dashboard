package handlers

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	"github.com/tdebuilt/nidus/internal/testutil"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	return testutil.SetupTestDB(t)
}

func setupRequest(t *testing.T, body interface{}) *http.Request {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(b))
}

func TestSetupCreatesUser(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	req := setupRequest(t, models.SetupRequest{
		Username: "admin",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()

	h.Setup(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.SetupResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.User.Username != "admin" {
		t.Errorf("expected username admin, got %s", resp.User.Username)
	}
	if resp.User.ID == 0 {
		t.Error("expected non-zero user ID")
	}

	// Verify user exists in DB
	ctx := context.Background()
	user, err := db.GetUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if user == nil {
		t.Fatal("user not found in DB")
	}
	if user.Username != "admin" {
		t.Errorf("expected admin, got %s", user.Username)
	}
}

func TestSetupPasswordIsHashed(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	password := "securepassword123"
	req := setupRequest(t, models.SetupRequest{
		Username: "admin",
		Password: password,
	})
	w := httptest.NewRecorder()

	h.Setup(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	ctx := context.Background()
	user, err := db.GetUserByUsername(ctx, "admin")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}

	// Password hash should NOT be the plaintext password
	if user.PasswordHash == password {
		t.Fatal("password stored in plaintext!")
	}

	// But bcrypt.CompareHashAndPassword should succeed
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		t.Errorf("bcrypt comparison failed: %v", err)
	}
}

func TestSetupRejectedIfAdminExists(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	// First setup
	req1 := setupRequest(t, models.SetupRequest{
		Username: "admin",
		Password: "securepassword123",
	})
	w1 := httptest.NewRecorder()
	h.Setup(w1, req1)

	if w1.Code != http.StatusCreated {
		t.Fatalf("first setup failed: %d", w1.Code)
	}

	// Second setup should be rejected
	req2 := setupRequest(t, models.SetupRequest{
		Username: "admin2",
		Password: "anotherpassword123",
	})
	w2 := httptest.NewRecorder()
	h.Setup(w2, req2)

	if w2.Code != http.StatusConflict {
		t.Errorf("expected 409 Conflict, got %d", w2.Code)
	}

	var errResp models.ErrorResponse
	json.NewDecoder(w2.Body).Decode(&errResp)
	if errResp.Error == "" {
		t.Error("expected error message")
	}

	// Only one user should exist
	ctx := context.Background()
	count, err := db.CountUsers(ctx)
	if err != nil {
		t.Fatalf("failed to count users: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 user, got %d", count)
	}
}

func TestSetupGeneratesJWTSecret(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	req := setupRequest(t, models.SetupRequest{
		Username: "admin",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()
	h.Setup(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// JWT secret should be stored
	ctx := context.Background()
	secret, err := db.GetSystemSetting(ctx, "jwt_secret")
	if err != nil {
		t.Fatalf("failed to get jwt_secret: %v", err)
	}
	if secret == "" {
		t.Fatal("jwt_secret not generated")
	}
	// Should be 64 hex chars (32 bytes)
	if len(secret) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(secret))
	}
}

func TestSetupGeneratesEncryptionKey(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	req := setupRequest(t, models.SetupRequest{
		Username: "admin",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()
	h.Setup(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}

	// Encryption key should be stored
	ctx := context.Background()
	key, err := db.GetSystemSetting(ctx, "encryption_key")
	if err != nil {
		t.Fatalf("failed to get encryption_key: %v", err)
	}
	if key == "" {
		t.Fatal("encryption_key not generated")
	}
	if len(key) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(key))
	}
}

func TestSetupEmptyUsername(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	req := setupRequest(t, models.SetupRequest{
		Username: "",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()
	h.Setup(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSetupShortPassword(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	req := setupRequest(t, models.SetupRequest{
		Username: "admin",
		Password: "short",
	})
	w := httptest.NewRecorder()
	h.Setup(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSetupInvalidJSON(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	h.Setup(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestSetupPasswordNotInResponse(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	req := setupRequest(t, models.SetupRequest{
		Username: "admin",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()
	h.Setup(w, req)

	// Response JSON should not contain the password or hash
	body := w.Body.String()
	if bytes.Contains([]byte(body), []byte("securepassword123")) {
		t.Error("password found in response body")
	}
	if bytes.Contains([]byte(body), []byte("password_hash")) {
		t.Error("password_hash found in response body")
	}
}

// --- Login endpoint tests ---

// setupAdminUser creates an admin user via the setup endpoint and returns the handler.
func setupAdminUser(t *testing.T, db *database.DB, username, password string) *AuthHandler {
	t.Helper()
	h := &AuthHandler{DB: db}

	req := setupRequest(t, models.SetupRequest{
		Username: username,
		Password: password,
	})
	w := httptest.NewRecorder()
	h.Setup(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("setup failed: %d: %s", w.Code, w.Body.String())
	}
	return h
}

func loginRequest(t *testing.T, body interface{}) *http.Request {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}
	return httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(b))
}

func TestLoginSuccess(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	req := loginRequest(t, models.LoginRequest{
		Username: "admin",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.LoginResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.User.Username != "admin" {
		t.Errorf("expected username admin, got %s", resp.User.Username)
	}
	if resp.Message != "login successful" {
		t.Errorf("expected 'login successful', got %s", resp.Message)
	}
}

func TestLoginSetsJWTCookie(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	req := loginRequest(t, models.LoginRequest{
		Username: "admin",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Check cookie
	cookies := w.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "nidus_token" {
			tokenCookie = c
			break
		}
	}
	if tokenCookie == nil {
		t.Fatal("nidus_token cookie not set")
	}
	if !tokenCookie.HttpOnly {
		t.Error("cookie should be HttpOnly")
	}
	if tokenCookie.Secure {
		t.Error("cookie should not be Secure over plain HTTP")
	}
	if tokenCookie.SameSite != http.SameSiteLaxMode {
		t.Error("cookie should be SameSite=Lax over plain HTTP")
	}
	if tokenCookie.MaxAge != 24*60*60 {
		t.Errorf("expected MaxAge %d, got %d", 24*60*60, tokenCookie.MaxAge)
	}

	// Verify JWT is valid and contains correct claims
	ctx := context.Background()
	jwtSecretHex, _ := db.GetSystemSetting(ctx, "jwt_secret")
	jwtSecret := make([]byte, 32)
	_, _ = hex.Decode(jwtSecret, []byte(jwtSecretHex))

	token, err := jwt.Parse(tokenCookie.Value, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		t.Fatalf("failed to parse JWT: %v", err)
	}
	if !token.Valid {
		t.Error("JWT token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		t.Fatal("failed to get claims")
	}
	// Verify user ID in claims
	if subVal, ok := claims["sub"].(float64); !ok || subVal != 1 {
		t.Errorf("expected sub=1, got %v", claims["sub"])
	}
}

func TestLoginWrongPassword(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	req := loginRequest(t, models.LoginRequest{
		Username: "admin",
		Password: "wrongpassword123",
	})
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var errResp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error != "invalid credentials" {
		t.Errorf("expected 'invalid credentials', got %s", errResp.Error)
	}
}

func TestLoginNonexistentUser(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	req := loginRequest(t, models.LoginRequest{
		Username: "notexist",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLoginEmptyFields(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	// Empty username
	req := loginRequest(t, models.LoginRequest{
		Username: "",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty username, got %d", w.Code)
	}

	// Empty password
	req2 := loginRequest(t, models.LoginRequest{
		Username: "admin",
		Password: "",
	})
	w2 := httptest.NewRecorder()
	h.Login(w2, req2)

	if w2.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty password, got %d", w2.Code)
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLoginTOTPRequiredWhenEnabled(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	// Manually enable TOTP for the user
	_, err := db.Exec("UPDATE users SET totp_enabled = 1, totp_secret = 'JBSWY3DPEHPK3PXP' WHERE username = 'admin'")
	if err != nil {
		t.Fatalf("failed to enable TOTP: %v", err)
	}

	// Login without TOTP code
	req := loginRequest(t, models.LoginRequest{
		Username: "admin",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}

	var errResp models.ErrorResponse
	json.NewDecoder(w.Body).Decode(&errResp)
	if errResp.Error != "totp_required" {
		t.Errorf("expected 'totp_required', got %s", errResp.Error)
	}
}

func TestLoginPasswordNotInResponse(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	req := loginRequest(t, models.LoginRequest{
		Username: "admin",
		Password: "securepassword123",
	})
	w := httptest.NewRecorder()
	h.Login(w, req)

	body := w.Body.String()
	if strings.Contains(body, "securepassword123") {
		t.Error("password found in response body")
	}
	if strings.Contains(body, "password_hash") {
		t.Error("password_hash found in response body")
	}
}

func TestLoginNoCookieOnFailure(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	req := loginRequest(t, models.LoginRequest{
		Username: "admin",
		Password: "wrongpassword",
	})
	w := httptest.NewRecorder()
	h.Login(w, req)

	cookies := w.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "nidus_token" {
			t.Error("nidus_token cookie should not be set on failed login")
		}
	}
}

// --- TOTP management tests ---

// loginAndGetCookie logs in and returns the JWT cookie for authenticated requests.
func loginAndGetCookie(t *testing.T, h *AuthHandler, username, password string) *http.Cookie {
	t.Helper()
	req := loginRequest(t, models.LoginRequest{
		Username: username,
		Password: password,
	})
	w := httptest.NewRecorder()
	h.Login(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d: %s", w.Code, w.Body.String())
	}

	for _, c := range w.Result().Cookies() {
		if c.Name == "nidus_token" {
			return c
		}
	}
	t.Fatal("nidus_token cookie not found")
	return nil
}

func TestTOTPGenerateReturnsSecretAndQR(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")
	cookie := loginAndGetCookie(t, h, "admin", "securepassword123")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/totp/generate", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	h.TOTPGenerate(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.TOTPGenerateResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Secret == "" {
		t.Error("expected non-empty secret")
	}
	if resp.URL == "" {
		t.Error("expected non-empty URL")
	}
	if !strings.HasPrefix(resp.QR, "data:image/png;base64,") {
		t.Errorf("expected QR to be base64 PNG data URL, got prefix: %s", resp.QR[:30])
	}
	if !strings.Contains(resp.URL, "otpauth://") {
		t.Error("expected URL to be otpauth:// URL")
	}
}

func TestTOTPEnableWithValidCode(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")
	cookie := loginAndGetCookie(t, h, "admin", "securepassword123")

	// Generate TOTP secret
	genReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/generate", nil)
	genReq.AddCookie(cookie)
	genW := httptest.NewRecorder()
	h.TOTPGenerate(genW, genReq)

	var genResp models.TOTPGenerateResponse
	json.NewDecoder(genW.Body).Decode(&genResp)

	// Generate a valid TOTP code
	code, err := totp.GenerateCode(genResp.Secret, time.Now())
	if err != nil {
		t.Fatalf("failed to generate TOTP code: %v", err)
	}

	// Enable TOTP with valid code
	enableBody, _ := json.Marshal(models.TOTPEnableRequest{Code: code})
	enableReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/enable", bytes.NewReader(enableBody))
	enableReq.AddCookie(cookie)
	enableW := httptest.NewRecorder()
	h.TOTPEnable(enableW, enableReq)

	if enableW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", enableW.Code, enableW.Body.String())
	}

	// Verify TOTP is enabled in DB
	ctx := context.Background()
	user, _ := db.GetUserByUsername(ctx, "admin")
	if !user.TOTPEnabled {
		t.Error("TOTP should be enabled")
	}
}

func TestTOTPEnableWithInvalidCode(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")
	cookie := loginAndGetCookie(t, h, "admin", "securepassword123")

	// Generate TOTP secret
	genReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/generate", nil)
	genReq.AddCookie(cookie)
	genW := httptest.NewRecorder()
	h.TOTPGenerate(genW, genReq)

	// Try to enable with invalid code
	enableBody, _ := json.Marshal(models.TOTPEnableRequest{Code: "000000"})
	enableReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/enable", bytes.NewReader(enableBody))
	enableReq.AddCookie(cookie)
	enableW := httptest.NewRecorder()
	h.TOTPEnable(enableW, enableReq)

	if enableW.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", enableW.Code)
	}

	// Verify TOTP is NOT enabled
	ctx := context.Background()
	user, _ := db.GetUserByUsername(ctx, "admin")
	if user.TOTPEnabled {
		t.Error("TOTP should not be enabled with invalid code")
	}
}

func TestTOTPDisable(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")
	cookie := loginAndGetCookie(t, h, "admin", "securepassword123")

	// Generate + enable TOTP
	genReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/generate", nil)
	genReq.AddCookie(cookie)
	genW := httptest.NewRecorder()
	h.TOTPGenerate(genW, genReq)

	var genResp models.TOTPGenerateResponse
	json.NewDecoder(genW.Body).Decode(&genResp)

	code, _ := totp.GenerateCode(genResp.Secret, time.Now())
	enableBody, _ := json.Marshal(models.TOTPEnableRequest{Code: code})
	enableReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/enable", bytes.NewReader(enableBody))
	enableReq.AddCookie(cookie)
	enableW := httptest.NewRecorder()
	h.TOTPEnable(enableW, enableReq)

	if enableW.Code != http.StatusOK {
		t.Fatalf("enable failed: %d", enableW.Code)
	}

	// Disable TOTP
	disableReq := httptest.NewRequest(http.MethodDelete, "/api/auth/totp", nil)
	disableReq.AddCookie(cookie)
	disableW := httptest.NewRecorder()
	h.TOTPDisable(disableW, disableReq)

	if disableW.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", disableW.Code, disableW.Body.String())
	}

	// Verify TOTP is disabled and secret cleared
	ctx := context.Background()
	user, _ := db.GetUserByUsername(ctx, "admin")
	if user.TOTPEnabled {
		t.Error("TOTP should be disabled")
	}
	if user.TOTPSecret != nil {
		t.Error("TOTP secret should be cleared")
	}
}

func TestTOTPGenerateRequiresAuth(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	// No cookie
	req := httptest.NewRequest(http.MethodPost, "/api/auth/totp/generate", nil)
	w := httptest.NewRecorder()
	h.TOTPGenerate(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestTOTPGenerateRejectsIfAlreadyEnabled(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")
	cookie := loginAndGetCookie(t, h, "admin", "securepassword123")

	// Generate + enable TOTP
	genReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/generate", nil)
	genReq.AddCookie(cookie)
	genW := httptest.NewRecorder()
	h.TOTPGenerate(genW, genReq)

	var genResp models.TOTPGenerateResponse
	json.NewDecoder(genW.Body).Decode(&genResp)

	code, _ := totp.GenerateCode(genResp.Secret, time.Now())
	enableBody, _ := json.Marshal(models.TOTPEnableRequest{Code: code})
	enableReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/enable", bytes.NewReader(enableBody))
	enableReq.AddCookie(cookie)
	enableW := httptest.NewRecorder()
	h.TOTPEnable(enableW, enableReq)

	// Try to generate again
	genReq2 := httptest.NewRequest(http.MethodPost, "/api/auth/totp/generate", nil)
	genReq2.AddCookie(cookie)
	genW2 := httptest.NewRecorder()
	h.TOTPGenerate(genW2, genReq2)

	if genW2.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", genW2.Code)
	}
}

func TestLoginRequiresTOTPAfterEnable(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")
	cookie := loginAndGetCookie(t, h, "admin", "securepassword123")

	// Generate + enable TOTP
	genReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/generate", nil)
	genReq.AddCookie(cookie)
	genW := httptest.NewRecorder()
	h.TOTPGenerate(genW, genReq)

	var genResp models.TOTPGenerateResponse
	json.NewDecoder(genW.Body).Decode(&genResp)

	code, _ := totp.GenerateCode(genResp.Secret, time.Now())
	enableBody, _ := json.Marshal(models.TOTPEnableRequest{Code: code})
	enableReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/enable", bytes.NewReader(enableBody))
	enableReq.AddCookie(cookie)
	enableW := httptest.NewRecorder()
	h.TOTPEnable(enableW, enableReq)

	// Login without TOTP → should require it
	loginReq := loginRequest(t, models.LoginRequest{
		Username: "admin",
		Password: "securepassword123",
	})
	loginW := httptest.NewRecorder()
	h.Login(loginW, loginReq)

	if loginW.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", loginW.Code)
	}

	var errResp models.ErrorResponse
	json.NewDecoder(loginW.Body).Decode(&errResp)
	if errResp.Error != "totp_required" {
		t.Errorf("expected 'totp_required', got %s", errResp.Error)
	}

	// Login with valid TOTP → should succeed
	validCode, _ := totp.GenerateCode(genResp.Secret, time.Now())
	loginReq2 := loginRequest(t, models.LoginRequest{
		Username: "admin",
		Password: "securepassword123",
		TOTPCode: validCode,
	})
	loginW2 := httptest.NewRecorder()
	h.Login(loginW2, loginReq2)

	if loginW2.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", loginW2.Code, loginW2.Body.String())
	}

	// Login with invalid TOTP → should fail
	loginReq3 := loginRequest(t, models.LoginRequest{
		Username: "admin",
		Password: "securepassword123",
		TOTPCode: "000000",
	})
	loginW3 := httptest.NewRecorder()
	h.Login(loginW3, loginReq3)

	if loginW3.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", loginW3.Code)
	}
}

func TestTOTPDisableNotEnabledReturnsError(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")
	cookie := loginAndGetCookie(t, h, "admin", "securepassword123")

	req := httptest.NewRequest(http.MethodDelete, "/api/auth/totp", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	h.TOTPDisable(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- Auth status tests ---

func TestStatusNoAdmin(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	req := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	w := httptest.NewRecorder()
	h.Status(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.AuthStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if resp.SetupCompleted {
		t.Error("expected setup_completed=false when no admin exists")
	}
}

func TestStatusAfterSetup(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")

	req := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	w := httptest.NewRecorder()
	h.Status(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.AuthStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if !resp.SetupCompleted {
		t.Error("expected setup_completed=true after admin setup")
	}
}

func TestStatusIsPublic(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := &AuthHandler{DB: db}

	// No cookie, no auth header — should still work
	req := httptest.NewRequest(http.MethodGet, "/api/auth/status", nil)
	w := httptest.NewRecorder()
	h.Status(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 (public endpoint), got %d", w.Code)
	}
}

// --- Logout tests ---

func TestLogoutClearsCookie(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")
	cookie := loginAndGetCookie(t, h, "admin", "securepassword123")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	h.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Check that the cookie is cleared (MaxAge=-1)
	cookies := w.Result().Cookies()
	var tokenCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "nidus_token" {
			tokenCookie = c
			break
		}
	}
	if tokenCookie == nil {
		t.Fatal("expected nidus_token cookie in response")
	}
	if tokenCookie.MaxAge != -1 {
		t.Errorf("expected MaxAge=-1 to clear cookie, got %d", tokenCookie.MaxAge)
	}
	if tokenCookie.Value != "" {
		t.Errorf("expected empty cookie value, got %s", tokenCookie.Value)
	}
}

func TestLogoutThenRequestUnauthorized(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	h := setupAdminUser(t, db, "admin", "securepassword123")
	cookie := loginAndGetCookie(t, h, "admin", "securepassword123")

	// Logout
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(cookie)
	logoutW := httptest.NewRecorder()
	h.Logout(logoutW, logoutReq)

	if logoutW.Code != http.StatusOK {
		t.Fatalf("logout failed: %d", logoutW.Code)
	}

	// Extract the cleared cookie from the logout response
	var clearedCookie *http.Cookie
	for _, c := range logoutW.Result().Cookies() {
		if c.Name == "nidus_token" {
			clearedCookie = c
			break
		}
	}

	// Use the cleared cookie for a protected request (TOTP generate as proxy)
	protectedReq := httptest.NewRequest(http.MethodPost, "/api/auth/totp/generate", nil)
	if clearedCookie != nil {
		protectedReq.AddCookie(clearedCookie)
	}
	protectedW := httptest.NewRecorder()
	h.TOTPGenerate(protectedW, protectedReq)

	if protectedW.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 after logout, got %d", protectedW.Code)
	}
}

// Ensure CGO is available for SQLite tests
func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
