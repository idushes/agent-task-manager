package scheduler

import (
	"context"
	"log"
	"time"

	"agent-task-manager/database"
	"agent-task-manager/models"

	"gorm.io/gorm"
)

// TaskCleanupScheduler управляет периодической очисткой задач
type TaskCleanupScheduler struct {
	interval time.Duration
	stopChan chan struct{}
}

// NewTaskCleanupScheduler создает новый планировщик очистки задач
func NewTaskCleanupScheduler(interval time.Duration) *TaskCleanupScheduler {
	return &TaskCleanupScheduler{
		interval: interval,
		stopChan: make(chan struct{}),
	}
}

// Start запускает периодическую очистку задач
func (s *TaskCleanupScheduler) Start() {
	log.Println("Starting task cleanup scheduler with interval:", s.interval)

	// Запускаем первую очистку сразу
	s.cleanupExpiredTasks()

	// Создаем тикер для периодического запуска
	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupExpiredTasks()
		case <-s.stopChan:
			log.Println("Stopping task cleanup scheduler")
			return
		}
	}
}

// Stop останавливает планировщик
func (s *TaskCleanupScheduler) Stop() {
	close(s.stopChan)
}

// cleanupExpiredTasks удаляет задачи с истекшим DeleteAt
func (s *TaskCleanupScheduler) cleanupExpiredTasks() {
	db := database.GetDB()
	if db == nil {
		log.Println("Database connection is not available")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Подсчитываем количество задач для удаления
	var count int64
	now := time.Now()

	// Сначала считаем количество задач для удаления
	err := db.WithContext(ctx).Model(&models.Task{}).
		Where("delete_at IS NOT NULL AND delete_at < ?", now).
		Count(&count).Error

	if err != nil {
		log.Printf("Error counting tasks for cleanup: %v", err)
		return
	}

	if count == 0 {
		log.Println("No tasks to cleanup")
		return
	}

	log.Printf("Found %d tasks to cleanup", count)

	// Удаляем задачи с истекшим DeleteAt
	// Используем транзакцию для безопасного удаления
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Удаляем задачи где DeleteAt меньше текущего времени
		result := tx.Where("delete_at IS NOT NULL AND delete_at < ?", now).
			Delete(&models.Task{})

		if result.Error != nil {
			return result.Error
		}

		log.Printf("Successfully deleted %d expired tasks", result.RowsAffected)
		return nil
	})

	if err != nil {
		log.Printf("Error cleaning up expired tasks: %v", err)
	}
}
