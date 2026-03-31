package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

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
		slog.Error("users: failed to list", "error", err)
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
		slog.Warn("users: role update failed, user not found", "user_id", userID)
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "user not found"})
		return
	}

	// Invalidate existing JWTs for this user
	if err := h.DB.IncrementTokenVersion(r.Context(), userID); err != nil {
		slog.Warn("failed to increment token version", "user_id", userID, "error", err)
	}

	slog.Info("users: role updated", "user_id", userID, "new_role", req.Role)
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
		slog.Warn("users: delete failed, user not found", "user_id", userID)
		writeJSON(w, http.StatusNotFound, models.ErrorResponse{Error: "user not found"})
		return
	}

	slog.Info("users: deleted", "user_id", userID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "user deleted"})
}
