package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Создаем новый роутер Gin
	router := gin.Default()
	config := LoadConfig()

	router.GET("/health", healthHandler)
	router.GET("/ready", readyHandler)

	// Запускаем сервер на порту 8080
	log.Println("Starting server on :" + config.Port)
	if err := router.Run(":" + config.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
