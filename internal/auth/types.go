package auth

import (
	"github.com/golang-jwt/jwt/v5"
)

type TelegramUser struct {
	ID int64 `json:"id"`
}

type Claims struct {
	jwt.RegisteredClaims
	TelegramUserID    string   `json:"tg_user_id"`
	Role              string   `json:"role"`
	AllowedNamespaces []string `json:"allowed_namespaces"`
}
