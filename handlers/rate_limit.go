package handlers

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter структура для хранения информации о rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	window   time.Duration
}

// NewRateLimiter создает новый rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}
}

// RateLimitMiddleware создает middleware для ограничения количества запросов
func RateLimitMiddleware(limit int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(limit, window)

	// Запускаем горутину для периодической очистки старых записей
	go func() {
		ticker := time.NewTicker(window)
		defer ticker.Stop()
		for range ticker.C {
			limiter.cleanOldRequests()
		}
	}()

	return func(c *gin.Context) {
		// Используем IP адрес клиента как ключ
		clientIP := c.ClientIP()

		if !limiter.Allow(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Allow проверяет, можно ли разрешить запрос
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Получаем список запросов для данного ключа
	requests, exists := rl.requests[key]
	if !exists {
		requests = []time.Time{}
	}

	// Фильтруем запросы, оставляя только те, которые попадают в окно
	var validRequests []time.Time
	for _, reqTime := range requests {
		if reqTime.After(windowStart) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Проверяем лимит
	if len(validRequests) >= rl.limit {
		rl.requests[key] = validRequests
		return false
	}

	// Добавляем текущий запрос
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests

	return true
}

// cleanOldRequests удаляет старые записи для экономии памяти
func (rl *RateLimiter) cleanOldRequests() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	for key, requests := range rl.requests {
		var validRequests []time.Time
		for _, reqTime := range requests {
			if reqTime.After(windowStart) {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) == 0 {
			delete(rl.requests, key)
		} else {
			rl.requests[key] = validRequests
		}
	}
}
