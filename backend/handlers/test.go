package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/middleware"
	"prompt-vault/models"
)

type TestHandler struct {
	db              *gorm.DB
	activityHandler *ActivityHandler
}

func NewTestHandler(db *gorm.DB, activityHandler *ActivityHandler) *TestHandler {
	return &TestHandler{db: db, activityHandler: activityHandler}
}

func (h *TestHandler) Test(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	var input models.TestRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	providerName := strings.ToLower(input.Provider)
	if providerName == "" {
		providerName = "openai"
	}

	provider := getProvider(providerName)
	apiKey := getProviderAPIKey(providerName)

	var response string
	var tokens int

	if apiKey == "" {
		response = mockAIResponse(input.Content)
		tokens = 100
	} else {
		var messages []map[string]string
		for _, m := range input.Messages {
			messages = append(messages, map[string]string{
				"role":    m.Role,
				"content": m.Content,
			})
		}
		if len(messages) == 0 {
			messages = append(messages, map[string]string{
				"role":    "user",
				"content": input.Content,
			})
		}

		response, tokens, err = provider.Call(messages, input.Model)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "AI provider request failed"})
			return
		}
	}

	record := models.TestRecord{
		PromptID:   uint(promptID),
		VersionID:  getLatestVersionID(h.db, uint(promptID)),
		Model:      input.Model,
		Provider:   provider.Name(),
		PromptText: input.Content,
		Response:   response,
		TokensUsed: tokens,
		LatencyMs:  0,
	}
	h.db.Create(&record)
	if h.activityHandler != nil {
		h.activityHandler.Log("test", record.ID, "tested", fmt.Sprintf(`{"prompt_id": %d, "model": "%s", "provider": "%s"}`, promptID, input.Model, provider.Name()))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"response":       response,
			"tokens_used":    tokens,
			"provider":       provider.Name(),
			"test_record_id": record.ID,
		},
	})
}

func (h *TestHandler) Optimize(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	var input models.OptimizeRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	providerName := strings.ToLower(input.Provider)
	if providerName == "" {
		providerName = "openai"
	}
	if input.Model == "" {
		input.Model = getDefaultModel(providerName)
	}

	provider := getProvider(providerName)
	apiKey := getProviderAPIKey(providerName)

	var optimized string
	if apiKey == "" {
		optimized = mockOptimizeResponse(input.Mode)
	} else {
		systemPrompt := buildOptimizeSystemPrompt(input.Mode)
		messages := []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": input.Content},
		}
		optimized, _, err = provider.Call(messages, input.Model)
		if err != nil {
			middleware.GetTraceLogger(c).Error("AI provider request failed", map[string]interface{}{
				"error": err.Error(),
				"prompt_id": promptID,
				"provider": providerName,
				"model": input.Model,
			})
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "AI provider request failed"})
			return
		}
	}

	record := models.TestRecord{
		PromptID:   uint(promptID),
		VersionID:  getLatestVersionID(h.db, uint(promptID)),
		Model:      input.Model,
		Provider:   provider.Name(),
		PromptText: input.Content,
		Response:   optimized,
		TokensUsed: 0,
	}
	h.db.Create(&record)
	if h.activityHandler != nil {
		h.activityHandler.Log("test", record.ID, "optimized", fmt.Sprintf(`{"prompt_id": %d, "mode": "%s", "provider": "%s"}`, promptID, input.Mode, provider.Name()))
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"optimized": optimized,
			"provider":  provider.Name(),
		},
	})
}

func (h *TestHandler) List(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	var records []models.TestRecord
	countQuery := h.db.Model(&models.TestRecord{}).Where("prompt_id = ?", promptID)
	query := h.db.Where("prompt_id = ?", promptID).Order("created_at DESC")

	offset, _, limit, _, meta := middleware.ParsePagination(c, countQuery, query)
	query.Offset(offset).Limit(limit).Find(&records)

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success: true,
		Data:    records,
		Meta:    meta,
	})
}

func (h *TestHandler) Compare(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	// Get all versions for this prompt
	var versions []models.PromptVersion
	h.db.Where("prompt_id = ?", promptID).Order("version DESC").Find(&versions)
	versionMap := make(map[uint]models.PromptVersion)
	for _, v := range versions {
		versionMap[v.ID] = v
	}

	// Get all test records for this prompt
	var records []models.TestRecord
	h.db.Where("prompt_id = ?", promptID).Order("created_at DESC").Find(&records)

	// Group records by version_id
	grouped := make(map[uint][]models.TestRecord)
	for _, r := range records {
		vid := r.VersionID
		if vid == 0 {
			// Unversioned tests go under the latest version
			vid = getLatestVersionID(h.db, uint(promptID))
		}
		grouped[vid] = append(grouped[vid], r)
	}

	// Build response: one entry per version with its tests
	var result []map[string]interface{}
	for _, v := range versions {
		tests := grouped[v.ID]
		var testSummaries []map[string]interface{}
		for _, t := range tests {
			testSummaries = append(testSummaries, map[string]interface{}{
				"id":          t.ID,
				"model":       t.Model,
				"provider":    t.Provider,
				"response":    t.Response,
				"tokens_used": t.TokensUsed,
				"created_at":  t.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		result = append(result, map[string]interface{}{
			"version_id": v.ID,
			"version":    v.Version,
			"content":    v.Content,
			"comment":    v.Comment,
			"tests":      testSummaries,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func (h *TestHandler) Analytics(c *gin.Context) {
	promptID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid prompt ID"})
		return
	}

	daysStr := c.DefaultQuery("days", "30")
	days, _ := strconv.Atoi(daysStr)
	if days <= 0 {
		days = 30
	}
	if days > 365 {
		days = 365
	}

	since := time.Now().AddDate(0, 0, -days)

	var records []models.TestRecord
	h.db.Where("prompt_id = ? AND created_at >= ?", promptID, since).
		Order("created_at ASC").
		Find(&records)

	if len(records) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": models.PromptAnalytics{
				TotalTests:  0,
				AvgTokens:   0,
				AvgLatency:  0,
				SuccessRate: 0,
				ByModel:     map[string]int{},
				ByDate:      []models.DailyStats{},
			},
		})
		return
	}

	totalTests := int64(len(records))
	var totalTokens, totalLatency int64
	byModel := make(map[string]int)

	for _, r := range records {
		totalTokens += int64(r.TokensUsed)
		totalLatency += r.LatencyMs
		byModel[r.Model]++
	}

	avgTokens := float64(totalTokens) / float64(totalTests)
	avgLatency := float64(totalLatency) / float64(totalTests)
	successRate := 1.0 // all records are "successful" since they returned a response

	// Group by date
	dailyMap := make(map[string]struct {
		count       int
		totalTokens int64
	})
	for _, r := range records {
		date := r.CreatedAt.Format("2006-01-02")
		entry := dailyMap[date]
		entry.count++
		entry.totalTokens += int64(r.TokensUsed)
		dailyMap[date] = entry
	}

	var byDate []models.DailyStats
	for date, entry := range dailyMap {
		avgT := float64(entry.totalTokens) / float64(entry.count)
		byDate = append(byDate, models.DailyStats{
			Date:      date,
			Count:     entry.count,
			AvgTokens: avgT,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": models.PromptAnalytics{
			TotalTests:  totalTests,
			AvgTokens:   avgTokens,
			AvgLatency:  avgLatency,
			SuccessRate: successRate,
			ByModel:     byModel,
			ByDate:      byDate,
		},
	})
}

func (h *TestHandler) ListModels(c *gin.Context) {
	provider := c.DefaultQuery("provider", "")

	if provider != "" {
		models := getModelsByProvider(strings.ToLower(provider))
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"provider": provider,
				"models":   models,
			},
		})
		return
	}

	// Return all models grouped by provider
	result := make(map[string][]string)
	for _, m := range availableModels {
		result[m.Provider] = append(result[m.Provider], m.Model)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// ----- Helpers -----

func getProviderAPIKey(provider string) string {
	switch provider {
	case "claude", "anthropic":
		return os.Getenv("ANTHROPIC_API_KEY")
	case "gemini", "google", "googleai":
		return os.Getenv("GEMINI_API_KEY")
	case "minimax":
		return os.Getenv("MINIMAX_API_KEY")
	default:
		return os.Getenv("OPENAI_API_KEY")
	}
}

func getDefaultModel(provider string) string {
	switch provider {
	case "claude":
		return os.Getenv("CLAUDE_DEFAULT_MODEL")
	case "gemini":
		return os.Getenv("GEMINI_DEFAULT_MODEL")
	case "minimax":
		return os.Getenv("MINIMAX_DEFAULT_MODEL")
	default:
		return os.Getenv("OPENAI_DEFAULT_MODEL")
	}
}

func getLatestVersionID(db *gorm.DB, promptID uint) uint {
	var v models.PromptVersion
	if err := db.Where("prompt_id = ?", promptID).Order("version DESC").First(&v).Error; err == nil {
		return v.ID
	}
	return 0
}

func buildOptimizeSystemPrompt(mode string) string {
	switch mode {
	case "improve":
		return "You are an expert at optimizing prompts for large language models. Improve the given prompt to be clearer, more specific, and more effective. Return ONLY the optimized prompt without any explanation."
	case "structure":
		return "You are an expert at structuring prompts. Add appropriate structure to the prompt including: role definition, context, task description, output format, and constraints. Return ONLY the structured prompt without any explanation."
	case "style":
		return "You are an expert at adjusting prompt style. Modify the prompt's tone, length, and style as appropriate. Return ONLY the adjusted prompt without any explanation."
	case "suggest":
		return "You are an expert at analyzing prompts and providing improvement suggestions. List 3-5 specific, actionable suggestions to improve the prompt. Be concise and practical."
	default:
		return "You are an expert at optimizing prompts. Provide an improved version of the following prompt. Return ONLY the optimized prompt without any explanation."
	}
}
