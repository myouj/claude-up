package handlers

import (
	"encoding/json"
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
	"prompt-vault/service"
)

const MaxSearchLength = 200

type PromptHandler struct {
	db              *gorm.DB
	promptService   *service.PromptService
	activityHandler *ActivityHandler
}

func NewPromptHandler(db *gorm.DB, activityHandler *ActivityHandler) *PromptHandler {
	return &PromptHandler{
		db:              db,
		promptService:   service.NewPromptService(db),
		activityHandler: activityHandler,
	}
}

func (h *PromptHandler) List(c *gin.Context) {
	var prompts []models.Prompt
	countQuery := h.db.Model(&models.Prompt{})
	query := h.db.Order("is_pinned DESC, updated_at DESC")

	if search := c.Query("search"); search != "" {
		if len(search) > MaxSearchLength {
			search = search[:MaxSearchLength]
		}
		query = query.Where("title LIKE ? OR content LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
		countQuery = countQuery.Where("title LIKE ? OR content LIKE ? OR description LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
		countQuery = countQuery.Where("category = ?", category)
	}
	if tag := c.Query("tag"); tag != "" {
		query = query.Where("tags LIKE ?", "%"+tag+"%")
		countQuery = countQuery.Where("tags LIKE ?", "%"+tag+"%")
	}
	if favorite := c.Query("favorite"); favorite == "true" {
		query = query.Where("is_favorite = ?", true)
		countQuery = countQuery.Where("is_favorite = ?", true)
	}

	offset, _, limit, _, meta := middleware.ParsePagination(c, countQuery, query)
	query.Offset(offset).Limit(limit).Find(&prompts)

	// Batch fetch version counts in a single query (eliminates N+1).
	versionCounts := make(map[uint]int)
	if len(prompts) > 0 {
		type countResult struct {
			PromptID uint
			Count    int64
		}
		var results []countResult
		promptIDs := make([]uint, len(prompts))
		for i, p := range prompts {
			promptIDs[i] = p.ID
		}
		h.db.Model(&models.PromptVersion{}).
			Select("prompt_id, COUNT(*) as count").
			Where("prompt_id IN ?", promptIDs).
			Group("prompt_id").
			Scan(&results)
		for _, r := range results {
			versionCounts[r.PromptID] = int(r.Count)
		}
	}

	var responses []models.PromptResponse
	for _, p := range prompts {
		responses = append(responses, models.PromptResponse{
			ID:           p.ID,
			Title:        p.Title,
			Content:      p.Content,
			ContentCN:    p.ContentCN,
			Description:  p.Description,
			Category:     p.Category,
			Tags:         parseTags(p.Tags),
			Variables:    parseVariables(p.Variables),
			IsFavorite:   p.IsFavorite,
			IsPinned:     p.IsPinned,
			VersionCount: versionCounts[p.ID],
			CreatedAt:    p.CreatedAt.Format("2006-01-02 15:04:05"),
			UpdatedAt:    p.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	c.JSON(http.StatusOK, models.PaginatedResponse{
		Success: true,
		Data:    responses,
		Meta:    meta,
	})
}

func (h *PromptHandler) Create(c *gin.Context) {
	var input struct {
		Title       string               `json:"title" binding:"required"`
		Content     string               `json:"content" binding:"required"`
		Description string               `json:"description"`
		Category    string               `json:"category"`
		Tags        []string             `json:"tags"`
		Variables   []models.Variable    `json:"variables"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}

	prompt := models.Prompt{
		Title:       input.Title,
		Content:     input.Content,
		Description: input.Description,
		Category:    input.Category,
		Tags:        marshalTags(input.Tags),
		Variables:   marshalVariables(input.Variables),
	}

	if err := h.db.Create(&prompt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("prompt", prompt.ID, "created", "")
	}

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
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Prompt not found"})
		return
	}

	versionCount, _ := h.promptService.CountVersions(prompt.ID)

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
			Variables:    parseVariables(prompt.Variables),
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
		Title       string               `json:"title"`
		Content     string               `json:"content"`
		Description string               `json:"description"`
		Category    string               `json:"category"`
		Tags        []string             `json:"tags"`
		Variables   []models.Variable    `json:"variables"`
		IsFavorite  *bool                `json:"is_favorite"`
		IsPinned    *bool                `json:"is_pinned"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}

	contentChanged := input.Content != "" && input.Content != prompt.Content

	// Build updates map - only non-empty fields are updated.
	updates := make(map[string]interface{})
	if input.Title != "" {
		updates["title"] = input.Title
	}
	if input.Content != "" {
		updates["content"] = input.Content
	}
	if input.Description != "" {
		updates["description"] = input.Description
	}
	if input.Category != "" {
		updates["category"] = input.Category
	}
	if input.Tags != nil {
		updates["tags"] = marshalTags(input.Tags)
	}
	if input.Variables != nil {
		updates["variables"] = marshalVariables(input.Variables)
	}
	if input.IsFavorite != nil {
		updates["is_favorite"] = *input.IsFavorite
	}
	if input.IsPinned != nil {
		updates["is_pinned"] = *input.IsPinned
	}

	if err := h.db.Model(&prompt).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("prompt", prompt.ID, "updated", "")
	}

	// Auto-create version if content changed (uses service layer).
	if contentChanged {
		h.promptService.EnsureVersion(prompt.ID, prompt.Content, c.DefaultQuery("comment", ""))
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

	if err := h.promptService.DeleteWithVersionsAndTests(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Prompt not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to delete prompt"})
		}
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("prompt", uint(id), "deleted", "")
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Prompt deleted successfully",
	})
}

func (h *PromptHandler) ToggleFavorite(c *gin.Context) {
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

	prompt.IsFavorite = !prompt.IsFavorite
	h.db.Save(&prompt)
	if h.activityHandler != nil {
		action := "unfavorited"
		if prompt.IsFavorite {
			action = "favorited"
		}
		h.activityHandler.Log("prompt", prompt.ID, action, "")
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"is_favorite": prompt.IsFavorite,
	})
}

func (h *PromptHandler) ListCategories(c *gin.Context) {
	var categories []string
	h.db.Model(&models.Prompt{}).
		Where("category != '' AND category IS NOT NULL").
		Distinct("category").
		Order("category ASC").
		Pluck("category", &categories)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"categories": categories,
	})
}

func (h *PromptHandler) Clone(c *gin.Context) {
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

	clone := models.Prompt{
		Title:       prompt.Title + " (Copy)",
		Content:      prompt.Content,
		ContentCN:   prompt.ContentCN,
		Description: prompt.Description,
		Category:    prompt.Category,
		Tags:        prompt.Tags,
		Variables:   prompt.Variables,
		IsFavorite:  false,
		IsPinned:    false,
	}

	if err := h.db.Create(&clone).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}
	if h.activityHandler != nil {
		h.activityHandler.Log("prompt", clone.ID, "cloned", fmt.Sprintf(`{"from_id": %d}`, id))
	}

	// Clone latest version
	var latestVersion models.PromptVersion
	if err := h.db.Where("prompt_id = ?", prompt.ID).Order("version DESC").First(&latestVersion).Error; err == nil {
		version := models.PromptVersion{
			PromptID: clone.ID,
			Version:  1,
			Content:  latestVersion.Content,
			Comment:  "Cloned from prompt #" + strconv.Itoa(int(prompt.ID)),
		}
		h.db.Create(&version)
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    clone,
	})
}

func (h *PromptHandler) Export(c *gin.Context) {
	var prompts []models.Prompt
	h.db.Order("updated_at DESC").Find(&prompts)

	export := models.ExportPayload{
		Version:    "1.0",
		ExportedAt: time.Now().Format("2006-01-02 15:04:05"),
		Prompts:    prompts,
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    export,
	})
}

func (h *PromptHandler) Import(c *gin.Context) {
	var payload struct {
		Prompts []models.Prompt `json:"prompts"`
	}

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "An internal error occurred"})
		return
	}

	imported := 0
	var failed []models.FailedItem
	for i, p := range payload.Prompts {
		clone := models.Prompt{
			Title:       p.Title,
			Content:     p.Content,
			ContentCN:   p.ContentCN,
			Description: p.Description,
			Category:    p.Category,
			Tags:        p.Tags,
			Variables:   p.Variables,
			IsFavorite:  false,
			IsPinned:    false,
		}
		if err := h.db.Create(&clone).Error; err != nil {
			failed = append(failed, models.FailedItem{
				Index: i,
				Title: p.Title,
				Error: err.Error(),
			})
			continue
		}
		version := models.PromptVersion{
			PromptID: clone.ID,
			Version:  1,
			Content:  clone.Content,
			Comment:  "Imported version",
		}
		h.db.Create(&version)
		if h.activityHandler != nil {
			h.activityHandler.Log("prompt", clone.ID, "imported", "")
		}
		imported++
	}

	c.JSON(http.StatusOK, models.ImportResult{
		Success:    true,
		Imported:   imported,
		Failed:     failed,
		TotalCount: len(payload.Prompts),
	})
}

// ----- Prefill -----

type PrefillRequest struct {
	Title string `json:"title" binding:"required"`
	Model string `json:"model"`
}

type PrefillResponse struct {
	Content     string   `json:"content"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
}

func (h *PromptHandler) Prefill(c *gin.Context) {
	var input PrefillRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "标题不能为空"})
		return
	}

	providerName := "minimax"
	model := input.Model
	if model == "" {
		model = "MiniMax-M2.7"
	}

	provider := getProvider(providerName)
	apiKey := os.Getenv("MINIMAX_API_KEY")

	var result PrefillResponse
	if apiKey == "" {
		result = mockPrefillResponse(input.Title)
	} else {
		systemPrompt := buildPrefillSystemPrompt(input.Title)
		messages := []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": input.Title},
		}
		response, _, err := provider.Call(messages, model)
		if err != nil {
			middleware.GetTraceLogger(c).Error("AI prefill request failed", map[string]interface{}{
				"error": err.Error(),
				"title": input.Title,
			})
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "AI 填充失败，请稍后重试"})
			return
		}
		result = parsePrefillResponse(response)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

func buildPrefillSystemPrompt(title string) string {
	return `You are an expert at creating structured prompts from a title.
Given a title: "` + title + `"
Generate a complete, well-structured prompt with:
1. **content**: The main prompt body with role definition, task, context, output format, and constraints
2. **description**: A brief 1-sentence description of this prompt's purpose (in Chinese)
3. **category**: A single category name (in Chinese, one of: 代码开发, 数据分析, 文档写作, 角色扮演, 翻译, 总结归纳, 问答助手, 其他)
4. **tags**: 2-3 short tags in Chinese, as a JSON array

Return ONLY a valid JSON object with keys: content, description, category, tags
Example output:
{"content": "## Role\nYou are...", "description": "用于代码审查的提示词", "category": "代码开发", "tags": ["代码审查", "review"]}`
}

func mockPrefillResponse(title string) PrefillResponse {
	category := "其他"
	var tags []string

	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "code") || strings.Contains(lower, "代码") || strings.Contains(lower, "review") || strings.Contains(lower, "审查") || strings.Contains(lower, "debug"):
		category = "代码开发"
		tags = []string{"代码", "开发"}
	case strings.Contains(lower, "data") || strings.Contains(lower, "数据") || strings.Contains(lower, "分析") || strings.Contains(lower, "统计"):
		category = "数据分析"
		tags = []string{"数据", "分析"}
	case strings.Contains(lower, "write") || strings.Contains(lower, "写作") || strings.Contains(lower, "文档") || strings.Contains(lower, "report"):
		category = "文档写作"
		tags = []string{"文档", "写作"}
	case strings.Contains(lower, "role") || strings.Contains(lower, "角色") || strings.Contains(lower, "扮演"):
		category = "角色扮演"
		tags = []string{"角色", "扮演"}
	case strings.Contains(lower, "translate") || strings.Contains(lower, "翻译"):
		category = "翻译"
		tags = []string{"翻译"}
	case strings.Contains(lower, "summar") || strings.Contains(lower, "总结") || strings.Contains(lower, "摘要"):
		category = "总结归纳"
		tags = []string{"总结", "归纳"}
	case strings.Contains(lower, "qa") || strings.Contains(lower, "问答") || strings.Contains(lower, "question") || strings.Contains(lower, "answer"):
		category = "问答助手"
		tags = []string{"问答"}
	default:
		tags = []string{"通用"}
	}

	content := "## Role\nYou are a helpful AI assistant" +
		"\n\n## Task\n" + title +
		"\n\n## Context\nProvide relevant background information and context for the task." +
		"\n\n## Output Format\nRespond in a clear, structured format as appropriate."

	return PrefillResponse{
		Content:     content,
		Description: "用于" + title + "的提示词",
		Category:    category,
		Tags:        tags,
	}
}

func parsePrefillResponse(raw string) PrefillResponse {
	var result PrefillResponse
	raw = strings.TrimSpace(raw)
	if strings.HasPrefix(raw, "```") {
		lines := strings.SplitN(raw, "\n", 3)
		if len(lines) >= 3 {
			raw = strings.TrimSuffix(lines[2], "```")
		}
	}
	raw = strings.TrimSpace(raw)

	if err := json.Unmarshal([]byte(raw), &result); err == nil {
		return result
	}
	return PrefillResponse{
		Content:     raw,
		Description: "",
		Category:    "其他",
		Tags:        []string{},
	}
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

func parseVariables(vars string) []models.Variable {
	if vars == "" {
		return []models.Variable{}
	}
	var result []models.Variable
	json.Unmarshal([]byte(vars), &result)
	return result
}

func marshalVariables(vars []models.Variable) string {
	if vars == nil {
		return "[]"
	}
	data, _ := json.Marshal(vars)
	return string(data)
}
