package tasks

import (
	"agent-task-manager/database"
	"agent-task-manager/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskWithoutCredentials представляет задачу без поля Credentials
type TaskWithoutCredentials struct {
	ID           uuid.UUID         `json:"id"`
	CreatedAt    time.Time         `json:"created_at"`
	DeleteAt     *time.Time        `json:"delete_at,omitempty"`
	CreatedBy    string            `json:"created_by"`
	Assignee     string            `json:"assignee"`
	Description  string            `json:"description"`
	RootTaskID   *uuid.UUID        `json:"root_task_id,omitempty"`
	ParentTaskID *uuid.UUID        `json:"parent_task_id,omitempty"`
	Result       string            `json:"result"`
	Status       models.TaskStatus `json:"status"`
}

// GetRootTasksHandler обработчик для получения всех задач по root_task_id
func GetRootTasksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем user_id из контекста (установлен в JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user_id not found in context",
			})
			return
		}

		// Получаем ID root задачи из параметра пути
		rootTaskIDStr := c.Param("id")
		rootTaskID, err := uuid.Parse(rootTaskIDStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid root task id format",
			})
			return
		}

		db := database.GetDB()

		// Сначала проверяем, что root задача существует и создана текущим пользователем
		var rootTask models.Task
		if err := db.First(&rootTask, "id = ?", rootTaskID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{
					"error": "root task not found",
				})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to find root task: " + err.Error(),
			})
			return
		}

		// Проверяем, что текущий пользователь является создателем root задачи
		if rootTask.CreatedBy != userID.(string) {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "access denied: you are not the creator of this root task",
			})
			return
		}

		// Получаем все задачи с данным root_task_id
		var tasks []models.Task
		if err := db.Where("root_task_id = ?", rootTaskID).Find(&tasks).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get tasks: " + err.Error(),
			})
			return
		}

		// Конвертируем задачи в структуры без Credentials
		tasksWithoutCreds := make([]TaskWithoutCredentials, len(tasks))
		for i, task := range tasks {
			tasksWithoutCreds[i] = TaskWithoutCredentials{
				ID:           task.ID,
				CreatedAt:    task.CreatedAt,
				DeleteAt:     task.DeleteAt,
				CreatedBy:    task.CreatedBy,
				Assignee:     task.Assignee,
				Description:  task.Description,
				RootTaskID:   task.RootTaskID,
				ParentTaskID: task.ParentTaskID,
				Result:       task.Result,
				Status:       task.Status,
			}
		}

		c.JSON(http.StatusOK, tasksWithoutCreds)
	}
}
