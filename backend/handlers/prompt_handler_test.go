package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func setupTestRouterWithPrefill(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	activityHandler := NewActivityHandler(db)
	promptHandler := NewPromptHandler(db, activityHandler)

	api := r.Group("/api")
	api.POST("/prompts/prefill", promptHandler.Prefill)

	return r
}

func TestPrefill_MockMode(t *testing.T) {
	os.Unsetenv("MINIMAX_API_KEY")
	db := newTestDB(t)
	router := setupTestRouterWithPrefill(db)

	body := map[string]interface{}{"title": "代码审查助手"}
	w := postJSON(router, "/api/prompts/prefill", body)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["success"] != true {
		t.Errorf("expected success=true, got %v", resp["success"])
	}

	data := resp["data"].(map[string]interface{})
	if data["content"] == "" {
		t.Error("expected non-empty content")
	}
	if data["description"] == "" {
		t.Error("expected non-empty description")
	}
	if data["category"] == "" {
		t.Error("expected non-empty category")
	}
	if data["tags"] == nil {
		t.Error("expected non-nil tags")
	}
}

func TestPrefill_EmptyTitle(t *testing.T) {
	os.Unsetenv("MINIMAX_API_KEY")
	db := newTestDB(t)
	router := setupTestRouterWithPrefill(db)

	body := map[string]interface{}{"title": ""}
	w := postJSON(router, "/api/prompts/prefill", body)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for empty title, got %d", w.Code)
	}
}

func TestPrefill_MissingTitle(t *testing.T) {
	os.Unsetenv("MINIMAX_API_KEY")
	db := newTestDB(t)
	router := setupTestRouterWithPrefill(db)

	body := map[string]interface{}{}
	w := postJSON(router, "/api/prompts/prefill", body)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400 for missing title, got %d", w.Code)
	}
}

func TestPrefill_CodeKeywordDetection(t *testing.T) {
	os.Unsetenv("MINIMAX_API_KEY")
	db := newTestDB(t)
	router := setupTestRouterWithPrefill(db)

	tests := []struct {
		title          string
		wantCategory   string
	}{
		{"代码审查助手", "代码开发"},
		{"Code Review Expert", "代码开发"},
		{"数据分析报告", "数据分析"},
		{"文档写作助手", "文档写作"},
		{"翻译助手", "翻译"},
		{"总结归纳工具", "总结归纳"},
		{"问答系统", "问答助手"},
		{"未分类助手", "其他"},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			body := map[string]interface{}{"title": tt.title}
			w := postJSON(router, "/api/prompts/prefill", body)

			if w.Code != http.StatusOK {
				t.Fatalf("expected status 200, got %d", w.Code)
			}

			var resp map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
				t.Fatalf("failed to unmarshal: %v", err)
			}

			data := resp["data"].(map[string]interface{})
			if data["category"] != tt.wantCategory {
				t.Errorf("category: got %s, want %s", data["category"], tt.wantCategory)
			}
		})
	}
}

func TestPrefill_TagsArray(t *testing.T) {
	os.Unsetenv("MINIMAX_API_KEY")
	db := newTestDB(t)
	router := setupTestRouterWithPrefill(db)

	body := map[string]interface{}{"title": "通用助手"}
	w := postJSON(router, "/api/prompts/prefill", body)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	data := resp["data"].(map[string]interface{})
	tags, ok := data["tags"].([]interface{})
	if !ok {
		t.Fatalf("expected tags to be array, got %T", data["tags"])
	}
	if len(tags) == 0 {
		t.Error("expected at least one tag")
	}
}
