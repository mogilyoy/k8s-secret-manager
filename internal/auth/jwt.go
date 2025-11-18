package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
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

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("could not sign token: %w", err)
	}

	return tokenString, nil
}

func GetUserRoleAndNamespaces(userID string) (role string, namespaces []string, err error) {

}
