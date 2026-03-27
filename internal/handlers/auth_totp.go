package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image/png"
	"net/http"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/models"
)

// decryptTOTPSecret decrypts an encrypted TOTP secret.
// Falls back to using the value as-is if decryption fails (legacy unencrypted secrets).
func (h *AuthHandler) decryptTOTPSecret(ctx context.Context, secret string) (string, error) {
	encKey, err := h.DB.GetSystemSetting(ctx, "encryption_key")
	if err != nil || encKey == "" {
		return secret, nil
	}
	decrypted, err := crypto.Decrypt(secret, encKey)
	if err != nil {
		return secret, nil
	}
	return decrypted, nil
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
	encKey, err := h.DB.GetSystemSetting(r.Context(), "encryption_key")
	if err != nil || encKey == "" {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "encryption key not configured"})
		return
	}
	encryptedSecret, err := crypto.Encrypt(key.Secret(), encKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to encrypt TOTP secret"})
		return
	}
	if err := h.DB.SetUserTOTPSecret(r.Context(), user.ID, encryptedSecret); err != nil {
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
	decryptedSecret, err := h.decryptTOTPSecret(r.Context(), *user.TOTPSecret)
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
	if err := h.DB.EnableUserTOTP(r.Context(), user.ID); err != nil {
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

	if err := h.DB.DisableUserTOTP(r.Context(), user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to disable TOTP"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"message": "TOTP disabled"})
}
