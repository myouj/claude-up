package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
	"prompt-vault/service"
)

// ScoringHandler handles scoring-related requests
type ScoringHandler struct {
	db              *gorm.DB
	scoringService  *service.ScoringService
}

// NewScoringHandler creates a new ScoringHandler
func NewScoringHandler(db *gorm.DB) *ScoringHandler {
	return &ScoringHandler{
		db:              db,
		scoringService:  service.NewScoringService(),
	}
}

// ScorePrompt returns quality scores for a prompt
func (h *ScoringHandler) ScorePrompt(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var prompt models.Prompt
	if err := h.db.First(&prompt, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Prompt not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to retrieve prompt"})
		}
		return
	}

	result := h.scoringService.Score(&prompt)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"prompt_id":  prompt.ID,
			"clarity":    result.Clarity,
			"completeness": result.Completeness,
			"example":    result.Example,
			"role":       result.Role,
			"overall":    result.Overall,
			"breakdown":  result.Breakdown,
		},
	})
}

// ScoreBatch returns quality scores for multiple prompts
func (h *ScoringHandler) ScoreBatch(c *gin.Context) {
	var input struct {
		PromptIDs []uint `json:"prompt_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid input"})
		return
	}

	if len(input.PromptIDs) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Maximum 50 prompts per request"})
		return
	}

	results := make(map[uint]map[string]interface{})
	for _, id := range input.PromptIDs {
		var prompt models.Prompt
		if err := h.db.First(&prompt, id).Error; err == nil {
			result := h.scoringService.Score(&prompt)
			results[id] = map[string]interface{}{
				"clarity":      result.Clarity,
				"completeness": result.Completeness,
				"example":      result.Example,
				"role":         result.Role,
				"overall":      result.Overall,
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}

// GetWeights returns the default scoring weights
func (h *ScoringHandler) GetWeights(c *gin.Context) {
	weights := h.scoringService.GetDefaultWeights()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"clarity":      weights.Clarity,
			"completeness": weights.Completeness,
			"example":      weights.Example,
			"role":         weights.Role,
		},
	})
}
