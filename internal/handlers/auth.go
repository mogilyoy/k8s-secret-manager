package handlers

import (
	"context"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
)

func (h *SecretHandler) AuthUser(ctx context.Context, request api.AuthUserRequestObject) (api.AuthUserResponseObject, error) {
	var seconds int64 = 86400
	return api.AuthUser200JSONResponse{
		Token:     "1234",
		ExpiresIn: &seconds,
	}, nil
}
