// internal/auth/service.go (заглушка)
package auth

type AuthService interface {
	Authenticate(token string) (bool, error)
}

func NewAuthService() AuthService {
	return &authService{}
}

type authService struct{}

func (h *authService) Authenticate(token string) (bool, error) {
	return true, nil
}
