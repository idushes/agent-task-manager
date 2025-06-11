package handlers

import (
	"agent-task-manager/database"
	"agent-task-manager/models"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateTaskRequest структура для запроса создания задачи
type CreateTaskRequest struct {
	Description  string          `json:"description" binding:"required"`
	Assignee     string          `json:"assignee"`
	ParentTaskID *uuid.UUID      `json:"parent_task_id"`
	DeleteAt     *time.Time      `json:"delete_at"`
	Credentials  json.RawMessage `json:"credentials"`
}

// CreateTaskHandler обработчик для создания новой задачи
func CreateTaskHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем user_id из контекста (установлен в JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user_id not found in context",
			})
			return
		}

		var req CreateTaskRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid request body: " + err.Error(),
			})
			return
		}

		// Валидация Credentials
		credentials := json.RawMessage("{}")
		if req.Credentials != nil && len(req.Credentials) > 0 {
			// Проверяем структуру credentials
			// Ожидаемая структура: { "service": { "ENV_VAR": "value" } }
			var credsMap map[string]map[string]string
			if err := json.Unmarshal(req.Credentials, &credsMap); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "invalid credentials format: " + err.Error(),
				})
				return
			}

			// Валидируем что все поля заполнены
			for serviceName, serviceVars := range credsMap {
				if serviceName == "" {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "service name cannot be empty",
					})
					return
				}
				if serviceVars == nil || len(serviceVars) == 0 {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "service credentials cannot be empty",
					})
					return
				}

				for envVar, value := range serviceVars {
					if envVar == "" {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "environment variable name cannot be empty",
						})
						return
					}
					if value == "" {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "environment variable value cannot be empty",
						})
						return
					}
				}
			}

			credentials = req.Credentials
		}

		// Устанавливаем DeleteAt по умолчанию на +3 месяца, если не указано
		deleteAt := req.DeleteAt
		if deleteAt == nil {
			threeMonthsLater := time.Now().AddDate(0, 3, 0)
			deleteAt = &threeMonthsLater
		}

		// Создаем задачу
		task := &models.Task{
			CreatedBy:    userID.(string),
			Description:  req.Description,
			Assignee:     req.Assignee,
			ParentTaskID: req.ParentTaskID,
			DeleteAt:     deleteAt,
			Credentials:  credentials,
			Status:       models.StatusSubmitted,
		}

		db := database.GetDB()

		// Если есть ParentTaskID, нужно получить RootTaskID из родительской задачи
		if req.ParentTaskID != nil {
			var parentTask models.Task
			if err := db.First(&parentTask, "id = ?", req.ParentTaskID).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "parent task not found",
				})
				return
			}

			// Устанавливаем RootTaskID из родительской задачи
			task.RootTaskID = parentTask.RootTaskID
			// Если у родительской задачи нет RootTaskID, используем ID родительской задачи
			if task.RootTaskID == nil {
				task.RootTaskID = &parentTask.ID
			}
		}

		// Создаем задачу в БД
		if err := db.Create(&task).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to create task: " + err.Error(),
			})
			return
		}

		// Если ParentTaskID == NULL, устанавливаем RootTaskID = ID созданной задачи
		if req.ParentTaskID == nil {
			task.RootTaskID = &task.ID
			if err := db.Save(&task).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to update root task id: " + err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusCreated, task)
	}
}
