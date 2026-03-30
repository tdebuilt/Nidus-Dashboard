package middleware

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"github.com/tdebuilt/nidus/internal/database"
)

// writeErrorJSON writes a JSON error response with proper Content-Type header.
func writeErrorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		slog.Error("writeErrorJSON: failed to encode response", "error", err)
	}
}

type contextKey string

const UserIDKey contextKey = "user_id"
const UserRoleKey contextKey = "user_role"

// Auth returns a middleware that validates JWT tokens from cookies or Bearer header.
// On success, it injects the user_id into the request context.
func Auth(db *database.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				writeErrorJSON(w, http.StatusUnauthorized, "authentication required")
				return
			}

			jwtSecretHex, err := db.GetSystemSetting(r.Context(), "jwt_secret")
			if err != nil || jwtSecretHex == "" {
				writeErrorJSON(w, http.StatusInternalServerError, "JWT secret not configured")
				return
			}

			jwtSecret, err := hex.DecodeString(jwtSecretHex)
			if err != nil {
				writeErrorJSON(w, http.StatusInternalServerError, "invalid JWT secret")
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return jwtSecret, nil
			})
			if err != nil || !token.Valid {
				writeErrorJSON(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeErrorJSON(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			userIDFloat, ok := claims["sub"].(float64)
			if !ok {
				writeErrorJSON(w, http.StatusUnauthorized, "invalid token claims")
				return
			}

			userID := int64(userIDFloat)

			// Verify user still exists
			user, err := db.GetUserByID(r.Context(), userID)
			if err != nil && !errors.Is(err, database.ErrNotFound) {
				writeErrorJSON(w, http.StatusInternalServerError, "database error")
				return
			}
			if user == nil {
				writeErrorJSON(w, http.StatusUnauthorized, "user not found")
				return
			}

			// Validate token version (invalidated on password/role change)
			if tvClaim, ok := claims["tv"].(float64); ok {
				if int64(tvClaim) != user.TokenVersion {
					writeErrorJSON(w, http.StatusUnauthorized, "token revoked")
					return
				}
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			ctx = context.WithValue(ctx, UserRoleKey, user.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// extractToken gets the JWT token from cookie or Authorization Bearer header.
func extractToken(r *http.Request) string {
	// Try cookie first
	if cookie, err := r.Cookie("nidus_token"); err == nil {
		return cookie.Value
	}

	// Try Authorization header
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}

	return ""
}

// GetUserID extracts the user ID from the request context.
func GetUserID(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(UserIDKey).(int64)
	return userID, ok
}

// GetUserRole extracts the user role from the request context.
func GetUserRole(ctx context.Context) string {
	role, _ := ctx.Value(UserRoleKey).(string)
	return role
}

// roleLevel returns the permission level for a role (higher = more permissions).
func roleLevel(role string) int {
	switch role {
	case "admin":
		return 3
	case "editor":
		return 2
	case "viewer":
		return 1
	default:
		return 0
	}
}

// RequireRole returns a middleware that only allows users with the given role or higher.
// Role hierarchy: admin > editor > viewer.
func RequireRole(minRole string) func(http.Handler) http.Handler {
	minLevel := roleLevel(minRole)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := GetUserRole(r.Context())
			if roleLevel(role) < minLevel {
				writeErrorJSON(w, http.StatusForbidden, "insufficient permissions")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
