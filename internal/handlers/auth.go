package handlers

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/png"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"

	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
)

// AuthHandler handles authentication-related HTTP requests.
type AuthHandler struct {
	DB *database.DB
}

// decryptTOTPSecret decrypts an encrypted TOTP secret.
// Falls back to using the value as-is if decryption fails (legacy unencrypted secrets).
func (h *AuthHandler) decryptTOTPSecret(secret string) (string, error) {
	encKey, err := h.DB.GetSystemSetting("encryption_key")
	if err != nil || encKey == "" {
		return secret, nil
	}
	decrypted, err := crypto.Decrypt(secret, encKey)
	if err != nil {
		return secret, nil
	}
	return decrypted, nil
}

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
	count, err := h.DB.CountUsers()
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
	userID, err := h.DB.CreateUser(req.Username, string(hash), models.RoleAdmin)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to create user"})
		return
	}

	// Generate JWT secret if not already set
	if existing, _ := h.DB.GetSystemSetting("jwt_secret"); existing == "" {
		jwtSecret, err := crypto.GenerateKey()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate JWT secret"})
			return
		}
		if err := h.DB.SetSystemSetting("jwt_secret", jwtSecret); err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to store JWT secret"})
			return
		}
	}

	// Generate encryption key if not already set
	if existing, _ := h.DB.GetSystemSetting("encryption_key"); existing == "" {
		encKey, err := crypto.GenerateKey()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate encryption key"})
			return
		}
		if err := h.DB.SetSystemSetting("encryption_key", encKey); err != nil {
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
	user, err := h.DB.GetUserByUsername(req.Username)
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
		decryptedSecret, err := h.decryptTOTPSecret(*user.TOTPSecret)
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

	jwtSecretHex, err := h.DB.GetSystemSetting("jwt_secret")
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

	return h.DB.GetUserByID(int64(userIDFloat))
}

// issueJWTCookie creates a JWT token for the user and sets it as an HTTP-only cookie.
func (h *AuthHandler) issueJWTCookie(w http.ResponseWriter, r *http.Request, user *models.User) error {
	jwtSecretHex, err := h.DB.GetSystemSetting("jwt_secret")
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
		"exp":  now.Add(24 * time.Hour).Unix(),
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
		MaxAge:   24 * 60 * 60,
	})
	return nil
}

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
		existing, err := h.DB.GetUserByUsername(*req.Username)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "database error"})
			return
		}
		if existing != nil {
			writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "username already taken"})
			return
		}
		if err := h.DB.UpdateUserUsername(user.ID, *req.Username); err != nil {
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
		if err := h.DB.UpdateUserPassword(user.ID, string(hash)); err != nil {
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
	h.DB.IncrementTokenVersion(user.ID)
	updatedUser, err := h.DB.GetUserByID(user.ID)
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

// TOTPGenerate godoc
// @Summary Generate a TOTP secret
// @Description Generates a new TOTP secret and returns it with a QR code for authenticator setup.
// @Tags auth
// @Produce json
// @Success 200 {object} models.TOTPGenerateResponse "TOTP secret and QR code"
// @Failure 401 {object} models.ErrorResponse "Authentication required"
// @Failure 409 {object} models.ErrorResponse "TOTP is already enabled"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /auth/totp/generate [post]
// @Security BearerAuth
func (h *AuthHandler) TOTPGenerate(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromJWT(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "authentication required"})
		return
	}

	if user.TOTPEnabled {
		writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "TOTP is already enabled"})
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Nidus",
		AccountName: user.Username,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate TOTP secret"})
		return
	}

	// Encrypt and store the secret (not yet enabled)
	encKey, err := h.DB.GetSystemSetting("encryption_key")
	if err != nil || encKey == "" {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
		return
	}
	encryptedSecret, err := crypto.Encrypt(key.Secret(), encKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to encrypt TOTP secret"})
		return
	}
	if err := h.DB.SetUserTOTPSecret(user.ID, encryptedSecret); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to store TOTP secret"})
		return
	}

	// Generate QR code as base64 PNG
	img, err := key.Image(200, 200)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to generate QR code"})
		return
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to encode QR code"})
		return
	}

	qrBase64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	writeJSON(w, http.StatusOK, models.TOTPGenerateResponse{
		Secret: key.Secret(),
		URL:    key.URL(),
		QR:     "data:image/png;base64," + qrBase64,
	})
}

// TOTPEnable godoc
// @Summary Enable TOTP for the current user
// @Description Verifies a TOTP code against the stored secret and enables TOTP two-factor authentication.
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.TOTPEnableRequest true "TOTP verification code"
// @Success 200 {object} map[string]string "TOTP enabled"
// @Failure 400 {object} models.ErrorResponse "Invalid request or secret not generated"
// @Failure 401 {object} models.ErrorResponse "Authentication required or invalid TOTP code"
// @Failure 409 {object} models.ErrorResponse "TOTP is already enabled"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /auth/totp/enable [post]
// @Security BearerAuth
func (h *AuthHandler) TOTPEnable(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromJWT(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "authentication required"})
		return
	}

	if user.TOTPEnabled {
		writeJSON(w, http.StatusConflict, models.ErrorResponse{Error: "TOTP is already enabled"})
		return
	}

	if user.TOTPSecret == nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "TOTP secret not generated, call generate first"})
		return
	}

	var req models.TOTPEnableRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "invalid request body"})
		return
	}

	if req.Code == "" {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "code is required"})
		return
	}

	// Decrypt and validate the code against the stored secret
	decryptedSecret, err := h.decryptTOTPSecret(*user.TOTPSecret)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to decrypt TOTP secret"})
		return
	}
	valid, err := totp.ValidateCustom(req.Code, decryptedSecret, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:     1,
		Digits:   otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil || !valid {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "invalid TOTP code"})
		return
	}

	// Enable TOTP
	if err := h.DB.EnableUserTOTP(user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to enable TOTP"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "TOTP enabled"})
}

// TOTPDisable godoc
// @Summary Disable TOTP for the current user
// @Description Disables TOTP two-factor authentication and clears the stored secret.
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "TOTP disabled"
// @Failure 400 {object} models.ErrorResponse "TOTP is not enabled"
// @Failure 401 {object} models.ErrorResponse "Authentication required"
// @Failure 500 {object} models.ErrorResponse "Server error"
// @Router /auth/totp [delete]
// @Security BearerAuth
func (h *AuthHandler) TOTPDisable(w http.ResponseWriter, r *http.Request) {
	user, err := h.getUserFromJWT(r)
	if err != nil || user == nil {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "authentication required"})
		return
	}

	if !user.TOTPEnabled {
		writeJSON(w, http.StatusBadRequest, models.ErrorResponse{Error: "TOTP is not enabled"})
		return
	}

	if err := h.DB.DisableUserTOTP(user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to disable TOTP"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "TOTP disabled"})
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
	count, err := h.DB.CountUsers()
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
