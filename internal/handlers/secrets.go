package handlers

import (
	"context"
	"log"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
)

func (h *SecretHandler) CreateSecret(ctx context.Context, request api.CreateSecretRequest) (api.CreateSecret201JSONResponse, error) {

	secretName := request.Name
	log.Printf("Attempting to create secret: %s", secretName)

	ok := true

	return api.CreateSecret201JSONResponse{
		OkResponseJSONResponse: api.OkResponseJSONResponse{
			Ok: &ok,
		},
	}, nil
}
