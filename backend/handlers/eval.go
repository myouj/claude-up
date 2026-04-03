package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
	"prompt-vault/service"
)

// EvalHandler handles evaluation set-related requests
type EvalHandler struct {
	db           *gorm.DB
	evalService  *service.EvalService
}

// NewEvalHandler creates a new EvalHandler
func NewEvalHandler(db *gorm.DB) *EvalHandler {
	return &EvalHandler{
		db:          db,
		evalService: service.NewEvalService(db),
	}
}

// CreateEvalSet creates a new evaluation set
func (h *EvalHandler) CreateEvalSet(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	var input struct {
		Name    string               `json:"name" binding:"required"`
		Cases   []models.EvalCase    `json:"cases" binding:"required"`
		Weights *models.EvalWeights  `json:"weights"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid input"})
		return
	}

	// Validate case count
	if len(input.Cases) < 5 || len(input.Cases) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Cases must be between 5 and 20"})
		return
	}

	// Use default weights if not provided
	weights := models.DefaultEvalWeights()
	if input.Weights != nil {
		weights = *input.Weights
	}

	evalSet, err := h.evalService.CreateEvalSet(uint(promptID), input.Name, input.Cases, weights)
	if err != nil {
		if err.Error() == "prompt not found" {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Prompt not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to create eval set"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    evalSet.ToResponse(),
	})
}

// GetEvalSet retrieves an evaluation set
func (h *EvalHandler) GetEvalSet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	evalSet, err := h.evalService.GetEvalSet(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Eval set not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    evalSet.ToResponse(),
	})
}

// ListEvalSets lists all evaluation sets for a prompt
func (h *EvalHandler) ListEvalSets(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	evalSets, err := h.evalService.ListEvalSetsByPrompt(uint(promptID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to list eval sets"})
		return
	}

	var responses []models.EvalSetResponse
	for _, es := range evalSets {
		responses = append(responses, es.ToResponse())
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
	})
}

// UpdateEvalSet updates an evaluation set
func (h *EvalHandler) UpdateEvalSet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var input struct {
		Name    string              `json:"name" binding:"required"`
		Cases   []models.EvalCase   `json:"cases" binding:"required"`
		Weights *models.EvalWeights `json:"weights"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid input"})
		return
	}

	// Validate case count
	if len(input.Cases) < 5 || len(input.Cases) > 20 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Cases must be between 5 and 20"})
		return
	}

	weights := models.DefaultEvalWeights()
	if input.Weights != nil {
		weights = *input.Weights
	}

	evalSet, err := h.evalService.UpdateEvalSet(uint(id), input.Name, input.Cases, weights)
	if err != nil {
		if err.Error() == "eval set not found" {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Eval set not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to update eval set"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    evalSet.ToResponse(),
	})
}

// DeleteEvalSet deletes an evaluation set
func (h *EvalHandler) DeleteEvalSet(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	if err := h.evalService.DeleteEvalSet(uint(id)); err != nil {
		if err.Error() == "eval set not found" {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Eval set not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete eval set"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Eval set deleted successfully",
	})
}

// GenerateAutoEvalSet generates an evaluation set with auto-generated cases
func (h *EvalHandler) GenerateAutoEvalSet(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	var input struct {
		Name      string `json:"name" binding:"required"`
		CaseCount int    `json:"case_count"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid input"})
		return
	}

	// Default to 10 cases if not specified
	caseCount := input.CaseCount
	if caseCount <= 0 {
		caseCount = 10
	}

	evalSet, err := h.evalService.GenerateAutoEvalSet(uint(promptID), input.Name, caseCount)
	if err != nil {
		if err.Error() == "prompt not found" {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Prompt not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to generate eval set"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    evalSet.ToResponse(),
	})
}

// RunEval runs evaluation on a prompt
func (h *EvalHandler) RunEval(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	evalSetID, err := strconv.Atoi(c.Param("eval_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid eval set ID"})
		return
	}

	result, err := h.evalService.RunEval(uint(promptID), uint(evalSetID))
	if err != nil {
		if err.Error() == "prompt not found" || err.Error() == "eval set not found" {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to run evaluation"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"prompt_id":   promptID,
			"eval_set_id": evalSetID,
			"clarity":     result.Clarity,
			"completeness": result.Completeness,
			"example":     result.Example,
			"role":        result.Role,
			"overall":     result.Overall,
			"breakdown":   result.Breakdown,
		},
	})
}
