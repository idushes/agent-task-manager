package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agent-task-manager/config"
	"agent-task-manager/database"
	"agent-task-manager/handlers"
	"agent-task-manager/handlers/tasks"

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

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Инициализируем подключение к базе данных
	if err := database.InitDB(cfg.PostgresURL); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	router.GET("/health", handlers.HealthHandler)
	router.GET("/ready", handlers.ReadyHandler)
	router.GET("/", handlers.InfoHandler())
	router.GET("/generate-jwt", handlers.GenerateJWTHandler(cfg))

	// Защищенные роуты с JWT аутентификацией
	router.GET("/me", handlers.JwtAuthMiddleware(cfg), handlers.MeHandler())
	router.POST("/task", handlers.JwtAuthMiddleware(cfg), tasks.CreateTaskHandler())
	router.GET("/task", handlers.JwtAuthMiddleware(cfg), tasks.GetTaskHandler())
	router.POST("/task/:id/complete", handlers.JwtAuthMiddleware(cfg), tasks.CompleteTaskHandler())
	router.POST("/task/:id/cancel", handlers.JwtAuthMiddleware(cfg), tasks.CancelTaskHandler())
	router.POST("/tasks/:id/fail", handlers.JwtAuthMiddleware(cfg), tasks.FailTaskHandler())
	router.GET("/root-task/:id/tasks", handlers.JwtAuthMiddleware(cfg), tasks.GetRootTasksHandler())

	// Создаем HTTP сервер
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	// Запускаем сервер в горутине
	go func() {
		log.Println("Starting server on :" + cfg.Port)
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

	// Закрываем соединение с базой данных
	if err := database.CloseDB(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	} else {
		log.Println("Database connection closed")
	}

	log.Println("Server exited")
}
