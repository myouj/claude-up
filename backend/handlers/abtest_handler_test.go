package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func setupTestRouterWithABTestHandler(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	abTestHandler := NewABTestHandler(db)

	api := r.Group("/api")
	api.GET("/prompts/:id/ab-tests", abTestHandler.ListByPrompt)
	api.POST("/prompts/:id/ab-tests", abTestHandler.Create)
	api.GET("/ab-tests", abTestHandler.List)
	api.GET("/ab-tests/:id", abTestHandler.Get)
	api.GET("/ab-tests/:id/results", abTestHandler.GetResults)
	api.GET("/ab-tests/:id/summary", abTestHandler.GetResultsSummary)
	api.POST("/ab-tests/:id/start", abTestHandler.Start)
	api.POST("/ab-tests/:id/stop", abTestHandler.Stop)
	api.POST("/ab-tests/:id/run", abTestHandler.RunIteration)
	api.DELETE("/ab-tests/:id", abTestHandler.Delete)

	return r
}

func TestABTestHandler_Create(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	// Create a prompt first
	db.Create(&models.Prompt{Title: "Test Prompt", Content: "Hello {{name}}}"})

	configJSON := `{"variant_a":"Hello A","variant_b":"Hello B","model":"gpt-4o","max_runs":10,"alpha":0.05}`

	tests := []struct {
		name       string
		promptID   string
		body       map[string]interface{}
		wantStatus int
		wantSucc   bool
	}{
		{
			name:       "valid A/B test",
			promptID:   "1",
			body:       map[string]interface{}{"name": "Test AB", "config": configJSON},
			wantStatus: http.StatusCreated,
			wantSucc:   true,
		},
		{
			name:       "missing name",
			promptID:   "1",
			body:       map[string]interface{}{"config": configJSON},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
		{
			name:       "missing config",
			promptID:   "1",
			body:       map[string]interface{}{"name": "Test AB"},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
		{
			name:       "invalid config JSON",
			promptID:   "1",
			body:       map[string]interface{}{"name": "Test AB", "config": "invalid-json"},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
		{
			name:       "non-existent prompt",
			promptID:   "999",
			body:       map[string]interface{}{"name": "Test AB", "config": configJSON},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postJSON(router, "/api/prompts/"+tt.promptID+"/ab-tests", tt.body)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}

			var resp APIResponse
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}
		})
	}
}

func TestABTestHandler_Get(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.Prompt{Title: "Test", Content: "Content"})
	db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B"}`})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantSucc   bool
	}{
		{"valid id", "/api/ab-tests/1", http.StatusOK, true},
		{"invalid id", "/api/ab-tests/999", http.StatusNotFound, false},
		{"non-numeric id", "/api/ab-tests/abc", http.StatusBadRequest, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := getJSON(router, tt.path)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}

			var resp APIResponse
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}
		})
	}
}

func TestABTestHandler_List(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.Prompt{Title: "Test", Content: "Content"})
	for i := 0; i < 5; i++ {
		db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B"}`})
	}

	tests := []struct {
		name      string
		query     string
		wantCount int
	}{
		{"all tests", "/api/ab-tests", 5},
		{"pagination page 1", "/api/ab-tests?page=1&limit=2", 2},
		{"pagination page 2", "/api/ab-tests?page=2&limit=2", 2},
		{"pagination page 3", "/api/ab-tests?page=3&limit=2", 1},
		{"empty result", "/api/ab-tests?page=10", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := getJSON(router, tt.query)
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
			if len(resp.Data) != tt.wantCount {
				t.Errorf("count: got %d, want %d", len(resp.Data), tt.wantCount)
			}
		})
	}
}

func TestABTestHandler_ListByPrompt(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.Prompt{Title: "Test", Content: "Content"})
	for i := 0; i < 3; i++ {
		db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B"}`})
	}

	w := getJSON(router, "/api/prompts/1/ab-tests")
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
	if len(resp.Data) != 3 {
		t.Errorf("count: got %d, want %d", len(resp.Data), 3)
	}
}

func TestABTestHandler_GetResults(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B"}`})
	db.Create(&models.ABTestResult{ABTestID: 1, RunIndex: 1, Variant: "A", Score: 0.8, LatencyMs: 150})
	db.Create(&models.ABTestResult{ABTestID: 1, RunIndex: 2, Variant: "B", Score: 0.75, LatencyMs: 180})

	w := getJSON(router, "/api/ab-tests/1/results")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Success bool                   `json:"success"`
		Data    []models.ABTestResult `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success")
	}
	if len(resp.Data) != 2 {
		t.Errorf("count: got %d, want %d", len(resp.Data), 2)
	}
}

func TestABTestHandler_GetResultsSummary(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B","alpha":0.05,"min_runs":2}`})
	db.Create(&models.ABTestResult{ABTestID: 1, RunIndex: 1, Variant: "A", Score: 0.8, LatencyMs: 150})
	db.Create(&models.ABTestResult{ABTestID: 1, RunIndex: 2, Variant: "B", Score: 0.75, LatencyMs: 180})

	w := getJSON(router, "/api/ab-tests/1/summary")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			TotalRunsA int     `json:"total_runs_a"`
			TotalRunsB int     `json:"total_runs_b"`
			AverageScoreA float64 `json:"average_score_a"`
			AverageScoreB float64 `json:"average_score_b"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success")
	}
	if resp.Data.TotalRunsA != 1 {
		t.Errorf("TotalRunsA: got %d, want %d", resp.Data.TotalRunsA, 1)
	}
	if resp.Data.TotalRunsB != 1 {
		t.Errorf("TotalRunsB: got %d, want %d", resp.Data.TotalRunsB, 1)
	}
}

func TestABTestHandler_Start(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B"}`, Status: models.ABTestStatusPending})

	w := postJSON(router, "/api/ab-tests/1/start", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var abTest models.ABTest
	db.First(&abTest, 1)
	if abTest.Status != models.ABTestStatusRunning {
		t.Errorf("status: got %s, want %s", abTest.Status, models.ABTestStatusRunning)
	}
}

func TestABTestHandler_Start_InvalidState(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B"}`, Status: models.ABTestStatusRunning})

	w := postJSON(router, "/api/ab-tests/1/start", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestABTestHandler_Stop(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B"}`, Status: models.ABTestStatusRunning})

	w := postJSON(router, "/api/ab-tests/1/stop", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var abTest models.ABTest
	db.First(&abTest, 1)
	if abTest.Status != models.ABTestStatusStopped {
		t.Errorf("status: got %s, want %s", abTest.Status, models.ABTestStatusStopped)
	}
}

func TestABTestHandler_Stop_InvalidState(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B"}`, Status: models.ABTestStatusPending})

	w := postJSON(router, "/api/ab-tests/1/stop", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestABTestHandler_Delete(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B"}`})
	db.Create(&models.ABTestResult{ABTestID: 1, RunIndex: 1, Variant: "A", Score: 0.8, LatencyMs: 150})

	w := deleteReq(router, "/api/ab-tests/1")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	// Verify ABTest is deleted
	var count int64
	db.Model(&models.ABTest{}).Count(&count)
	if count != 0 {
		t.Errorf("ABTest count: got %d, want 0", count)
	}

	// Verify results are deleted
	db.Model(&models.ABTestResult{}).Count(&count)
	if count != 0 {
		t.Errorf("ABTestResult count: got %d, want 0", count)
	}
}

func TestABTestHandler_Delete_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	w := deleteReq(router, "/api/ab-tests/999")
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestABTestHandler_RunIteration(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B","max_runs":5,"min_runs":1}`, Status: models.ABTestStatusRunning})

	w := postJSON(router, "/api/ab-tests/1/run", nil)
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Result struct {
				ID        uint   `json:"id"`
				Variant   string `json:"variant"`
				RunIndex  int    `json:"run_index"`
			} `json:"result"`
			Status string `json:"status"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success")
	}
	if resp.Data.Result.RunIndex != 1 {
		t.Errorf("run_index: got %d, want %d", resp.Data.Result.RunIndex, 1)
	}
}

func TestABTestHandler_RunIteration_NotRunning(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithABTestHandler(db)

	db.Create(&models.ABTest{PromptID: 1, Name: "Test AB", Config: `{"variant_a":"A","variant_b":"B"}`, Status: models.ABTestStatusPending})

	w := postJSON(router, "/api/ab-tests/1/run", nil)
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}
