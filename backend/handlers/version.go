package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/middleware"
	"prompt-vault/models"
)

type VersionHandler struct {
	db *gorm.DB
}

func NewVersionHandler(db *gorm.DB) *VersionHandler {
	return &VersionHandler{db: db}
}

func (h *VersionHandler) List(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	var versions []models.PromptVersion
	countQuery := h.db.Model(&models.PromptVersion{}).Where("prompt_id = ?", promptID)
	query := h.db.Where("prompt_id = ?", promptID).Order("version DESC")

	offset, _, limit, _, meta := middleware.ParsePagination(c, countQuery, query)
	query.Offset(offset).Limit(limit).Find(&versions)

	var responses []models.VersionResponse
	for _, v := range versions {
		responses = append(responses, models.VersionResponse{
			ID:        v.ID,
			PromptID:  v.PromptID,
			Version:   v.Version,
			Content:   v.Content,
			Comment:   v.Comment,
			CreatedAt: v.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success: true,
		Data:    responses,
		Meta:    meta,
	})
}

func (h *VersionHandler) Create(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	// 验证 prompt 是否存在
	var prompt models.Prompt
	if err := h.db.First(&prompt, promptID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Prompt not found"})
		return
	}

	var input struct {
		Content string `json:"content" binding:"required"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	var newVersion models.PromptVersion
	err = h.db.Transaction(func(tx *gorm.DB) error {
		var maxVersion int
		tx.Model(&models.PromptVersion{}).
			Where("prompt_id = ?", promptID).
			Select("COALESCE(MAX(version), 0)").
			Scan(&maxVersion)

		newVersion = models.PromptVersion{
			PromptID: uint(promptID),
			Version:  maxVersion + 1,
			Content:  input.Content,
			Comment:  input.Comment,
		}
		if err := tx.Create(&newVersion).Error; err != nil {
			return err
		}
		return tx.Model(&prompt).Updates(map[string]interface{}{
			"content": input.Content,
		}).Error
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": models.VersionResponse{
			ID:        newVersion.ID,
			PromptID:  newVersion.PromptID,
			Version:   newVersion.Version,
			Content:   newVersion.Content,
			Comment:   newVersion.Comment,
			CreatedAt: newVersion.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}

func (h *VersionHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var version models.PromptVersion
	if err := h.db.First(&version, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Version not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": models.VersionResponse{
			ID:        version.ID,
			PromptID:  version.PromptID,
			Version:   version.Version,
			Content:   version.Content,
			Comment:   version.Comment,
			CreatedAt: version.CreatedAt.Format("2006-01-02 15:04:05"),
		},
	})
}
