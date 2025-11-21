package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/mogilyoy/k8s-secret-manager/internal/auth"
)

func JWTAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		claims, err := auth.GetClaimsFromToken(token)

		if err != nil {
			if errors.Is(err, auth.ErrUnauthorizedToken) {
				sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized", "Invalid or expired token.")
				return
			}

			sendErrorResponse(w, http.StatusInternalServerError, "InternalError", "Server error during authentication.")
			return
		}

		ctx := context.WithValue(r.Context(), auth.ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
