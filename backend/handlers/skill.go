package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/middleware"
	"prompt-vault/models"
	"prompt-vault/service"
)

type SkillHandler struct {
	db            *gorm.DB
	skillService *service.SkillService
	activityHandler *ActivityHandler
}

func NewSkillHandler(db *gorm.DB, activityHandler *ActivityHandler) *SkillHandler {
	return &SkillHandler{
		db:              db,
		skillService:    service.NewSkillService(db),
		activityHandler: activityHandler,
	}
}

func (h *SkillHandler) List(c *gin.Context) {
	var skills []models.Skill
	countQuery := h.db.Model(&models.Skill{})
	query := h.db.Order("source ASC, updated_at DESC")

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
		countQuery = countQuery.Where("category = ?", category)
	}
	if source := c.Query("source"); source != "" {
		query = query.Where("source = ?", source)
		countQuery = countQuery.Where("source = ?", source)
	}

	offset, _, limit, _, meta := middleware.ParsePagination(c, countQuery, query)
	query.Offset(offset).Limit(limit).Find(&skills)

	var responses []models.SkillResponse
	for _, s := range skills {
		responses = append(responses, toSkillResponse(s))
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success: true,
		Data:    responses,
		Meta:    meta,
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "An internal error occurred"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("skill", skill.ID, "created", "")
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "An internal error occurred"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("skill", skill.ID, "updated", "")
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

	if err := h.skillService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Skill not found"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("skill", uint(id), "deleted", "")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Skill deleted successfully",
	})
}

func (h *SkillHandler) ListCategories(c *gin.Context) {
	var categories []string
	h.db.Model(&models.Skill{}).
		Where("category != '' AND category IS NOT NULL").
		Distinct("category").
		Order("category ASC").
		Pluck("category", &categories)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"categories": categories,
	})
}

func (h *SkillHandler) Clone(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	clone, details, err := h.skillService.CloneWithActivity(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Skill not found"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("skill", clone.ID, "cloned", details)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    toSkillResponse(*clone),
	})
}

func (h *SkillHandler) Export(c *gin.Context) {
	var skills []models.Skill
	h.db.Order("updated_at DESC").Find(&skills)

	export := models.ExportPayload{
		Version:    "1.0",
		ExportedAt: time.Now().Format("2006-01-02 15:04:05"),
		Skills:     skills,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    export,
	})
}

func (h *SkillHandler) Import(c *gin.Context) {
	var payload struct {
		Skills []models.Skill `json:"skills"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}

	imported := 0
	var failed []models.FailedItem
	for i, s := range payload.Skills {
		clone := models.Skill{
			Name:        s.Name,
			Description: s.Description,
			Content:     s.Content,
			ContentCN:   s.ContentCN,
			Category:    s.Category,
			Source:      "custom",
		}
		if err := h.db.Create(&clone).Error; err != nil {
			failed = append(failed, models.FailedItem{
				Index: i,
				Title: s.Name,
				Error: err.Error(),
			})
			continue
		}
		if h.activityHandler != nil {
			h.activityHandler.Log("skill", clone.ID, "imported", "")
		}
		imported++
	}

	c.JSON(http.StatusOK, models.ImportResult{
		Success:    true,
		Imported:   imported,
		Failed:     failed,
		TotalCount: len(payload.Skills),
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
