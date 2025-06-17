package cache

import (
	"agent-task-manager/database"
	"agent-task-manager/models"
	"log"
	"sync"
	"time"
)

// UsersCache хранилище для пользователей с активными задачами
type UsersCache struct {
	mu         sync.RWMutex
	users      map[string]struct{} // Set пользователей
	stopSync   chan struct{}       // Канал для остановки синхронизации
	syncTicker *time.Ticker        // Ticker для периодической синхронизации
}

// Global instance
var usersCache *UsersCache

// InitUsersCache инициализирует кэш пользователей
func InitUsersCache() error {
	usersCache = &UsersCache{
		users:    make(map[string]struct{}),
		stopSync: make(chan struct{}),
	}

	// Синхронизируем с базой данных при старте
	return SyncUsersWithTasks()
}

// StartPeriodicSync запускает периодическую синхронизацию кэша с БД
func StartPeriodicSync(interval time.Duration) {
	if usersCache == nil {
		log.Printf("Warning: cannot start periodic sync - cache is not initialized")
		return
	}

	usersCache.syncTicker = time.NewTicker(interval)

	go func() {
		log.Printf("Started periodic cache sync every %v", interval)

		for {
			select {
			case <-usersCache.syncTicker.C:
				log.Println("Running periodic cache sync...")
				if err := SyncUsersWithTasks(); err != nil {
					log.Printf("Error during periodic sync: %v", err)
				} else {
					log.Println("Periodic cache sync completed successfully")
				}
			case <-usersCache.stopSync:
				log.Println("Stopping periodic cache sync")
				return
			}
		}
	}()
}

// StopPeriodicSync останавливает периодическую синхронизацию
func StopPeriodicSync() {
	if usersCache != nil && usersCache.syncTicker != nil {
		usersCache.syncTicker.Stop()
		close(usersCache.stopSync)
		log.Println("Periodic cache sync stopped")
	}
}

// AddUserWithTask добавляет пользователя в кэш
func AddUserWithTask(userID string) {
	if usersCache == nil {
		log.Printf("Warning: users cache is not initialized")
		return
	}

	usersCache.mu.Lock()
	defer usersCache.mu.Unlock()

	usersCache.users[userID] = struct{}{}
	log.Printf("Added user to cache: %s", userID)
}

// RemoveUserWithTask удаляет пользователя из кэша
func RemoveUserWithTask(userID string) {
	if usersCache == nil {
		log.Printf("Warning: users cache is not initialized")
		return
	}

	usersCache.mu.Lock()
	defer usersCache.mu.Unlock()

	delete(usersCache.users, userID)
	log.Printf("Removed user from cache: %s", userID)
}

// GetUsersWithTasks возвращает список всех пользователей с активными задачами
func GetUsersWithTasks() []string {
	if usersCache == nil {
		log.Printf("Warning: users cache is not initialized")
		return []string{}
	}

	usersCache.mu.RLock()
	defer usersCache.mu.RUnlock()

	users := make([]string, 0, len(usersCache.users))
	for user := range usersCache.users {
		users = append(users, user)
	}

	return users
}

// CheckUserInCache проверяет, есть ли пользователь в кэше
func CheckUserInCache(userID string) bool {
	if usersCache == nil {
		return false
	}

	usersCache.mu.RLock()
	defer usersCache.mu.RUnlock()

	_, exists := usersCache.users[userID]
	return exists
}

// SyncUsersWithTasks синхронизирует кэш с базой данных
func SyncUsersWithTasks() error {
	db := database.GetDB()

	// Получаем всех уникальных пользователей с активными задачами
	var users []string
	if err := db.Model(&models.Task{}).
		Select("DISTINCT assignee").
		Where("status IN ?", []models.TaskStatus{
			models.StatusSubmitted,
			models.StatusWorking,
			models.StatusWaiting,
		}).
		Pluck("assignee", &users).Error; err != nil {
		return err
	}

	// Заменяем весь кэш новыми данными
	newUsers := make(map[string]struct{})
	for _, user := range users {
		newUsers[user] = struct{}{}
	}

	// Атомарно заменяем кэш
	usersCache.mu.Lock()
	usersCache.users = newUsers
	usersCache.mu.Unlock()

	if len(users) > 0 {
		log.Printf("Synced %d users with active tasks to cache", len(users))
	} else {
		log.Println("No users with active tasks found during sync")
	}

	return nil
}
