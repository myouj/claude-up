package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
)

type TestHandler struct {
	db *gorm.DB
}

func NewTestHandler(db *gorm.DB) *TestHandler {
	return &TestHandler{db: db}
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

	// 调用 AI API
	response, tokens, err := h.callAI(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 保存测试记录
	record := models.TestRecord{
		PromptID:   uint(promptID),
		VersionID:   0,
		Model:      input.Model,
		PromptText: input.Content,
		Response:   response,
		TokensUsed: tokens,
	}
	h.db.Create(&record)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"response":     response,
			"tokens_used":  tokens,
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

	// 调用 AI 进行优化
	optimized, err := h.callAIOptimize(input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// 保存优化记录
	record := models.TestRecord{
		PromptID:   uint(promptID),
		VersionID:   0,
		Model:      "optimize",
		PromptText: input.Content,
		Response:   optimized,
		TokensUsed: 0,
	}
	h.db.Create(&record)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"optimized": optimized,
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
	h.db.Where("prompt_id = ?", promptID).Order("created_at DESC").Limit(50).Find(&records)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    records,
	})
}

func (h *TestHandler) callAI(input models.TestRequest) (string, int, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return mockAIResponse(input), 100, nil
	}

	var messages []map[string]string
	if len(input.Messages) > 0 {
		for _, m := range input.Messages {
			messages = append(messages, map[string]string{
				"role":    m.Role,
				"content": m.Content,
			})
		}
	} else {
		messages = append(messages, map[string]string{
			"role":    "user",
			"content": input.Content,
		})
	}

	reqBody := map[string]interface{}{
		"model": input.Model,
		"messages": messages,
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", 0, err
	}

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				content := msg["content"].(string)
				usage := result["usage"].(map[string]interface{})
				tokens := int(usage["total_tokens"].(float64))
				return content, tokens, nil
			}
		}
	}

	return "", 0, fmt.Errorf("unexpected response format")
}

func (h *TestHandler) callAIOptimize(input models.OptimizeRequest) (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")

	var systemPrompt, userPrompt string
	switch input.Mode {
	case "improve":
		systemPrompt = "You are an expert at optimizing prompts for large language models. Improve the given prompt to be clearer, more specific, and more effective. Return ONLY the optimized prompt without any explanation."
		userPrompt = input.Content
	case "structure":
		systemPrompt = "You are an expert at structuring prompts. Add appropriate structure to the prompt including: role definition, context, task description, output format, and constraints. Return ONLY the structured prompt without any explanation."
		userPrompt = input.Content
	case "style":
		systemPrompt = "You are an expert at adjusting prompt style. Modify the prompt's tone, length, and style as appropriate. Return ONLY the adjusted prompt without any explanation."
		userPrompt = "Original prompt:\n" + input.Content + "\n\nProvide a version with adjusted style."
	case "suggest":
		systemPrompt = "You are an expert at analyzing prompts and providing improvement suggestions. List 3-5 specific, actionable suggestions to improve the prompt. Be concise and practical."
		userPrompt = "Analyze this prompt and suggest improvements:\n\n" + input.Content
	default:
		systemPrompt = "You are an expert at optimizing prompts. Provide an improved version of the following prompt. Return ONLY the optimized prompt without any explanation."
		userPrompt = input.Content
	}

	if apiKey == "" {
		return mockOptimizeResponse(input), nil
	}

	reqBody := map[string]interface{}{
		"model": "gpt-4",
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
	}

	jsonData, _ := json.Marshal(reqBody)
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				return msg["content"].(string), nil
			}
		}
	}

	return "", fmt.Errorf("unexpected response format")
}

func mockAIResponse(input models.TestRequest) string {
	if strings.Contains(strings.ToLower(input.Content), "hello") ||
		strings.Contains(strings.ToLower(input.Content), "hi") {
		return "Hello! I'm Claude, an AI assistant. How can I help you today?"
	}
	if strings.Contains(strings.ToLower(input.Content), "write") &&
		(strings.Contains(strings.ToLower(input.Content), "code") ||
			strings.Contains(strings.ToLower(input.Content), "function")) {
		return "// Example code response\nfunction hello() {\n  console.log('Hello, World!');\n}\n\n// This is a mock response for testing purposes."
	}
	return "This is a mock AI response for testing. In production, this would be the actual AI response based on your prompt and selected model."
}

func mockOptimizeResponse(input models.OptimizeRequest) string {
	switch input.Mode {
	case "improve":
		return "## Improved Prompt\n\nYou are a helpful AI assistant with expertise in software development.\n\n**Task**: Help the user with their coding questions by providing clear, accurate, and well-structured responses.\n\n**Requirements**:\n- Be concise and to the point\n- Include code examples when relevant\n- Explain your reasoning\n\n**Output Format**: Provide your response in clear sections with headers."
	case "structure":
		return "**Role**: You are an expert software developer.\n\n**Context**: The user needs assistance with a technical problem or question.\n\n**Task**: Provide a clear, accurate, and helpful response.\n\n**Constraints**:\n- Be concise\n- Include examples when helpful\n- Focus on practical solutions\n\n**Output Format**: Markdown with code blocks if applicable."
	case "suggest":
		return "## Improvement Suggestions\n\n1. **Add role definition**: Specify who the AI should be (e.g., 'You are an expert developer...')\n\n2. **Define output format**: Specify how the response should be structured (e.g., 'Respond in markdown with headers...')\n\n3. **Add constraints**: Define any limitations or requirements (e.g., 'Keep responses under 200 words...')\n\n4. **Include examples**: Add a sample input/output pair to illustrate expected behavior"
	default:
		return "## Optimized Prompt\n\nYou are a highly capable AI assistant specialized in helping users with their tasks.\n\n**Objective**: Provide the most helpful and accurate response possible.\n\n**Approach**:\n- Understand the user's intent\n- Provide structured, clear answers\n- Include relevant examples\n- Be precise and actionable"
	}
}
