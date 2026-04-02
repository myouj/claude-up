package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func setupTestRouterWithTestHandler(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	activityHandler := NewActivityHandler(db)
	promptHandler := NewPromptHandler(db, activityHandler)
	versionHandler := NewVersionHandler(db)
	testHandler := NewTestHandler(db, activityHandler)

	api := r.Group("/api")
	api.GET("/prompts", promptHandler.List)
	api.POST("/prompts", promptHandler.Create)
	api.GET("/prompts/:id", promptHandler.Get)
	api.POST("/prompts/:id/test", testHandler.Test)
	api.POST("/prompts/:id/optimize", testHandler.Optimize)
	api.GET("/prompts/:id/tests", testHandler.List)
	api.GET("/prompts/:id/compare", testHandler.Compare)
	api.GET("/prompts/:id/analytics", testHandler.Analytics)
	api.GET("/prompts/:id/versions", versionHandler.List)
	api.POST("/prompts/:id/versions", versionHandler.Create)
	api.GET("/versions/:id", versionHandler.Get)
	api.GET("/models", testHandler.ListModels)

	return r
}

func TestTestHandler_Test(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTestHandler(db)

	prompt := models.Prompt{Title: "Test Prompt", Content: "You are a helpful assistant."}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 1, Content: prompt.Content})

	tests := []struct {
		name       string
		path       string
		body       map[string]interface{}
		wantStatus int
		wantSucc   bool
	}{
		{
			name:       "test with mock response (no API key)",
			path:       "/api/prompts/1/test",
			body:       map[string]interface{}{"content": "Hello", "model": "gpt-4o"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "test with hello triggers mock",
			path:       "/api/prompts/1/test",
			body:       map[string]interface{}{"content": "Say hello", "model": "gpt-4o", "provider": "openai"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "test with code write triggers mock",
			path:       "/api/prompts/1/test",
			body:       map[string]interface{}{"content": "Write a function", "model": "gpt-4o", "provider": "claude"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "invalid prompt ID",
			path:       "/api/prompts/abc/test",
			body:       map[string]interface{}{"content": "test", "model": "gpt-4o"},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
		{
			name:       "missing required content",
			path:       "/api/prompts/1/test",
			body:       map[string]interface{}{"model": "gpt-4o"},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
		{
			name:       "with messages array",
			path:       "/api/prompts/1/test",
			body: map[string]interface{}{
				"content": "prompt content",
				"model":   "gpt-4o",
				"messages": []map[string]string{
					{"role": "user", "content": "Hello"},
				},
			},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postJSON(router, tt.path, tt.body)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d, body: %s", w.Code, tt.wantStatus, w.Body.String())
			}

			var resp APIResponse
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}
		})
	}

	// Verify test record was created
	var count int64
	db.Model(&models.TestRecord{}).Count(&count)
	if count == 0 {
		t.Error("expected test records to be created")
	}
}

func TestTestHandler_Optimize(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTestHandler(db)

	prompt := models.Prompt{Title: "Optimize Target", Content: "Help me code."}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 1, Content: prompt.Content})

	modes := []string{"improve", "structure", "suggest", "style", "unknown"}

	for _, mode := range modes {
		t.Run("mode_"+mode, func(t *testing.T) {
			w := postJSON(router, "/api/prompts/1/optimize", map[string]interface{}{
				"content": "Help me write better code",
				"mode":    mode,
				"model":   "gpt-4o",
			})
			if w.Code != http.StatusOK {
				t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
			}

			var resp APIResponse
			json.Unmarshal(w.Body.Bytes(), &resp)
			if !resp.Success {
				t.Errorf("success: got false, want true")
			}
		})
	}

	// Test with provider
	t.Run("with provider claude", func(t *testing.T) {
		w := postJSON(router, "/api/prompts/1/optimize", map[string]interface{}{
			"content":  "test",
			"mode":     "improve",
			"provider": "claude",
		})
		if w.Code != http.StatusOK {
			t.Errorf("status: got %d", w.Code)
		}
	})

	// Test default provider and model
	t.Run("default provider and model", func(t *testing.T) {
		w := postJSON(router, "/api/prompts/1/optimize", map[string]interface{}{
			"content": "test",
			"mode":    "improve",
		})
		if w.Code != http.StatusOK {
			t.Errorf("status: got %d", w.Code)
		}
	})
}

func TestTestHandler_List(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTestHandler(db)

	prompt := models.Prompt{Title: "List Tests", Content: "Content"}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 1, Content: "v1"})

	// Create test records
	for i := 0; i < 5; i++ {
		db.Create(&models.TestRecord{PromptID: prompt.ID, Model: "gpt-4o", Provider: "openai", TokensUsed: 100})
	}

	tests := []struct {
		name      string
		path      string
		wantCount int
	}{
		{"all tests", "/api/prompts/1/tests", 5},
		{"with pagination", "/api/prompts/1/tests?page=1&limit=2", 2},
		{"page 2", "/api/prompts/1/tests?page=2&limit=2", 2},
		{"page 3", "/api/prompts/1/tests?page=3&limit=2", 1},
		{"invalid prompt ID", "/api/prompts/abc/tests", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := getJSON(router, tt.path)
			if tt.path == "/api/prompts/abc/tests" {
				if w.Code != http.StatusBadRequest {
					t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
				}
				return
			}
			if w.Code != http.StatusOK {
				t.Errorf("status: got %d", w.Code)
				return
			}

			var resp struct {
				Success bool              `json:"success"`
				Data    []json.RawMessage `json:"data"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if len(resp.Data) != tt.wantCount {
				t.Errorf("count: got %d, want %d", len(resp.Data), tt.wantCount)
			}
		})
	}
}

func TestTestHandler_Compare(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTestHandler(db)

	prompt := models.Prompt{Title: "Compare", Content: "Content"}
	db.Create(&prompt)

	// Create two versions
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 1, Content: "Version 1 content"})
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 2, Content: "Version 2 content"})

	// Create test records for each version
	db.Create(&models.TestRecord{PromptID: prompt.ID, VersionID: 1, Model: "gpt-4o", TokensUsed: 50})
	db.Create(&models.TestRecord{PromptID: prompt.ID, VersionID: 2, Model: "gpt-4o", TokensUsed: 75})

	w := getJSON(router, "/api/prompts/1/compare")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Success bool              `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success")
	}
	if len(resp.Data) != 2 {
		t.Errorf("expected 2 versions, got %d", len(resp.Data))
	}
}

func TestTestHandler_Analytics(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTestHandler(db)

	prompt := models.Prompt{Title: "Analytics", Content: "Content"}
	db.Create(&prompt)

	// Create test records
	for i := 0; i < 3; i++ {
		db.Create(&models.TestRecord{
			PromptID:  prompt.ID,
			Model:     "gpt-4o",
			Provider:  "openai",
			TokensUsed: 100,
			LatencyMs: 50,
		})
	}

	tests := []struct {
		name      string
		path      string
		wantSucc  bool
		wantTests int64
	}{
		{"default 30 days", "/api/prompts/1/analytics", true, 3},
		{"7 days", "/api/prompts/1/analytics?days=7", true, 3},
		{"365 days", "/api/prompts/1/analytics?days=365", true, 3},
		{"invalid days", "/api/prompts/1/analytics?days=abc", true, 3},
		{"negative days", "/api/prompts/1/analytics?days=-5", true, 3},
		{"zero days", "/api/prompts/1/analytics?days=0", true, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := getJSON(router, tt.path)
			if w.Code != http.StatusOK {
				t.Errorf("status: got %d", w.Code)
				return
			}

			var resp struct {
				Success    bool `json:"success"`
				Data       models.PromptAnalytics
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v", resp.Success)
			}
		})
	}
}

func TestTestHandler_ListModels(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTestHandler(db)

	tests := []struct {
		name     string
		path     string
		wantSucc bool
	}{
		{"all models", "/api/models", true},
		{"openai models", "/api/models?provider=openai", true},
		{"claude models", "/api/models?provider=claude", true},
		{"gemini models", "/api/models?provider=gemini", true},
		{"minimax models", "/api/models?provider=minimax", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := getJSON(router, tt.path)
			if w.Code != http.StatusOK {
				t.Errorf("status: got %d", w.Code)
				return
			}

			var resp APIResponse
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v", resp.Success)
			}
		})
	}
}

func TestBuildOptimizeSystemPrompt(t *testing.T) {
	modes := []struct {
		mode string
	}{
		{"improve"},
		{"structure"},
		{"style"},
		{"suggest"},
		{"unknown"},
	}

	for _, m := range modes {
		t.Run(m.mode, func(t *testing.T) {
			got := buildOptimizeSystemPrompt(m.mode)
			if got == "" {
				t.Error("expected non-empty prompt")
			}
		})
	}
}

func TestGetLatestVersionID(t *testing.T) {
	db := newTestDB(t)

	// No versions yet
	id := getLatestVersionID(db, 999)
	if id != 0 {
		t.Errorf("expected 0 for non-existent prompt, got %d", id)
	}

	prompt := models.Prompt{Title: "V", Content: "C"}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 1, Content: "v1"})
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 2, Content: "v2"})

	id = getLatestVersionID(db, prompt.ID)
	if id == 0 {
		t.Error("expected non-zero version ID")
	}
}

func TestGetProviderAPIKey(t *testing.T) {
	tests := []struct {
		provider string
	}{
		{"openai"},
		{"claude"},
		{"anthropic"},
		{"gemini"},
		{"google"},
		{"minimax"},
		{"unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := getProviderAPIKey(tt.provider)
			// Just verify it returns without panicking
			_ = got
		})
	}
}

func TestGetDefaultModel(t *testing.T) {
	tests := []struct {
		provider string
		want     string
	}{
		{"claude", "claude-3-5-sonnet-20241022"},
		{"gemini", "gemini-2.0-flash"},
		{"minimax", "MiniMax-Text-01"},
		{"openai", "gpt-4o"},
		{"unknown", "gpt-4o"},
	}

	for _, tt := range tests {
		t.Run(tt.provider, func(t *testing.T) {
			got := getDefaultModel(tt.provider)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
