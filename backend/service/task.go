package service

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"

	"prompt-vault/models"
)

// TaskService handles business logic for Task entities.
type TaskService struct {
	db *gorm.DB
}

// NewTaskService creates a new TaskService.
func NewTaskService(db *gorm.DB) *TaskService {
	return &TaskService{db: db}
}

// CreateTaskRequest represents the request to create a task.
type CreateTaskRequest struct {
	Type    string                 `json:"type" binding:"required"`
	Payload map[string]interface{} `json:"payload"`
	RunAt   *time.Time             `json:"run_at"` // Optional scheduled time
}

// Create creates a new task.
func (s *TaskService) Create(req CreateTaskRequest) (*models.Task, error) {
	payloadJSON := ""
	if req.Payload != nil {
		// Serialize payload to JSON
		data, err := json.Marshal(req.Payload)
		if err != nil {
			return nil, err
		}
		payloadJSON = string(data)
	}

	runAt := time.Now()
	if req.RunAt != nil {
		runAt = *req.RunAt
	}

	task := &models.Task{
		Type:       req.Type,
		Status:     models.TaskStatusPending,
		Payload:    payloadJSON,
		Progress:   0,
		RetryCount: 0,
		RunAt:      runAt,
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

// GetByID retrieves a task by its ID.
func (s *TaskService) GetByID(id uint) (*models.Task, error) {
	var task models.Task
	if err := s.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// List returns a list of tasks with pagination.
func (s *TaskService) List(limit, offset int) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	// Count total
	if err := s.db.Model(&models.Task{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch paginated results
	if err := s.db.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// ListByStatus returns tasks filtered by status.
func (s *TaskService) ListByStatus(status string, limit, offset int) ([]models.Task, int64, error) {
	var tasks []models.Task
	var total int64

	query := s.db.Model(&models.Task{}).Where("status = ?", status)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&tasks).Error; err != nil {
		return nil, 0, err
	}

	return tasks, total, nil
}

// Cancel cancels a task by ID.
func (s *TaskService) Cancel(id uint) error {
	result := s.db.Model(&models.Task{}).
		Where("id = ? AND status IN ?", id, []string{models.TaskStatusPending, models.TaskStatusRunning}).
		Update("status", models.TaskStatusCancelled)

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return result.Error
}

// GetProgress returns the progress of a task.
func (s *TaskService) GetProgress(id uint) (int, error) {
	var task models.Task
	if err := s.db.Select("progress").First(&task, id).Error; err != nil {
		return 0, err
	}
	return task.Progress, nil
}

// Delete deletes a task by ID.
func (s *TaskService) Delete(id uint) error {
	return s.db.Delete(&models.Task{}, id).Error
}
