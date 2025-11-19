package k8s

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *K8sSecretManager) CreateSecret(secret *v1.Secret) error {

	createOptions := metav1.CreateOptions{}

	_, err := m.Clientset.CoreV1().Secrets(secret.Namespace).Create(context.TODO(), secret, createOptions)

	if err != nil {
		return fmt.Errorf("error creating the secret: %w", err)
	}
	return nil

}
