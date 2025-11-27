package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/mogilyoy/k8s-secret-manager/internal/cfg"
	"github.com/mogilyoy/k8s-secret-manager/internal/handlers"
	"github.com/mogilyoy/k8s-secret-manager/internal/k8s"
	authMiddleware "github.com/mogilyoy/k8s-secret-manager/internal/middleware"
	"github.com/mogilyoy/k8s-secret-manager/internal/observability"
)

func main() {

	tp := observability.InitTracer()
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			slog.Error("❌ Error shutting down tracer provider: %v", slog.Any("error", err))
		}
	}()

	slog.Info("✅ OpenTelemetry Tracer Provider initialized.")

	logger := observability.NewContextualLogger(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})

	slog.SetDefault(logger)

	configPath := os.Getenv("CONFIG_PATH")
	config, err := cfg.LoadConfig(configPath)
	if err != nil {
		slog.Error("❌ FATAL: Failed to load config: %v", slog.Any("error", err))
		os.Exit(1)
	}
	if config.JWT.Secret == "" {
		slog.Error("❌ FATAL: JWT secret is empty. Set via config.yaml or JWT_SECRET environment variable.")
		os.Exit(1)
	}
	slog.Info("✅ Configuration loaded successfully.")

	k8sManager, err := k8s.NewK8sSecretManager(logger, tp)
	if err != nil {
		slog.Error("❌ FATAL: Failed to initialize Kubernetes manager: %v", slog.Any("error", err))
	}
	slog.Info("✅ Kubernetes Client initialized successfully.")

	tracer := tp.Tracer(cfg.AppConfig.Service.Name)
	secretHandler := handlers.NewSecretHandler(k8sManager, *config, logger, tracer)

	router := chi.NewRouter()
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(observability.NewOTelMiddleware(cfg.AppConfig.Service.Name))
	router.Use(observability.NewSlogMiddleware(logger))
	router.Use(observability.SlogRequestLogger())

	strictServer := api.NewStrictHandler(secretHandler, nil)

	baseAPIMux := chi.NewMux()
	api.HandlerFromMux(strictServer, baseAPIMux)

	router.Post("/user/auth", func(w http.ResponseWriter, r *http.Request) {
		baseAPIMux.ServeHTTP(w, r)
	})

	router.Group(func(r chi.Router) {
		jwtMiddlewareFunc := authMiddleware.JWTAuthMiddleware(config.JWT.Secret)
		r.Use(jwtMiddlewareFunc)
		r.Mount("/", baseAPIMux)
	})

	// 7. Запуск сервера
	srv := &http.Server{
		Addr:         cfg.AppConfig.Service.Port,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	slog.Info("🚀 Starting REST API server on %s", slog.String("port", cfg.AppConfig.Service.Port))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		slog.Error("❌ Could not listen on %s: %v", slog.Any("port", cfg.AppConfig.Service.Port), slog.Any("error", err))
	}
}
