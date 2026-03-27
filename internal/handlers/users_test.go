package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	nidusmw "github.com/tdebuilt/nidus/internal/middleware"
	"github.com/tdebuilt/nidus/internal/models"
)

func usersHandler(t *testing.T) *UsersHandler {
	t.Helper()
	db := setupTestDB(t)
	return &UsersHandler{DB: db}
}

func usersRouter(h *UsersHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Get("/api/users", h.List)
	r.Put("/api/users/{id}/role", h.UpdateRole)
	r.Delete("/api/users/{id}", h.Delete)
	r.Post("/api/users/{id}/reset", h.CreateReset)
	r.Get("/api/invites", h.ListInvites)
	r.Post("/api/invites", h.CreateInvite)
	r.Delete("/api/invites/{id}", h.DeleteInvite)
	r.Post("/api/auth/register", h.Register)
	r.Post("/api/auth/reset-password", h.ResetPassword)
	return r
}

// createAdminUser creates an admin user with a hashed password and returns the user ID.
func createAdminUser(t *testing.T, h *UsersHandler) int64 {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte("adminpassword1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	userID, err := h.DB.CreateUser(context.Background(), "admin", string(hash), "admin")
	if err != nil {
		t.Fatalf("failed to create admin user: %v", err)
	}
	return userID
}

// requestWithUser creates an HTTP request with the given user ID injected into context.
func requestWithUser(method, url string, body []byte, userID int64) *http.Request {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, url, bytes.NewReader(body))
	} else {
		req = httptest.NewRequest(method, url, nil)
	}
	ctx := context.WithValue(req.Context(), nidusmw.UserIDKey, userID)
	return req.WithContext(ctx)
}

// createInviteViaDB creates an invitation directly in the DB and returns the code.
func createInviteViaDB(t *testing.T, h *UsersHandler, role string, createdBy int64) string {
	t.Helper()
	code := fmt.Sprintf("testcode%d", time.Now().UnixNano())
	expiresAt := time.Now().Add(7 * 24 * time.Hour).Format(time.RFC3339)
	if err := h.DB.CreateInvitation(context.Background(), code, role, createdBy, expiresAt); err != nil {
		t.Fatalf("failed to create invitation: %v", err)
	}
	return code
}

// --- List Users ---

func TestListUsers_Empty(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var users []models.User
	if err := json.NewDecoder(w.Body).Decode(&users); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(users) != 0 {
		t.Errorf("expected empty list, got %d users", len(users))
	}
}

func TestListUsers_WithData(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)

	createAdminUser(t, h)
	hash, _ := bcrypt.GenerateFromPassword([]byte("viewerpass1"), bcrypt.DefaultCost)
	h.DB.CreateUser(context.Background(), "viewer", string(hash), "viewer")

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var users []models.User
	if err := json.NewDecoder(w.Body).Decode(&users); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

// --- Create Invite ---

func TestCreateInvite(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	body, _ := json.Marshal(models.CreateInviteRequest{Role: "viewer"})
	req := requestWithUser(http.MethodPost, "/api/invites", body, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.CreateInviteResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code == "" {
		t.Error("expected non-empty invite code")
	}
	if resp.Role != "viewer" {
		t.Errorf("expected role 'viewer', got '%s'", resp.Role)
	}
	if resp.ExpiresAt == "" {
		t.Error("expected non-empty expiration date")
	}
}

func TestCreateInvite_DefaultRole(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	body, _ := json.Marshal(models.CreateInviteRequest{})
	req := requestWithUser(http.MethodPost, "/api/invites", body, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.CreateInviteResponse
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.Role != "viewer" {
		t.Errorf("expected default role 'viewer', got '%s'", resp.Role)
	}
}

func TestCreateInvite_InvalidRole(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	body, _ := json.Marshal(models.CreateInviteRequest{Role: "superuser"})
	req := requestWithUser(http.MethodPost, "/api/invites", body, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- List Invites ---

func TestListInvites_Empty(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)

	req := httptest.NewRequest(http.MethodGet, "/api/invites", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var invites []models.Invitation
	json.NewDecoder(w.Body).Decode(&invites)
	if len(invites) != 0 {
		t.Errorf("expected empty list, got %d", len(invites))
	}
}

// --- Delete Invite ---

func TestDeleteInvite(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	code := createInviteViaDB(t, h, "viewer", adminID)

	// Get the invite to find its ID
	invites, _ := h.DB.ListInvitations(context.Background())
	var invID int64
	for _, inv := range invites {
		if inv.Code == code {
			invID = inv.ID
			break
		}
	}
	if invID == 0 {
		t.Fatal("could not find created invitation")
	}

	req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/invites/%d", invID), nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify the invite is gone
	invites, _ = h.DB.ListInvitations(context.Background())
	for _, inv := range invites {
		if inv.ID == invID {
			t.Error("invitation should have been deleted")
		}
	}
}

func TestDeleteInvite_InvalidID(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)

	req := httptest.NewRequest(http.MethodDelete, "/api/invites/abc", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- Register ---

func TestRegister(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	code := createInviteViaDB(t, h, "editor", adminID)

	body, _ := json.Marshal(models.RegisterRequest{
		Username: "newuser",
		Password: "securepassword123",
		Code:     code,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.SetupResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.User.Username != "newuser" {
		t.Errorf("expected username 'newuser', got '%s'", resp.User.Username)
	}
	if resp.User.Role != "editor" {
		t.Errorf("expected role 'editor', got '%s'", resp.User.Role)
	}
	if resp.User.ID == 0 {
		t.Error("expected non-zero user ID")
	}

	// Verify user exists in DB
	user, err := h.DB.GetUserByUsername(context.Background(), "newuser")
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	}
	if user == nil {
		t.Fatal("expected user to exist in DB")
	}
}

func TestRegister_InvalidInvite(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)

	body, _ := json.Marshal(models.RegisterRequest{
		Username: "newuser",
		Password: "securepassword123",
		Code:     "nonexistentcode",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRegister_MissingUsername(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)
	code := createInviteViaDB(t, h, "viewer", adminID)

	body, _ := json.Marshal(models.RegisterRequest{
		Username: "",
		Password: "securepassword123",
		Code:     code,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_ShortPassword(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)
	code := createInviteViaDB(t, h, "viewer", adminID)

	body, _ := json.Marshal(models.RegisterRequest{
		Username: "newuser",
		Password: "short",
		Code:     code,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_MissingCode(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)

	body, _ := json.Marshal(models.RegisterRequest{
		Username: "newuser",
		Password: "securepassword123",
		Code:     "",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)
	code := createInviteViaDB(t, h, "viewer", adminID)

	body, _ := json.Marshal(models.RegisterRequest{
		Username: "admin",
		Password: "securepassword123",
		Code:     code,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

// --- Update User Role ---

func TestUpdateUserRole(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	hash, _ := bcrypt.GenerateFromPassword([]byte("viewerpass1"), bcrypt.DefaultCost)
	viewerID, _ := h.DB.CreateUser(context.Background(), "viewer", string(hash), "viewer")

	body, _ := json.Marshal(models.UpdateUserRoleRequest{Role: "editor"})
	req := requestWithUser(http.MethodPut, fmt.Sprintf("/api/users/%d/role", viewerID), body, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify role changed in DB
	user, _ := h.DB.GetUserByID(context.Background(), viewerID)
	if user.Role != "editor" {
		t.Errorf("expected role 'editor', got '%s'", user.Role)
	}
}

func TestUpdateUserRole_InvalidRole(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	hash, _ := bcrypt.GenerateFromPassword([]byte("viewerpass1"), bcrypt.DefaultCost)
	viewerID, _ := h.DB.CreateUser(context.Background(), "viewer", string(hash), "viewer")

	body, _ := json.Marshal(models.UpdateUserRoleRequest{Role: "superuser"})
	req := requestWithUser(http.MethodPut, fmt.Sprintf("/api/users/%d/role", viewerID), body, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateUserRole_SelfModification(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	body, _ := json.Marshal(models.UpdateUserRoleRequest{Role: "viewer"})
	req := requestWithUser(http.MethodPut, fmt.Sprintf("/api/users/%d/role", adminID), body, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateUserRole_InvalidUserID(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	body, _ := json.Marshal(models.UpdateUserRoleRequest{Role: "editor"})
	req := requestWithUser(http.MethodPut, "/api/users/abc/role", body, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpdateUserRole_UserNotFound(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	body, _ := json.Marshal(models.UpdateUserRoleRequest{Role: "editor"})
	req := requestWithUser(http.MethodPut, "/api/users/999/role", body, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// --- Delete User ---

func TestDeleteUser(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	hash, _ := bcrypt.GenerateFromPassword([]byte("viewerpass1"), bcrypt.DefaultCost)
	viewerID, _ := h.DB.CreateUser(context.Background(), "viewer", string(hash), "viewer")

	req := requestWithUser(http.MethodDelete, fmt.Sprintf("/api/users/%d", viewerID), nil, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify user is deleted
	user, _ := h.DB.GetUserByID(context.Background(), viewerID)
	if user != nil {
		t.Error("user should have been deleted")
	}
}

func TestDeleteUser_SelfDeletion(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	req := requestWithUser(http.MethodDelete, fmt.Sprintf("/api/users/%d", adminID), nil, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	// Verify user still exists
	user, _ := h.DB.GetUserByID(context.Background(), adminID)
	if user == nil {
		t.Error("admin user should not have been deleted")
	}
}

func TestDeleteUser_NotFound(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	req := requestWithUser(http.MethodDelete, "/api/users/999", nil, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDeleteUser_InvalidID(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	req := requestWithUser(http.MethodDelete, "/api/users/abc", nil, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- Create Reset (admin generates reset code) ---

func TestCreateReset(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	hash, _ := bcrypt.GenerateFromPassword([]byte("viewerpass1"), bcrypt.DefaultCost)
	viewerID, _ := h.DB.CreateUser(context.Background(), "viewer", string(hash), "viewer")

	req := requestWithUser(http.MethodPost, fmt.Sprintf("/api/users/%d/reset", viewerID), nil, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}

	var resp models.CreateResetResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Code == "" {
		t.Error("expected non-empty reset code")
	}
	if resp.ExpiresAt == "" {
		t.Error("expected non-empty expiration date")
	}
}

func TestCreateReset_SelfReset(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	req := requestWithUser(http.MethodPost, fmt.Sprintf("/api/users/%d/reset", adminID), nil, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateReset_UserNotFound(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	req := requestWithUser(http.MethodPost, "/api/users/999/reset", nil, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// --- Reset Password (user resets with code) ---

func TestResetPassword(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	hash, _ := bcrypt.GenerateFromPassword([]byte("viewerpass1"), bcrypt.DefaultCost)
	viewerID, _ := h.DB.CreateUser(context.Background(), "viewer", string(hash), "viewer")

	// Create a reset code via DB
	resetCode := "testresetcode123"
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	if err := h.DB.CreatePasswordReset(context.Background(), viewerID, resetCode, adminID, expiresAt); err != nil {
		t.Fatalf("failed to create password reset: %v", err)
	}

	body, _ := json.Marshal(models.ResetPasswordRequest{
		Code:        resetCode,
		NewPassword: "newpassword123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify user can authenticate with new password
	user, _ := h.DB.GetUserByID(context.Background(), viewerID)
	if user == nil {
		t.Fatal("user should still exist")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("newpassword123")); err != nil {
		t.Error("password should have been updated to new value")
	}
}

func TestResetPassword_InvalidCode(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)

	body, _ := json.Marshal(models.ResetPasswordRequest{
		Code:        "invalidcode",
		NewPassword: "newpassword123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestResetPassword_MissingCode(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)

	body, _ := json.Marshal(models.ResetPasswordRequest{
		Code:        "",
		NewPassword: "newpassword123",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestResetPassword_ShortPassword(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	hash, _ := bcrypt.GenerateFromPassword([]byte("viewerpass1"), bcrypt.DefaultCost)
	viewerID, _ := h.DB.CreateUser(context.Background(), "viewer", string(hash), "viewer")

	resetCode := "testresetcode456"
	expiresAt := time.Now().Add(24 * time.Hour).Format(time.RFC3339)
	h.DB.CreatePasswordReset(context.Background(), viewerID, resetCode, adminID, expiresAt)

	body, _ := json.Marshal(models.ResetPasswordRequest{
		Code:        resetCode,
		NewPassword: "short",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/auth/reset-password", bytes.NewReader(body))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// --- Token Revocation (via UpdateRole side-effect) ---

func TestUpdateUserRole_InvalidatesToken(t *testing.T) {
	t.Parallel()
	h := usersHandler(t)
	r := usersRouter(h)
	adminID := createAdminUser(t, h)

	hash, _ := bcrypt.GenerateFromPassword([]byte("viewerpass1"), bcrypt.DefaultCost)
	viewerID, _ := h.DB.CreateUser(context.Background(), "viewer", string(hash), "viewer")

	// Get initial token version
	userBefore, _ := h.DB.GetUserByID(context.Background(), viewerID)
	versionBefore := userBefore.TokenVersion

	body, _ := json.Marshal(models.UpdateUserRoleRequest{Role: "editor"})
	req := requestWithUser(http.MethodPut, fmt.Sprintf("/api/users/%d/role", viewerID), body, adminID)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// Verify token version was incremented
	userAfter, _ := h.DB.GetUserByID(context.Background(), viewerID)
	if userAfter.TokenVersion <= versionBefore {
		t.Errorf("expected token version to be incremented: before=%d, after=%d",
			versionBefore, userAfter.TokenVersion)
	}
}
