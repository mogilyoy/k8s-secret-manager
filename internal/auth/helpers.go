package auth

import (
	"fmt"

	"github.com/mogilyoy/k8s-secret-manager/internal/cfg"
)

func GetUserPermissions(userID string) (role string, namespaces []string, err error) {
	roleMappings := map[string][]cfg.UserCfg{
		"admin":     cfg.AppConfig.RoleConfig.Admin,
		"developer": cfg.AppConfig.RoleConfig.Developer,
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

func IsNamespaceAllowed(requestedNamespace string, allowedNamespaces []string) bool {
	for _, allowed := range allowedNamespaces {
		if allowed == "*" {
			return true
		}
	}
	if len(allowedNamespaces) > 0 {
		for _, allowed := range allowedNamespaces {
			if requestedNamespace == allowed {
				return true
			}
		}
	}
	return false
}
