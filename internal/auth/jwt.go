package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mogilyoy/k8s-secret-manager/internal/cfg"
)

var ErrUnauthorizedToken = errors.New("authentication failed: token is invalid or expired")

func GenerateJWT(user *cfg.User, expiresIn int64, JWTSecret string) (string, error) {
	if JWTSecret == "" {
		return "", fmt.Errorf("jwt secret is empty")
	}
	now := time.Now()
	expirationTime := time.Now().Add(time.Duration(expiresIn) * time.Second)

	claims := Claims{
		Username:          user.Username,
		Role:              user.Role,
		AllowedNamespaces: user.AllowedNamespaces,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   user.ID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}
