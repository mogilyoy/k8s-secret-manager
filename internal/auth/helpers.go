package auth

import "golang.org/x/crypto/bcrypt"

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

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
