package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/service"
	"prompt-vault/worker"
)

// BatchHandler handles batch operations.
type BatchHandler struct {
	db         *gorm.DB
	batchSvc   *service.BatchService
	sseManager *worker.SSEManager
}

// NewBatchHandler creates a new BatchHandler.
func NewBatchHandler(db *gorm.DB, sseManager *worker.SSEManager) *BatchHandler {
	return &BatchHandler{
		db:         db,
		batchSvc:   service.NewBatchService(db),
		sseManager: sseManager,
	}
}

// CreateBatchTest creates a new batch test task.
// POST /api/batch/test
func (h *BatchHandler) CreateBatchTest(c *gin.Context) {
	var req service.BatchTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Create task
	task, err := h.batchSvc.CreateBatchTestTask(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"task_id": task.ID,
			"status":  task.Status,
		},
	})
}

// GetBatchTestResult gets the result of a batch test task.
// GET /api/batch/test/:task_id
func (h *BatchHandler) GetBatchTestResult(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid task ID",
		})
		return
	}

	task, err := h.batchSvc.GetTask(uint(taskID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"task_id":  task.ID,
			"type":     task.Type,
			"status":   task.Status,
			"progress": task.Progress,
			"result":   task.Result,
			"error":    task.Error,
		},
	})
}

// SSEBatchTest streams batch test progress via SSE.
// GET /api/batch/test/:task_id/stream
func (h *BatchHandler) SSEBatchTest(c *gin.Context) {
	taskID, err := strconv.ParseUint(c.Param("task_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid task ID",
		})
		return
	}

	// Use SSEHandler from worker package
	worker.SSEHandler(h.sseManager, uint(taskID))(c)
}

// ListBatchTests lists batch test tasks for a prompt.
// GET /api/batch/tests?prompt_id=123
func (h *BatchHandler) ListBatchTests(c *gin.Context) {
	promptIDStr := c.Query("prompt_id")
	promptID, err := strconv.ParseUint(promptIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid prompt ID",
		})
		return
	}

	tasks, err := h.batchSvc.GetTaskByPrompt(uint(promptID), "batch_test")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tasks,
	})
}

// RunBatchTestSync runs a batch test synchronously and returns results.
// POST /api/batch/test/sync
func (h *BatchHandler) RunBatchTestSync(c *gin.Context) {
	var req service.BatchTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Run batch test with progress callback
	result, err := h.batchSvc.RunBatchTest(req, func(current, total int) {
		// Send SSE progress if manager is available
		if h.sseManager != nil {
			h.sseManager.SendSSEProgress(req.PromptID, current, total, "running")
		}
	})

	if err != nil {
		if h.sseManager != nil {
			h.sseManager.SendSSEError(req.PromptID, err.Error())
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Send completion event
	if h.sseManager != nil {
		h.sseManager.SendSSEComplete(req.PromptID, result)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}
