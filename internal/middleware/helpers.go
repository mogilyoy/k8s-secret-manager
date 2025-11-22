package middleware

import (
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mogilyoy/k8s-secret-manager/internal/auth"
	"github.com/mogilyoy/k8s-secret-manager/internal/cfg"
)

func StrPnc(v string) *string {
	return &v
}

func IntPnc(v int) *int {
	return &v
}

func GetClaimsFromToken(tokenString string) (*auth.Claims, error) {
	claims := &auth.Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return cfg.AppConfig.JWT.Secret, nil
		},
	)

	if err != nil || !token.Valid {
		return nil, auth.ErrUnauthorizedToken
	}
	return claims, nil
}
