package main

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mogilyoy/k8s-secret-manager/internal/api"
	"github.com/mogilyoy/k8s-secret-manager/internal/auth"
	"github.com/mogilyoy/k8s-secret-manager/internal/handlers"
	"github.com/mogilyoy/k8s-secret-manager/internal/k8s"
)

const (
	// PORT - порт, на котором слушает REST API
	PORT = ":8080"
)

func main() {

	k8sManager, err := k8s.NewK8sSecretManager()
	if err != nil {
		log.Fatalf("❌ FATAL: Failed to initialize Kubernetes manager: %v", err)
	}
	log.Println("✅ Kubernetes Client (controller-runtime) initialized successfully.")

	// 2. Инициализация Сервисов
	authService := auth.NewAuthService()

	// 3. Инициализация Хэндлеров
	secretHandler := handlers.NewSecretHandler(k8sManager, authService)

	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	// 4a. Создаем StrictServerInterface, обернутый в HTTP-адаптер
	// secretHandler - это ваша реализация StrictServerInterface.
	// Мы передаем пустой слайс мидлваров, если не используем их.
	strictServer := api.NewStrictHandler(
		secretHandler, // <-- Ваш реализатор интерфейса
		nil,           // <-- Мидлвары StrictServer (можно добавить аутентификацию)
	)

	// 4b. Используем сгенерированный Chi-адаптер для подключения к роутеру
	// Эта функция (HandlerFromMux) берет сгенерированный адаптер и регистрирует все роуты Chi.
	// Она сама знает, как преобразовать вызов из http.ResponseWriter в сигнатуру Go-интерфейса.
	apiRouter := api.HandlerFromMux(strictServer, router)
	// 5. Запуск Сервера
	srv := &http.Server{
		Addr:         PORT,
		Handler:      apiRouter,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("🚀 Starting REST API server on %s", PORT)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("❌ Could not listen on %s: %v", PORT, err)
	}
}
