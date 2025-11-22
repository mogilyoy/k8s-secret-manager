package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	jwt.RegisteredClaims
	Username          string   `json:"username"`
	Role              string   `json:"role"`
	AllowedNamespaces []string `json:"allowed_namespaces"`
}
