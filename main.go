package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Создаем новый роутер Gin без дефолтного middleware
	router := gin.New()

	// Добавляем Recovery middleware
	router.Use(gin.Recovery())

	// Добавляем кастомный Logger middleware, исключающий health-check пути
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/health", "/ready"},
	}))

	config := LoadConfig()

	router.GET("/health", healthHandler)
	router.GET("/ready", readyHandler)
	router.GET("/generate-jwt", generateJWTHandler(config))

	// Защищенные роуты с JWT аутентификацией
	router.GET("/me", jwtAuthMiddleware(config), meHandler())

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:    ":" + config.Port,
		Handler: router,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Println("Starting server on :" + config.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// Канал для получения сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Ждем сигнал завершения
	<-quit
	log.Println("Shutting down server...")

	// Создаем контекст с таймаутом для graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Корректно завершаем сервер
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
