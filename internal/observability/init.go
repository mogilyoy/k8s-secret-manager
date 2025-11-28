package observability

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/mogilyoy/k8s-secret-manager/internal/cfg"
	stdout "go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
)

type ContextHandler struct {
	slog.Handler
}

func InitTracer() *sdktrace.TracerProvider {

	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.AppConfig.Service.Name),
			semconv.ServiceVersion(cfg.AppConfig.Service.Version),
		),
	)
	if err != nil {
		slog.Error("failed to create resource")
	}

	exporter, err := stdout.New(stdout.WithPrettyPrint())
	if err != nil {
		slog.Error("failed to create stdout exporter")
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	slog.Info("OpenTelemetry initialized successfully")

	return tp
}

func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		r.Add("trace_id", slog.StringValue(spanCtx.TraceID().String()))
		r.Add("span_id", slog.StringValue(spanCtx.SpanID().String()))
	}

	return h.Handler.Handle(ctx, r)
}

func NewContextualLogger(w io.Writer, opts *slog.HandlerOptions) *slog.Logger {
	baseHandler := slog.NewJSONHandler(w, opts)

	contextualHandler := ContextHandler{
		Handler: baseHandler,
	}
	return slog.New(contextualHandler)
}

func NewOTelMiddleware(serviceName string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return otelhttp.NewHandler(next, serviceName)
	}
}

func NewSlogMiddleware(baseLogger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := middleware.GetReqID(r.Context())

			if span := trace.SpanFromContext(r.Context()); span != nil {
				span.SetAttributes(attribute.String("request.id", reqID))
			}

			requestLogger := baseLogger.With(
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.String("request_id", reqID),
			)

			ctx := context.WithValue(r.Context(), LoggerContextKey, requestLogger)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func LoggerFromContext(ctx context.Context) *slog.Logger {
	if logger, ok := ctx.Value(LoggerContextKey).(*slog.Logger); ok {
		return logger
	}
	return slog.Default()
}

func SlogRequestLogger() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			logger := LoggerFromContext(r.Context())

			entry := &LogEntry{logger: logger}
			wrap := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()

			next.ServeHTTP(wrap, r.WithContext(context.WithValue(r.Context(), middleware.LogEntryCtxKey, entry)))

			entry.Write(wrap.Status(), wrap.BytesWritten(), wrap.Header(), time.Since(t1))
		}
		return http.HandlerFunc(fn)
	}
}

type LogEntry struct {
	logger *slog.Logger
}

func (l *LogEntry) Write(status, bytes int, header http.Header, elapsed time.Duration) {
	l.logger.Info("Request completed",
		slog.Int("status", status),
		slog.Int("bytes", bytes),
		slog.Duration("elapsed", elapsed),
	)
}

func (l *LogEntry) Panic(v interface{}, stack []byte) {
	l.logger.Error("Request panicked",
		slog.Any("recover", v),
		slog.String("stack", string(stack)),
	)
}

func ResponseHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := middleware.GetReqID(r.Context())
		if requestID != "" {
			w.Header().Set("X-Request-ID", requestID)
		}
		next.ServeHTTP(w, r)
	})
}
