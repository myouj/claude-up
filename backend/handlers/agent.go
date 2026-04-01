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

type AgentHandler struct {
	db             *gorm.DB
	agentService  *service.AgentService
	activityHandler *ActivityHandler
}

func NewAgentHandler(db *gorm.DB, activityHandler *ActivityHandler) *AgentHandler {
	return &AgentHandler{
		db:              db,
		agentService:    service.NewAgentService(db),
		activityHandler: activityHandler,
	}
}

func (h *AgentHandler) List(c *gin.Context) {
	var agents []models.Agent
	countQuery := h.db.Model(&models.Agent{})
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
	query.Offset(offset).Limit(limit).Find(&agents)

	var responses []models.AgentResponse
	for _, a := range agents {
		responses = append(responses, toAgentResponse(a))
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success: true,
		Data:    responses,
		Meta:    meta,
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "An internal error occurred"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("agent", agent.ID, "created", "")
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "An internal error occurred"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("agent", agent.ID, "updated", "")
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

	if err := h.agentService.Delete(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Agent not found"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("agent", uint(id), "deleted", "")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Agent deleted successfully",
	})
}

func (h *AgentHandler) ListCategories(c *gin.Context) {
	var categories []string
	h.db.Model(&models.Agent{}).
		Where("category != '' AND category IS NOT NULL").
		Distinct("category").
		Order("category ASC").
		Pluck("category", &categories)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"categories": categories,
	})
}

func (h *AgentHandler) Clone(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid ID"})
		return
	}

	clone, details, err := h.agentService.CloneWithActivity(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Agent not found"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("agent", clone.ID, "cloned", details)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    toAgentResponse(*clone),
	})
}

func (h *AgentHandler) Export(c *gin.Context) {
	var agents []models.Agent
	h.db.Order("updated_at DESC").Find(&agents)

	export := models.ExportPayload{
		Version:    "1.0",
		ExportedAt: time.Now().Format("2006-01-02 15:04:05"),
		Agents:     agents,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    export,
	})
}

func (h *AgentHandler) Import(c *gin.Context) {
	var payload struct {
		Agents []models.Agent `json:"agents"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}

	imported := 0
	var failed []models.FailedItem
	for i, a := range payload.Agents {
		clone := models.Agent{
			Name:         a.Name,
			Role:         a.Role,
			Content:      a.Content,
			ContentCN:    a.ContentCN,
			Capabilities: a.Capabilities,
			Category:     a.Category,
			Source:       "custom",
		}
		if err := h.db.Create(&clone).Error; err != nil {
			failed = append(failed, models.FailedItem{
				Index: i,
				Title: a.Name,
				Error: err.Error(),
			})
			continue
		}
		if h.activityHandler != nil {
			h.activityHandler.Log("agent", clone.ID, "imported", "")
		}
		imported++
	}

	c.JSON(http.StatusOK, models.ImportResult{
		Success:    true,
		Imported:   imported,
		Failed:     failed,
		TotalCount: len(payload.Agents),
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
