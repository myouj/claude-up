package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
)

type SkillHandler struct {
	db *gorm.DB
}

func NewSkillHandler(db *gorm.DB) *SkillHandler {
	return &SkillHandler{db: db}
}

func (h *SkillHandler) List(c *gin.Context) {
	var skills []models.Skill
	query := h.db.Order("source ASC, updated_at DESC")

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	if source := c.Query("source"); source != "" {
		query = query.Where("source = ?", source)
	}

	query.Find(&skills)

	var responses []models.SkillResponse
	for _, s := range skills {
		responses = append(responses, toSkillResponse(s))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
	})
}

func (h *SkillHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var skill models.Skill
	if err := h.db.First(&skill, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Skill not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    toSkillResponse(skill),
	})
}

func (h *SkillHandler) Create(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Content     string `json:"content" binding:"required"`
		ContentCN   string `json:"content_cn"`
		Category    string `json:"category"`
		Source      string `json:"source"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	source := input.Source
	if source == "" {
		source = "custom"
	}

	skill := models.Skill{
		Name:        input.Name,
		Description: input.Description,
		Content:     input.Content,
		ContentCN:   input.ContentCN,
		Category:    input.Category,
		Source:      source,
	}

	if err := h.db.Create(&skill).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    toSkillResponse(skill),
	})
}

func (h *SkillHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var skill models.Skill
	if err := h.db.First(&skill, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Skill not found"})
		return
	}

	var input struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Content     string `json:"content"`
		ContentCN   string `json:"content_cn"`
		Category    string `json:"category"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Content != "" {
		updates["content"] = input.Content
	}
	if input.ContentCN != "" {
		updates["content_cn"] = input.ContentCN
	}
	if input.Category != "" {
		updates["category"] = input.Category
	}

	if err := h.db.Model(&skill).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    toSkillResponse(skill),
	})
}

func (h *SkillHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	if err := h.db.Delete(&models.Skill{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Skill deleted successfully",
	})
}

func toSkillResponse(s models.Skill) models.SkillResponse {
	return models.SkillResponse{
		ID:          s.ID,
		Name:        s.Name,
		Description: s.Description,
		Content:     s.Content,
		ContentCN:   s.ContentCN,
		Category:    s.Category,
		Source:      s.Source,
		CreatedAt:   s.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   s.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
