package auth

import (
	"context"
	"errors"
)

type contextKey string

const ClaimsContextKey contextKey = "claims"

func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, ClaimsContextKey, claims)
}

func GetClaimsFromContext(ctx context.Context) (*Claims, error) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	if !ok || claims == nil {
		return nil, errors.New("claims not found in context (middleware failed)")
	}
	return claims, nil
}
