package handlers

import (
	"encoding/hex"
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/golang-jwt/jwt/v5"
	ws "github.com/gorilla/websocket"

	"github.com/tdebuilt/nidus/internal/database"
	"github.com/tdebuilt/nidus/internal/models"
	nidusws "github.com/tdebuilt/nidus/internal/websocket"
)

// WebSocketHandler handles WebSocket upgrade requests with manual JWT auth.
type WebSocketHandler struct {
	DB      *database.DB
	Hub     *nidusws.Hub
	BaseURL string
}

// newUpgrader creates a WebSocket upgrader that validates the origin against the configured base URL.
func (h *WebSocketHandler) newUpgrader() ws.Upgrader {
	return ws.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			origin := r.Header.Get("Origin")
			if origin == "" {
				return true
			}
			if h.BaseURL != "" {
				if parsed, err := url.Parse(h.BaseURL); err == nil {
					allowed := parsed.Scheme + "://" + parsed.Host
					return origin == allowed
				}
			}
			host := r.Host
			return origin == "https://"+host || origin == "http://"+host
		},
	}
}

// HandleWS upgrades the HTTP connection to WebSocket after JWT validation.
func (h *WebSocketHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authenticateWS(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, models.ErrorResponse{Error: "authentication required"})
		return
	}

	upgrader := h.newUpgrader()
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("websocket upgrade failed", "error", err)
		return
	}

	client := nidusws.NewClient(h.Hub, conn, userID)
	h.Hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// authenticateWS extracts and validates JWT from cookie or query param.
func (h *WebSocketHandler) authenticateWS(r *http.Request) (int64, error) {
	cookie, err := r.Cookie("nidus_token")
	if err != nil || cookie.Value == "" {
		return 0, jwt.ErrTokenNotValidYet
	}

	jwtSecretHex, err := h.DB.GetSystemSetting(r.Context(), "jwt_secret")
	if err != nil || jwtSecretHex == "" {
		return 0, jwt.ErrTokenNotValidYet
	}
	jwtSecret, err := hex.DecodeString(jwtSecretHex)
	if err != nil {
		return 0, err
	}

	userID, tokenVersion, err := parseWSToken(cookie.Value, jwtSecret)
	if err != nil {
		return 0, err
	}

	user, err := h.DB.GetUserByID(r.Context(), userID)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return 0, err
	}
	if user == nil {
		return 0, jwt.ErrTokenInvalidClaims
	}

	if tokenVersion >= 0 && tokenVersion != user.TokenVersion {
		return 0, jwt.ErrTokenInvalidClaims
	}
	return userID, nil
}

// parseWSToken validates a JWT string and extracts the user ID and token version from claims.
// Returns tokenVersion=-1 if the claim is absent (legacy tokens).
func parseWSToken(tokenString string, jwtSecret []byte) (int64, int64, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return 0, 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, 0, jwt.ErrTokenInvalidClaims
	}
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, 0, jwt.ErrTokenInvalidClaims
	}
	var tokenVersion int64 = -1
	if tvClaim, ok := claims["tv"].(float64); ok {
		tokenVersion = int64(tvClaim)
	}
	return int64(userIDFloat), tokenVersion, nil
}
