package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TaskStatus представляет статус задачи
type TaskStatus string

const (
	StatusFailed        TaskStatus = "failed"
	StatusCanceled      TaskStatus = "canceled"
	StatusCompleted     TaskStatus = "completed"
	StatusSubmitted     TaskStatus = "submitted"
	StatusRejected      TaskStatus = "rejected"
	StatusWorking       TaskStatus = "working"
	StatusInputRequired TaskStatus = "input-required"
	StatusUnknown       TaskStatus = "unknown"
)

// Scan реализует интерфейс Scanner для TaskStatus
func (s *TaskStatus) Scan(value interface{}) error {
	if value == nil {
		*s = StatusUnknown
		return nil
	}

	switch v := value.(type) {
	case string:
		*s = TaskStatus(v)
		return nil
	case []byte:
		*s = TaskStatus(v)
		return nil
	default:
		return errors.New("cannot scan TaskStatus")
	}
}

// Value реализует интерфейс driver.Valuer для TaskStatus
func (s TaskStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// Task представляет модель задачи
type Task struct {
	ID           uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CreatedAt    time.Time       `json:"created_at"`
	DeleteAt     *time.Time      `gorm:"index" json:"delete_at,omitempty"` // Время, когда задачу нужно удалить из истории
	CreatedBy    string          `gorm:"not null" json:"created_by"`
	Assignee     string          `json:"assignee"`
	Description  string          `gorm:"type:text" json:"description"`
	RootTaskID   *uuid.UUID      `gorm:"type:uuid;index;constraint:OnDelete:CASCADE" json:"root_task_id,omitempty"`
	ParentTaskID *uuid.UUID      `gorm:"type:uuid;index;constraint:OnDelete:CASCADE" json:"parent_task_id,omitempty"`
	Result       string          `gorm:"type:text" json:"result"`
	Credentials  json.RawMessage `gorm:"type:jsonb" json:"credentials,omitempty"`
	Status       TaskStatus      `gorm:"type:varchar(20);not null;default:'submitted'" json:"status"`

	// Связи для каскадного удаления
	RootTask   *Task `gorm:"foreignKey:RootTaskID;constraint:OnDelete:CASCADE" json:"-"`
	ParentTask *Task `gorm:"foreignKey:ParentTaskID;constraint:OnDelete:CASCADE" json:"-"`
}

// BeforeCreate hook для генерации UUID перед созданием записи
func (t *Task) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	if t.Status == "" {
		t.Status = StatusSubmitted
	}
	return nil
}

// TableName возвращает имя таблицы для модели
func (Task) TableName() string {
	return "tasks"
}
