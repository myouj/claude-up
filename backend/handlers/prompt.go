package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
)

type PromptHandler struct {
	db *gorm.DB
}

func NewPromptHandler(db *gorm.DB) *PromptHandler {
	return &PromptHandler{db: db}
}

func (h *PromptHandler) List(c *gin.Context) {
	var prompts []models.Prompt
	query := h.db.Order("is_pinned DESC, updated_at DESC")

	// 搜索过滤
	if search := c.Query("search"); search != "" {
		query = query.Where("title LIKE ? OR content LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// 分类过滤
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	// 标签过滤
	if tag := c.Query("tag"); tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
	}

	// 收藏过滤
	if favorite := c.Query("favorite"); favorite == "true" {
		query = query.Where("is_favorite = ?", true)
	}

	query.Find(&prompts)

	// 转换为响应格式，包含版本数
	var responses []models.PromptResponse
	for _, p := range prompts {
		var versionCount int64
		h.db.Model(&models.PromptVersion{}).Where("prompt_id = ?", p.ID).Count(&versionCount)

		tags := parseTags(p.Tags)
		responses = append(responses, models.PromptResponse{
			ID:           p.ID,
			Title:        p.Title,
			Content:      p.Content,
			ContentCN:    p.ContentCN,
			Description:  p.Description,
			Category:     p.Category,
			Tags:         tags,
			IsFavorite:   p.IsFavorite,
			IsPinned:     p.IsPinned,
			VersionCount: int(versionCount),
			CreatedAt:    p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    p.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
	})
}

func (h *PromptHandler) Create(c *gin.Context) {
	var input struct {
		Title       string   `json:"title" binding:"required"`
		Content     string   `json:"content" binding:"required"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Tags        []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	prompt := models.Prompt{
		Title:       input.Title,
		Content:     input.Content,
		Description: input.Description,
		Category:    input.Category,
		Tags:        marshalTags(input.Tags),
	}

	if err := h.db.Create(&prompt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 自动创建第一个版本
	version := models.PromptVersion{
		PromptID: prompt.ID,
		Version:  1,
		Content:  prompt.Content,
		Comment:  "Initial version",
	}
	h.db.Create(&version)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    prompt,
	})
}

func (h *PromptHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var prompt models.Prompt
	if err := h.db.First(&prompt, id).Error; err != nil {
		fmt.Printf("Error fetching prompt: %v", err)
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Prompt not found"})
		return
	}

	var versionCount int64
	h.db.Model(&models.PromptVersion{}).Where("prompt_id = ?", prompt.ID).Count(&versionCount)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": models.PromptResponse{
			ID:           prompt.ID,
			Title:        prompt.Title,
			Content:      prompt.Content,
			ContentCN:    prompt.ContentCN,
			Description:  prompt.Description,
			Category:     prompt.Category,
			Tags:         parseTags(prompt.Tags),
			IsFavorite:   prompt.IsFavorite,
			IsPinned:     prompt.IsPinned,
			VersionCount: int(versionCount),
			CreatedAt:    prompt.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    prompt.UpdatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func (h *PromptHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var prompt models.Prompt
	if err := h.db.First(&prompt, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Prompt not found"})
		return
	}

	var input struct {
		Title       string   `json:"title"`
		Content     string   `json:"content"`
		Description string   `json:"description"`
		Category    string   `json:"category"`
		Tags        []string `json:"tags"`
		IsFavorite  *bool    `json:"is_favorite"`
		IsPinned    *bool    `json:"is_pinned"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	contentChanged := input.Content != "" && input.Content != prompt.Content

	if input.Title != "" {
		prompt.Title = input.Title
	}
	if input.Content != "" {
		prompt.Content = input.Content
	}
	if input.Description != "" {
		prompt.Description = input.Description
	}
	if input.Category != "" {
		prompt.Category = input.Category
	}
	if input.Tags != nil {
		prompt.Tags = marshalTags(input.Tags)
	}
	if input.IsFavorite != nil {
		prompt.IsFavorite = *input.IsFavorite
	}
	if input.IsPinned != nil {
		prompt.IsPinned = *input.IsPinned
	}

	if err := h.db.Save(&prompt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 如果内容变更，自动创建新版本
	if contentChanged {
		var maxVersion int
		h.db.Model(&models.PromptVersion{}).Where("prompt_id = ?", prompt.ID).Select("COALESCE(MAX(version), 0)").Scan(&maxVersion)

		version := models.PromptVersion{
			PromptID: prompt.ID,
			Version:  maxVersion + 1,
			Content:  prompt.Content,
			Comment:  c.DefaultQuery("comment", ""),
		}
		h.db.Create(&version)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    prompt,
	})
}

func (h *PromptHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	// 删除关联的版本和测试记录
	h.db.Where("prompt_id = ?", id).Delete(&models.PromptVersion{})
	h.db.Where("prompt_id = ?", id).Delete(&models.TestRecord{})

	if err := h.db.Delete(&models.Prompt{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prompt deleted successfully",
	})
}

func parseTags(tags string) []string {
	if tags == "" {
		return []string{}
	}
	var result []string
	json.Unmarshal([]byte(tags), &result)
	return result
}

func marshalTags(tags []string) string {
	if tags == nil {
		return "[]"
	}
	data, _ := json.Marshal(tags)
	return string(data)
}
