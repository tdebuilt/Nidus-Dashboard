package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/tdebuilt/nidus/internal/database"
	nidusmw "github.com/tdebuilt/nidus/internal/middleware"
	"github.com/tdebuilt/nidus/internal/models"
)

const (
	MinPasswordLength = 8
	InviteExpiry      = 7 * 24 * time.Hour
	ResetExpiry       = 24 * time.Hour
)

// UsersHandler handles user management HTTP requests.
type UsersHandler struct {
	DB *database.DB
}

// List godoc
// @Summary List all users
// @Description Returns all registered users (admin only).
// @Tags users
// @Produce json
// @Success 200 {array} models.User "List of users"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /users [get]
// @Security BearerAuth
func (h *UsersHandler) List(w http.ResponseWriter, r *http.Request) {
	users, err := h.DB.ListUsers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list users"})
		return
	}
	if users == nil {
		users = []models.User{}
	}
	writeJSON(w, http.StatusOK, users)
}

// UpdateRole godoc
// @Summary Update a user's role
// @Description Changes a user's role (admin only). Cannot change your own role.
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param request body models.UpdateUserRoleRequest true "New role"
// @Success 200 {object} map[string]string "Role updated"
// @Failure 400 {object} models.ErrorResponse "Invalid user ID, invalid role, or self-modification"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Router /users/{id}/role [put]
// @Security BearerAuth
func (h *UsersHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid user ID"})
		return
	}

	// Prevent changing own role
	currentUserID, _ := nidusmw.GetUserID(r.Context())
	if userID == currentUserID {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "cannot change your own role"})
		return
	}

	var req models.UpdateUserRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if !models.IsValidRole(req.Role) {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid role, must be admin, editor, or viewer"})
		return
	}

	if err := h.DB.UpdateUserRole(r.Context(), userID, req.Role); err != nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "user not found"})
		return
	}

	// Invalidate existing JWTs for this user
	if err := h.DB.IncrementTokenVersion(r.Context(), userID); err != nil {
		log.Printf("warning: failed to increment token version for user %d: %v", userID, err)
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "role updated"})
}

// Delete godoc
// @Summary Delete a user
// @Description Deletes a user account (admin only). Cannot delete your own account.
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} map[string]string "User deleted"
// @Failure 400 {object} models.ErrorResponse "Invalid user ID or self-deletion"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Router /users/{id} [delete]
// @Security BearerAuth
func (h *UsersHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid user ID"})
		return
	}

	// Prevent deleting self
	currentUserID, _ := nidusmw.GetUserID(r.Context())
	if userID == currentUserID {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "cannot delete your own account"})
		return
	}

	if err := h.DB.DeleteUser(r.Context(), userID); err != nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "user not found"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}

// CreateInvite godoc
// @Summary Create an invitation code
// @Description Creates a new invitation code for user registration (admin only). Expires in 7 days.
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.CreateInviteRequest true "Invitation details"
// @Success 201 {object} models.CreateInviteResponse "Invitation created"
// @Failure 400 {object} models.ErrorResponse "Invalid request body or role"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /invites [post]
// @Security BearerAuth
func (h *UsersHandler) CreateInvite(w http.ResponseWriter, r *http.Request) {
	var req models.CreateInviteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Role == "" {
		req.Role = models.RoleViewer
	}
	if !models.IsValidRole(req.Role) {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid role"})
		return
	}

	// Generate random code (16 bytes = 32 hex chars)
	codeBytes := make([]byte, 16)
	if _, err := rand.Read(codeBytes); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate code"})
		return
	}
	code := hex.EncodeToString(codeBytes)

	createdBy, _ := nidusmw.GetUserID(r.Context())
	expiresAt := time.Now().Add(InviteExpiry)

	if err := h.DB.CreateInvitation(r.Context(), code, req.Role, createdBy, expiresAt.Format(time.RFC3339)); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create invitation"})
		return
	}

	writeJSON(w, http.StatusCreated, models.CreateInviteResponse{
		Code:      code,
		Role:      req.Role,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

// ListInvites godoc
// @Summary List all invitations
// @Description Returns all invitation codes (admin only).
// @Tags users
// @Produce json
// @Success 200 {array} models.Invitation "List of invitations"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /invites [get]
// @Security BearerAuth
func (h *UsersHandler) ListInvites(w http.ResponseWriter, r *http.Request) {
	invitations, err := h.DB.ListInvitations(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to list invitations"})
		return
	}
	if invitations == nil {
		invitations = []models.Invitation{}
	}
	writeJSON(w, http.StatusOK, invitations)
}

// DeleteInvite godoc
// @Summary Delete an invitation
// @Description Deletes an invitation code (admin only).
// @Tags users
// @Produce json
// @Param id path int true "Invitation ID"
// @Success 200 {object} map[string]string "Invitation deleted"
// @Failure 400 {object} models.ErrorResponse "Invalid invitation ID"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /invites/{id} [delete]
// @Security BearerAuth
func (h *UsersHandler) DeleteInvite(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	invID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid invitation ID"})
		return
	}

	if err := h.DB.DeleteInvitation(r.Context(), invID); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to delete invitation"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "invitation deleted"})
}

// Register godoc
// @Summary Register a new user
// @Description Creates a new user account using a valid invitation code.
// @Tags users
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "Registration details with invitation code"
// @Success 201 {object} models.SetupResponse "Account created"
// @Failure 400 {object} models.ErrorResponse "Invalid request or expired invitation"
// @Failure 409 {object} models.ErrorResponse "Username already taken"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /auth/register [post]
func (h *UsersHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Username == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "username is required"})
		return
	}
	if len(req.Password) < MinPasswordLength {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "password must be at least 8 characters"})
		return
	}
	if req.Code == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invitation code is required"})
		return
	}

	// Check invitation
	inv, err := h.DB.GetInvitationByCode(r.Context(), req.Code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if inv == nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid or expired invitation code"})
		return
	}

	// Check username not taken
	existing, err := h.DB.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if existing != nil {
		writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "username already taken"})
		return
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to hash password"})
		return
	}

	// Create user with the invitation's role
	userID, err := h.DB.CreateUser(r.Context(), req.Username, string(hash), inv.Role)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create user"})
		return
	}

	// Mark invitation as used
	if err := h.DB.UseInvitation(r.Context(), inv.ID, userID); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to use invitation"})
		return
	}

	writeJSON(w, http.StatusCreated, models.SetupResponse{
		Message: "account created",
		User: models.User{
			ID:       userID,
			Username: req.Username,
			Role:     inv.Role,
		},
	})
}

// CreateReset generates a password reset code for a user (admin only).
func (h *UsersHandler) CreateReset(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid user ID"})
		return
	}

	// Prevent resetting own account
	currentUserID, _ := nidusmw.GetUserID(r.Context())
	if userID == currentUserID {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "cannot reset your own account"})
		return
	}

	// Verify target user exists
	target, err := h.DB.GetUserByID(r.Context(), userID)
	if err != nil || target == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "user not found"})
		return
	}

	// Generate random code (16 bytes = 32 hex chars)
	codeBytes := make([]byte, 16)
	if _, err := rand.Read(codeBytes); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate code"})
		return
	}
	code := hex.EncodeToString(codeBytes)

	expiresAt := time.Now().Add(ResetExpiry)

	if err := h.DB.CreatePasswordReset(r.Context(), userID, code, currentUserID, expiresAt.Format(time.RFC3339)); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create reset code"})
		return
	}

	writeJSON(w, http.StatusCreated, models.CreateResetResponse{
		Code:      code,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

// ResetPassword allows a user to set a new password using a valid reset code.
func (h *UsersHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req models.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Code == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "reset code is required"})
		return
	}
	if len(req.NewPassword) < 8 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "password must be at least 8 characters"})
		return
	}

	// Validate reset code
	reset, err := h.DB.GetPasswordResetByCode(r.Context(), req.Code)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if reset == nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid or expired reset code"})
		return
	}

	// Verify user still exists
	user, err := h.DB.GetUserByID(r.Context(), reset.UserID)
	if err != nil || user == nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "user not found"})
		return
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to hash password"})
		return
	}

	// Update password
	if err := h.DB.UpdateUserPassword(r.Context(), user.ID, string(hash)); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update password"})
		return
	}

	// Always disable TOTP on reset
	if user.TOTPEnabled {
		if err := h.DB.DisableUserTOTP(r.Context(), user.ID); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to disable TOTP"})
			return
		}
	}

	// Invalidate all existing sessions
	if err := h.DB.IncrementTokenVersion(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to invalidate sessions"})
		return
	}

	// Mark reset code as used
	if err := h.DB.UsePasswordReset(r.Context(), reset.ID); err != nil {
		log.Printf("warning: failed to mark password reset as used: %v", err)
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successful"})
}
