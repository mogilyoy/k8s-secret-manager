package handlers

import (
	"log/slog"

	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/mogilyoy/k8s-secret-manager/internal/cfg"
	"github.com/mogilyoy/k8s-secret-manager/internal/k8s"
	"go.opentelemetry.io/otel/trace"
)

var _ api.StrictServerInterface = &SecretHandler{}

type SecretHandler struct {
	K8sManager k8s.SecretClaimsInterface
	Logger     *slog.Logger
	Tracer     trace.Tracer
	cfg        cfg.Config
}

func NewSecretHandler(k8sMgr k8s.SecretClaimsInterface, config cfg.Config, logger *slog.Logger, tracer trace.Tracer) *SecretHandler {
	return &SecretHandler{
		K8sManager: k8sMgr,
		cfg:        config,
		Logger:     logger,
		Tracer:     tracer,
	}
}
