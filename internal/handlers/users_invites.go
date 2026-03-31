package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/tdebuilt/nidus/internal/database"
	nidusmw "github.com/tdebuilt/nidus/internal/middleware"
	"github.com/tdebuilt/nidus/internal/models"
)

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
		slog.Error("invites: failed to generate code", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate code"})
		return
	}
	code := hex.EncodeToString(codeBytes)

	createdBy, _ := nidusmw.GetUserID(r.Context())
	expiresAt := time.Now().Add(InviteExpiry)

	if err := h.DB.CreateInvitation(r.Context(), code, req.Role, createdBy, expiresAt.Format(time.RFC3339)); err != nil {
		slog.Error("invites: failed to create invitation", "error", err)
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
		slog.Error("invites: failed to list invitations", "error", err)
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
		slog.Error("invites: failed to delete invitation", "id", invID, "error", err)
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

	inv, err := h.validateInvitation(r.Context(), req.Code, r.RemoteAddr)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	userID, err := h.createUserFromInvite(r.Context(), inv, req.Username, req.Password)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	slog.Info("register: new user created", "user_id", userID, "username", req.Username, "role", inv.Role)
	writeJSON(w, http.StatusCreated, models.SetupResponse{
		Message: "account created",
		User:    models.User{ID: userID, Username: req.Username, Role: inv.Role},
	})
}

// validateInvitation checks the invitation code and returns the invitation if valid.
func (h *UsersHandler) validateInvitation(ctx context.Context, code, remoteAddr string) (*models.Invitation, error) {
	inv, err := h.DB.GetInvitationByCode(ctx, code)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, &accountError{http.StatusInternalServerError, "database error"}
	}
	if inv == nil {
		slog.Warn("register: invalid or expired invitation code", "ip", remoteAddr)
		return nil, &accountError{http.StatusBadRequest, "invalid or expired invitation code"}
	}
	return inv, nil
}

// createUserFromInvite creates a user from an invitation, checking username uniqueness.
func (h *UsersHandler) createUserFromInvite(ctx context.Context, inv *models.Invitation, username, password string) (int64, error) {
	existing, err := h.DB.GetUserByUsername(ctx, username)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return 0, &accountError{http.StatusInternalServerError, "database error"}
	}
	if existing != nil {
		return 0, &accountError{http.StatusConflict, "username already taken"}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), BcryptCost)
	if err != nil {
		return 0, &accountError{http.StatusInternalServerError, "failed to hash password"}
	}
	userID, err := h.DB.CreateUser(ctx, username, string(hash), inv.Role)
	if err != nil {
		return 0, &accountError{http.StatusInternalServerError, "failed to create user"}
	}
	if err := h.DB.UseInvitation(ctx, inv.ID, userID); err != nil {
		return 0, &accountError{http.StatusInternalServerError, "failed to use invitation"}
	}
	return userID, nil
}

// CreateReset generates a password reset code for a user (admin only).
func (h *UsersHandler) CreateReset(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	userID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid user ID"})
		return
	}

	currentUserID, _ := nidusmw.GetUserID(r.Context())
	if userID == currentUserID {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "cannot reset your own account"})
		return
	}

	target, err := h.DB.GetUserByID(r.Context(), userID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		slog.Error("password-reset: failed to fetch user", "user_id", userID, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if target == nil {
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "user not found"})
		return
	}

	codeBytes := make([]byte, 16)
	if _, err := rand.Read(codeBytes); err != nil {
		slog.Error("password-reset: failed to generate code", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate code"})
		return
	}
	code := hex.EncodeToString(codeBytes)

	expiresAt := time.Now().Add(ResetExpiry)

	if err := h.DB.CreatePasswordReset(r.Context(), userID, code, currentUserID, expiresAt.Format(time.RFC3339)); err != nil {
		slog.Error("password-reset: failed to create reset code", "user_id", userID, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create reset code"})
		return
	}

	writeJSON(w, http.StatusCreated, models.CreateResetResponse{
		Code:      code,
		ExpiresAt: expiresAt.Format(time.RFC3339),
	})
}

// applyPasswordReset hashes the new password, updates the user record,
// disables TOTP if enabled, and invalidates existing sessions.
func (h *UsersHandler) applyPasswordReset(ctx context.Context, user *models.User, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), BcryptCost)
	if err != nil {
		return fmt.Errorf("failed to hash password")
	}
	if err := h.DB.UpdateUserPassword(ctx, user.ID, string(hash)); err != nil {
		return fmt.Errorf("failed to update password")
	}
	if user.TOTPEnabled {
		if err := h.DB.DisableUserTOTP(ctx, user.ID); err != nil {
			return fmt.Errorf("failed to disable TOTP")
		}
	}
	if err := h.DB.IncrementTokenVersion(ctx, user.ID); err != nil {
		return fmt.Errorf("failed to invalidate sessions")
	}
	return nil
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

	reset, err := h.DB.GetPasswordResetByCode(r.Context(), req.Code)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		slog.Error("password-reset: failed to fetch reset code", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if reset == nil {
		slog.Warn("password-reset: invalid or expired code", "ip", r.RemoteAddr)
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid or expired reset code"})
		return
	}

	user, err := h.DB.GetUserByID(r.Context(), reset.UserID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		slog.Error("password-reset: failed to fetch user for reset", "user_id", reset.UserID, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if user == nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "user not found"})
		return
	}

	if err := h.applyPasswordReset(r.Context(), user, req.NewPassword); err != nil {
		slog.Error("password-reset: failed to apply reset", "user_id", user.ID, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: sanitizeError(err)})
		return
	}

	if err := h.DB.UsePasswordReset(r.Context(), reset.ID); err != nil {
		slog.Warn("failed to mark password reset as used", "error", err)
	}

	slog.Info("password-reset: completed", "user_id", user.ID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "password reset successful"})
}
