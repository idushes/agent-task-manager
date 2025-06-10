package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config структура для хранения конфигурации приложения
type Config struct {
	Port        string
	SecretKey   string
	PostgresURL string
	RedisURL    string
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
		RedisURL:    getEnvOrDefault("REDIS_URL", ""),
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
