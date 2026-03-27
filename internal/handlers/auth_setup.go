package handlers

import (
	"encoding/json"
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
	// Check if admin already exists
	count, err := h.DB.CountUsers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if count > 0 {
		writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "admin account already exists"})
		return
	}

	// Parse request
	var req models.SetupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	// Validate
	if req.Username == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "username is required"})
		return
	}
	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "password must be at least 8 characters"})
		return
	}

	// Hash password with bcrypt
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to hash password"})
		return
	}

	// Create user
	userID, err := h.DB.CreateUser(r.Context(), req.Username, string(hash), models.RoleAdmin)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create user"})
		return
	}

	// Generate JWT secret if not already set
	if existing, _ := h.DB.GetSystemSetting(r.Context(), "jwt_secret"); existing == "" {
		jwtSecret, err := crypto.GenerateKey()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate JWT secret"})
			return
		}
		if err := h.DB.SetSystemSetting(r.Context(), "jwt_secret", jwtSecret); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to store JWT secret"})
			return
		}
	}

	// Generate encryption key if not already set
	if existing, _ := h.DB.GetSystemSetting(r.Context(), "encryption_key"); existing == "" {
		encKey, err := crypto.GenerateKey()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate encryption key"})
			return
		}
		if err := h.DB.SetSystemSetting(r.Context(), "encryption_key", encKey); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to store encryption key"})
			return
		}
	}

	user := models.User{
		ID:       userID,
		Username: req.Username,
	}

	writeJSON(w, http.StatusCreated, models.SetupResponse{
		Message: "admin account created",
		User:    user,
	})
}
