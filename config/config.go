package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config структура для хранения конфигурации приложения
type Config struct {
	Port              string
	SecretKey         string
	PostgresURL       string
	AllowedOrigins    []string
	BlacklistedUsers  []string
	CleanupInterval   time.Duration
	CacheSyncInterval time.Duration
}

// LoadConfig загружает конфигурацию из переменных окружения
func LoadConfig() (*Config, error) {
	// Проверяем, существует ли файл .env
	if _, err := os.Stat(".env"); err == nil {
		// Файл существует, пытаемся его загрузить
		if errLoad := godotenv.Load(".env"); errLoad != nil {
			log.Printf("Warning: error loading .env file: %s", errLoad)
		} else {
			log.Println(".env file loaded successfully")
		}
	} else if os.IsNotExist(err) {
		log.Println("No .env file found, using environment variables only")
	} else {
		// Другая ошибка при проверке файла .env (например, нет прав доступа)
		log.Printf("Warning: error checking .env file: %s", err)
	}

	config := &Config{
		Port:        getEnvOrDefault("PORT", "8081"),
		SecretKey:   getEnvOrDefault("SECRET_KEY", ""),
		PostgresURL: getEnvOrDefault("POSTGRES_URL", ""),
	}

	// Загружаем интервал очистки (по умолчанию 1 час)
	cleanupIntervalStr := getEnvOrDefault("CLEANUP_INTERVAL", "1h")
	cleanupInterval, err := time.ParseDuration(cleanupIntervalStr)
	if err != nil {
		log.Printf("Invalid CLEANUP_INTERVAL format, using default (1h): %v", err)
		cleanupInterval = 1 * time.Hour
	}
	config.CleanupInterval = cleanupInterval

	// Загружаем интервал синхронизации кэша (по умолчанию 10 минут)
	cacheSyncIntervalStr := getEnvOrDefault("CACHE_SYNC_INTERVAL", "10m")
	cacheSyncInterval, err := time.ParseDuration(cacheSyncIntervalStr)
	if err != nil {
		log.Printf("Invalid CACHE_SYNC_INTERVAL format, using default (10m): %v", err)
		cacheSyncInterval = 10 * time.Minute
	}
	config.CacheSyncInterval = cacheSyncInterval

	// Загружаем список разрешенных доменов
	allowedOriginsStr := getEnvOrDefault("ALLOWED_ORIGINS", "*")
	if allowedOriginsStr == "*" {
		config.AllowedOrigins = []string{"*"}
	} else {
		// Разделяем строку по запятым и удаляем пробелы
		origins := strings.Split(allowedOriginsStr, ",")
		for i, origin := range origins {
			origins[i] = strings.TrimSpace(origin)
		}
		config.AllowedOrigins = origins
	}

	// Загружаем список заблокированных пользователей
	blacklistedUsersStr := getEnvOrDefault("BLACKLISTED_USERS", "")
	if blacklistedUsersStr != "" {
		// Разделяем строку по запятым и удаляем пробелы
		users := strings.Split(blacklistedUsersStr, ",")
		for i, user := range users {
			users[i] = strings.TrimSpace(user)
		}
		config.BlacklistedUsers = users
	} else {
		config.BlacklistedUsers = []string{}
	}

	// Проверяем обязательные параметры
	if config.SecretKey == "" {
		return nil, fmt.Errorf("SECRET_KEY environment variable is required")
	}

	return config, nil
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
