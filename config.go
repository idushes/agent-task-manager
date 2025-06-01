package main

import (
	"fmt"
	"os"
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
