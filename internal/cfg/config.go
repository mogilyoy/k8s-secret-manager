package cfg

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type UserCfg struct {
	ID                string   `yaml:"id"`
	AllowedNamespaces []string `yaml:"allowed_namespaces"`
}

type UserRolesConfig struct {
	Admin     []UserCfg `yaml:"admin"`
	Developer []UserCfg `yaml:"developer"`
	Readonly  []UserCfg `yaml:"readonly"`
}

type Auth struct {
	TelegramBotToken string
	JWTSecret        string
}

type Config struct {
	RoleConfig UserRolesConfig
	AuthConfig Auth
}

var AppConfig Config

func LoadConfig(configPath string) (*Config, error) {
	fileContent, err := os.ReadFile(configPath)

	if err != nil {
		return nil, fmt.Errorf("error reading configPath: %w", err)
	}
	var userRolesCfg UserRolesConfig
	if err := yaml.Unmarshal(fileContent, &userRolesCfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML into struct: %w", err)
	}

	telegramToken := os.Getenv("TELEGRAM_TOKEN")
	if telegramToken == "" {
		return nil, fmt.Errorf("required environment variable TELEGRAM_TOKEN is not set")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("required environment variable JWT_SECRET is not set")
	}

	authCfg := Auth{
		TelegramBotToken: telegramToken,
		JWTSecret:        jwtSecret,
	}

	AppConfig = Config{
		RoleConfig: userRolesCfg,
		AuthConfig: authCfg,
	}

	return &AppConfig, nil
}
