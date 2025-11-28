package handlers

import (
	"time"

	secretsv1alpha1 "github.com/mogilyoy/k8s-secret-manager/api/v1alpha1"
	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	corev1 "k8s.io/api/core/v1"
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

func mapClaimToSecretResponse(claim *secretsv1alpha1.SecretClaim, secret *corev1.Secret) api.SecretResponse {
	externalStatus := "Pending"
	var errorMessage *string = nil
	var secretName *string = nil
	var lastSyncTime *time.Time = nil

	if claim.Status.ErrorMessage != "" {
		externalStatus = "Error"
		errorMessage = &claim.Status.ErrorMessage
	} else if claim.Status.Synced {
		externalStatus = "Ready"
	}

	if claim.Status.CreatedSecretName != "" {
		secretName = &claim.Status.CreatedSecretName
	}
	if claim.Status.LastUpdate != nil && !claim.Status.LastUpdate.IsZero() {
		lastSyncTime = &claim.Status.LastUpdate.Time
	}

	var generationConfig *api.GenerationConfig
	if claim.Spec.Generation != nil {
		enc := api.GenerationConfigEncoding(claim.Spec.Generation.Encoding)
		dKeys := claim.Spec.Generation.DataKeys
		generationConfig = &api.GenerationConfig{
			Length:   int32(claim.Spec.Generation.Length),
			Encoding: &enc,
			DataKeys: &dKeys,
		}
	}

	secretData := make(map[string]string)
	if secret != nil {
		for k, v := range secret.Data {
			secretData[k] = string(v)
		}
	}

	return api.SecretResponse{
		Name:      claim.Name,
		Namespace: &claim.Namespace,
		Type:      claim.Spec.Type,

		Uid:               StrPnc(string(claim.UID)),
		CreationTimestamp: &claim.CreationTimestamp.Time,
		ResourceVersion:   &claim.ResourceVersion,
		Labels:            &claim.Labels,
		Annotations:       &claim.Annotations,
		// -----------------------------------

		Data:             &secretData,
		GenerationConfig: generationConfig,

		Status: api.SecretStatus{
			CurrentStatus: api.SecretStatusCurrentStatus(externalStatus),
			Synced:        claim.Status.Synced,
			SecretName:    secretName,
			LastSyncTime:  lastSyncTime,
			ErrorMessage:  errorMessage,
		},
	}
}

func mapSecretListToResponseList(claims []secretsv1alpha1.SecretClaim) []api.SecretSummary {
	result := make([]api.SecretSummary, 0, len(claims))
	for _, claim := range claims {
		externalStatus := "Pending"
		if claim.Status.ErrorMessage != "" {
			externalStatus = "Error"
		} else if claim.Status.Synced {
			externalStatus = "Ready"
		}

		result = append(result, api.SecretSummary{
			Name:              claim.Name,
			Namespace:         &claim.Namespace,
			Type:              claim.Spec.Type,
			CreationTimestamp: &claim.CreationTimestamp.Time,

			Status: api.SimpleSecretStatus{
				CurrentStatus: api.SimpleSecretStatusCurrentStatus(externalStatus),
				Synced:        &claim.Status.Synced,
			},
		})
	}
	return result
}
