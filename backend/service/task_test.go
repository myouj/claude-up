package service

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Task{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestTaskService_Create(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTaskService(db)

	tests := []struct {
		name    string
		req     CreateTaskRequest
		wantErr bool
	}{
		{
			name: "create batch_test task",
			req: CreateTaskRequest{
				Type: models.TaskTypeBatchTest,
				Payload: map[string]interface{}{
					"prompt_ids": []uint{1, 2, 3},
				},
			},
			wantErr: false,
		},
		{
			name: "create task with run_at",
			req: CreateTaskRequest{
				Type:  models.TaskTypeABTest,
				RunAt: func() *time.Time { t := time.Now().Add(1 * time.Hour); return &t }(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task, err := svc.Create(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if task.ID == 0 {
					t.Error("Create() returned task with zero ID")
				}
				if task.Status != models.TaskStatusPending {
					t.Errorf("Create() status = %v, want %v", task.Status, models.TaskStatusPending)
				}
				if task.Type != tt.req.Type {
					t.Errorf("Create() type = %v, want %v", task.Type, tt.req.Type)
				}
			}
		})
	}
}

func TestTaskService_GetByID(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTaskService(db)

	// Create a task first
	task, err := svc.Create(CreateTaskRequest{Type: models.TaskTypeBatchTest})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "get existing task",
			id:      task.ID,
			wantErr: false,
		},
		{
			name:    "get non-existing task",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := svc.GetByID(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("GetByID() got ID = %v, want %v", got.ID, tt.id)
			}
		})
	}
}

func TestTaskService_List(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTaskService(db)

	// Create multiple tasks
	for i := 0; i < 5; i++ {
		_, err := svc.Create(CreateTaskRequest{Type: models.TaskTypeBatchTest})
		if err != nil {
			t.Fatalf("failed to create task: %v", err)
		}
	}

	tests := []struct {
		name       string
		limit      int
		offset     int
		wantCount  int
		wantTotal  int64
	}{
		{
			name:      "list first 3 tasks",
			limit:     3,
			offset:    0,
			wantCount: 3,
			wantTotal: 5,
		},
		{
			name:      "list next 3 tasks",
			limit:     3,
			offset:    3,
			wantCount: 2,
			wantTotal: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, total, err := svc.List(tt.limit, tt.offset)
			if err != nil {
				t.Errorf("List() error = %v", err)
				return
			}
			if total != tt.wantTotal {
				t.Errorf("List() total = %v, want %v", total, tt.wantTotal)
			}
			if len(tasks) != tt.wantCount {
				t.Errorf("List() returned %v tasks, want %v", len(tasks), tt.wantCount)
			}
		})
	}
}

func TestTaskService_Cancel(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTaskService(db)

	// Create a pending task
	task, err := svc.Create(CreateTaskRequest{Type: models.TaskTypeBatchTest})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "cancel pending task",
			id:      task.ID,
			wantErr: false,
		},
		{
			name:    "cancel non-existing task",
			id:      9999,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Cancel(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Cancel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify the task is cancelled
				task, _ := svc.GetByID(tt.id)
				if task.Status != models.TaskStatusCancelled {
					t.Errorf("Cancel() task status = %v, want %v", task.Status, models.TaskStatusCancelled)
				}
			}
		})
	}
}

func TestTaskService_ListByStatus(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTaskService(db)

	// Create tasks with different statuses
	_, _ = svc.Create(CreateTaskRequest{Type: models.TaskTypeBatchTest})
	task2, _ := svc.Create(CreateTaskRequest{Type: models.TaskTypeABTest})

	// Cancel task2
	svc.Cancel(task2.ID)

	tests := []struct {
		name      string
		status    string
		wantCount int
		wantTotal int64
	}{
		{
			name:      "list pending tasks",
			status:    models.TaskStatusPending,
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:      "list cancelled tasks",
			status:    models.TaskStatusCancelled,
			wantCount: 1,
			wantTotal: 1,
		},
		{
			name:      "list running tasks",
			status:    models.TaskStatusRunning,
			wantCount: 0,
			wantTotal: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tasks, total, err := svc.ListByStatus(tt.status, 10, 0)
			if err != nil {
				t.Errorf("ListByStatus() error = %v", err)
				return
			}
			if total != tt.wantTotal {
				t.Errorf("ListByStatus() total = %v, want %v", total, tt.wantTotal)
			}
			if len(tasks) != tt.wantCount {
				t.Errorf("ListByStatus() returned %v tasks, want %v", len(tasks), tt.wantCount)
			}
		})
	}
}

func TestTaskService_Delete(t *testing.T) {
	db := setupTestDB(t)
	svc := NewTaskService(db)

	// Create a task
	task, err := svc.Create(CreateTaskRequest{Type: models.TaskTypeBatchTest})
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}

	tests := []struct {
		name    string
		id      uint
		wantErr bool
	}{
		{
			name:    "delete existing task",
			id:      task.ID,
			wantErr: false,
		},
		{
			name:    "delete non-existing task",
			id:      9999,
			wantErr: false, // GORM Delete doesn't return error for non-existing records
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := svc.Delete(tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Verify the task is deleted
				_, err := svc.GetByID(tt.id)
				if err == nil {
					t.Error("Delete() task still exists")
				}
			}
		})
	}
}
