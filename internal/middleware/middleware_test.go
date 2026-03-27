package middleware

import (
	"context"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/testutil"
)

func setupTestDB(t *testing.T) *database.DB {
	t.Helper()
	return testutil.SetupTestDB(t)
}

// setupUserAndSecret creates a test user and JWT secret, returns the secret bytes.
func setupUserAndSecret(t *testing.T, db *database.DB) []byte {
	t.Helper()
	ctx := context.Background()
	_, err := db.CreateUser(ctx, "admin", "$2a$10$abcdefghijklmnopqrstuuABCDEFGHIJKLMNOPQRSTUVWXYZ012", "admin")
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Store a known JWT secret (32 bytes = 64 hex chars)
	secret := make([]byte, 32)
	for i := range secret {
		secret[i] = byte(i)
	}
	secretHex := hex.EncodeToString(secret)
	if err := db.SetSystemSetting(ctx, "jwt_secret", secretHex); err != nil {
		t.Fatalf("failed to set jwt_secret: %v", err)
	}
	return secret
}

func generateToken(t *testing.T, secret []byte, userID int64, expiry time.Duration) string {
	t.Helper()
	now := time.Now()
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": now.Unix(),
		"exp": now.Add(expiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}
	return tokenString
}

// protectedHandler is a simple handler that returns 200 and the user ID.
func protectedHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserID(r.Context())
	if !ok {
		http.Error(w, "no user id", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"user_id":` + itoa(userID) + `}`))
}

func itoa(n int64) string {
	return fmt.Sprintf("%d", n)
}

func TestAuthMiddlewareValidCookie(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	secret := setupUserAndSecret(t, db)

	tokenStr := generateToken(t, secret, 1, 7*24*time.Hour)

	handler := Auth(db)(http.HandlerFunc(protectedHandler))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "nidus_token", Value: tokenStr})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddlewareValidBearerHeader(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	secret := setupUserAndSecret(t, db)

	tokenStr := generateToken(t, secret, 1, 7*24*time.Hour)

	handler := Auth(db)(http.HandlerFunc(protectedHandler))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuthMiddlewareExpiredToken(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	secret := setupUserAndSecret(t, db)

	// Token expired 1 hour ago
	tokenStr := generateToken(t, secret, 1, -1*time.Hour)

	handler := Auth(db)(http.HandlerFunc(protectedHandler))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "nidus_token", Value: tokenStr})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareMissingToken(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	setupUserAndSecret(t, db)

	handler := Auth(db)(http.HandlerFunc(protectedHandler))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareInvalidToken(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	setupUserAndSecret(t, db)

	handler := Auth(db)(http.HandlerFunc(protectedHandler))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "nidus_token", Value: "not.a.valid.jwt"})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareWrongSigningKey(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	setupUserAndSecret(t, db)

	// Sign with a different key
	wrongSecret := make([]byte, 32)
	for i := range wrongSecret {
		wrongSecret[i] = byte(i + 100)
	}
	tokenStr := generateToken(t, wrongSecret, 1, 7*24*time.Hour)

	handler := Auth(db)(http.HandlerFunc(protectedHandler))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "nidus_token", Value: tokenStr})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareNonexistentUser(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	secret := setupUserAndSecret(t, db)

	// Token for user ID 999 (doesn't exist)
	tokenStr := generateToken(t, secret, 999, 7*24*time.Hour)

	handler := Auth(db)(http.HandlerFunc(protectedHandler))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "nidus_token", Value: tokenStr})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareInjectsUserID(t *testing.T) {
	t.Parallel()
	db := setupTestDB(t)
	secret := setupUserAndSecret(t, db)

	tokenStr := generateToken(t, secret, 1, 7*24*time.Hour)

	var capturedUserID int64
	var capturedOK bool
	handler := Auth(db)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserID, capturedOK = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.AddCookie(&http.Cookie{Name: "nidus_token", Value: tokenStr})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if !capturedOK {
		t.Fatal("user_id not found in context")
	}
	if capturedUserID != 1 {
		t.Errorf("expected user_id 1, got %d", capturedUserID)
	}
}

func TestGetUserIDWithoutMiddleware(t *testing.T) {
	t.Parallel()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	_, ok := GetUserID(req.Context())
	if ok {
		t.Error("expected no user_id in context without middleware")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
