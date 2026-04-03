package worker

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func setupTestDB(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	db, err := gorm.Open(sqlite.Open(dbPath+"?_busy_timeout=30000&_journal_mode=DELETE"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	err = db.AutoMigrate(&models.Task{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	// Close DB connection at cleanup. The OS will clean up the temp dir.
	t.Cleanup(func() {
		if db != nil {
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		}
	})

	return db
}

func TestWorkerConfig_DefaultConfig(t *testing.T) {
	cfg := DefaultWorkerConfig()

	if cfg.PoolSize != 5 {
		t.Errorf("PoolSize: got %d, want 5", cfg.PoolSize)
	}
	if cfg.PollInterval != 3*time.Second {
		t.Errorf("PollInterval: got %v, want 3s", cfg.PollInterval)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries: got %d, want 3", cfg.MaxRetries)
	}
}

func TestPool_StartStop(t *testing.T) {
	db := setupTestDB(t)
	cfg := DefaultWorkerConfig()
	cfg.PoolSize = 2
	cfg.PollInterval = 100 * time.Millisecond

	pool := NewPool(cfg, db)

	// Start pool
	err := pool.Start()
	if err != nil {
		t.Fatalf("Start() error = %v", err)
	}

	if !pool.IsRunning() {
		t.Error("IsRunning() should return true after Start()")
	}

	// Stop pool
	pool.Stop()

	if pool.IsRunning() {
		t.Error("IsRunning() should return false after Stop()")
	}
}

func TestPool_StartTwice(t *testing.T) {
	db := setupTestDB(t)
	cfg := DefaultWorkerConfig()
	pool := NewPool(cfg, db)

	// Start twice should not panic
	pool.Start()
	pool.Start()
	pool.Stop()
}

func TestPool_RecoverRunningTasks(t *testing.T) {
	db := setupTestDB(t)
	cfg := DefaultWorkerConfig()
	cfg.PoolSize = 2
	cfg.PollInterval = 100 * time.Millisecond

	// Create a task marked as running (simulating crash)
	task := &models.Task{
		Type:    models.TaskTypeBatchTest,
		Status:  models.TaskStatusRunning,
		RunAt:   time.Now(),
	}
	db.Create(task)

	pool := NewPool(cfg, db)
	pool.Start()

	// Wait for recovery and processing
	time.Sleep(400 * time.Millisecond)

	// Verify the running task was recovered and eventually completed
	var recoveredTask models.Task
	db.First(&recoveredTask, task.ID)
	// After recovery + processing, task should be done (not still running)
	if recoveredTask.Status == models.TaskStatusRunning {
		t.Errorf("task should not be running after recovery: got %s", recoveredTask.Status)
	}

	pool.Stop()
}

func TestPool_PollsPendingTasks(t *testing.T) {
	db := setupTestDB(t)
	cfg := DefaultWorkerConfig()
	cfg.PoolSize = 1
	cfg.PollInterval = 100 * time.Millisecond

	// Create pending tasks
	for i := 0; i < 2; i++ {
		db.Create(&models.Task{
			Type:   models.TaskTypeBatchTest,
			Status: models.TaskStatusPending,
			RunAt:  time.Now().Add(-1 * time.Hour), // due in the past
		})
	}

	pool := NewPool(cfg, db)
	pool.Start()

	// Wait for polling - need to wait longer for SQLite
	time.Sleep(500 * time.Millisecond)

	// Verify at least some tasks were picked up (running or done)
	var doneCount int64
	var runningCount int64
	db.Model(&models.Task{}).Where("status = ?", models.TaskStatusDone).Count(&doneCount)
	db.Model(&models.Task{}).Where("status = ?", models.TaskStatusRunning).Count(&runningCount)

	if doneCount+runningCount == 0 {
		t.Errorf("no tasks were processed: done=%d running=%d", doneCount, runningCount)
	}

	pool.Stop()
}

func TestTaskExecutor_ExecuteBatchTest(t *testing.T) {
	executor := NewTaskExecutor()

	task := &models.Task{
		ID:      1,
		Type:    models.TaskTypeBatchTest,
		Status:  models.TaskStatusPending,
		Payload: `{"prompt_ids":[1,2,3]}`,
	}

	updateCalled := false
	updateFn := func(t *models.Task) error {
		updateCalled = true
		return nil
	}

	err := executor.Execute(task, updateFn)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	if !updateCalled {
		t.Error("updateFn was not called")
	}
	if task.Status != models.TaskStatusDone {
		t.Errorf("task status: got %s, want %s", task.Status, models.TaskStatusDone)
	}
	if task.Progress != 100 {
		t.Errorf("task progress: got %d, want 100", task.Progress)
	}
}

func TestTaskExecutor_ExecuteAllTaskTypes(t *testing.T) {
	executor := NewTaskExecutor()

	taskTypes := []string{
		models.TaskTypeBatchTest,
		models.TaskTypeABTest,
		models.TaskTypeEvalGen,
		models.TaskTypeRegression,
		models.TaskTypeMultiTurn,
	}

	for _, taskType := range taskTypes {
		t.Run(taskType, func(t *testing.T) {
			task := &models.Task{
				ID:     1,
				Type:   taskType,
				Status: models.TaskStatusPending,
			}

			updateFn := func(t *models.Task) error {
				return nil
			}

			err := executor.Execute(task, updateFn)
			if err != nil {
				t.Errorf("Execute() error = %v", err)
			}
			if task.Status != models.TaskStatusDone {
				t.Errorf("task status: got %s, want %s", task.Status, models.TaskStatusDone)
			}
			if task.Progress != 100 {
				t.Errorf("task progress: got %d, want 100", task.Progress)
			}
		})
	}
}

func TestTaskExecutor_ExecuteUnknownType(t *testing.T) {
	executor := NewTaskExecutor()

	task := &models.Task{
		ID:     1,
		Type:   "unknown_type",
		Status: models.TaskStatusPending,
	}

	updateFn := func(t *models.Task) error {
		return nil
	}

	err := executor.Execute(task, updateFn)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
	if task.Status != models.TaskStatusDone {
		t.Errorf("task status: got %s, want %s", task.Status, models.TaskStatusDone)
	}
}

func TestTaskExecutor_ExecuteWithInvalidPayload(t *testing.T) {
	executor := NewTaskExecutor()

	task := &models.Task{
		ID:      1,
		Type:    models.TaskTypeBatchTest,
		Status:  models.TaskStatusPending,
		Payload: "invalid json",
	}

	updateFn := func(t *models.Task) error {
		return nil
	}

	// Should not error even with invalid payload
	err := executor.Execute(task, updateFn)
	if err != nil {
		t.Errorf("Execute() error = %v", err)
	}
}
