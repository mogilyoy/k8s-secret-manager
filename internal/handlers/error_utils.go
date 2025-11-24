package handlers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/mogilyoy/k8s-secret-manager/internal/observability"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

type ErrorResult struct {
	StatusCode   int
	ErrorCode    string
	ErrorMessage string
}

func HandleK8sError(ctx context.Context, err error) ErrorResult {
	span := trace.SpanFromContext(ctx)
	logger := observability.LoggerFromContext(ctx)

	span.RecordError(err)

	var statusError *k8serrors.StatusError

	if errors.As(err, &statusError) {

		status := int(statusError.ErrStatus.Code)
		k8sMessage := statusError.ErrStatus.Message

		switch status {
		case http.StatusForbidden: // 403
			span.SetStatus(codes.Error, "K8s API: Forbidden")
			logger.Warn("K8s API Access Denied (403)", slog.Any("k8s_status", status), slog.Any("k8s_message", k8sMessage), slog.Any("error", err.Error()))
			return ErrorResult{
				StatusCode:   status,
				ErrorCode:    "Forbidden",
				ErrorMessage: fmt.Sprintf("Access denied by Kubernetes RBAC or policy: %s", k8sMessage),
			}

		case http.StatusConflict: // 409
			span.SetStatus(codes.Error, "K8s API: Conflict")
			logger.Warn("K8s API Resource Conflict (409)", slog.Any("k8s_status", status), slog.Any("k8s_message", k8sMessage), slog.Any("error", err.Error()))
			return ErrorResult{
				StatusCode:   status,
				ErrorCode:    "Conflict",
				ErrorMessage: fmt.Sprintf("Secret already exists: %s", k8sMessage),
			}

		case http.StatusNotFound: // 404
			span.SetStatus(codes.Error, "K8s API: Not Found")
			logger.Warn("K8s API Resource Not Found (404)", slog.Any("k8s_status", status), slog.Any("k8s_message", k8sMessage), slog.Any("error", err.Error()))
			return ErrorResult{
				StatusCode:   status,
				ErrorCode:    "NotFound",
				ErrorMessage: fmt.Sprintf("Resource not found: %s", k8sMessage),
			}

		default:
			span.SetStatus(codes.Error, "K8s API: Unexpected Error")
			logger.Error("K8s API Unexpected Error", slog.Any("k8s_status", status), slog.Any("k8s_message", k8sMessage), slog.Any("error", err.Error()))
			return ErrorResult{
				StatusCode:   http.StatusInternalServerError,
				ErrorCode:    "KubernetesAPIError",
				ErrorMessage: fmt.Sprintf("Unexpected K8s API error [%d]: %s", status, k8sMessage),
			}
		}
	}
	span.SetStatus(codes.Error, "Internal K8s communication error")
	logger.Error("Internal error communicating with K8s", slog.Any("error", err))

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
func BuildListSecretsErrorResponse(res ErrorResult) api.ListSecretsResponseObject {
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

func BuildAuthErrorResponse(res ErrorResult) api.AuthUserResponseObject {
	commonBody := getCommonBody(res)
	switch res.StatusCode {
	case 400:
		return api.AuthUser400JSONResponse{BadRequestJSONResponse: api.BadRequestJSONResponse(commonBody)}
	case 401:
		return api.AuthUser401JSONResponse{UnauthorizedJSONResponse: api.UnauthorizedJSONResponse(commonBody)}
	default:
		return api.AuthUser500JSONResponse{InternalJSONResponse: commonBody}
	}
}
