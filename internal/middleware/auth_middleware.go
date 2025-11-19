package middleware

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/mogilyoy/k8s-secret-manager/internal/auth"
)

// AuthMiddleware проверяет JWT, декодирует Claims и помещает их в контекст.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")

		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Unauthorized: Bearer token required", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := auth.GetClaimsFromToken(tokenString)
		if err != nil {
			log.Printf("JWT validation error: %v", err)
			http.Error(w, "Unauthorized: Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), auth.ClaimsContextKey, claims)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
