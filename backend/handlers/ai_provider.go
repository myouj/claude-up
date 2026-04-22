package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// AIProvider defines the interface for calling different AI providers.
type AIProvider interface {
	Name() string
	Call(messages []map[string]string, model string) (response string, tokens int, err error)
	DefaultModel() string
}

// ----- OpenAI -----

type OpenAIProvider struct {
	baseURL string
	model   string
}

func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		baseURL: os.Getenv("OPENAI_BASE_URL"),
		model:   os.Getenv("OPENAI_DEFAULT_MODEL"),
	}
}

func (p *OpenAIProvider) DefaultModel() string {
	return p.model
}

func (p *OpenAIProvider) Name() string { return "openai" }
func (p *OpenAIProvider) BaseURL() string {
	return p.baseURL
}

func (p *OpenAIProvider) Call(messages []map[string]string, model string) (string, int, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	baseURL := p.baseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1/chat/completions"
	}
	if model == "" {
		model = p.model
	}
	if model == "" {
		model = "gpt-4o"
	}

	reqBody := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("openai request marshal error: %w", err)
	}
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("openai response read error: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("openai response parse error")
	}

	if resp.StatusCode >= 400 {
		return "", 0, fmt.Errorf("openai error: %s", http.StatusText(resp.StatusCode))
	}

	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if msg, ok := choice["message"].(map[string]interface{}); ok {
				content, _ := msg["content"].(string)
				usage, _ := result["usage"].(map[string]interface{})
				var tokens int
				if usage != nil {
					if t, ok := usage["total_tokens"].(float64); ok {
						tokens = int(t)
					}
				}
				return content, tokens, nil
			}
		}
	}
	return "", 0, fmt.Errorf("unexpected openai response format")
}

// ----- Claude (Anthropic) -----

type ClaudeProvider struct {
	baseURL string
	model   string
}

func NewClaudeProvider() *ClaudeProvider {
	return &ClaudeProvider{
		baseURL: os.Getenv("ANTHROPIC_BASE_URL"),
		model:   os.Getenv("ANTHROPIC_DEFAULT_MODEL"),
	}
}

func (p *ClaudeProvider) Name() string { return "claude" }

func (p *ClaudeProvider) DefaultModel() string {
	if p.model == "" {
		return "claude-3-5-sonnet-20241022"
	}
	return p.model
}

func (p *ClaudeProvider) Call(messages []map[string]string, model string) (string, int, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	baseURL := p.baseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1/messages"
	}
	if model == "" {
		model = p.model
	}
	if model == "" {
		model = "claude-3-5-sonnet-20241022"
	}

	// Convert messages to Anthropic format
	var anthropicMessages []map[string]string
	for _, m := range messages {
		if m["role"] == "system" {
			continue // handled separately
		}
		role := "user"
		if m["role"] == "assistant" {
			role = "assistant"
		}
		anthropicMessages = append(anthropicMessages, map[string]string{
			"role":    role,
			"content": m["content"],
		})
	}

	reqBody := map[string]interface{}{
		"model":      model,
		"max_tokens": 4096,
		"messages":   anthropicMessages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("claude request marshal error: %w", err)
	}
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("anthropic-dangerous-direct-browser-access", "true")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("claude response read error: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("claude response parse error")
	}

	if resp.StatusCode >= 400 {
		return "", 0, fmt.Errorf("claude error: %s", http.StatusText(resp.StatusCode))
	}

	if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
		if block, ok := content[0].(map[string]interface{}); ok {
			if text, ok := block["text"].(string); ok {
				var tokens int
				if usage, ok := result["usage"].(map[string]interface{}); ok {
					if v, ok := usage["input_tokens"].(float64); ok {
						tokens += int(v)
					}
					if v, ok := usage["output_tokens"].(float64); ok {
						tokens += int(v)
					}
				}
				return text, tokens, nil
			}
		}
	}
	return "", 0, fmt.Errorf("unexpected claude response format")
}

// ----- Gemini (Google AI) -----

type GeminiProvider struct {
	baseURL string
	model   string
}

func NewGeminiProvider() *GeminiProvider {
	return &GeminiProvider{
		baseURL: os.Getenv("GEMINI_BASE_URL"),
		model:   os.Getenv("GEMINI_DEFAULT_MODEL"),
	}
}

func (p *GeminiProvider) Name() string { return "gemini" }

func (p *GeminiProvider) DefaultModel() string {
	if p.model == "" {
		return "gemini-2.0-flash"
	}
	return p.model
}

func (p *GeminiProvider) Call(messages []map[string]string, model string) (string, int, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	baseURL := p.baseURL
	if model == "" {
		model = p.model
	}
	if model == "" {
		model = "gemini-2.0-flash"
	}
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent", model)
	}
	url := baseURL
	if !strings.Contains(url, "?") {
		url = fmt.Sprintf("%s?key=%s", baseURL, apiKey)
	}

	// Convert to Gemini format
	var contents []map[string]interface{}
	for _, m := range messages {
		if m["role"] == "system" {
			continue
		}
		role := "user"
		if m["role"] == "assistant" {
			role = "model"
		}
		contents = append(contents, map[string]interface{}{
			"role": role,
			"parts": []map[string]string{
				{"text": m["content"]},
			},
		})
	}

	reqBody := map[string]interface{}{
		"contents": contents,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("gemini request marshal error: %w", err)
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("gemini response read error: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("gemini response parse error")
	}

	if resp.StatusCode >= 400 {
		return "", 0, fmt.Errorf("gemini error: %s", http.StatusText(resp.StatusCode))
	}

	if candidates, ok := result["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if cand, ok := candidates[0].(map[string]interface{}); ok {
			if content, ok := cand["content"].(map[string]interface{}); ok {
				if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]interface{}); ok {
						if text, ok := part["text"].(string); ok {
							// Gemini doesn't expose token counts in the same way
							return text, 0, nil
						}
					}
				}
			}
		}
	}
	return "", 0, fmt.Errorf("unexpected gemini response format")
}

// ----- MiniMax -----

type MiniMaxProvider struct {
	baseURL string
	model   string
}

type Role string

const (
	RoleUser       Role = "user"
	RoleAssistant  Role = "assistant"
	RoleSystem     Role = "system"
	RoleUserSystem Role = "user_system"
)

type MiniMaxMessage struct {
	Role    Role   `json:"role"`
	Content string `json:"content,omitempty"`
}

type MiniMaxRequest struct {
	Model               string           `json:"model"`
	MaxCompletionTokens int              `json:"max_completion_tokens,omitempty"`
	Messages            []MiniMaxMessage `json:"messages"`
}

type MiniMaxResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func NewMiniMaxProvider() *MiniMaxProvider {
	return &MiniMaxProvider{
		baseURL: os.Getenv("MINIMAX_BASE_URL"),
		model:   os.Getenv("MINIMAX_DEFAULT_MODEL"),
	}
}

func (p *MiniMaxProvider) Name() string { return "minimax" }

func (p *MiniMaxProvider) DefaultModel() string {
	if p.model == "" {
		return "MiniMax-M2.7"
	}
	return p.model
}

func (p *MiniMaxProvider) Call(messages []map[string]string, model string) (string, int, error) {
	apiKey := os.Getenv("MINIMAX_API_KEY")
	baseURL := p.baseURL
	if baseURL == "" {
		baseURL = "https://api.minimaxi.com/v1"
	}

	if model == "" {
		model = p.DefaultModel()
	}

	// Convert messages
	var minimaxMessages []MiniMaxMessage
	for _, m := range messages {
		role := "user"
		if m["role"] == "assistant" {
			role = "assistant"
		} else if m["role"] == "system" {
			role = "system"
		}
		minimaxMessages = append(minimaxMessages, MiniMaxMessage{
			Role:    Role(role),
			Content: m["content"],
		})
	}

	reqBody := MiniMaxRequest{
		Model:               model,
		MaxCompletionTokens: 4096,
		Messages:            minimaxMessages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("minimax request marshal error: %w", err)
	}
	url := baseURL + "/text/chatcompletion_v2"

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("minimax response read error: %w", err)
	}
	var result MiniMaxResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("minimax response parse error")
	}

	if resp.StatusCode >= 400 {
		return "", 0, fmt.Errorf("minimax error: %s", http.StatusText(resp.StatusCode))
	}
	choices := result.Choices
	if len(choices) > 0 {
		choice := choices[0]
		msg := choice.Message
		content := msg.Content
		return content, 0, nil
	}
	return "", 0, fmt.Errorf("unexpected minimax response format")
}

// ----- Alibaba (阿里百炼) -----

type QwenResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Usage struct {
		TotalTokens int `json:"total_tokens"`
	} `json:"usage"`
}

type AlibabaProvider struct {
	baseURL string
	model   string
}

func NewAlibabaProvider() *AlibabaProvider {
	return &AlibabaProvider{
		baseURL: os.Getenv("ALIBABA_BASE_URL"),
		model:   os.Getenv("ALIBABA_DEFAULT_MODEL"),
	}
}

func (p *AlibabaProvider) Name() string { return "alibaba" }

func (p *AlibabaProvider) DefaultModel() string {
	if p.model != "" {
		return p.model
	}
	return "qwen3.5-plus"
}

func (p *AlibabaProvider) Call(messages []map[string]string, model string) (string, int, error) {
	apiKey := os.Getenv("ALIBABA_API_KEY")
	baseURL := p.baseURL
	if baseURL == "" {
		baseURL = "https://coding.dashscope.aliyuncs.com/v1"
	}
	if model == "" {
		model = p.DefaultModel()
	}

	// Convert all messages, mapping roles correctly
	alibabaMessages := make([]map[string]interface{}, 0, len(messages))
	for _, m := range messages {
		role := "user"
		if m["role"] == "system" {
			role = "system"
		} else if m["role"] == "assistant" {
			role = "assistant"
		}
		alibabaMessages = append(alibabaMessages, map[string]interface{}{
			"role":    role,
			"content": m["content"],
		})
	}

	reqBody := map[string]interface{}{
		"model":    model,
		"messages": alibabaMessages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, fmt.Errorf("alibaba request marshal error: %w", err)
	}
	url := baseURL + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, fmt.Errorf("alibaba response read error: %w", err)
	}

	if resp.StatusCode >= 400 {
		return "", 0, fmt.Errorf("alibaba error: %s", http.StatusText(resp.StatusCode))
	}

	var result QwenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("alibaba response parse error")
	}

	choices := result.Choices
	if len(choices) > 0 {
		message := choices[0].Message
		return message.Content, result.Usage.TotalTokens, nil
	}
	return "", 0, fmt.Errorf("unexpected alibaba response format")
}

// ----- Provider Registry -----

func getProvider(name string) AIProvider {
	switch strings.ToLower(name) {
	case "openai", "gpt":
		return NewOpenAIProvider()
	case "minimax":
		return NewMiniMaxProvider()
	case "alibaba", "qwen", "dashscope":
		return NewAlibabaProvider()
	case "claude", "anthropic":
		return NewClaudeProvider()
	case "gemini", "google", "googleai":
		return NewGeminiProvider()
	default:
		return NewOpenAIProvider()
	}
}

// ----- Mock Response -----

func mockAIResponse(content string) string {
	lower := strings.ToLower(content)
	if strings.Contains(lower, "hello") || strings.Contains(lower, "hi") {
		return "Hello! I'm Claude, an AI assistant. How can I help you today?"
	}
	if strings.Contains(lower, "write") &&
		(strings.Contains(lower, "code") || strings.Contains(lower, "function")) {
		return "// Example code response\nfunction hello() {\n  console.log('Hello, World!');\n}\n\n// This is a mock response for testing purposes."
	}
	return "This is a mock AI response for testing. In production, this would be the actual AI response based on your prompt and selected model."
}

func mockOptimizeResponse(mode string) string {
	switch mode {
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

// ----- Available Models -----

type ModelInfo struct {
	Provider   string  `json:"provider"`
	Model      string  `json:"model"`
	InputCost  float64 `json:"input_cost_per_1m"` // per 1M tokens
	OutputCost float64 `json:"output_cost_per_1m"`
}

var availableModels = []ModelInfo{
	// MiniMax
	{"minimax", "MiniMax-M2.7", 0.99, 0.99},
	// Alibaba (阿里百炼)
	{"alibaba", "qwen-turbo", 0.00, 0.00},
	{"alibaba", "qwen-plus", 0.00, 0.00},
	{"alibaba", "qwen-max", 0.00, 0.00},
}

func getModelsByProvider(provider string) []string {
	var models []string
	for _, m := range availableModels {
		if m.Provider == provider {
			models = append(models, m.Model)
		}
	}
	return models
}
