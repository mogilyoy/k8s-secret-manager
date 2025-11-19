package auth

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/mogilyoy/k8s-secret-manager/internal/cfg"
)

func GenerateJWT(userID, role string, namespaces []string) (string, error) {

	expirationTime := time.Now().Add(time.Duration(3) * time.Hour)

	claims := &Claims{
		TelegramUserID:    userID,
		Role:              role,
		AllowedNamespaces: namespaces,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(cfg.AppConfig.AuthConfig.JWTSecret)
	if err != nil {
		return "", fmt.Errorf("could not sign token: %w", err)
	}
	return tokenString, nil
}

func GetUserPermissions(userID string) (role string, namespaces []string, err error) {
	roleMappings := map[string][]cfg.UserCfg{
		"admin":     cfg.AppConfig.RoleConfig.Admin,
		"developer": cfg.AppConfig.RoleConfig.Developer,
		"readonly":  cfg.AppConfig.RoleConfig.Readonly,
	}

	for role, users := range roleMappings {
		for _, user := range users {
			if user.ID == userID {
				return role, user.AllowedNamespaces, nil
			}

		}
	}
	return "", nil, fmt.Errorf("user not found in config")
}

func GetClaimsFromToken(tokenString string) (*Claims, error) {
	claims := &Claims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return cfg.AppConfig.AuthConfig.JWTSecret, nil
		},
	)

	if err != nil {
		return nil, fmt.Errorf("token validation failed: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("token is not valid")
	}

	return claims, nil
}
