package handlers

import (
	"encoding/json"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/tdebuilt/nidus/internal/models"
)

// UpdateAccount allows the authenticated user to change their username and/or password.
func (h *AuthHandler) UpdateAccount(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromJWT(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "authentication required"})
		return
	}

	var req models.UpdateAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.CurrentPassword == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "current password is required"})
		return
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "incorrect current password"})
		return
	}

	hasChanges := false
	newUsername := user.Username

	// Update username if provided
	if req.Username != nil && *req.Username != user.Username {
		if *req.Username == "" {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "username cannot be empty"})
			return
		}
		existing, err := h.DB.GetUserByUsername(r.Context(), *req.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
			return
		}
		if existing != nil {
			writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "username already taken"})
			return
		}
		if err := h.DB.UpdateUserUsername(r.Context(), user.ID, *req.Username); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update username"})
			return
		}
		newUsername = *req.Username
		hasChanges = true
	}

	// Update password if provided
	if req.NewPassword != nil {
		if len(*req.NewPassword) < 8 {
			writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "password must be at least 8 characters"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to hash password"})
			return
		}
		if err := h.DB.UpdateUserPassword(r.Context(), user.ID, string(hash)); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to update password"})
			return
		}
		hasChanges = true
	}

	if !hasChanges {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "no changes provided"})
		return
	}

	// Invalidate old JWTs and re-issue a new one
	h.DB.IncrementTokenVersion(r.Context(), user.ID)
	updatedUser, err := h.DB.GetUserByID(r.Context(), user.ID)
	if err != nil || updatedUser == nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to refresh user"})
		return
	}

	if err := h.issueJWTCookie(w, r, updatedUser); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, models.UpdateAccountResponse{
		Message: "account updated",
		User: models.User{
			ID:          updatedUser.ID,
			Username:    newUsername,
			Role:        updatedUser.Role,
			TOTPEnabled: updatedUser.TOTPEnabled,
		},
	})
}
