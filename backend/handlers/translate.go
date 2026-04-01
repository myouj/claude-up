package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/middleware"
	"prompt-vault/models"
)

type TranslateHandler struct {
	db *gorm.DB
}

func NewTranslateHandler(db *gorm.DB) *TranslateHandler {
	return &TranslateHandler{db: db}
}

func (h *TranslateHandler) Translate(c *gin.Context) {
	var input models.TranslateRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	sourceLang := input.SourceLang
	if sourceLang == "" {
		sourceLang = "en"
	}
	targetLang := input.TargetLang
	if targetLang == "" {
		targetLang = "zh"
	}

	result, err := h.callTranslateAPI(input.Text, sourceLang, targetLang)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Translation service failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": models.TranslateResponse{
			SourceText: input.Text,
			TargetText: result,
		},
	})
}

func (h *TranslateHandler) TranslateEntity(c *gin.Context) {
	entityType := c.Param("type")
	id := c.Param("id")

	var input struct {
		SourceLang string `json:"source_lang"`
		TargetLang string `json:"target_lang"`
	}
	c.ShouldBindJSON(&input)

	sourceLang := input.SourceLang
	if sourceLang == "" {
		sourceLang = "en"
	}
	targetLang := input.TargetLang
	if targetLang == "" {
		targetLang = "zh"
	}

	var content string
	var entityID uint

	switch entityType {
	case "prompt":
		var prompt models.Prompt
		if err := h.db.First(&prompt, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Prompt not found"})
			return
		}
		content = prompt.Content
		entityID = prompt.ID
		if sourceLang == "zh" {
			content = prompt.ContentCN
		}

	case "skill":
		var skill models.Skill
		if err := h.db.First(&skill, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Skill not found"})
			return
		}
		content = skill.Content
		entityID = skill.ID
		if sourceLang == "zh" {
			content = skill.ContentCN
		}

	case "agent":
		var agent models.Agent
		if err := h.db.First(&agent, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Agent not found"})
			return
		}
		content = agent.Content
		entityID = agent.ID
		if sourceLang == "zh" {
			content = agent.ContentCN
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid entity type"})
		return
	}

	result, err := h.callTranslateAPI(content, sourceLang, targetLang)
	if err != nil {
		middleware.GetTraceLogger(c).Error("translation service failed", map[string]interface{}{
			"error": err.Error(),
			"entity_type": entityType,
			"entity_id": entityID,
		})

		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Translation service failed"})
		return
	}

	// 保存翻译记录
	translation := models.Translation{
		EntityType: entityType,
		EntityID:   entityID,
		SourceLang: sourceLang,
		TargetLang: targetLang,
		SourceText: content,
		TargetText: result,
	}
	h.db.Create(&translation)

	// 更新实体的翻译字段
	switch entityType {
	case "prompt":
		if targetLang == "zh" {
			h.db.Model(&models.Prompt{}).Where("id = ?", entityID).Update("content_cn", result)
		} else {
			// 不支持反向翻译更新原字段
		}
	case "skill":
		if targetLang == "zh" {
			h.db.Model(&models.Skill{}).Where("id = ?", entityID).Update("content_cn", result)
		}
	case "agent":
		if targetLang == "zh" {
			h.db.Model(&models.Agent{}).Where("id = ?", entityID).Update("content_cn", result)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": models.TranslateResponse{
			SourceText: content,
			TargetText: result,
		},
	})
}

func (h *TranslateHandler) callTranslateAPI(text, sourceLang, targetLang string) (string, error) {
	providerName := strings.ToLower(os.Getenv("TRANSLATE_PROVIDER"))
	model := os.Getenv("TRANSLATE_MODEL")
	if providerName == "" {
		providerName = "openai"
	}
	if model == "" {
		model = getDefaultTranslateModel(providerName)
	}

	provider := getProvider(providerName)
	apiKey := getProviderAPIKey(providerName)

	if apiKey == "" {
		return mockTranslate(text, sourceLang, targetLang), nil
	}

	systemPrompt := fmt.Sprintf("You are a professional translator. Translate the following text from %s to %s. Only output the translation, no explanations.", sourceLang, targetLang)
	messages := []map[string]string{
		{"role": "system", "content": systemPrompt},
		{"role": "user", "content": text},
	}

	response, _, err := provider.Call(messages, model)
	if err != nil {
		return "", fmt.Errorf("translate call failed: %w", err)
	}
	return response, nil
}

func getDefaultTranslateModel(provider string) string {
	switch provider {
	case "claude":
		return "claude-3-5-sonnet-20241022"
	case "gemini":
		return "gemini-2.0-flash"
	case "minimax":
		return "MiniMax-Text-01"
	default:
		return "gpt-4o"
	}
}

func mockTranslate(text, sourceLang, targetLang string) string {
	if targetLang == "zh" {
		return "【翻译内容】" + text
	}
	// Mock 反向翻译
	return strings.TrimPrefix(text, "【翻译内容】")
}
