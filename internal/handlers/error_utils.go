package handlers

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

type ErrorResult struct {
	StatusCode   int
	ErrorCode    string
	ErrorMessage string
}

func HandleK8sError(err error) ErrorResult {
	var statusError *k8serrors.StatusError

	if errors.As(err, &statusError) {

		status := int(statusError.ErrStatus.Code)
		k8sMessage := statusError.ErrStatus.Message

		switch status {
		case http.StatusForbidden: // 403
			return ErrorResult{
				StatusCode:   status,
				ErrorCode:    "Forbidden",
				ErrorMessage: fmt.Sprintf("Access denied by Kubernetes RBAC or policy: %s", k8sMessage),
			}

		case http.StatusConflict: // 409
			return ErrorResult{
				StatusCode:   status,
				ErrorCode:    "Conflict",
				ErrorMessage: fmt.Sprintf("Secret already exists: %s", k8sMessage),
			}

		case http.StatusNotFound: // 404
			return ErrorResult{
				StatusCode:   status,
				ErrorCode:    "NotFound",
				ErrorMessage: fmt.Sprintf("Resource not found: %s", k8sMessage),
			}

		default:
			return ErrorResult{
				StatusCode:   http.StatusInternalServerError,
				ErrorCode:    "KubernetesAPIError",
				ErrorMessage: fmt.Sprintf("Unexpected K8s API error [%d]: %s", status, k8sMessage),
			}
		}
	}

	return ErrorResult{
		StatusCode:   http.StatusInternalServerError,
		ErrorCode:    "InternalError",
		ErrorMessage: fmt.Sprintf("Internal error communicating with K8s: %v", err),
	}
}

func getCommonBody(res ErrorResult) api.InternalJSONResponse {
	return api.InternalJSONResponse{
		ErrorMessage: StrPnc(res.ErrorMessage),
		ErrorCode:    StrPnc(res.ErrorCode),
		StatusCode:   IntPnc(res.StatusCode),
	}
}

func BuildCreateSecretErrorResponse(res ErrorResult) api.CreateSecretResponseObject {
	commonBody := getCommonBody(res)
	switch res.StatusCode {
	case 400:
		return api.CreateSecret400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse(commonBody)}
	case 403:
		return api.CreateSecret403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse(commonBody)}
	case 409:
		return api.CreateSecret409JSONResponse{ConflictJSONResponse: api.ConflictJSONResponse(commonBody)}
	case 404:
		return api.CreateSecret404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse(commonBody)}
	default:
		return api.CreateSecret500JSONResponse{InternalJSONResponse: commonBody}
	}
}
func BuildListSecretErrorResponse(res ErrorResult) api.ListSecretsResponseObject {
	commonBody := getCommonBody(res)
	switch res.StatusCode {
	case 400:
		return api.ListSecrets400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse(commonBody)}
	case 403:
		return api.ListSecrets403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse(commonBody)}
	case 404:
		return api.ListSecrets404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse(commonBody)}
	default:
		return api.ListSecrets500JSONResponse{InternalJSONResponse: commonBody}
	}
}

func BuildGetSecretErrorResponse(res ErrorResult) api.GetSecretResponseObject {
	commonBody := getCommonBody(res)
	switch res.StatusCode {
	case 400:
		return api.GetSecret400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse(commonBody)}
	case 403:
		return api.GetSecret403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse(commonBody)}
	case 404:
		return api.GetSecret404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse(commonBody)}
	default:
		return api.GetSecret500JSONResponse{InternalJSONResponse: commonBody}
	}
}

func BuildUpdateSecretErrorResponse(res ErrorResult) api.UpdateSecretResponseObject {
	commonBody := getCommonBody(res)
	switch res.StatusCode {
	case 400:
		return api.UpdateSecret400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse(commonBody)}
	case 403:
		return api.UpdateSecret403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse(commonBody)}
	case 404:
		return api.UpdateSecret404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse(commonBody)}
	default:
		return api.UpdateSecret500JSONResponse{InternalJSONResponse: commonBody}
	}
}

func BuildDeleteSecretErrorResponse(res ErrorResult) api.DeleteSecretResponseObject {
	commonBody := getCommonBody(res)
	switch res.StatusCode {
	case 400:
		return api.DeleteSecret400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse(commonBody)}
	case 403:
		return api.DeleteSecret403JSONResponse{ForbiddenJSONResponse: api.ForbiddenJSONResponse(commonBody)}
	case 404:
		return api.DeleteSecret404JSONResponse{NotFoundJSONResponse: api.NotFoundJSONResponse(commonBody)}
	default:
		return api.DeleteSecret500JSONResponse{InternalJSONResponse: commonBody}
	}
}
