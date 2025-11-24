package observability

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

type contextKey string

const (
	K8sTraceAnnotationKey                  = "k8s-secret-manager/trace-id"
	K8sTraceparentAnnotationKey            = "k8s-secret-manager/traceparent"
	LoggerContextKey            contextKey = "slog-logger"
)

func GetTraceID(ctx context.Context) string {
	spanCtx := trace.SpanFromContext(ctx).SpanContext()
	if spanCtx.IsValid() {
		return spanCtx.TraceID().String()
	}
	return ""
}

func GetTraceParentHeader(ctx context.Context) string {
	carrier := propagation.MapCarrier{}

	otel.GetTextMapPropagator().Inject(ctx, carrier)

	return carrier["traceparent"]
}
