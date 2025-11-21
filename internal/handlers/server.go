package handlers

import (
	"github.com/mogilyoy/k8s-secret-manager/internal/auth"
	"github.com/mogilyoy/k8s-secret-manager/internal/k8s"
)

type SecretHandler struct {
	K8sManager  k8s.SecretClaimsInterface
	AuthService auth.AuthService
}

func NewSecretHandler(k8sMgr k8s.SecretClaimsInterface, authSvc auth.AuthService) *SecretHandler {
	return &SecretHandler{
		K8sManager:  k8sMgr,
		AuthService: authSvc,
	}
}
