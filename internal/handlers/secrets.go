package handlers

import (
	"context"
	"errors"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/mogilyoy/k8s-secret-manager/internal/auth"
	"github.com/mogilyoy/k8s-secret-manager/internal/k8s"
)

func (h *SecretHandler) CreateSecret(ctx context.Context, request api.CreateSecretRequest) (api.CreateSecretResponseObject, error) {
	claims, err := auth.GetClaimsFromContext(ctx)
	if err != nil {
		return BuildCreateSecretErrorResponse(ErrorResult{
			ErrorMessage: "Cannot parse request context",
			ErrorCode:    "InternalServerError",
			StatusCode:   500,
		}), err
	}

	if !auth.IsNamespaceAllowed(request.Namespace, claims.AllowedNamespaces) || claims.Role == "guest" {
		return BuildCreateSecretErrorResponse(ErrorResult{
			ErrorMessage: "Access denied: insufficient role or namespace permissions",
			ErrorCode:    "Forbidden",
			StatusCode:   403,
		}), errors.New("forbidden")
	}

	secret, err := k8s.ToK8sSecret(request)
	if err != nil {
		return BuildCreateSecretErrorResponse(ErrorResult{
			ErrorMessage: "Error parsing request body",
			ErrorCode:    "BadRequest",
			StatusCode:   400,
		}), err
	}

	_, err = h.k8sManager.CreateSecret(ctx, secret)
	if err != nil {
		errorResult := HandleK8sError(err)
		return BuildCreateSecretErrorResponse(errorResult), err
	}

	return api.CreateSecret201JSONResponse{
		OkResponseJSONResponse: api.OkResponseJSONResponse{
			Ok: BoolPnc(true),
		},
	}, nil
}

func (h *SecretHandler) ListSecrets(ctx context.Context, request api.ListSecretsParams) (api.ListSecretsResponseObject, error) {
	claims, err := auth.GetClaimsFromContext(ctx)
	if err != nil {
		return BuildListSecretErrorResponse(ErrorResult{
			ErrorMessage: "Cannot parse request context",
			ErrorCode:    "InternalServerError",
			StatusCode:   500,
		}), err
	}
	if !auth.IsNamespaceAllowed(request.Namespace, claims.AllowedNamespaces) {
		return BuildListSecretErrorResponse(ErrorResult{
			ErrorMessage: "Access denied: insufficient role or namespace permissions",
			ErrorCode:    "Forbidden",
			StatusCode:   403,
		}), err
	}

	listSecrets, err := h.k8sManager.ListSecrets(ctx, request.Namespace)
	if err != nil {
		errorResult := HandleK8sError(err)
		return BuildListSecretErrorResponse(errorResult), err
	}

	response := ToSecretListResponse(listSecrets)

	return api.ListSecrets200JSONResponse(*response), err
}

func (h *SecretHandler) GetSecret(ctx context.Context, request api.GetSecretRequestObject) (api.GetSecretResponseObject, error) {
	claims, err := auth.GetClaimsFromContext(ctx)
	if err != nil {
		return BuildGetSecretErrorResponse(ErrorResult{
			ErrorMessage: "Cannot parse request context",
			ErrorCode:    "InternalServerError",
			StatusCode:   500,
		}), err
	}

	if !auth.IsNamespaceAllowed(request.Params.Namespace, claims.AllowedNamespaces) {
		return BuildGetSecretErrorResponse(ErrorResult{
			ErrorMessage: "Access denied: insufficient role or namespace permissions",
			ErrorCode:    "Forbidden",
			StatusCode:   403,
		}), err
	}

	secret, err := h.k8sManager.GetSecret(ctx, request)

	if err != nil {
		errorResult := HandleK8sError(err)
		return BuildGetSecretErrorResponse(errorResult), err
	}

	return api.GetSecret200JSONResponse{
		Data:            MapStrStrPnc(EncodeSecretData(secret.Data)),
		Name:            &secret.Name,
		Labels:          &secret.Labels,
		Namespace:       &secret.Namespace,
		ResourceVersion: &secret.ResourceVersion,
		Type:            StrPnc(string(secret.Type)),
	}, nil

}
