package handlers

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"log/slog"
	"net/http"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"

	"github.com/tdebuilt/nidus/internal/crypto"
	"github.com/tdebuilt/nidus/internal/models"
)

// decryptTOTPSecret decrypts an encrypted TOTP secret.
func (h *AuthHandler) decryptTOTPSecret(ctx context.Context, secret string) (string, error) {
	encKey, err := h.DB.GetSystemSetting(ctx, "encryption_key")
	if err != nil {
		return "", fmt.Errorf("retrieving encryption key: %w", err)
	}
	if encKey == "" {
		return "", fmt.Errorf("encryption key not configured")
	}
	decrypted, err := crypto.Decrypt(secret, encKey)
	if err != nil {
		return "", fmt.Errorf("decrypting TOTP secret: %w", err)
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

	if err := h.encryptAndStoreTOTPSecret(r.Context(), user.ID, key.Secret()); err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	qr, err := generateQRBase64(key)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, models.TOTPGenerateResponse{
		Secret: key.Secret(),
		URL:    key.URL(),
		QR:     qr,
	})
}

// encryptAndStoreTOTPSecret encrypts the TOTP secret and saves it for the user.
func (h *AuthHandler) encryptAndStoreTOTPSecret(ctx context.Context, userID int64, secret string) error {
	encKey, err := h.DB.GetSystemSetting(ctx, "encryption_key")
	if err != nil || encKey == "" {
		return fmt.Errorf("encryption key not configured")
	}
	encrypted, err := crypto.Encrypt(secret, encKey)
	if err != nil {
		return fmt.Errorf("encrypting TOTP secret: %w", err)
	}
	if err := h.DB.SetUserTOTPSecret(ctx, userID, encrypted); err != nil {
		return fmt.Errorf("storing TOTP secret: %w", err)
	}
	return nil
}

// generateQRBase64 renders the TOTP key as a base64-encoded PNG data URI.
func generateQRBase64(key *otp.Key) (string, error) {
	img, err := key.Image(200, 200)
	if err != nil {
		return "", fmt.Errorf("generating QR code: %w", err)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", fmt.Errorf("encoding QR code: %w", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()), nil
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

	if err := h.validateTOTPCode(r.Context(), *user.TOTPSecret, req.Code); err != nil {
		slog.Warn("totp-enable: invalid verification code", "user_id", user.ID)
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "invalid TOTP code"})
		return
	}

	if err := h.DB.EnableUserTOTP(r.Context(), user.ID); err != nil {
		slog.Error("totp-enable: failed to enable TOTP", "user_id", user.ID, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to enable TOTP"})
		return
	}

	slog.Info("totp: enabled", "user_id", user.ID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "TOTP enabled"})
}

// validateTOTPCode decrypts the stored secret and validates the code against it.
func (h *AuthHandler) validateTOTPCode(ctx context.Context, encryptedSecret, code string) error {
	decrypted, err := h.decryptTOTPSecret(ctx, encryptedSecret)
	if err != nil {
		return fmt.Errorf("failed to decrypt TOTP secret: %w", err)
	}
	valid, err := totp.ValidateCustom(code, decrypted, time.Now(), totp.ValidateOpts{
		Period:    30,
		Skew:     1,
		Digits:   otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	})
	if err != nil || !valid {
		return fmt.Errorf("invalid TOTP code")
	}
	return nil
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
		slog.Error("totp-disable: failed to disable TOTP", "user_id", user.ID, "error", err)
		writeJSON(w, http.StatusInternalServerError, models.ErrorResponse{Error: "failed to disable TOTP"})
		return
	}

	slog.Info("totp: disabled", "user_id", user.ID)
	writeJSON(w, http.StatusOK, map[string]string{"message": "TOTP disabled"})
}
