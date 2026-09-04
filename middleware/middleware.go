// Package middleware contains HTTP middleware used by the API.
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"1-basic-api/jwt"
)

type contextKey string

const (
	usernameKey contextKey = "username"
	roleKey     contextKey = "role"
)

// UsernameFromContext returns the authenticated username, if any.
func UsernameFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(usernameKey).(string)
	return v, ok
}

// RoleFromContext returns the authenticated user's role, if any.
func RoleFromContext(ctx context.Context) (string, bool) {
	v, ok := ctx.Value(roleKey).(string)
	return v, ok
}

// Auth verifies the Bearer token and stores claims in the request context.
func Auth(tokens *jwt.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				unauthorized(w, "Missing authorization header")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
				unauthorized(w, "Invalid authorization format. Use 'Bearer <token>'")
				return
			}

			claims, err := tokens.VerifyToken(parts[1])
			if err != nil {
				unauthorized(w, "Invalid token")
				return
			}

			ctx := context.WithValue(r.Context(), usernameKey, claims.Username)
			ctx = context.WithValue(ctx, roleKey, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func unauthorized(w http.ResponseWriter, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authenticate", "Bearer")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
