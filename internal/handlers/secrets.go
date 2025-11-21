package handlers

import (
	"context"
	"errors"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/mogilyoy/k8s-secret-manager/internal/auth"
)

func (h *SecretHandler) CreateSecret(ctx context.Context, request api.CreateSecretRequestObject) (api.CreateSecretResponseObject, error) {
	claims, err := auth.GetClaimsFromContext(ctx)
	if err != nil {
		return BuildCreateSecretErrorResponse(ErrorResult{
			ErrorMessage: "Cannot parse request context",
			ErrorCode:    "InternalServerError",
			StatusCode:   500,
		}), err
	}

	if !auth.IsNamespaceAllowed(*request.Body.Namespace, claims.AllowedNamespaces) || claims.Role == "developer" {
		return BuildCreateSecretErrorResponse(ErrorResult{
			ErrorMessage: "Access denied: insufficient role or namespace permissions",
			ErrorCode:    "Forbidden",
			StatusCode:   403,
		}), errors.New("forbidden")
	}

	err = h.K8sManager.CreateSecretClaim(ctx, request.Body.Name, *request.Body.Namespace, string(request.Body.Type), *request.Body.Data)
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

func (h *SecretHandler) ListSecrets(ctx context.Context, request api.ListSecretsRequestObject) (api.ListSecretsResponseObject, error) {
	claims, err := auth.GetClaimsFromContext(ctx)
	if err != nil {
		return BuildListSecretErrorResponse(ErrorResult{
			ErrorMessage: "Cannot parse request context",
			ErrorCode:    "InternalServerError",
			StatusCode:   500,
		}), err
	}
	if !auth.IsNamespaceAllowed(request.Params.Namespace, claims.AllowedNamespaces) {
		return BuildListSecretErrorResponse(ErrorResult{
			ErrorMessage: "Access denied: insufficient role or namespace permissions",
			ErrorCode:    "Forbidden",
			StatusCode:   403,
		}), err
	}

	secretClaimList, err := h.K8sManager.ListSecretClaim(ctx, request.Params.Namespace)
	if err != nil {
		errorResult := HandleK8sError(err)
		return BuildListSecretErrorResponse(errorResult), err
	}

	secretResponses := make([]api.SecretResponse, 0, len(secretClaimList.Items))

	for _, claim := range secretClaimList.Items {
		response := mapClaimToSecretResponse(&claim)
		secretResponses = append(secretResponses, response)
	}

	return api.ListSecrets200JSONResponse{
		Items: secretResponses,
	}, err
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

	secret, err := h.K8sManager.GetSecretClaim(ctx, request.Name, request.Params.Namespace)

	if err != nil {
		errorResult := HandleK8sError(err)
		return BuildGetSecretErrorResponse(errorResult), err
	}

	return api.GetSecret200JSONResponse{
		Data:      MapStrStrPnc(secret.Spec.Data),
		Name:      secret.Name,
		Namespace: &secret.Namespace,
		Type:      string(secret.Spec.Type),
	}, nil

}

func (h *SecretHandler) UpdateSecret(ctx context.Context, request api.UpdateSecretRequestObject) (api.UpdateSecretResponseObject, error) {
	claims, err := auth.GetClaimsFromContext(ctx)
	if err != nil {
		return BuildUpdateSecretErrorResponse(ErrorResult{
			ErrorMessage: "Cannot parse request context",
			ErrorCode:    "InternalServerError",
			StatusCode:   500,
		}), err
	}

	if !auth.IsNamespaceAllowed(request.Params.Namespace, claims.AllowedNamespaces) || claims.Role == "developer" {
		return BuildUpdateSecretErrorResponse(ErrorResult{
			ErrorMessage: "Access denied: insufficient role or namespace permissions",
			ErrorCode:    "Forbidden",
			StatusCode:   403,
		}), errors.New("forbidden")
	}

	updatedSecret, err := h.K8sManager.UpdateSecretClaim(ctx, request.Name, request.Params.Namespace, string(*request.Body.Type), *request.Body.Regenerate, *request.Body.Data)
	if err != nil {
		errResult := HandleK8sError(err)
		return BuildUpdateSecretErrorResponse(errResult), err
	}

	secretResponse := mapClaimToSecretResponse(updatedSecret)

	return api.UpdateSecret200JSONResponse(secretResponse), nil
}

func (h *SecretHandler) DeleteSecret(ctx context.Context, request api.DeleteSecretRequestObject) (api.DeleteSecretResponseObject, error) {
	claims, err := auth.GetClaimsFromContext(ctx)
	if err != nil {
		return BuildDeleteSecretErrorResponse(ErrorResult{
			ErrorMessage: "Cannot parse request context",
			ErrorCode:    "InternalServerError",
			StatusCode:   500,
		}), err
	}

	if !auth.IsNamespaceAllowed(request.Params.Namespace, claims.AllowedNamespaces) || claims.Role == "developer" {
		return BuildDeleteSecretErrorResponse(ErrorResult{
			ErrorMessage: "Access denied: insufficient role or namespace permissions",
			ErrorCode:    "Forbidden",
			StatusCode:   403,
		}), errors.New("forbidden")
	}

	err = h.K8sManager.DeleteSecretClaim(ctx, request.Name, request.Params.Namespace)
	if err != nil {
		errorResult := HandleK8sError(err)
		return BuildDeleteSecretErrorResponse(errorResult), err
	}

	return api.DeleteSecret200JSONResponse{
		OkResponseJSONResponse: api.OkResponseJSONResponse{
			Ok: BoolPnc(true),
		},
	}, nil

}
