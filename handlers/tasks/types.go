package tasks

import (
	"encoding/json"
	"time"

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

// CompleteTaskRequest структура для запроса завершения задачи
type CompleteTaskRequest struct {
	Description string     `json:"description" binding:"required"`
	DeleteAt    *time.Time `json:"delete_at"`
}

// FailTaskRequest структура для запроса неудачного завершения задачи
type FailTaskRequest struct {
	Reason string `json:"reason" binding:"required"`
}
