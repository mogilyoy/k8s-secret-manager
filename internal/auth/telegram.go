package auth

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	initdata "github.com/telegram-mini-apps/init-data-golang"
)

func VerifyTelegramInitData(initData, botToken string) (userID string, err error) {
	expIn := 24 * time.Hour

	err = initdata.Validate(initData, botToken, expIn)
	if err != nil {
		return "", fmt.Errorf("telegram initData validation failed: %w", err)
	}

	u, err := url.ParseQuery(initData)
	if err != nil {
		return "", fmt.Errorf("could not parse initData query string: %w", err)
	}

	userJSONStr := u.Get("user")
	if userJSONStr == "" {
		return "", fmt.Errorf("initData is valid, but 'user' field is missing")
	}

	var tgUser TelegramUser
	err = json.Unmarshal([]byte(userJSONStr), &tgUser)
	if err != nil {
		return "", fmt.Errorf("could not unmarshal user JSON from initData: %w", err)
	}

	return strconv.FormatInt(tgUser.ID, 10), nil
}
