package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// JWTResponse структура для ответа с JWT токеном
type JWTResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	UserID    string `json:"user_id"`
}

// Claims структура для JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// generateJWTHandler обработчик для генерации JWT токена
func generateJWTHandler(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем secret из query параметра
		secretParam := c.Query("secret")

		// Проверяем что secret передан
		if secretParam == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "secret parameter is required",
			})
			return
		}

		// Проверяем что secret совпадает с секретом из конфига
		if secretParam != config.SecretKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid secret",
			})
			return
		}

		// Проверяем что SecretKey в конфиге не пустой
		if config.SecretKey == "" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "server secret key is not configured",
			})
			return
		}

		// Получаем параметр expires_in (в часах), по умолчанию год (8760 часов)
		expiresInStr := c.DefaultQuery("expires_in", "8760") // 365 дней * 24 часа = 8760 часов
		expiresInHours, err := strconv.Atoi(expiresInStr)
		if err != nil || expiresInHours <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "expires_in must be a positive integer (hours)",
			})
			return
		}

		// Получаем параметр user_id, по умолчанию "anonymous"
		userID := c.DefaultQuery("user_id", "anonymous")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "user_id cannot be empty",
			})
			return
		}

		// Вычисляем время истечения
		expirationTime := time.Now().Add(time.Duration(expiresInHours) * time.Hour)

		// Создаем claims для токена
		claims := &Claims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				Issuer:    "agent-task-manager",
				Subject:   userID,
			},
		}

		// Создаем токен
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Подписываем токен секретным ключом
		tokenString, err := token.SignedString([]byte(config.SecretKey))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to generate token",
			})
			return
		}

		// Возвращаем токен
		c.JSON(http.StatusOK, JWTResponse{
			Token:     tokenString,
			ExpiresAt: expirationTime.Unix(),
			UserID:    userID,
		})
	}
}
