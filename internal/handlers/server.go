package handlers

import (
	"github.com/mogilyoy/k8s-secret-manager/internal/cfg"
	"github.com/mogilyoy/k8s-secret-manager/internal/k8s"
)

type SecretHandler struct {
	K8sManager k8s.SecretClaimsInterface
	cfg        cfg.Config
}

func NewSecretHandler(k8sMgr k8s.SecretClaimsInterface, config cfg.Config) *SecretHandler {
	return &SecretHandler{
		K8sManager: k8sMgr,
		cfg:        config,
	}
}
