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

// userInfo holds the verified user identity extracted from a JWT token.
type userInfo struct {
	ID   int64
	Role string
}

// Auth returns a middleware that validates JWT tokens from cookies or Bearer header.
// On success, it injects the user_id and user_role into the request context.
func Auth(db *database.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := extractToken(r)
			if tokenString == "" {
				writeErrorJSON(w, http.StatusUnauthorized, "authentication required")
				return
			}

			jwtSecret, err := loadJWTSecret(r.Context(), db)
			if err != nil {
				slog.Error("auth: JWT secret error", "error", err)
				writeErrorJSON(w, http.StatusInternalServerError, err.Error())
				return
			}

			claims, err := validateJWTToken(tokenString, jwtSecret)
			if err != nil {
				writeErrorJSON(w, http.StatusUnauthorized, err.Error())
				return
			}

			info, err := verifyTokenUser(r.Context(), db, claims)
			if err != nil {
				var ae *authError
				if errors.As(err, &ae) {
					writeErrorJSON(w, http.StatusUnauthorized, ae.Error())
				} else {
					writeErrorJSON(w, http.StatusInternalServerError, err.Error())
				}
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, info.ID)
			ctx = context.WithValue(ctx, UserRoleKey, info.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// loadJWTSecret retrieves and decodes the JWT secret from the database.
func loadJWTSecret(ctx context.Context, db *database.DB) ([]byte, error) {
	jwtSecretHex, err := db.GetSystemSetting(ctx, "jwt_secret")
	if err != nil {
		return nil, errors.New("JWT secret not configured")
	}
	if jwtSecretHex == "" {
		return nil, errors.New("JWT secret not configured")
	}

	secret, err := hex.DecodeString(jwtSecretHex)
	if err != nil {
		return nil, errors.New("invalid JWT secret")
	}
	return secret, nil
}

// validateJWTToken parses and validates a JWT token string, returning the claims.
// Returns a user-facing error message distinguishing expired vs invalid tokens.
func validateJWTToken(tokenString string, jwtSecret []byte) (jwt.MapClaims, error) {
	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithExpirationRequired(),
	)
	token, err := parser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, errors.New("token expired")
		}
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}

// authError represents an authentication failure (401) with a user-facing message.
// It is distinguished from server errors (500) via errors.As in the caller.
type authError struct{ msg string }

func (e *authError) Error() string { return e.msg }

// verifyTokenUser checks that the user from the JWT claims still exists and that
// the token version is current. Returns *authError for 401 cases (invalid claims,
// user not found, token revoked) and plain errors for 500 cases (database error).
func verifyTokenUser(ctx context.Context, db *database.DB, claims jwt.MapClaims) (*userInfo, error) {
	userIDFloat, ok := claims["sub"].(float64)
	if !ok {
		return nil, &authError{"invalid token claims"}
	}
	userID := int64(userIDFloat)

	user, err := db.GetUserByID(ctx, userID)
	if errors.Is(err, database.ErrNotFound) || user == nil {
		return nil, &authError{"user not found"}
	}
	if err != nil {
		return nil, errors.New("database error")
	}

	// Validate token version (invalidated on password/role change)
	if tvClaim, ok := claims["tv"].(float64); ok {
		if int64(tvClaim) != user.TokenVersion {
			return nil, &authError{"token revoked"}
		}
	}

	return &userInfo{ID: userID, Role: user.Role}, nil
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
