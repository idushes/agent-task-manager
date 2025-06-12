package handlers

import (
	"agent-task-manager/database"
	"agent-task-manager/models"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// StatsResponse структура для ответа со статистикой
type StatsResponse struct {
	Period       string `json:"period"`
	PendingTasks int    `json:"pending_tasks"` // Задачи в ожидании (submitted)
	InProgress   int    `json:"in_progress"`   // Задачи в работе (working)
	NewTasks     int    `json:"new_tasks"`     // Новые задачи за период
	FailedTasks  int    `json:"failed_tasks"`  // Зафейленные задачи за период
}

// StatsHandler обработчик для получения статистики по задачам
func StatsHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем параметр периода из query string
		period := c.DefaultQuery("period", "all-time")

		// Проверяем валидность периода
		validPeriods := map[string]bool{
			"today":     true,
			"yesterday": true,
			"week":      true,
			"month":     true,
			"year":      true,
			"all-time":  true,
		}

		if !validPeriods[period] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid period parameter. Valid values: today, yesterday, week, month, year, all-time",
			})
			return
		}

		// Получаем user_id из контекста (установлен в JWT middleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "user_id not found in context",
			})
			return
		}

		db := database.GetDB()

		// Вычисляем временные границы для периода
		now := time.Now()
		var startTime time.Time

		switch period {
		case "today":
			startTime = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		case "yesterday":
			yesterday := now.AddDate(0, 0, -1)
			startTime = time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
			now = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		case "week":
			startTime = now.AddDate(0, 0, -7)
		case "month":
			startTime = now.AddDate(0, -1, 0)
		case "year":
			startTime = now.AddDate(-1, 0, 0)
		case "all-time":
			// Для all-time берем очень старую дату
			startTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
		}

		// Считаем задачи в ожидании (submitted) для текущего пользователя
		var pendingCount int64
		if err := db.Model(&models.Task{}).
			Where("created_by = ? AND status = ?", userID.(string), models.StatusSubmitted).
			Count(&pendingCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to count pending tasks: " + err.Error(),
			})
			return
		}

		// Считаем задачи в работе (working) для текущего пользователя
		var inProgressCount int64
		if err := db.Model(&models.Task{}).
			Where("created_by = ? AND status = ?", userID.(string), models.StatusWorking).
			Count(&inProgressCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to count in-progress tasks: " + err.Error(),
			})
			return
		}

		// Считаем новые задачи за период
		var newTasksCount int64
		query := db.Model(&models.Task{}).Where("created_by = ?", userID.(string))
		if period == "yesterday" {
			query = query.Where("created_at >= ? AND created_at < ?", startTime, now)
		} else if period != "all-time" {
			query = query.Where("created_at >= ?", startTime)
		}
		if err := query.Count(&newTasksCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to count new tasks: " + err.Error(),
			})
			return
		}

		// Считаем зафейленные задачи за период
		// Поскольку у нас нет поля updated_at, считаем failed задачи по времени создания
		// Это означает, что мы считаем задачи, которые были созданы и зафейлены в указанный период
		var failedCount int64
		failedQuery := db.Model(&models.Task{}).
			Where("created_by = ? AND status = ?", userID.(string), models.StatusFailed)

		if period == "yesterday" {
			failedQuery = failedQuery.Where("created_at >= ? AND created_at < ?", startTime, now)
		} else if period != "all-time" {
			failedQuery = failedQuery.Where("created_at >= ?", startTime)
		}

		if err := failedQuery.Count(&failedCount).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to count failed tasks: " + err.Error(),
			})
			return
		}

		// Формируем ответ
		response := StatsResponse{
			Period:       period,
			PendingTasks: int(pendingCount),
			InProgress:   int(inProgressCount),
			NewTasks:     int(newTasksCount),
			FailedTasks:  int(failedCount),
		}

		c.JSON(http.StatusOK, response)
	}
}
