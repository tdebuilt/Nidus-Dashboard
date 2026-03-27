package handlers

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
)

const (
	JWTExpiry    = 24 * time.Hour
	CookieMaxAge = 24 * 60 * 60 // seconds
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	DB *database.DB
}

// Login godoc
// @Summary Authenticate a user
// @Description Verifies credentials, checks TOTP if enabled, and returns a JWT cookie.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "Login credentials"
// @Success 200 {object} models.LoginResponse "Login successful"
// @Failure 400 {object} models.ErrorResponse "Invalid request body"
// @Failure 401 {object} models.ErrorResponse "Invalid credentials or TOTP required"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "username and password are required"})
		return
	}

	// Fetch user
	user, err := h.DB.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}

	// Check TOTP if enabled
	if user.TOTPEnabled {
		if req.TOTPCode == "" {
			writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "totp_required"})
			return
		}
		if user.TOTPSecret == nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "TOTP secret not found"})
			return
		}
		decryptedSecret, err := h.decryptTOTPSecret(r.Context(), *user.TOTPSecret)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to decrypt TOTP secret"})
			return
		}
		valid := totp.Validate(req.TOTPCode, decryptedSecret)
		if !valid {
			writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "invalid TOTP code"})
			return
		}
	}

	if err := h.issueJWTCookie(w, r, user); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate token"})
		return
	}

	writeJSON(w, http.StatusOK, models.LoginResponse{
		Message: "login successful",
		User: models.User{
			ID:          user.ID,
			Username:    user.Username,
			Role:        user.Role,
			TOTPEnabled: user.TOTPEnabled,
		},
	})
}

// getUserFromJWT extracts the authenticated user from the JWT cookie.
// This is a temporary helper until the auth middleware is implemented in step 11.
func (h *AuthHandler) getUserFromJWT(r *http.Request) (*models.User, error) {
	cookie, err := r.Cookie("nidus_token")
	if err != nil {
		return nil, err
	}

	jwtSecretHex, err := h.DB.GetSystemSetting(r.Context(), "jwt_secret")
	if err != nil || jwtSecretHex == "" {
		return nil, err
	}

	jwtSecret, err := hex.DecodeString(jwtSecretHex)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(cookie.Value, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, jwt.ErrSignatureInvalid
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return nil, jwt.ErrSignatureInvalid
	}

	return h.DB.GetUserByID(r.Context(), int64(userIDFloat))
}

// issueJWTCookie creates a JWT token for the user and sets it as an HTTP-only cookie.
func (h *AuthHandler) issueJWTCookie(w http.ResponseWriter, r *http.Request, user *models.User) error {
	jwtSecretHex, err := h.DB.GetSystemSetting(r.Context(), "jwt_secret")
	if err != nil || jwtSecretHex == "" {
		return fmt.Errorf("JWT secret not configured")
	}

	jwtSecret, err := hex.DecodeString(jwtSecretHex)
	if err != nil {
		return fmt.Errorf("invalid JWT secret")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"role": user.Role,
		"tv":   user.TokenVersion,
		"iat":  now.Unix(),
		"exp":  now.Add(JWTExpiry).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return fmt.Errorf("failed to sign token")
	}

	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteStrictMode
	if !secure {
		sameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "nidus_token",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   CookieMaxAge,
	})
	return nil
}

// Status godoc
// @Summary Check setup status
// @Description Returns whether the initial setup has been completed (whether any users exist).
// @Tags auth
// @Produce json
// @Success 200 {object} models.AuthStatusResponse "Setup status"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /auth/status [get]
func (h *AuthHandler) Status(w http.ResponseWriter, r *http.Request) {
	count, err := h.DB.CountUsers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
		return
	}

	writeJSON(w, http.StatusOK, models.AuthStatusResponse{
		SetupCompleted: count > 0,
	})
}

// Logout godoc
// @Summary Log out the current user
// @Description Clears the JWT authentication cookie.
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "Logged out"
// @Router /auth/logout [post]
// @Security BearerAuth
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	secure := r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
	sameSite := http.SameSiteStrictMode
	if !secure {
		sameSite = http.SameSiteLaxMode
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "nidus_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSite,
		MaxAge:   -1,
	})

	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
