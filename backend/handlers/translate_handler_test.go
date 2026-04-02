package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func setupTestRouterWithTranslateHandler(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	translateHandler := NewTranslateHandler(db)
	promptHandler := NewPromptHandler(db, nil)
	skillHandler := NewSkillHandler(db, nil)
	agentHandler := NewAgentHandler(db, nil)

	api := r.Group("/api")
	api.POST("/translate", translateHandler.Translate)
	api.POST("/translate/:type/:id", translateHandler.TranslateEntity)
	api.GET("/prompts", promptHandler.List)
	api.POST("/prompts", promptHandler.Create)
	api.GET("/prompts/:id", promptHandler.Get)
	api.GET("/skills", skillHandler.List)
	api.POST("/skills", skillHandler.Create)
	api.GET("/skills/:id", skillHandler.Get)
	api.GET("/agents", agentHandler.List)
	api.POST("/agents", agentHandler.Create)
	api.GET("/agents/:id", agentHandler.Get)

	return r
}

func TestTranslateHandler_Translate(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTranslateHandler(db)

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
		wantSucc   bool
	}{
		{
			name:       "translate to zh (mock)",
			body:       map[string]interface{}{"text": "Hello world", "source_lang": "en", "target_lang": "zh"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "translate to en (mock reverse)",
			body:       map[string]interface{}{"text": "【翻译内容】你好世界", "source_lang": "zh", "target_lang": "en"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "default source lang",
			body:       map[string]interface{}{"text": "test"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "empty text fails binding",
			body:       map[string]interface{}{"text": ""},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
		{
			name:       "invalid JSON",
			body:       map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postJSON(router, "/api/translate", tt.body)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}

			var resp struct {
				Success bool `json:"success"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}
		})
	}
}

func TestTranslateHandler_TranslateEntity_Prompt(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTranslateHandler(db)

	prompt := models.Prompt{Title: "Translate Test", Content: "You are a helpful assistant."}
	db.Create(&prompt)

	tests := []struct {
		name       string
		path       string
		body       map[string]interface{}
		wantStatus int
		wantSucc   bool
	}{
		{
			name:       "translate prompt to zh",
			path:       "/api/translate/prompt/1",
			body:       map[string]interface{}{"source_lang": "en", "target_lang": "zh"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "translate prompt from zh",
			path:       "/api/translate/prompt/1",
			body:       map[string]interface{}{"source_lang": "zh", "target_lang": "en"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "non-existent prompt",
			path:       "/api/translate/prompt/999",
			body:       map[string]interface{}{"source_lang": "en", "target_lang": "zh"},
			wantStatus: http.StatusNotFound,
			wantSucc:   false,
		},
		{
			name:       "default langs",
			path:       "/api/translate/prompt/1",
			body:       map[string]interface{}{},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postJSON(router, tt.path, tt.body)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}

			var resp struct {
				Success bool `json:"success"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}
		})
	}

	// Verify translation record was created
	var count int64
	db.Model(&models.Translation{}).Count(&count)
	if count == 0 {
		t.Error("expected translation records to be created")
	}
}

func TestTranslateHandler_TranslateEntity_Skill(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTranslateHandler(db)

	skill := models.Skill{Name: "/translate-skill", Content: "Skill content here"}
	db.Create(&skill)

	w := postJSON(router, "/api/translate/skill/1", map[string]interface{}{
		"source_lang": "en",
		"target_lang": "zh",
	})
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Success bool `json:"success"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestTranslateHandler_TranslateEntity_Agent(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTranslateHandler(db)

	agent := models.Agent{Name: "translate-agent", Content: "Agent persona content"}
	db.Create(&agent)

	w := postJSON(router, "/api/translate/agent/1", map[string]interface{}{
		"source_lang": "en",
		"target_lang": "zh",
	})
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Success bool `json:"success"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success")
	}
}

func TestTranslateHandler_TranslateEntity_InvalidType(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithTranslateHandler(db)

	w := postJSON(router, "/api/translate/invalid/1", map[string]interface{}{
		"source_lang": "en",
		"target_lang": "zh",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}

	var resp struct {
		Success bool `json:"success"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Success {
		t.Error("expected failure for invalid entity type")
	}
}

func TestMockTranslate(t *testing.T) {
	tests := []struct {
		text       string
		sourceLang string
		targetLang string
		wantHas    string
	}{
		{"Hello world", "en", "zh", "【翻译内容】"},
		{"Test content", "en", "zh", "【翻译内容】"},
		{"【翻译内容】你好", "zh", "en", "你好"},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			got := mockTranslate(tt.text, tt.sourceLang, tt.targetLang)
			if tt.targetLang == "zh" && got == "" {
				t.Error("expected non-empty translation for zh")
			}
			if tt.targetLang == "zh" && got[:len(tt.wantHas)] != tt.wantHas {
				t.Errorf("got %s, want prefix %s", got, tt.wantHas)
			}
		})
	}
}

func TestGetDefaultTranslateModel(t *testing.T) {
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
			got := getDefaultTranslateModel(tt.provider)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
