package handlers

import (
	"context"
	"fmt"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/mogilyoy/k8s-secret-manager/internal/auth"
)

func (h *SecretHandler) AuthUser(ctx context.Context, request api.AuthUserRequestObject) (api.AuthUserResponseObject, error) {
	user := h.cfg.FindUser(request.Body.Username)

	if user == nil || !auth.CheckPasswordHash(request.Body.Password, user.PasswordHash) {
		return BuildAuthErrorResponse(ErrorResult{
			ErrorCode:    "Unauthorized",
			StatusCode:   401,
			ErrorMessage: "Wrong username/password",
		}), nil
	}

	expiresIn := int64(24 * 60 * 60)

	token, err := auth.GenerateJWT(user, expiresIn, h.cfg.JWT.Secret)
	if err != nil {
		return BuildAuthErrorResponse(ErrorResult{
			ErrorCode:    "Internal Server Error",
			StatusCode:   500,
			ErrorMessage: "Something went wrong",
		}), fmt.Errorf("could not generate jwt: %w", err)
	}

	return api.AuthUser200JSONResponse{
		Token:     token,
		ExpiresIn: &expiresIn,
	}, nil
}
