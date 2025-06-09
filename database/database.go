package database

import (
	"fmt"
	"log"

	"agent-task-manager/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB инициализирует подключение к базе данных PostgreSQL
func InitDB(postgresURL string) error {
	if postgresURL == "" {
		return fmt.Errorf("POSTGRES_URL is required")
	}

	// Настройки GORM
	gormConfig := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Подключаемся к PostgreSQL
	db, err := gorm.Open(postgres.Open(postgresURL), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}
	log.Println("Connected to PostgreSQL database")

	// Автомиграция
	if err := db.AutoMigrate(&models.Task{}); err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}
	log.Println("Database migration completed")

	DB = db
	return nil
}

// GetDB возвращает экземпляр базы данных
func GetDB() *gorm.DB {
	return DB
}

// CloseDB закрывает соединение с базой данных
func CloseDB() error {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}
