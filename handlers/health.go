package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthHandler обрабатывает запросы на проверку жизнеспособности сервиса
func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "alive",
		"message": "Service is running",
	})
}

// ReadyHandler обрабатывает запросы на проверку готовности сервиса
func ReadyHandler(c *gin.Context) {
	// Здесь можно добавить проверки подключения к БД, внешним сервисам и т.д.
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"message": "Service is ready to accept requests",
	})
}
