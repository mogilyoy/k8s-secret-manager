package cfg

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Users []User `yaml:"users"`

	JWT JWTConfig `yaml:"jwt"`
}

type User struct {
	ID                string   `yaml:"id"`
	Username          string   `yaml:"username"`
	PasswordHash      string   `yaml:"pwd"`
	Role              string   `yaml:"role"`
	AllowedNamespaces []string `yaml:"allowed_namespaces"`
}

type JWTConfig struct {
	Secret string `yaml:"secret"`
}

var AppConfig Config

func LoadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	if err := yaml.Unmarshal(data, &AppConfig); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}
	if envSecret := os.Getenv("JWT_SECRET"); envSecret != "" {
		AppConfig.JWT.Secret = envSecret
	}
	return &AppConfig, nil
}

func (c *Config) FindUser(username string) *User {
	for _, u := range c.Users {
		if u.Username == username {
			return &u
		}
	}
	return nil
}
