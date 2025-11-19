package k8s

import (
	"encoding/base64"
	"fmt"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func decodeSecretData(data map[string]string) (map[string][]byte, error) {
	// Turn base64 data into map[string][]byte
	k8sData := make(map[string][]byte, len(data))

	for key, base64Value := range data {

		decodedBytes, err := base64.StdEncoding.DecodeString(base64Value)
		if err != nil {

			return nil, fmt.Errorf("data key '%s' has invalid Base64 encoding: %w", key, err)
		}
		k8sData[key] = decodedBytes
	}

	return k8sData, nil
}

func ToK8sSecret(req api.CreateSecretRequest) (*v1.Secret, error) {

	k8sData, err := decodeSecretData(req.Data)
	if err != nil {
		return nil, err
	}

	secretType := v1.SecretTypeOpaque // default type
	if req.Type != nil {
		secretType = v1.SecretType(*req.Type)
	}

	labels := make(map[string]string)
	if req.Labels != nil {
		labels = *req.Labels
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      req.Name,
			Namespace: req.Namespace,
			Labels:    labels,
		},
		Data: k8sData,
		Type: secretType,
	}

	return secret, nil
}
