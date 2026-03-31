package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/tdebuilt/nidus/internal/database"
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

	if err := verifyCurrentPassword(r, user, req.CurrentPassword); err != nil {
		writeHandlerError(w, err)
		return
	}

	newUsername, hasChanges, err := h.applyAccountChanges(r, user, &req)
	if err != nil {
		writeHandlerError(w, err)
		return
	}
	if !hasChanges {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "no changes provided"})
		return
	}

	updatedUser, err := h.refreshUserToken(w, r, user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: sanitizeError(err)})
		return
	}

	slog.Info("account: updated", "user_id", user.ID, "changed_username", req.Username != nil, "changed_password", req.NewPassword != nil)
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

// verifyCurrentPassword checks that the provided password matches the user's hash.
func verifyCurrentPassword(r *http.Request, user *models.User, password string) error {
	if password == "" {
		return &accountError{http.StatusBadRequest, "current password is required"}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		slog.Warn("account: incorrect current password", "user_id", user.ID, "ip", r.RemoteAddr)
		return &accountError{http.StatusUnauthorized, "incorrect current password"}
	}
	return nil
}

type accountError struct {
	Status  int
	Message string
}

func (e *accountError) Error() string { return e.Message }

// applyAccountChanges updates username and/or password, returning the new username and whether changes were made.
func (h *AuthHandler) applyAccountChanges(r *http.Request, user *models.User, req *models.UpdateAccountRequest) (string, bool, error) {
	hasChanges := false
	newUsername := user.Username

	if req.Username != nil && *req.Username != user.Username {
		if *req.Username == "" {
			return "", false, &accountError{http.StatusBadRequest, "username cannot be empty"}
		}
		existing, err := h.DB.GetUserByUsername(r.Context(), *req.Username)
		if err != nil && !errors.Is(err, database.ErrNotFound) {
			return "", false, &accountError{http.StatusInternalServerError, "database error"}
		}
		if existing != nil {
			return "", false, &accountError{http.StatusConflict, "username already taken"}
		}
		if err := h.DB.UpdateUserUsername(r.Context(), user.ID, *req.Username); err != nil {
			return "", false, &accountError{http.StatusInternalServerError, "failed to update username"}
		}
		newUsername = *req.Username
		hasChanges = true
	}

	if req.NewPassword != nil {
		if len(*req.NewPassword) < 8 {
			return "", false, &accountError{http.StatusBadRequest, "password must be at least 8 characters"}
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(*req.NewPassword), BcryptCost)
		if err != nil {
			return "", false, &accountError{http.StatusInternalServerError, "failed to hash password"}
		}
		if err := h.DB.UpdateUserPassword(r.Context(), user.ID, string(hash)); err != nil {
			return "", false, &accountError{http.StatusInternalServerError, "failed to update password"}
		}
		hasChanges = true
	}

	return newUsername, hasChanges, nil
}

// refreshUserToken invalidates old JWTs, re-issues a new one, and returns the updated user.
func (h *AuthHandler) refreshUserToken(w http.ResponseWriter, r *http.Request, userID int64) (*models.User, error) {
	if err := h.DB.IncrementTokenVersion(r.Context(), userID); err != nil {
		slog.Error("account: failed to increment token version", "user_id", userID, "error", err)
	}
	updatedUser, err := h.DB.GetUserByID(r.Context(), userID)
	if err != nil || updatedUser == nil {
		return nil, fmt.Errorf("failed to refresh user")
	}
	if err := h.issueJWTCookie(w, r, updatedUser); err != nil {
		slog.Error("account: failed to issue JWT after update", "user_id", userID, "error", err)
		return nil, fmt.Errorf("failed to generate token")
	}
	return updatedUser, nil
}
