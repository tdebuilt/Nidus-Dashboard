package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"golang.org/x/crypto/bcrypt"

	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/models"
)

// Setup godoc
// @Summary Create the first admin account
// @Description Creates the first admin user and generates JWT secret + encryption key. Only works when no users exist.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.SetupRequest true "Admin credentials"
// @Success 201 {object} models.SetupResponse "Admin account created"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 409 {object} models.ErrorResponse "Admin account already exists"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /auth/setup [post]
func (h *AuthHandler) Setup(w http.ResponseWriter, r *http.Request) {
	count, err := h.DB.CountUsers(r.Context())
	if err != nil {
		slog.Error("setup: database error counting users", "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if count > 0 {
		slog.Warn("setup: attempted setup when admin already exists", "ip", r.RemoteAddr)
		writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "admin account already exists"})
		return
	}

	var req models.SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	hash, err := validateSetupCredentials(req)
	if err != nil {
		writeHandlerError(w, err)
		return
	}

	userID, err := h.DB.CreateUser(r.Context(), req.Username, string(hash), models.RoleAdmin)
	if err != nil {
		slog.Error("setup: failed to create admin user", "username", req.Username, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create user"})
		return
	}

	if err := h.ensureSystemKey(r, "jwt_secret"); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: sanitizeError(err)})
		return
	}
	if err := h.ensureSystemKey(r, "encryption_key"); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: sanitizeError(err)})
		return
	}

	slog.Info("setup: admin account created", "user_id", userID, "username", req.Username)
	writeJSON(w, http.StatusCreated, models.SetupResponse{
		Message: "admin account created",
		User:    models.User{ID: userID, Username: req.Username},
	})
}

// validateSetupCredentials validates username/password and returns the bcrypt hash.
func validateSetupCredentials(req models.SetupRequest) ([]byte, error) {
	if req.Username == "" {
		return nil, &accountError{http.StatusBadRequest, "username is required"}
	}
	if len(req.Password) < 8 {
		return nil, &accountError{http.StatusBadRequest, "password must be at least 8 characters"}
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), BcryptCost)
	if err != nil {
		slog.Error("setup: failed to hash password", "error", err)
		return nil, &accountError{http.StatusInternalServerError, "failed to hash password"}
	}
	return hash, nil
}

// ensureSystemKey generates and stores a cryptographic key if not already set.
func (h *AuthHandler) ensureSystemKey(r *http.Request, settingName string) error {
	if existing, _ := h.DB.GetSystemSetting(r.Context(), settingName); existing != "" {
		return nil
	}
	key, err := crypto.GenerateKey()
	if err != nil {
		return fmt.Errorf("failed to generate %s: %w", settingName, err)
	}
	if err := h.DB.SetSystemSetting(r.Context(), settingName, key); err != nil {
		return fmt.Errorf("failed to store %s: %w", settingName, err)
	}
	return nil
}
