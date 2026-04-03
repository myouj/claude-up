package service

import (
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"

	"prompt-vault/models"
)

// BatchService handles batch operations for prompts.
type BatchService struct {
	db *gorm.DB
}

// BatchTestRequest represents a batch test request.
type BatchTestRequest struct {
	PromptID     uint             `json:"prompt_id" binding:"required"`
	Model        string           `json:"model" binding:"required"`
	TestCases    []TestCase       `json:"test_cases" binding:"required,min=1"`
	VariableSets []map[string]string `json:"variable_sets"`
}

// TestCase represents a single test case in a batch test.
type TestCase struct {
	Name     string            `json:"name"`
	Input    map[string]string `json:"input"`
	Expected string            `json:"expected,omitempty"`
}

// BatchTestResult represents the result of a batch test.
type BatchTestResult struct {
	CaseName   string            `json:"case_name"`
	Model      string            `json:"model"`
	Input      map[string]string `json:"input"`
	Response   string            `json:"response"`
	TokensUsed int               `json:"tokens_used"`
	LatencyMs  int64             `json:"latency_ms"`
	Score      float64           `json:"score"`
	Passed     bool              `json:"passed"`
	Error      string            `json:"error,omitempty"`
}

// BatchTestResponse represents the response for a batch test.
type BatchTestResponse struct {
	TaskID       uint               `json:"task_id"`
	PromptID     uint               `json:"prompt_id"`
	Model        string             `json:"model"`
	TotalCases   int                `json:"total_cases"`
	PassedCases  int                `json:"passed_cases"`
	FailedCases  int                `json:"failed_cases"`
	AvgScore     float64            `json:"avg_score"`
	TotalTokens  int               `json:"total_tokens"`
	AvgLatencyMs float64            `json:"avg_latency_ms"`
	Results      []BatchTestResult  `json:"results"`
}

// NewBatchService creates a new BatchService.
func NewBatchService(db *gorm.DB) *BatchService {
	return &BatchService{db: db}
}

// CreateBatchTestTask creates a new batch test task.
func (s *BatchService) CreateBatchTestTask(req BatchTestRequest) (*models.Task, error) {
	// Verify prompt exists
	var prompt models.Prompt
	if err := s.db.First(&prompt, req.PromptID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("prompt not found")
		}
		return nil, err
	}

	// Serialize request to payload
	payloadBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	task := &models.Task{
		Type:     models.TaskTypeBatchTest,
		Status:   models.TaskStatusPending,
		Payload:  string(payloadBytes),
		Progress: 0,
		RunAt:    time.Now(),
	}

	if err := s.db.Create(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

// GetTask retrieves a task by ID.
func (s *BatchService) GetTask(taskID uint) (*models.Task, error) {
	var task models.Task
	if err := s.db.First(&task, taskID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("task not found")
		}
		return nil, err
	}
	return &task, nil
}

// GetTaskByPrompt retrieves tasks for a specific prompt.
func (s *BatchService) GetTaskByPrompt(promptID uint, taskType string) ([]models.Task, error) {
	var tasks []models.Task
	query := s.db.Where("prompt_id = (SELECT id FROM prompts WHERE id = ?)", promptID)
	if taskType != "" {
		query = query.Where("type = ?", taskType)
	}
	if err := query.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// ParseBatchTestRequest parses a batch test request from a task payload.
func ParseBatchTestRequest(payload string) (*BatchTestRequest, error) {
	var req BatchTestRequest
	if err := json.Unmarshal([]byte(payload), &req); err != nil {
		return nil, err
	}
	return &req, nil
}

// RunBatchTest runs a batch test synchronously (for direct execution without task queue).
func (s *BatchService) RunBatchTest(req BatchTestRequest, progressCallback func(current, total int)) (*BatchTestResponse, error) {
	// Get prompt content
	var prompt models.Prompt
	if err := s.db.First(&prompt, req.PromptID).Error; err != nil {
		return nil, errors.New("prompt not found")
	}

	totalCases := len(req.TestCases)
	results := make([]BatchTestResult, 0, totalCases)
	totalTokens := 0
	totalLatency := int64(0)
	passedCases := 0
	totalScore := 0.0

	for i, tc := range req.TestCases {
		// Apply variables to prompt
		content := prompt.Content
		for k, v := range tc.Input {
			content = replaceVariables(content, k, v)
		}

		// Call AI provider (mock for now)
		response := mockBatchTestResponse(content, req.Model)
		tokens := len(response) / 4 // Approximate tokens
		latency := int64(100 + i*10) // Mock latency

		// Calculate score (simplified)
		score := calculateScore(response, tc.Expected)
		passed := score >= 0.7

		if passed {
			passedCases++
		}

		result := BatchTestResult{
			CaseName:   tc.Name,
			Model:      req.Model,
			Input:      tc.Input,
			Response:   response,
			TokensUsed: tokens,
			LatencyMs:  latency,
			Score:      score,
			Passed:     passed,
		}
		results = append(results, result)

		totalTokens += tokens
		totalLatency += latency
		totalScore += score

		// Report progress
		if progressCallback != nil {
			progressCallback(i+1, totalCases)
		}
	}

	avgScore := 0.0
	avgLatency := 0.0
	if totalCases > 0 {
		avgScore = totalScore / float64(totalCases)
		avgLatency = float64(totalLatency) / float64(totalCases)
	}

	return &BatchTestResponse{
		PromptID:     req.PromptID,
		Model:        req.Model,
		TotalCases:   totalCases,
		PassedCases:  passedCases,
		FailedCases:  totalCases - passedCases,
		AvgScore:     avgScore,
		TotalTokens:  totalTokens,
		AvgLatencyMs: avgLatency,
		Results:      results,
	}, nil
}

// replaceVariables replaces {{variable}} placeholders in content.
func replaceVariables(content, name, value string) string {
	placeholder := "{{" + name + "}}"
	for {
		idx := -1
		for i := 0; i <= len(content)-len(placeholder); i++ {
			if content[i:i+len(placeholder)] == placeholder {
				idx = i
				break
			}
		}
		if idx == -1 {
			break
		}
		content = content[:idx] + value + content[idx+len(placeholder):]
	}
	return content
}

// calculateScore calculates a similarity score between response and expected.
func calculateScore(response, expected string) float64 {
	if expected == "" {
		return 0.8 // Default score if no expected value
	}

	// Simple length-based score
	respLen := float64(len(response))
	expLen := float64(len(expected))

	if respLen == 0 || expLen == 0 {
		return 0.0
	}

	// Levenshtein-like similarity (simplified)
	similarity := 1.0 - (float64(absDiff(respLen, expLen)) / max(respLen, expLen))
	return similarity
}

func absDiff(a, b float64) float64 {
	if a > b {
		return a - b
	}
	return b - a
}

func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// mockBatchTestResponse returns a mock AI response for batch testing.
func mockBatchTestResponse(content, model string) string {
	return "Mock batch test response for: " + model + ". Content length: " + strconv.Itoa(len(content))
}
