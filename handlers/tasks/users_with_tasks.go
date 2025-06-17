package tasks

import (
	"agent-task-manager/cache"
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetUsersWithTasksHandler возвращает список всех пользователей с активными задачами
func GetUsersWithTasksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем список пользователей из кэша
		users := cache.GetUsersWithTasks()

		// Возвращаем список пользователей
		c.JSON(http.StatusOK, gin.H{
			"users": users,
			"count": len(users),
		})
	}
}
