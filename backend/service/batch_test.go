package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func setupBatchTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Migrate schema
	err = db.AutoMigrate(&models.Prompt{}, &models.Task{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

func TestBatchService_CreateBatchTestTask(t *testing.T) {
	db := setupBatchTestDB(t)
	svc := NewBatchService(db)

	// Create a test prompt
	prompt := models.Prompt{
		Title:   "Test Prompt",
		Content: "Hello {{name}}, how are you?",
	}
	db.Create(&prompt)

	req := BatchTestRequest{
		PromptID: prompt.ID,
		Model:    "gpt-4",
		TestCases: []TestCase{
			{
				Name:  "Test 1",
				Input: map[string]string{"name": "Alice"},
			},
			{
				Name:  "Test 2",
				Input: map[string]string{"name": "Bob"},
			},
		},
	}

	task, err := svc.CreateBatchTestTask(req)
	if err != nil {
		t.Fatalf("CreateBatchTestTask failed: %v", err)
	}

	if task.ID == 0 {
		t.Error("Task ID should not be zero")
	}
	if task.Type != models.TaskTypeBatchTest {
		t.Errorf("Expected task type %s, got %s", models.TaskTypeBatchTest, task.Type)
	}
	if task.Status != models.TaskStatusPending {
		t.Errorf("Expected task status %s, got %s", models.TaskStatusPending, task.Status)
	}
}

func TestBatchService_CreateBatchTestTask_NotFound(t *testing.T) {
	db := setupBatchTestDB(t)
	svc := NewBatchService(db)

	req := BatchTestRequest{
		PromptID: 9999,
		Model:    "gpt-4",
		TestCases: []TestCase{
			{
				Name:  "Test 1",
				Input: map[string]string{"name": "Alice"},
			},
		},
	}

	_, err := svc.CreateBatchTestTask(req)
	if err == nil {
		t.Error("Expected error for non-existent prompt")
	}
	if err.Error() != "prompt not found" {
		t.Errorf("Expected 'prompt not found' error, got: %v", err)
	}
}

func TestBatchService_GetTask(t *testing.T) {
	db := setupBatchTestDB(t)
	svc := NewBatchService(db)

	// Create a task
	task := models.Task{
		Type:    models.TaskTypeBatchTest,
		Status:  models.TaskStatusPending,
		Payload: "{}",
	}
	db.Create(&task)

	// Get the task
	retrieved, err := svc.GetTask(task.ID)
	if err != nil {
		t.Fatalf("GetTask failed: %v", err)
	}

	if retrieved.ID != task.ID {
		t.Errorf("Expected task ID %d, got %d", task.ID, retrieved.ID)
	}
}

func TestBatchService_GetTask_NotFound(t *testing.T) {
	db := setupBatchTestDB(t)
	svc := NewBatchService(db)

	_, err := svc.GetTask(9999)
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
	if err.Error() != "task not found" {
		t.Errorf("Expected 'task not found' error, got: %v", err)
	}
}

func TestBatchService_RunBatchTest(t *testing.T) {
	db := setupBatchTestDB(t)
	svc := NewBatchService(db)

	// Create a test prompt
	prompt := models.Prompt{
		Title:   "Test Prompt",
		Content: "Hello {{name}}, how are you?",
	}
	db.Create(&prompt)

	req := BatchTestRequest{
		PromptID: prompt.ID,
		Model:    "gpt-4",
		TestCases: []TestCase{
			{
				Name:  "Test 1",
				Input: map[string]string{"name": "Alice"},
			},
			{
				Name:  "Test 2",
				Input: map[string]string{"name": "Bob"},
			},
		},
	}

	progressCount := 0
	result, err := svc.RunBatchTest(req, func(current, total int) {
		progressCount++
	})

	if err != nil {
		t.Fatalf("RunBatchTest failed: %v", err)
	}

	if result.TotalCases != 2 {
		t.Errorf("Expected 2 total cases, got %d", result.TotalCases)
	}
	if progressCount != 2 {
		t.Errorf("Expected 2 progress callbacks, got %d", progressCount)
	}
}

func TestParseBatchTestRequest(t *testing.T) {
	payload := `{"prompt_id":1,"model":"gpt-4","test_cases":[{"name":"Test 1","input":{"name":"Alice"}}]}`

	req, err := ParseBatchTestRequest(payload)
	if err != nil {
		t.Fatalf("ParseBatchTestRequest failed: %v", err)
	}

	if req.PromptID != 1 {
		t.Errorf("Expected prompt ID 1, got %d", req.PromptID)
	}
	if req.Model != "gpt-4" {
		t.Errorf("Expected model 'gpt-4', got '%s'", req.Model)
	}
	if len(req.TestCases) != 1 {
		t.Errorf("Expected 1 test case, got %d", len(req.TestCases))
	}
}

func TestReplaceVariables(t *testing.T) {
	content := "Hello {{name}}, your order {{order_id}} is ready."
	result := replaceVariables(content, "name", "Alice")
	expected := "Hello Alice, your order {{order_id}} is ready."
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	result = replaceVariables(result, "order_id", "12345")
	expected = "Hello Alice, your order 12345 is ready."
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		response string
		expected string
		minScore float64
	}{
		{"Hello world", "Hello world", 0.99},
		{"Hello", "Hello world", 0.4},
		{"", "", 0.8},
		{"Hi there", "Hello world", 0.0},
	}

	for _, tt := range tests {
		score := calculateScore(tt.response, tt.expected)
		if score < tt.minScore {
			t.Errorf("calculateScore(%q, %q) = %f, expected >= %f", tt.response, tt.expected, score, tt.minScore)
		}
	}
}
