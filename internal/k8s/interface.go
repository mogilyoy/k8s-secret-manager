package k8s

import (
	"context"
	"fmt"

	secretsv1alpha1 "github.com/mogilyoy/k8s-secret-manager/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type SecretClaimsInterface interface {
	CreateSecretClaim(ctx context.Context, name, namespace, claimType string, data map[string]string) error
	GetSecretClaim(ctx context.Context, name, namespace string) (*secretsv1alpha1.SecretClaim, error)
	GetActualSecretData(ctx context.Context, name, namespace string) (map[string]string, error)
	ListSecretClaim(ctx context.Context, namespace string) (*secretsv1alpha1.SecretClaimList, error)
	UpdateSecretClaim(ctx context.Context, name, namespace, claimType string, regenerate bool, data map[string]string) (*secretsv1alpha1.SecretClaim, error)
	DeleteSecretClaim(ctx context.Context, name, namespace string) error
}

type K8sDynamicClient struct {
	Client client.Client
}

func NewK8sSecretManager() (*K8sDynamicClient, error) {

	scheme := runtime.NewScheme()

	if err := secretsv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add scheme: %w", err)
	}

	cl, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &K8sDynamicClient{Client: cl}, nil
}
