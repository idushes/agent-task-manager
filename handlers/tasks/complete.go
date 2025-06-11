package tasks

import (
	"agent-task-manager/database"
	"agent-task-manager/models"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// cancelSubtasksRecursive рекурсивно отменяет все активные подзадачи
func cancelSubtasksRecursive(tx *gorm.DB, parentID uuid.UUID) error {
	// Получаем все подзадачи, которые нужно отменить
	var subtasks []models.Task
	if err := tx.Where("parent_task_id = ? AND status IN ?", parentID, []models.TaskStatus{
		models.StatusSubmitted,
		models.StatusWorking,
		models.StatusWaiting,
	}).Find(&subtasks).Error; err != nil {
		return err
	}

	// Отменяем найденные подзадачи
	if len(subtasks) > 0 {
		subtaskIDs := make([]uuid.UUID, len(subtasks))
		for i, subtask := range subtasks {
			subtaskIDs[i] = subtask.ID
		}

		if err := tx.Model(&models.Task{}).
			Where("id IN ?", subtaskIDs).
			Update("status", models.StatusCanceled).Error; err != nil {
			return err
		}

		// Рекурсивно отменяем подзадачи каждой подзадачи
		for _, subtask := range subtasks {
			if err := cancelSubtasksRecursive(tx, subtask.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// CompleteTaskHandler обработчик для завершения задачи
func CompleteTaskHandler() gin.HandlerFunc {
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

		var req CompleteTaskRequest
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
				"error": "only assignee can complete the task",
			})
			return
		}

		// Проверяем текущий статус задачи
		if task.Status != models.StatusWorking {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":          "task must be in working status to complete",
				"current_status": task.Status,
			})
			return
		}

		// Обновляем задачу
		task.Status = models.StatusCompleted
		task.Result = req.Description
		if req.DeleteAt != nil {
			task.DeleteAt = req.DeleteAt
		}

		if err := tx.Save(&task).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to update task: " + err.Error(),
			})
			return
		}

		// Рекурсивно отменяем все активные подзадачи этой задачи
		if err := cancelSubtasksRecursive(tx, task.ID); err != nil {
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
				if err := tx.Model(&models.Task{}).
					Where("id = ?", task.ParentTaskID).
					Update("status", models.StatusSubmitted).Error; err != nil {
					tx.Rollback()
					c.JSON(http.StatusInternalServerError, gin.H{
						"error": "failed to update parent task status: " + err.Error(),
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
