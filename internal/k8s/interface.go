package k8s

import (
	"context"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	v1 "k8s.io/api/core/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (m *K8sSecretManager) CreateSecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error) {
	return m.Clientset.CoreV1().Secrets(secret.Namespace).Create(ctx, secret, metav1.CreateOptions{})
}

func (m *K8sSecretManager) GetSecret(ctx context.Context, request api.GetSecretRequestObject) (*v1.Secret, error) {
	return m.Clientset.CoreV1().Secrets(request.Params.Namespace).Get(ctx, request.Name, metav1.GetOptions{})

}

func (m *K8sSecretManager) ListSecrets(ctx context.Context, namespace string) (*v1.SecretList, error) {
	return m.Clientset.CoreV1().Secrets(namespace).List(ctx, metav1.ListOptions{})
}
