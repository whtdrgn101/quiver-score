package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/quiverscore/backend-go/internal/auth"
)

type contextKey string

const UserIDKey contextKey = "user_id"

// RequireAuth is middleware that validates the Bearer token and injects
// the user ID into the request context. Returns 401 if invalid.
func RequireAuth(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := extractUserID(r, secretKey)
			if !ok {
				http.Error(w, `{"detail":"Not authenticated"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// OptionalAuth is middleware that extracts the user ID if a valid token
// is present, but does not reject the request if missing/invalid.
func OptionalAuth(secretKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if userID, ok := extractUserID(r, secretKey); ok {
				ctx := context.WithValue(r.Context(), UserIDKey, userID)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID retrieves the authenticated user's ID from the context.
// Returns empty string if not authenticated.
func GetUserID(ctx context.Context) string {
	if v, ok := ctx.Value(UserIDKey).(string); ok {
		return v
	}
	return ""
}

func extractUserID(r *http.Request, secretKey string) (string, bool) {
	header := r.Header.Get("Authorization")
	if header == "" {
		return "", false
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", false
	}

	claims, err := auth.DecodeToken(parts[1], secretKey)
	if err != nil {
		return "", false
	}

	if claims.Type != string(auth.TokenTypeAccess) {
		return "", false
	}

	if claims.Subject == "" {
		return "", false
	}

	return claims.Subject, true
}
