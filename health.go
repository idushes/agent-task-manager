package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)


// healthHandler обрабатывает запросы на проверку жизнеспособности сервиса
func healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "alive",
		"message": "Service is running",
	})
}

// readyHandler обрабатывает запросы на проверку готовности сервиса
func readyHandler(c *gin.Context) {
	// Здесь можно добавить проверки подключения к БД, внешним сервисам и т.д.
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"message": "Service is ready to accept requests",
	})
}
