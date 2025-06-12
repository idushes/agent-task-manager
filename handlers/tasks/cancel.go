package tasks

import (
	"agent-task-manager/database"
	"agent-task-manager/models"
	"agent-task-manager/redis"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// CancelTaskHandler обработчик для отмены задачи
func CancelTaskHandler() gin.HandlerFunc {
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

		// Проверяем, что пользователь является исполнителем задачи или создателем
		if task.Assignee != userID.(string) && task.CreatedBy != userID.(string) {
			tx.Rollback()
			c.JSON(http.StatusForbidden, gin.H{
				"error": "only assignee or creator can cancel the task",
			})
			return
		}

		// Проверяем текущий статус задачи - нельзя отменить уже завершенную или отмененную задачу
		if task.Status == models.StatusCompleted || task.Status == models.StatusCanceled || task.Status == models.StatusFailed {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "cannot cancel task with status: " + string(task.Status),
				"current_status": task.Status,
			})
			return
		}

		// Обновляем задачу
		task.Status = models.StatusCanceled

		if err := tx.Save(&task).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update task: " + err.Error(),
			})
			return
		}

		// Рекурсивно отменяем все активные подзадачи этой задачи
		if err := CancelSubtasksRecursive(tx, task.ID); err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to cancel subtasks: " + err.Error(),
			})
			return
		}

		// Если у задачи есть родитель, проверяем все задачи с таким же parent
		if task.ParentTaskID != nil {
			// Подсчитываем задачи с таким же parent
			var totalCount int64
			var completedOrCanceledCount int64

			tx.Model(&models.Task{}).Where("parent_task_id = ?", task.ParentTaskID).Count(&totalCount)
			tx.Model(&models.Task{}).Where("parent_task_id = ? AND status IN ?", task.ParentTaskID, []models.TaskStatus{models.StatusCompleted, models.StatusCanceled}).Count(&completedOrCanceledCount)

			// Если все подзадачи завершены или отменены, обновляем статус родительской задачи
			if totalCount > 0 && totalCount == completedOrCanceledCount {
				// Получаем информацию о родительской задаче для отправки уведомления
				var parentTask models.Task
				if err := tx.First(&parentTask, "id = ?", task.ParentTaskID).Error; err == nil {
					// Обновляем статус родительской задачи
					if err := tx.Model(&models.Task{}).
						Where("id = ?", task.ParentTaskID).
						Update("status", models.StatusSubmitted).Error; err != nil {
						tx.Rollback()
						c.JSON(http.StatusInternalServerError, gin.H{
							"error": "failed to update parent task status: " + err.Error(),
						})
						return
					}

					// Отправляем уведомление в Redis очередь для родительской задачи
					if err := redis.SendTaskNotification(parentTask.ID.String(), parentTask.Assignee); err != nil {
						// Логируем ошибку, но не прерываем выполнение
						c.Error(err)
					}
				} else {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "failed to get parent task: " + err.Error(),
					})
					return
				}
			}
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
