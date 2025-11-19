package handlers

import (
	"github.com/mogilyoy/k8s-secret-manager/internal/k8s"
)

type SecretHandler struct {
	k8sManager k8s.SecretManager
}
