/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	secretsv1alpha1 "github.com/mogilyoy/k8s-secret-manager/api/v1alpha1"
	"github.com/mogilyoy/k8s-secret-manager/internal/observability"
)

// SecretClaimReconciler reconciles a SecretClaim object
type SecretClaimReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Tracer trace.Tracer
	Log    *slog.Logger
}

// +kubebuilder:rbac:groups=secrets.myapp.io,resources=secretclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=secrets.myapp.io,resources=secretclaims/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=secrets.myapp.io,resources=secretclaims/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SecretClaim object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.4/pkg/reconcile
func (r *SecretClaimReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log
	if logger == nil {
		logger = slog.Default()
	}
	logger = logger.With(
		slog.String("namespace", req.Namespace),
		slog.String("name", req.Name),
	)

	if r.Tracer == nil {
		r.Tracer = otel.Tracer("k8s-secret-manager")
	}

	var span trace.Span
	var reconcileError error = nil

	tempCtx := context.WithValue(ctx, observability.LoggerContextKey, logger)

	var claim secretsv1alpha1.SecretClaim
	if err := r.Get(tempCtx, req.NamespacedName, &claim); err != nil {
		if client.IgnoreNotFound(err) != nil {
			logger.Error("Failed to fetch SecretClaim", slog.Any("error", err))
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	traceParent := claim.GetAnnotations()[observability.K8sTraceparentAnnotationKey]

	if traceParent != "" {
		logger = logger.With(slog.String("api_traceparent", traceParent), slog.String("claim_type", claim.Spec.Type))

		carrier := propagation.MapCarrier{"traceparent": traceParent}
		parentCtx := otel.GetTextMapPropagator().Extract(ctx, carrier)

		ctx, span = r.Tracer.Start(parentCtx, "SecretClaimReconciler.Reconcile",
			trace.WithSpanKind(trace.SpanKindInternal))

		span.SetAttributes(attribute.String("api.request.traceparent", traceParent))
	} else {
		logger.Info("Starting Reconcile without upstream trace context. Starting new root span.")
		ctx, span = r.Tracer.Start(ctx, "SecretClaimReconciler.Reconcile",
			trace.WithSpanKind(trace.SpanKindInternal))
	}

	defer func() {
		if reconcileError != nil {
			span.RecordError(reconcileError)
			span.SetStatus(codes.Error, reconcileError.Error())
		} else {
			span.SetStatus(codes.Ok, "Reconciliation successful")
		}
		span.End()
	}()

	ctx = context.WithValue(ctx, observability.LoggerContextKey, logger)
	logger.Info("Starting reconciliation cycle.")

	targetSecretName := claim.Name
	var secret corev1.Secret

	err := r.Get(ctx, client.ObjectKey{
		Name:      targetSecretName,
		Namespace: claim.Namespace,
	}, &secret)

	if err != nil && errors.IsNotFound(err) {
		// Секрета нет -> нужно создать
		logger.Info("K8s Secret not found, creating new Secret.", slog.String("secret_name", targetSecretName))

		if reconcileError = r.createSecret(ctx, &claim); reconcileError != nil {
			logger.Error("Failed to create Secret", slog.Any("error", reconcileError))
			r.updateStatus(ctx, &claim, false, reconcileError.Error())
			return ctrl.Result{}, reconcileError
		}
		logger.Info("K8s Secret created successfully.")
		r.updateStatus(ctx, &claim, true, "")
		return ctrl.Result{}, nil

	} else if err != nil {
		logger.Error("Failed to get Secret", slog.Any("error", err))
		reconcileError = err
		return ctrl.Result{RequeueAfter: time.Minute}, reconcileError
	}

	if !metav1.IsControlledBy(&secret.ObjectMeta, &claim) {
		logger.Warn("Secret exists but is not controlled by SecretClaim. Skipping.", slog.String("secret_name", targetSecretName))
		return ctrl.Result{}, nil
	}

	needsSecretUpdate := false
	if claim.Spec.Type == "AutoGenerated" {
		if claim.Spec.Generation == nil {
			reconcileError = fmt.Errorf("generationConfig spec is nil for AutoGenerated claim")
			logger.Error("Invalid generation spec", slog.Any("error", reconcileError))
			r.updateStatus(ctx, &claim, false, reconcileError.Error())
			return ctrl.Result{}, reconcileError
		}

		currentTrigger := claim.Spec.Generation.ReconcileTrigger

		if currentTrigger != "" && currentTrigger != claim.Status.LastReconcileTrigger {
			logger.Info("ReconcileTrigger changed. Starting regeneration.", slog.String("old_trigger", claim.Status.LastReconcileTrigger), slog.String("new_trigger", currentTrigger))
			needsSecretUpdate = true
		}
	}

	if claim.Spec.Type == "Opaque" {
		if needsUpdate(claim.Spec.Data, secret.Data) {
			logger.Info("Opaque data changed. Starting secret update.")
			needsSecretUpdate = true
		}
	}

	if needsSecretUpdate {
		if reconcileError = r.updateSecret(ctx, &claim, &secret); reconcileError != nil {
			logger.Error("Failed to update Secret", slog.Any("error", reconcileError))
			r.updateStatus(ctx, &claim, false, reconcileError.Error())
			return ctrl.Result{}, reconcileError
		}

		logger.Info("K8s Secret updated successfully.")
		r.updateStatus(ctx, &claim, true, "")
		return ctrl.Result{}, nil
	}

	if !claim.Status.Synced {
		logger.Info("SecretClaim not synced yet, updating status")
		r.updateStatus(ctx, &claim, true, "")
		return ctrl.Result{}, nil
	}

	logger.Info("Reconciliation complete")
	return ctrl.Result{}, nil
}

func (r *SecretClaimReconciler) createSecret(ctx context.Context, claim *secretsv1alpha1.SecretClaim) error {

	logger := observability.LoggerFromContext(ctx)

	ctx, span := r.Tracer.Start(ctx, "SecretClaimReconciler.createSecret")
	defer span.End()

	secretData := make(map[string][]byte)
	var err error
	switch claim.Spec.Type {
	case "Opaque":
		for k, v := range claim.Spec.Data {
			secretData[k] = []byte(v)
		}
		span.AddEvent("Opaque data copied")

	case "AutoGenerated":

		if claim.Spec.Generation == nil {
			err = fmt.Errorf("generation spec is nil for AutoGenerated claim")
			span.RecordError(err)
			span.SetStatus(codes.Error, "Missing Generation Spec")
			return err
		}

		span.AddEvent("Starting keys generation")
		for _, key := range claim.Spec.Generation.DataKeys {
			if claim.Spec.Generation.Length < 8 {
				err = fmt.Errorf("secrets should be at least 8 symbols")
				span.RecordError(err)
				span.SetStatus(codes.Error, "Secret lenght <8")
				return err
			}

			password, err := generatePassword(claim.Spec.Generation.Length, claim.Spec.Generation.Encoding)
			if err != nil {
				err = fmt.Errorf("failed to generate password for key %s: %w", key, err)
				span.RecordError(err)
				span.SetStatus(codes.Error, "Password Generation Failed")
				return err
			}
			secretData[key] = []byte(password)
		}
		span.AddEvent("Password generation complete", trace.WithAttributes(attribute.Int("data_keys", len(claim.Spec.Generation.DataKeys))))

	default:
		err = fmt.Errorf("unknown claim type: %s", claim.Spec.Type)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Unknown Claim Type")
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:        claim.Name,
			Namespace:   claim.Namespace,
			Labels:      claim.Labels,
			Annotations: claim.Annotations,
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretData,
	}

	if err := ctrl.SetControllerReference(claim, secret, r.Scheme); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Controller Reference Failed")
		return err
	}

	if createErr := r.Create(ctx, secret); createErr != nil {
		span.RecordError(createErr)
		span.SetStatus(codes.Error, "K8s Create Secret Failed")
		logger.Error("K8s API call failed to create Secret", slog.Any("error", createErr))
		return createErr
	}
	span.SetStatus(codes.Ok, "Success")
	return nil
}

func (r *SecretClaimReconciler) updateSecret(ctx context.Context, claim *secretsv1alpha1.SecretClaim, existingSecret *corev1.Secret) error {
	logger := observability.LoggerFromContext(ctx)

	ctx, span := r.Tracer.Start(ctx, "SecretClaimReconciler.updateSecret")
	defer span.End()

	span.SetAttributes(attribute.String("claim.type", claim.Spec.Type))

	secretData := make(map[string][]byte)
	var err error

	switch claim.Spec.Type {
	case "Opaque":
		for k, v := range claim.Spec.Data {
			secretData[k] = []byte(v)
		}
		span.AddEvent("Opaque data copied for update")

	case "AutoGenerated":
		if claim.Spec.Generation == nil {
			err = fmt.Errorf("generation spec is nil for AutoGenerated claim")
			span.RecordError(err)
			span.SetStatus(codes.Error, "Missing Generation Spec")
			return err
		}

		span.AddEvent("Starting password regeneration")
		for _, key := range claim.Spec.Generation.DataKeys {

			if claim.Spec.Generation.Length < 8 {
				err = fmt.Errorf("secrets should be at least 8 symbols")
				span.RecordError(err)
				span.SetStatus(codes.Error, "Secret lenght <8")
				return err
			}
			password, genErr := generatePassword(claim.Spec.Generation.Length, claim.Spec.Generation.Encoding)
			if genErr != nil {
				err = fmt.Errorf("failed to generate password for key %s: %w", key, genErr)
				span.RecordError(err)
				span.SetStatus(codes.Error, "Password Regeneration Failed")
				return err
			}
			secretData[key] = []byte(password)
		}
		span.AddEvent("Password regeneration complete")

	default:
		err = fmt.Errorf("unknown claim type: %s", claim.Spec.Type)
		span.RecordError(err)
		span.SetStatus(codes.Error, "Unknown Claim Type")
		return err
	}

	existingSecret.Type = corev1.SecretTypeOpaque
	existingSecret.Data = secretData

	logger.Debug("Attempting K8s Update Secret API call")
	if updateErr := r.Update(ctx, existingSecret); updateErr != nil {
		span.RecordError(updateErr)
		span.SetStatus(codes.Error, "K8s Update Secret Failed")
		logger.Error("K8s API call failed to update Secret", slog.Any("error", updateErr))
		return updateErr
	}

	span.SetStatus(codes.Ok, "Secret Updated")
	return nil

}

func (r *SecretClaimReconciler) updateStatus(ctx context.Context, claim *secretsv1alpha1.SecretClaim, synced bool, msg string) {
	logger := observability.LoggerFromContext(ctx)

	_, span := r.Tracer.Start(ctx, "SecretClaimReconciler.updateStatus")
	defer span.End()

	claim.Status.Synced = synced

	if msg != "" {
		claim.Status.ErrorMessage = msg
	} else {
		claim.Status.ErrorMessage = ""
	}

	if synced {
		lastUpdate := metav1.NewTime(time.Now())
		claim.Status.CreatedSecretName = claim.Name
		claim.Status.LastUpdate = &lastUpdate

		if claim.Spec.Type == "AutoGenerated" && claim.Spec.Generation != nil {
			claim.Status.LastReconcileTrigger = claim.Spec.Generation.ReconcileTrigger
		}
	} else {
		claim.Status.LastReconcileTrigger = ""
	}

	logger.Debug("Attempting K8s Status Update API call")
	if err := r.Status().Update(ctx, claim); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "K8s Status Update Failed")
		logger.Error("Failed to update status", slog.Any("error", err))
	} else {
		span.SetStatus(codes.Ok, "Status Updated")
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *SecretClaimReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.Log == nil {
		r.Log = slog.Default()
	}
	if r.Tracer == nil {
		r.Tracer = otel.Tracer("k8s-secret-manager")
	}
	return ctrl.NewControllerManagedBy(mgr).
		For(&secretsv1alpha1.SecretClaim{}).
		Named("secretclaim").
		Owns(&corev1.Secret{}).
		Complete(r)
}
