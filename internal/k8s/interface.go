package k8s

import (
	"context"
	"fmt"
	"log/slog"

	secretsv1alpha1 "github.com/mogilyoy/k8s-secret-manager/api/v1alpha1"
	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type SecretClaimsInterface interface {
	CreateSecretClaim(ctx context.Context, name, namespace, claimType string, data *map[string]string, generationConfig *api.GenerationConfig) error
	GetSecretClaim(ctx context.Context, name, namespace string) (*secretsv1alpha1.SecretClaim, error)
	GetActualSecretData(ctx context.Context, name, namespace string) (map[string]string, error)
	ListSecretClaim(ctx context.Context, namespace string) (*secretsv1alpha1.SecretClaimList, error)
	UpdateSecretClaim(ctx context.Context, name, namespace, claimType string, regenerate bool, data *map[string]string, generationConfig *api.GenerationConfig) (*secretsv1alpha1.SecretClaim, error)
	DeleteSecretClaim(ctx context.Context, name, namespace string) error
}

type K8sDynamicClient struct {
	Client client.Client
	Logger *slog.Logger
	Tracer trace.Tracer
}

func NewK8sSecretManager(logger *slog.Logger, tp *sdktrace.TracerProvider) (*K8sDynamicClient, error) {

	tracer := tp.Tracer("k8s-manager")

	scheme := runtime.NewScheme()

	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add native K8s scheme: %w", err)
	}

	if err := secretsv1alpha1.AddToScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to add scheme: %w", err)
	}

	cl, err := client.New(config.GetConfigOrDie(), client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return &K8sDynamicClient{Client: cl, Logger: logger, Tracer: tracer}, nil
}
