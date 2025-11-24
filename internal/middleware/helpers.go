package middleware

import (
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mogilyoy/k8s-secret-manager/internal/auth"
)

func StrPnc(v string) *string {
	return &v
}

func IntPnc(v int) *int {
	return &v
}

func GetClaimsFromToken(tokenString, jwtSecret string) (*auth.Claims, error) {
	claims := &auth.Claims{}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(jwtSecret), nil
		},
	)

	if err != nil || !token.Valid {
		return nil, auth.ErrUnauthorizedToken
	}
	return claims, nil
}
