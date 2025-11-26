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

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
