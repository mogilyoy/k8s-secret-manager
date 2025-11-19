package handlers

import (
	"encoding/base64"

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

func EncodeSecretData(data map[string][]byte) map[string]string {
	outputData := make(map[string]string)

	for key, bytesArray := range data {
		encodedBytes := base64.StdEncoding.EncodeToString(bytesArray)
		outputData[key] = encodedBytes
	}
	return outputData
}

func ToSecretListResponse(k8sList *corev1.SecretList) *api.ListSecretsResponse {

	if k8sList == nil || len(k8sList.Items) == 0 {

		emptyItems := make([]api.SecretSummary, 0)
		return &api.ListSecretsResponse{
			Items: &emptyItems,

			Namespace: StrPnc(k8sList.ResourceVersion),
		}
	}

	summaries := make([]api.SecretSummary, 0, len(k8sList.Items))

	for _, item := range k8sList.Items {

		summary := api.SecretSummary{

			CreationTimestamp: &item.ObjectMeta.CreationTimestamp.Time,

			Name: StrPnc(item.Name),

			Type: StrPnc(string(item.Type)),
		}

		summaries = append(summaries, summary)
	}

	namespace := k8sList.Items[0].Namespace

	return &api.ListSecretsResponse{

		Items:     &summaries,
		Namespace: StrPnc(namespace),
	}
}
