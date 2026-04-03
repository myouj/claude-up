package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/middleware"
	"prompt-vault/models"
	"prompt-vault/service"
)

type ABTestHandler struct {
	db         *gorm.DB
	abTestSvc  *service.ABTestService
}

func NewABTestHandler(db *gorm.DB) *ABTestHandler {
	return &ABTestHandler{
		db:        db,
		abTestSvc: service.NewABTestService(db),
	}
}

// Create creates a new A/B test.
func (h *ABTestHandler) Create(c *gin.Context) {
	promptIDStr := c.Param("id")
	promptID, err := strconv.ParseUint(promptIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid prompt ID"})
		return
	}

	// Verify prompt exists
	var count int64
	h.db.Model(&models.Prompt{}).Where("id = ?", promptID).Count(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "prompt not found"})
		return
	}

	var body struct {
		Name   string `json:"name" binding:"required"`
		Config string `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Validate config JSON is valid
	var config models.ABTestConfig
	if err := json.Unmarshal([]byte(body.Config), &config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid config JSON"})
		return
	}

	abTest, err := h.abTestSvc.Create(uint(promptID), body.Name, body.Config)
	if err != nil {
		middleware.Error("failed to create AB test", map[string]interface{}{
			"error": err.Error(),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to create A/B test"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    abTest,
	})
}

// Get retrieves an A/B test by ID.
func (h *ABTestHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid ID"})
		return
	}

	abTest, err := h.abTestSvc.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "A/B test not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to get A/B test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    abTest,
	})
}

// List returns all A/B tests with pagination.
func (h *ABTestHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	abTests, total, err := h.abTestSvc.List(offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to list A/B tests"})
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	meta := models.PaginationMeta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, models.ABTestListResponse{
		Success: true,
		Data:    abTests,
		Meta:    meta,
	})
}

// ListByPrompt returns all A/B tests for a specific prompt.
func (h *ABTestHandler) ListByPrompt(c *gin.Context) {
	promptIDStr := c.Param("id")
	promptID, err := strconv.ParseUint(promptIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid prompt ID"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	abTests, total, err := h.abTestSvc.ListByPromptID(uint(promptID), offset, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to list A/B tests"})
		return
	}

	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	meta := models.PaginationMeta{
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}

	c.JSON(http.StatusOK, models.ABTestListResponse{
		Success: true,
		Data:    abTests,
		Meta:    meta,
	})
}

// GetResults retrieves all results for an A/B test.
func (h *ABTestHandler) GetResults(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid ID"})
		return
	}

	results, err := h.abTestSvc.GetResults(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to get results"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// GetResultsSummary retrieves summary statistics for an A/B test.
func (h *ABTestHandler) GetResultsSummary(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid ID"})
		return
	}

	summary, err := h.abTestSvc.GetResultsSummary(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to get summary"})
		return
	}

	// Also get SPRT significance
	sprt, err := h.abTestSvc.CheckSignificance(uint(id))
	if err != nil {
		// If significance check fails, just return summary
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    summary,
		})
		return
	}

	// Merge SPRT result into summary
	summary.IsSignificant = sprt.IsSignificant
	summary.Winner = sprt.Winner
	summary.PA = sprt.PA
	summary.PB = sprt.PB
	summary.LambdaA = sprt.LambdaA
	summary.LambdaB = sprt.LambdaB

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// Start starts an A/B test (changes status from pending to running).
func (h *ABTestHandler) Start(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid ID"})
		return
	}

	abTest, err := h.abTestSvc.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "A/B test not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to get A/B test"})
		return
	}

	if abTest.Status != models.ABTestStatusPending && abTest.Status != models.ABTestStatusStopped {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "test cannot be started in current state"})
		return
	}

	if err := h.abTestSvc.UpdateStatus(uint(id), models.ABTestStatusRunning); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to start test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"status": models.ABTestStatusRunning},
	})
}

// Stop stops an A/B test.
func (h *ABTestHandler) Stop(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid ID"})
		return
	}

	abTest, err := h.abTestSvc.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "A/B test not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to get A/B test"})
		return
	}

	if abTest.Status != models.ABTestStatusRunning {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "test cannot be stopped in current state"})
		return
	}

	if err := h.abTestSvc.UpdateStatus(uint(id), models.ABTestStatusStopped); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to stop test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    gin.H{"status": models.ABTestStatusStopped},
	})
}

// Delete deletes an A/B test and its results.
func (h *ABTestHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid ID"})
		return
	}

	if err := h.abTestSvc.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to delete A/B test"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// RunIteration runs a single iteration of the A/B test.
func (h *ABTestHandler) RunIteration(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid ID"})
		return
	}

	abTest, err := h.abTestSvc.GetByID(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "A/B test not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to get A/B test"})
		return
	}

	if abTest.Status != models.ABTestStatusRunning {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "test is not running"})
		return
	}

	result, err := h.abTestSvc.RunIteration(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to run iteration"})
		return
	}

	// Get updated test status
	abTest, _ = h.abTestSvc.GetByID(uint(id))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"result": result,
			"status": abTest.Status,
		},
	})
}
