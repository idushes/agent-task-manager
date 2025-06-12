package tasks

import (
	"agent-task-manager/database"
	"agent-task-manager/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUserRootTasksHandler обработчик для получения списка корневых задач пользователя
func GetUserRootTasksHandler() gin.HandlerFunc {
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

		var tasks []models.Task

		// Ищем задачи где created_by == userID и id == root_task_id (корневые задачи)
		// Корневая задача - это задача где ID равен RootTaskID
		err := db.Where("created_by = ? AND id = root_task_id", userID.(string)).
			Find(&tasks).Error

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to fetch root tasks: " + err.Error(),
			})
			return
		}

		// Преобразуем задачи в формат RootTaskSummary
		summaries := make([]RootTaskSummary, len(tasks))
		for i, task := range tasks {
			summaries[i] = RootTaskSummary{
				RootTaskID:  task.ID,
				CreatedAt:   task.CreatedAt,
				DeleteAt:    task.DeleteAt,
				Assignee:    task.Assignee,
				Description: task.Description,
				Status:      task.Status,
			}
		}

		c.JSON(http.StatusOK, summaries)
	}
}
