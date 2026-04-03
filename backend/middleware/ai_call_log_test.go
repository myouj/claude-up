package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newTestDB creates a file-based SQLite database for testing.
func newTestDB(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
	db, err := gorm.Open(sqlite.Open(dbPath+"?_busy_timeout=30000&_journal_mode=DELETE"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&models.AICallLog{})
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	t.Cleanup(func() {
		if db != nil {
			sqlDB, _ := db.DB()
			if sqlDB != nil {
				sqlDB.Close()
			}
		}
	})
	return db
}

func TestAICallLogMiddleware_Basic(t *testing.T) {
	db := newTestDB(t)

	router := gin.New()
	router.Use(TraceMiddleware())
	router.Use(AICallLogMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(HeaderAIProvider, "openai")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	// Verify log was created
	var count int64
	db.Model(&models.AICallLog{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 log entry, got %d", count)
	}

	var log models.AICallLog
	db.First(&log)
	if log.Provider != "openai" {
		t.Errorf("provider: got %s, want openai", log.Provider)
	}
	if log.TraceID == "" {
		t.Error("trace_id should not be empty")
	}
}

func TestAICallLogMiddleware_NoHeader(t *testing.T) {
	db := newTestDB(t)

	router := gin.New()
	router.Use(TraceMiddleware())
	router.Use(AICallLogMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	// No X-AI-Provider header
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	// Verify no log was created
	var count int64
	db.Model(&models.AICallLog{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 log entries, got %d", count)
	}
}

func TestAICallLogMiddleware_WithContextValues(t *testing.T) {
	db := newTestDB(t)

	router := gin.New()
	router.Use(TraceMiddleware())
	router.Use(AICallLogMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		// Set AI call details in context
		SetAICallLog(c, &models.AICallLog{
			Model:        "gpt-4",
			InputTokens:  100,
			OutputTokens: 200,
			Cost:         0.003,
			PromptID:     42,
		})
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(HeaderAIProvider, "openai")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	// Verify log was created with correct values
	var count int64
	db.Model(&models.AICallLog{}).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 log entry, got %d", count)
	}

	var log models.AICallLog
	db.First(&log)
	if log.Provider != "openai" {
		t.Errorf("provider: got %s, want openai", log.Provider)
	}
	if log.Model != "gpt-4" {
		t.Errorf("model: got %s, want gpt-4", log.Model)
	}
	if log.InputTokens != 100 {
		t.Errorf("input_tokens: got %d, want 100", log.InputTokens)
	}
	if log.OutputTokens != 200 {
		t.Errorf("output_tokens: got %d, want 200", log.OutputTokens)
	}
	if log.Cost != 0.003 {
		t.Errorf("cost: got %f, want 0.003", log.Cost)
	}
	if log.PromptID != 42 {
		t.Errorf("prompt_id: got %d, want 42", log.PromptID)
	}
	if log.LatencyMs < 0 {
		t.Errorf("latency_ms should not be negative, got %d", log.LatencyMs)
	}
}

func TestAICallLogMiddleware_AllProviders(t *testing.T) {
	providers := []string{"openai", "claude", "gemini", "minimax"}

	for _, provider := range providers {
		t.Run(provider, func(t *testing.T) {
			db := newTestDB(t)

			router := gin.New()
			router.Use(TraceMiddleware())
			router.Use(AICallLogMiddleware(db))
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"success": true})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set(HeaderAIProvider, provider)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var log models.AICallLog
			db.First(&log)
			if log.Provider != provider {
				t.Errorf("provider: got %s, want %s", log.Provider, provider)
			}
		})
	}
}

func TestAICallLogMiddleware_LatencyRecorded(t *testing.T) {
	db := newTestDB(t)

	router := gin.New()
	router.Use(TraceMiddleware())
	router.Use(AICallLogMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		// Simulate some processing time
		c.JSON(200, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(HeaderAIProvider, "openai")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var log models.AICallLog
	db.First(&log)
	if log.LatencyMs < 0 {
		t.Errorf("latency_ms should not be negative, got %d", log.LatencyMs)
	}
}

func TestAICallLogMiddleware_TraceIDPropagation(t *testing.T) {
	db := newTestDB(t)

	router := gin.New()
	router.Use(TraceMiddleware())
	router.Use(AICallLogMiddleware(db))
	router.GET("/test", func(c *gin.Context) {
		traceID := GetTraceID(c)
		c.JSON(200, gin.H{"trace_id": traceID})
	})

	// Provide a trace ID in the header
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set(HeaderAIProvider, "openai")
	req.Header.Set(HeaderTraceID, "test-trace-123")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var log models.AICallLog
	db.First(&log)
	if log.TraceID != "test-trace-123" {
		t.Errorf("trace_id: got %s, want test-trace-123", log.TraceID)
	}
}

func TestAICallLogMiddleware_DifferentModels(t *testing.T) {
	testCases := []struct {
		provider string
		model    string
	}{
		{"openai", "gpt-4o"},
		{"claude", "claude-3-5-sonnet-20241022"},
		{"gemini", "gemini-2.0-flash"},
		{"minimax", "MiniMax-Text-01"},
	}

	for _, tc := range testCases {
		t.Run(tc.provider+"-"+tc.model, func(t *testing.T) {
			// Fresh DB for each test
			db := newTestDB(t)

			router := gin.New()
			router.Use(TraceMiddleware())
			router.Use(AICallLogMiddleware(db))
			router.GET("/test", func(c *gin.Context) {
				SetAICallLog(c, &models.AICallLog{
					Model: tc.model,
				})
				c.JSON(200, gin.H{"success": true})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.Header.Set(HeaderAIProvider, tc.provider)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var log models.AICallLog
			db.First(&log)
			if log.Provider != tc.provider {
				t.Errorf("provider: got %s, want %s", log.Provider, tc.provider)
			}
			if log.Model != tc.model {
				t.Errorf("model: got %s, want %s", log.Model, tc.model)
			}
		})
	}
}

func TestSetAndGetAICallLog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	original := &models.AICallLog{
		Model:        "test-model",
		InputTokens:  50,
		OutputTokens: 100,
		Cost:         0.001,
		PromptID:     1,
	}

	SetAICallLog(c, original)
	retrieved := GetAICallLog(c)

	if retrieved == nil {
		t.Fatal("expected non-nil AICallLog")
	}
	if retrieved.Model != original.Model {
		t.Errorf("model: got %s, want %s", retrieved.Model, original.Model)
	}
	if retrieved.InputTokens != original.InputTokens {
		t.Errorf("input_tokens: got %d, want %d", retrieved.InputTokens, original.InputTokens)
	}
	if retrieved.OutputTokens != original.OutputTokens {
		t.Errorf("output_tokens: got %d, want %d", retrieved.OutputTokens, original.OutputTokens)
	}
	if retrieved.Cost != original.Cost {
		t.Errorf("cost: got %f, want %f", retrieved.Cost, original.Cost)
	}
	if retrieved.PromptID != original.PromptID {
		t.Errorf("prompt_id: got %d, want %d", retrieved.PromptID, original.PromptID)
	}
}

func TestGetAICallLog_NotSet(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	retrieved := GetAICallLog(c)
	if retrieved != nil {
		t.Error("expected nil when not set")
	}
}

func TestAICallLog_JSONSerialization(t *testing.T) {
	log := models.AICallLog{
		ID:           1,
		Provider:     "openai",
		Model:        "gpt-4",
		InputTokens:  100,
		OutputTokens: 200,
		LatencyMs:    150,
		Cost:         0.003,
		TraceID:      "trace-123",
		PromptID:     42,
	}

	data, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var unmarshaled models.AICallLog
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if unmarshaled.Provider != log.Provider {
		t.Errorf("provider: got %s, want %s", unmarshaled.Provider, log.Provider)
	}
	if unmarshaled.Model != log.Model {
		t.Errorf("model: got %s, want %s", unmarshaled.Model, log.Model)
	}
	if unmarshaled.InputTokens != log.InputTokens {
		t.Errorf("input_tokens: got %d, want %d", unmarshaled.InputTokens, log.InputTokens)
	}
	if unmarshaled.OutputTokens != log.OutputTokens {
		t.Errorf("output_tokens: got %d, want %d", unmarshaled.OutputTokens, log.OutputTokens)
	}
	if unmarshaled.LatencyMs != log.LatencyMs {
		t.Errorf("latency_ms: got %d, want %d", unmarshaled.LatencyMs, log.LatencyMs)
	}
	if unmarshaled.Cost != log.Cost {
		t.Errorf("cost: got %f, want %f", unmarshaled.Cost, log.Cost)
	}
	if unmarshaled.TraceID != log.TraceID {
		t.Errorf("trace_id: got %s, want %s", unmarshaled.TraceID, log.TraceID)
	}
	if unmarshaled.PromptID != log.PromptID {
		t.Errorf("prompt_id: got %d, want %d", unmarshaled.PromptID, log.PromptID)
	}
}
