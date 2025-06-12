package handlers

import (
	"agent-task-manager/config"
	"net/http"
	"strings"
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

// GenerateJWTRequest структура для запроса генерации JWT токена
type GenerateJWTRequest struct {
	Secret    string `json:"secret" binding:"required"`
	UserID    string `json:"user_id,omitempty"`
	ExpiresIn int    `json:"expires_in,omitempty"`
}

// UserInfoResponse структура для ответа с информацией о пользователе
type UserInfoResponse struct {
	UserID    string `json:"user_id"`
	ExpiresAt int64  `json:"expires_at"`
}

// Claims структура для JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// JwtAuthMiddleware middleware для проверки JWT токена
func JwtAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	// Создаем карту для быстрой проверки blacklist пользователей
	blacklistedUsers := make(map[string]bool)
	for _, user := range cfg.BlacklistedUsers {
		blacklistedUsers[user] = true
	}

	return func(c *gin.Context) {
		// Получаем токен из заголовка Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header is required",
			})
			c.Abort()
			return
		}

		// Проверяем формат Bearer токена
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format. Use: Bearer <token>",
			})
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Парсим и валидируем токен
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			// Проверяем алгоритм подписи
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.SecretKey), nil
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token: " + err.Error(),
			})
			c.Abort()
			return
		}

		// Проверяем что токен валидный
		claims, ok := token.Claims.(*Claims)
		if !ok || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token claims",
			})
			c.Abort()
			return
		}

		// Проверяем, не заблокирован ли пользователь
		if blacklistedUsers[claims.UserID] {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "user has been blocked",
			})
			c.Abort()
			return
		}

		// Сохраняем claims в контексте для дальнейшего использования
		c.Set("claims", claims)
		c.Set("user_id", claims.UserID)

		c.Next()
	}
}

// MeHandler обработчик для получения информации о текущем пользователе
func MeHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем claims из контекста (они были установлены в middleware)
		claimsInterface, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "claims not found in context",
			})
			return
		}

		claims, ok := claimsInterface.(*Claims)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "invalid claims type",
			})
			return
		}

		// Формируем ответ с информацией о пользователе
		response := UserInfoResponse{
			UserID:    claims.UserID,
			ExpiresAt: claims.ExpiresAt.Unix(),
		}

		c.JSON(http.StatusOK, response)
	}
}

// GenerateJWTHandler обработчик для генерации JWT токена
func GenerateJWTHandler(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GenerateJWTRequest

		// Парсим JSON из тела запроса
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request format: " + err.Error(),
			})
			return
		}

		// Проверяем что secret совпадает с секретом из конфига
		if req.Secret != cfg.SecretKey {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid secret",
			})
			return
		}

		// Устанавливаем значения по умолчанию
		userID := req.UserID
		if userID == "" {
			userID = "anonymous"
		}

		expiresInHours := req.ExpiresIn
		if expiresInHours <= 0 {
			expiresInHours = 8760 // 365 дней * 24 часа = 8760 часов
		}

		// Вычисляем время истечения
		expirationTime := time.Now().Add(time.Duration(expiresInHours) * time.Hour)

		// Создаем claims для токена
		claims := &Claims{
			UserID: userID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		// Создаем токен
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

		// Подписываем токен секретным ключом
		tokenString, err := token.SignedString([]byte(cfg.SecretKey))
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
