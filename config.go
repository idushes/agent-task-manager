package main

import (
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
func LoadConfig() *Config {
	config := &Config{
		Port:        getEnvOrDefault("PORT", "8081"),
		SecretKey:   getEnvOrDefault("SECRET_KEY", ""),
		PostgresURL: getEnvOrDefault("POSTGRES_URL", ""),
		RedisURL:    getEnvOrDefault("REDIS_URL", ""),
	}

	return config
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
