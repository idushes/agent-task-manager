package tasks

import (
	"agent-task-manager/database"
	"agent-task-manager/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetTaskHandler обработчик для получения задачи в работу
func GetTaskHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем user_id из контекста (установлен в JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user_id not found in context",
			})
			return
		}

		db := database.GetDB()

		// Начинаем транзакцию для атомарного обновления
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to start transaction: " + tx.Error.Error(),
			})
			return
		}

		var task models.Task

		// Ищем задачу где assignee == userID и status == submitted
		// Используем FOR UPDATE для блокировки строки
		err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("assignee = ? AND status = ?", userID.(string), models.StatusSubmitted).
			First(&task).Error

		if err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "no tasks available for assignment",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to find task: " + err.Error(),
			})
			return
		}

		// Меняем статус на working
		task.Status = models.StatusWorking
		if err := tx.Save(&task).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update task status: " + err.Error(),
			})
			return
		}

		// Коммитим транзакцию
		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to commit transaction: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, task)
	}
}
