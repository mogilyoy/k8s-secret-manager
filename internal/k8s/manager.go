package k8s

import (
	"context"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

type SecretManager interface {
	CreateSecret(ctx context.Context, secret *v1.Secret) (*v1.Secret, error)
	GetSecret(ctx context.Context, request api.GetSecretRequestObject) (*v1.Secret, error)
	ListSecrets(ctx context.Context, namespace string) (*v1.SecretList, error)
	UpdateSecret(name, namespace, secretType string, data map[string]string) error
	DeleteSecret(name, namespace string)
}

type K8sSecretManager struct {
	Clientset *kubernetes.Clientset
}
