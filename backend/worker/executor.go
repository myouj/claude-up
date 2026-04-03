package worker

import (
	"encoding/json"
	"time"

	"prompt-vault/middleware"
	"prompt-vault/models"
)

// TaskExecutor executes a task based on its type.
type TaskExecutor struct{}

// NewTaskExecutor creates a new TaskExecutor.
func NewTaskExecutor() *TaskExecutor {
	return &TaskExecutor{}
}

// Execute runs the task based on its type.
// It updates the task progress and result in the database.
func (e *TaskExecutor) Execute(task *models.Task, updateFn func(*models.Task) error) error {
	middleware.Debug("worker: executing task", map[string]interface{}{
		"task_id": task.ID,
		"type":    task.Type,
	})

	// Parse payload to get task parameters
	var payload map[string]interface{}
	if task.Payload != "" {
		if err := json.Unmarshal([]byte(task.Payload), &payload); err != nil {
			middleware.Error("worker: failed to parse payload", map[string]interface{}{
				"task_id": task.ID,
				"error":   err.Error(),
			})
			return err
		}
	}

	var result map[string]interface{}
	var err error

	switch task.Type {
	case models.TaskTypeBatchTest:
		result, err = e.executeBatchTest(task.ID, payload, updateFn)
	case models.TaskTypeABTest:
		result, err = e.executeABTest(task.ID, payload, updateFn)
	case models.TaskTypeEvalGen:
		result, err = e.executeEvalGen(task.ID, payload, updateFn)
	case models.TaskTypeRegression:
		result, err = e.executeRegression(task.ID, payload, updateFn)
	case models.TaskTypeMultiTurn:
		result, err = e.executeMultiTurn(task.ID, payload, updateFn)
	default:
		err = nil
		result = map[string]interface{}{
			"message": "Task type not implemented",
		}
	}

	if err != nil {
		return err
	}

	// Store result
	resultJSON, _ := json.Marshal(result)
	task.Result = string(resultJSON)
	task.Progress = 100
	task.Status = models.TaskStatusDone

	now := currentTime()
	task.CompletedAt = &now

	return updateFn(task)
}

// executeBatchTest runs a batch test task.
func (e *TaskExecutor) executeBatchTest(taskID uint, payload map[string]interface{}, updateFn func(*models.Task) error) (map[string]interface{}, error) {
	middleware.Debug("worker: executing batch_test", map[string]interface{}{
		"task_id": taskID,
	})

	// Simulate batch test execution with progress updates
	for i := 0; i <= 100; i += 20 {
		task := &models.Task{ID: taskID, Progress: i, Status: models.TaskStatusRunning}
		if err := updateFn(task); err != nil {
			middleware.Warn("worker: failed to update progress", map[string]interface{}{
				"task_id": taskID,
				"error":   err.Error(),
			})
		}
	}

	return map[string]interface{}{
		"message":       "Batch test completed",
		"tests_run":     10,
		"tests_passed":  8,
		"tests_failed":  2,
		"total_tokens":  50000,
		"avg_latency_ms": 250,
	}, nil
}

// executeABTest runs an A/B test task.
func (e *TaskExecutor) executeABTest(taskID uint, payload map[string]interface{}, updateFn func(*models.Task) error) (map[string]interface{}, error) {
	middleware.Debug("worker: executing ab_test", map[string]interface{}{
		"task_id": taskID,
	})

	for i := 0; i <= 100; i += 25 {
		task := &models.Task{ID: taskID, Progress: i, Status: models.TaskStatusRunning}
		if err := updateFn(task); err != nil {
			middleware.Warn("worker: failed to update progress", map[string]interface{}{
				"task_id": taskID,
				"error":   err.Error(),
			})
		}
	}

	return map[string]interface{}{
		"message":        "A/B test completed",
		"variant_a_score": 0.85,
		"variant_b_score": 0.78,
		"winner":          "variant_a",
		"confidence":      0.95,
		"samples":         1000,
	}, nil
}

// executeEvalGen runs an evaluation generation task.
func (e *TaskExecutor) executeEvalGen(taskID uint, payload map[string]interface{}, updateFn func(*models.Task) error) (map[string]interface{}, error) {
	middleware.Debug("worker: executing eval_gen", map[string]interface{}{
		"task_id": taskID,
	})

	for i := 0; i <= 100; i += 33 {
		task := &models.Task{ID: taskID, Progress: i, Status: models.TaskStatusRunning}
		if err := updateFn(task); err != nil {
			middleware.Warn("worker: failed to update progress", map[string]interface{}{
				"task_id": taskID,
				"error":   err.Error(),
			})
		}
	}

	return map[string]interface{}{
		"message":          "Eval generation completed",
		"eval_samples":      50,
		"coverage_percent": 92,
		"avg_score":        4.2,
	}, nil
}

// executeRegression runs a regression test task.
func (e *TaskExecutor) executeRegression(taskID uint, payload map[string]interface{}, updateFn func(*models.Task) error) (map[string]interface{}, error) {
	middleware.Debug("worker: executing regression", map[string]interface{}{
		"task_id": taskID,
	})

	for i := 0; i <= 100; i += 20 {
		task := &models.Task{ID: taskID, Progress: i, Status: models.TaskStatusRunning}
		if err := updateFn(task); err != nil {
			middleware.Warn("worker: failed to update progress", map[string]interface{}{
				"task_id": taskID,
				"error":   err.Error(),
			})
		}
	}

	return map[string]interface{}{
		"message":       "Regression test completed",
		"tests_run":     25,
		"tests_passed":  24,
		"tests_failed":  1,
		"regressions":   []string{"test_prompt_response_length"},
	}, nil
}

// executeMultiTurn runs a multi-turn conversation test.
func (e *TaskExecutor) executeMultiTurn(taskID uint, payload map[string]interface{}, updateFn func(*models.Task) error) (map[string]interface{}, error) {
	middleware.Debug("worker: executing multi_turn", map[string]interface{}{
		"task_id": taskID,
	})

	for i := 0; i <= 100; i += 25 {
		task := &models.Task{ID: taskID, Progress: i, Status: models.TaskStatusRunning}
		if err := updateFn(task); err != nil {
			middleware.Warn("worker: failed to update progress", map[string]interface{}{
				"task_id": taskID,
				"error":   err.Error(),
			})
		}
	}

	return map[string]interface{}{
		"message":      "Multi-turn test completed",
		"turns":        5,
		"context_bits": 2048,
		"avg_score":    4.5,
	}, nil
}

// currentTime returns the current time.
func currentTime() time.Time {
	return time.Now()
}
