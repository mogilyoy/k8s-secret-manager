package k8s

import (
	"k8s.io/client-go/kubernetes"
)

type SecretManager interface {
	CreateSecret(name, namespace, secretType string, data map[string]string) error
	GetSecret(name, namespace string) error
	ListSecrets(namespace string) error
	UpdateSecret(name, namespace, secretType string, data map[string]string) error
	DeleteSecret(name, namespace string)
}

type K8sSecretManager struct {
	Clientset *kubernetes.Clientset // Главный клиент для взаимодействия с K8s
}
