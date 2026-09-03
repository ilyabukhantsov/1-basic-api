package middleware

import (
	"context"
	"net/http"
	"strings"

	"1-basic-api/jwt"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" {
			http.Error(w, `{"error":"Missing authorization header"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(
				w,
				`{"error":"Invalid authorization format. Use 'Bearer <token>'"}`,
				http.StatusUnauthorized,
			)
			return
		}

		tokenString := parts[1]

		claims, err := jwt.VerifyToken(tokenString)
		if err != nil {
			http.Error(
				w,
				`{"error":"Invalid token"}`,
				http.StatusUnauthorized,
			)
			return
		}

		ctx := r.Context()

		ctx = context.WithValue(ctx, "username", claims.Username)
		ctx = context.WithValue(ctx, "role", claims.Role)

		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
