package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Создаем новый роутер Gin
	router := gin.Default()

	router.GET("/health", healthHandler)
	router.GET("/ready", readyHandler)

	// Запускаем сервер на порту 8080
	log.Println("Starting server on :8081")
	if err := router.Run(":8081"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
