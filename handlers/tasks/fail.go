package tasks

import (
	"agent-task-manager/database"
	"agent-task-manager/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FailTaskHandler обработчик для пометки задачи как неудачной
func FailTaskHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем user_id из контекста (установлен в JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user_id not found in context",
			})
			return
		}

		// Получаем ID задачи из параметра пути
		taskIDStr := c.Param("id")
		taskID, err := uuid.Parse(taskIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid task id format",
			})
			return
		}

		var req FailTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		db := database.GetDB()

		// Начинаем транзакцию
		tx := db.Begin()
		if tx.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to start transaction: " + tx.Error.Error(),
			})
			return
		}

		// Получаем задачу и проверяем права
		var task models.Task
		if err := tx.First(&task, "id = ?", taskID).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "task not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to find task: " + err.Error(),
			})
			return
		}

		// Проверяем, что пользователь является исполнителем задачи
		if task.Assignee != userID.(string) {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{
				"error": "only assignee can fail the task",
			})
			return
		}

		// Проверяем текущий статус задачи
		if task.Status != models.StatusWorking {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "task must be in working status to fail",
				"current_status": task.Status,
			})
			return
		}

		// Обновляем задачу
		task.Status = models.StatusFailed
		task.Result = "FAILURE REASON: " + req.Reason

		if err := tx.Save(&task).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update task: " + err.Error(),
			})
			return
		}

		// В отличие от complete, при fail родительская задача остается в статусе waiting

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
