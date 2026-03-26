package handlers

import (
	"encoding/hex"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	ws "github.com/gorilla/websocket"

	"github.com/tdebuilt/nidus/internal/database"
	nidusws "github.com/tdebuilt/nidus/internal/websocket"
)

var upgrader = ws.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		host := r.Host
		return origin == "https://"+host || origin == "http://"+host
	},
}

// WebSocketHandler handles WebSocket upgrade requests with manual JWT auth.
type WebSocketHandler struct {
	DB  *database.DB
	Hub *nidusws.Hub
}

// HandleWS upgrades the HTTP connection to WebSocket after JWT validation.
func (h *WebSocketHandler) HandleWS(w http.ResponseWriter, r *http.Request) {
	userID, err := h.authenticateWS(r)
	if err != nil {
		http.Error(w, `{"error":"authentication required"}`, http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("websocket: upgrade error: %v", err)
		return
	}

	client := nidusws.NewClient(h.Hub, conn, userID)
	h.Hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}

// authenticateWS extracts and validates JWT from cookie or query param.
func (h *WebSocketHandler) authenticateWS(r *http.Request) (int64, error) {
	tokenString := ""

	// Try cookie first
	if cookie, err := r.Cookie("nidus_token"); err == nil {
		tokenString = cookie.Value
	}

	if tokenString == "" {
		return 0, jwt.ErrTokenNotValidYet
	}

	jwtSecretHex, err := h.DB.GetSystemSetting("jwt_secret")
	if err != nil || jwtSecretHex == "" {
		return 0, jwt.ErrTokenNotValidYet
	}

	jwtSecret, err := hex.DecodeString(jwtSecretHex)
	if err != nil {
		return 0, err
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return 0, jwt.ErrSignatureInvalid
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}

	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return 0, jwt.ErrTokenInvalidClaims
	}

	userID := int64(userIDFloat)

	user, err := h.DB.GetUserByID(userID)
	if err != nil || user == nil {
		return 0, jwt.ErrTokenInvalidClaims
	}

	return userID, nil
}
