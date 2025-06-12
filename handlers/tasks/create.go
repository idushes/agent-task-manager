package tasks

import (
	"agent-task-manager/database"
	"agent-task-manager/models"
	"agent-task-manager/redis"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

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
			credsMap, err := validateCredentials(req.Credentials)
			if err != nil {
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

			// Проверяем, что parent задача находится в разрешенном статусе
			if !isParentStatusAllowed(parentTask.Status) {
				c.JSON(http.StatusBadRequest, gin.H{
					"error":         "parent task must be in waiting, working or submitted status",
					"parent_status": parentTask.Status,
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

		// Отправляем уведомление в Redis очередь для новой задачи со статусом submitted
		if err := redis.SendTaskNotification(task.ID.String(), task.Assignee); err != nil {
			// Логируем ошибку, но не прерываем выполнение
			c.Error(err)
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
		} else {
			// Если есть parent, переводим его в статус waiting
			if err := db.Model(&models.Task{}).
				Where("id = ?", req.ParentTaskID).
				Update("status", models.StatusWaiting).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "failed to update parent task status: " + err.Error(),
				})
				return
			}
		}

		c.JSON(http.StatusCreated, task)
	}
}
