package tasks

import (
	"agent-task-manager/models"
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

// TaskWithSubtasks структура для ответа с задачей и её завершенными подзадачами
type TaskWithSubtasks struct {
	models.Task
	CompletedSubtasks []models.Task `json:"completed_subtasks,omitempty"`
}

// RootTaskSummary структура для ответа со списком корневых задач с ограниченными полями
type RootTaskSummary struct {
	RootTaskID  uuid.UUID         `json:"root_task_id"`
	CreatedAt   time.Time         `json:"created_at"`
	DeleteAt    *time.Time        `json:"delete_at,omitempty"`
	Assignee    string            `json:"assignee"`
	Description string            `json:"description"`
	Status      models.TaskStatus `json:"status"`
}
