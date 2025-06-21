package tasks

import (
	"agent-task-manager/cache"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetUsersWithTasksHandler возвращает список всех пользователей с активными задачами
func GetUsersWithTasksHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Получаем список пользователей из кэша
		allUsers := cache.GetUsersWithTasks()

		// Получаем параметр filter из query string
		filterParam := c.Query("filter")

		var filteredUsers []string = make([]string, 0)

		if filterParam != "" {
			// Разбиваем filter по запятым и очищаем от пробелов
			filterUsers := make(map[string]bool)
			for _, user := range strings.Split(filterParam, ",") {
				user = strings.TrimSpace(user)
				if user != "" {
					filterUsers[user] = true
				}
			}

			// Фильтруем пользователей
			for _, user := range allUsers {
				if filterUsers[user] {
					filteredUsers = append(filteredUsers, user)
				}
			}
		} else {
			// Если фильтр не задан, возвращаем всех пользователей
			filteredUsers = allUsers
		}

		// Возвращаем список пользователей
		c.JSON(http.StatusOK, gin.H{
			"users": filteredUsers,
			"count": len(filteredUsers),
		})
	}
}
