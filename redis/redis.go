package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	// QueueName имя очереди для уведомлений о задачах
	QueueName = "task_notifications"
	// MaxRetries максимальное количество попыток подключения
	MaxRetries = 5
	// RetryDelay задержка между попытками подключения
	RetryDelay = 2 * time.Second
)

// TaskNotification структура уведомления о задаче
type TaskNotification struct {
	TaskID    string    `json:"task_id"`
	Assignee  string    `json:"assignee"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	client *redis.Client
	ctx    = context.Background()
)

// InitRedis инициализирует подключение к Redis
func InitRedis(redisURL string) error {
	// Если URL не указан, возвращаем ошибку
	if redisURL == "" {
		return fmt.Errorf("redis URL is empty")
	}

	// Парсим Redis URL
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return fmt.Errorf("failed to parse redis URL: %w", err)
	}

	// Создаем клиент
	client = redis.NewClient(opt)

	// Пытаемся подключиться с повторными попытками
	for i := 0; i < MaxRetries; i++ {
		if err := client.Ping(ctx).Err(); err != nil {
			log.Printf("Failed to connect to Redis (attempt %d/%d): %v", i+1, MaxRetries, err)
			if i < MaxRetries-1 {
				time.Sleep(RetryDelay)
				continue
			}
			return fmt.Errorf("failed to connect to Redis after %d attempts: %w", MaxRetries, err)
		}
		break
	}

	log.Println("Successfully connected to Redis")

	// Проверяем существование очереди
	exists, err := client.Exists(ctx, QueueName).Result()
	if err != nil {
		return fmt.Errorf("failed to check queue existence: %w", err)
	}

	// Если очереди нет, создаем её (добавляем пустое значение и сразу удаляем)
	if exists == 0 {
		// Redis списки создаются автоматически при первом добавлении элемента
		// Но мы можем добавить и удалить элемент для явного создания
		if err := client.LPush(ctx, QueueName, "init").Err(); err != nil {
			return fmt.Errorf("failed to initialize queue: %w", err)
		}
		if err := client.LPop(ctx, QueueName).Err(); err != nil {
			return fmt.Errorf("failed to initialize queue: %w", err)
		}
		log.Printf("Created Redis queue: %s", QueueName)
	} else {
		log.Printf("Redis queue already exists: %s", QueueName)
	}

	return nil
}

// SendTaskNotification отправляет уведомление о задаче в очередь
func SendTaskNotification(taskID, assignee string) error {
	if client == nil {
		return fmt.Errorf("redis client is not initialized")
	}

	notification := TaskNotification{
		TaskID:    taskID,
		Assignee:  assignee,
		Timestamp: time.Now(),
	}

	// Сериализуем уведомление в JSON
	data, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// Отправляем в очередь Redis (используем RPUSH для добавления в конец очереди)
	if err := client.RPush(ctx, QueueName, data).Err(); err != nil {
		return fmt.Errorf("failed to push notification to queue: %w", err)
	}

	log.Printf("Sent task notification to queue: taskID=%s, assignee=%s", taskID, assignee)
	return nil
}

// CloseRedis закрывает соединение с Redis
func CloseRedis() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// GetClient возвращает Redis клиент (для тестирования или других операций)
func GetClient() *redis.Client {
	return client
}
