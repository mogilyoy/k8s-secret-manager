package handlers

import (
	secretsv1alpha1 "github.com/mogilyoy/k8s-secret-manager/api/v1alpha1"
	"github.com/mogilyoy/k8s-secret-manager/internal/api"
)

func StrPnc(v string) *string {
	return &v
}

func IntPnc(v int) *int {
	return &v
}

func BoolPnc(v bool) *bool {
	return &v
}

func MapStrStrPnc(v map[string]string) *map[string]string {
	return &v
}

func mapClaimToSecretResponse(claim *secretsv1alpha1.SecretClaim) api.SecretResponse {
	externalStatus := "Pending"
	var errorMessage *string = nil

	if claim.Status.ErrorMessage != "" {
		externalStatus = "Error"
		errorMessage = &claim.Status.ErrorMessage
	} else if claim.Status.Synced {
		externalStatus = "Ready"
	}

	var emptyData *map[string]string = nil

	return api.SecretResponse{
		Name:         claim.Name,
		Namespace:    &claim.Namespace,
		Type:         string(claim.Spec.Type),
		Status:       api.SecretResponseStatus(externalStatus),
		ErrorMessage: errorMessage,
		Data:         emptyData,
	}
}
