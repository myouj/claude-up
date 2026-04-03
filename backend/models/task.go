package models

import (
	"time"
)

// Task types
const (
	TaskTypeBatchTest    = "batch_test"
	TaskTypeABTest       = "ab_test"
	TaskTypeEvalGen      = "eval_gen"
	TaskTypeRegression   = "regression"
	TaskTypeMultiTurn    = "multi_turn"
)

// Task status
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusDone      = "done"
	TaskStatusFailed    = "failed"
	TaskStatusCancelled = "cancelled"
)

// Task represents an asynchronous background task.
type Task struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Type        string     `gorm:"size:50;not null" json:"type"` // batch_test | ab_test | eval_gen | regression | multi_turn
	Status      string     `gorm:"size:20;not null;default:pending" json:"status"` // pending | running | done | failed | cancelled
	Payload     string     `gorm:"type:text" json:"payload"` // JSON
	Progress    int        `gorm:"default:0" json:"progress"` // 0-100
	Result      string     `gorm:"type:text" json:"result,omitempty"` // JSON
	Error       string     `gorm:"size:500" json:"error,omitempty"`
	RetryCount  int        `gorm:"default:0" json:"retry_count"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	RunAt       time.Time  `json:"run_at"` // 计划执行时间
	StartedAt   *time.Time `json:"started_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

// TaskResponse is the API response format for a task.
type TaskResponse struct {
	ID          uint      `json:"id"`
	Type        string    `json:"type"`
	Status      string    `json:"status"`
	Payload     string    `json:"payload"`
	Progress    int       `json:"progress"`
	Result      string    `json:"result,omitempty"`
	Error       string    `json:"error,omitempty"`
	RetryCount  int       `json:"retry_count"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
	RunAt       string    `json:"run_at"`
	StartedAt   string    `json:"started_at,omitempty"`
	CompletedAt string    `json:"completed_at,omitempty"`
}

// ToResponse converts a Task to TaskResponse.
func (t *Task) ToResponse() TaskResponse {
	resp := TaskResponse{
		ID:         t.ID,
		Type:       t.Type,
		Status:     t.Status,
		Payload:    t.Payload,
		Progress:   t.Progress,
		Result:     t.Result,
		Error:      t.Error,
		RetryCount: t.RetryCount,
		CreatedAt:  t.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:  t.UpdatedAt.Format("2006-01-02 15:04:05"),
		RunAt:      t.RunAt.Format("2006-01-02 15:04:05"),
	}
	if t.StartedAt != nil {
		resp.StartedAt = t.StartedAt.Format("2006-01-02 15:04:05")
	}
	if t.CompletedAt != nil {
		resp.CompletedAt = t.CompletedAt.Format("2006-01-02 15:04:05")
	}
	return resp
}
