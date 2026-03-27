package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
)

type AgentHandler struct {
	db *gorm.DB
}

func NewAgentHandler(db *gorm.DB) *AgentHandler {
	return &AgentHandler{db: db}
}

func (h *AgentHandler) List(c *gin.Context) {
	var agents []models.Agent
	query := h.db.Order("source ASC, updated_at DESC")

	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}

	if source := c.Query("source"); source != "" {
		query = query.Where("source = ?", source)
	}

	query.Find(&agents)

	var responses []models.AgentResponse
	for _, a := range agents {
		responses = append(responses, toAgentResponse(a))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    responses,
	})
}

func (h *AgentHandler) Get(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var agent models.Agent
	if err := h.db.First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Agent not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    toAgentResponse(agent),
	})
}

func (h *AgentHandler) Create(c *gin.Context) {
	var input struct {
		Name         string `json:"name" binding:"required"`
		Role         string `json:"role"`
		Content      string `json:"content" binding:"required"`
		ContentCN    string `json:"content_cn"`
		Capabilities string `json:"capabilities"`
		Category     string `json:"category"`
		Source       string `json:"source"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	source := input.Source
	if source == "" {
		source = "custom"
	}

	agent := models.Agent{
		Name:         input.Name,
		Role:         input.Role,
		Content:      input.Content,
		ContentCN:    input.ContentCN,
		Capabilities: input.Capabilities,
		Category:     input.Category,
		Source:       source,
	}

	if err := h.db.Create(&agent).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    toAgentResponse(agent),
	})
}

func (h *AgentHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	var agent models.Agent
	if err := h.db.First(&agent, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Agent not found"})
		return
	}

	var input struct {
		Name         string `json:"name"`
		Role         string `json:"role"`
		Content      string `json:"content"`
		ContentCN    string `json:"content_cn"`
		Capabilities string `json:"capabilities"`
		Category     string `json:"category"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	updates := make(map[string]interface{})
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Role != "" {
		updates["role"] = input.Role
	}
	if input.Content != "" {
		updates["content"] = input.Content
	}
	if input.ContentCN != "" {
		updates["content_cn"] = input.ContentCN
	}
	if input.Capabilities != "" {
		updates["capabilities"] = input.Capabilities
	}
	if input.Category != "" {
		updates["category"] = input.Category
	}

	if err := h.db.Model(&agent).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    toAgentResponse(agent),
	})
}

func (h *AgentHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	if err := h.db.Delete(&models.Agent{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Agent deleted successfully",
	})
}

func toAgentResponse(a models.Agent) models.AgentResponse {
	return models.AgentResponse{
		ID:           a.ID,
		Name:         a.Name,
		Role:         a.Role,
		Content:      a.Content,
		ContentCN:    a.ContentCN,
		Capabilities: a.Capabilities,
		Category:     a.Category,
		Source:       a.Source,
		CreatedAt:    a.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:    a.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}
