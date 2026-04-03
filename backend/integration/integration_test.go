package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/handlers"
	"prompt-vault/models"
)

// init sets TESTING=1 before any tests run so that ActivityHandler.Log
// executes synchronously and does not leak goroutines that write to closed DBs.
func init() {
	os.Setenv("TESTING", "1")
}

func setupTestDB(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/integration_test.db"
	db, err := gorm.Open(sqlite.Open(dbPath+"?_busy_timeout=30000&_journal_mode=DELETE"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(
		&models.Prompt{},
		&models.PromptVersion{},
		&models.TestRecord{},
		&models.Skill{},
		&models.Agent{},
		&models.Translation{},
		&models.ActivityLog{},
		&models.Setting{},
	)
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

func setupRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	activityHandler := handlers.NewActivityHandler(db)
	promptHandler := handlers.NewPromptHandler(db, activityHandler)
	skillHandler := handlers.NewSkillHandler(db, activityHandler)
	agentHandler := handlers.NewAgentHandler(db, activityHandler)
	versionHandler := handlers.NewVersionHandler(db)
	testHandler := handlers.NewTestHandler(db, activityHandler)
	translateHandler := handlers.NewTranslateHandler(db)
	settingHandler := handlers.NewSettingHandler(db)

	api := r.Group("/api")
	// Prompt CRUD
	api.GET("/prompts", promptHandler.List)
	api.POST("/prompts", promptHandler.Create)
	api.GET("/prompts/:id", promptHandler.Get)
	api.PUT("/prompts/:id", promptHandler.Update)
	api.DELETE("/prompts/:id", promptHandler.Delete)
	api.POST("/prompts/:id/favorite", promptHandler.ToggleFavorite)
	api.GET("/prompts/categories", promptHandler.ListCategories)
	api.POST("/prompts/:id/clone", promptHandler.Clone)
	api.GET("/prompts/export", promptHandler.Export)
	api.POST("/prompts/import", promptHandler.Import)
	// Version management
	api.GET("/prompts/:id/versions", versionHandler.List)
	api.POST("/prompts/:id/versions", versionHandler.Create)
	api.GET("/versions/:id", versionHandler.Get)
	// Testing
	api.POST("/prompts/:id/test", testHandler.Test)
	api.POST("/prompts/:id/optimize", testHandler.Optimize)
	api.GET("/prompts/:id/tests", testHandler.List)
	api.GET("/prompts/:id/test-compare", testHandler.Compare)
	api.GET("/prompts/:id/analytics", testHandler.Analytics)
	api.GET("/models", testHandler.ListModels)
	// Skills CRUD
	api.GET("/skills", skillHandler.List)
	api.POST("/skills", skillHandler.Create)
	api.GET("/skills/:id", skillHandler.Get)
	api.PUT("/skills/:id", skillHandler.Update)
	api.DELETE("/skills/:id", skillHandler.Delete)
	api.GET("/skills/categories", skillHandler.ListCategories)
	api.POST("/skills/:id/clone", skillHandler.Clone)
	api.GET("/skills/export", skillHandler.Export)
	api.POST("/skills/import", skillHandler.Import)
	// Agents CRUD
	api.GET("/agents", agentHandler.List)
	api.POST("/agents", agentHandler.Create)
	api.GET("/agents/:id", agentHandler.Get)
	api.PUT("/agents/:id", agentHandler.Update)
	api.DELETE("/agents/:id", agentHandler.Delete)
	api.GET("/agents/categories", agentHandler.ListCategories)
	api.POST("/agents/:id/clone", agentHandler.Clone)
	api.GET("/agents/export", agentHandler.Export)
	api.POST("/agents/import", agentHandler.Import)
	// Translation
	api.POST("/translate", translateHandler.Translate)
	api.POST("/translate/:type/:id", translateHandler.TranslateEntity)
	// Activity logs
	api.GET("/activity-logs", activityHandler.List)
	// Settings
	api.GET("/settings", settingHandler.List)
	api.GET("/settings/:key", settingHandler.Get)
	api.PUT("/settings/:key", settingHandler.Set)
	api.DELETE("/settings/:key", settingHandler.Delete)
	// Stats
	api.GET("/stats", func(c *gin.Context) {
		var promptCount, skillCount, agentCount int64
		db.Model(&models.Prompt{}).Count(&promptCount)
		db.Model(&models.Skill{}).Count(&skillCount)
		db.Model(&models.Agent{}).Count(&agentCount)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"prompts": promptCount,
				"skills":  skillCount,
				"agents":  agentCount,
			},
		})
	})
	// Full export
	api.GET("/export", func(c *gin.Context) {
		var prompts []models.Prompt
		var skills []models.Skill
		var agents []models.Agent
		db.Order("updated_at DESC").Find(&prompts)
		db.Order("updated_at DESC").Find(&skills)
		db.Order("updated_at DESC").Find(&agents)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": models.ExportPayload{
				Version:    "1.0",
				Prompts:    prompts,
				Skills:     skills,
				Agents:     agents,
			},
		})
	})

	return r
}

func postJSON(r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func putJSON(r *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func getJSON(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func deleteReq(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ----- Integration Scenarios -----

// TestPromptLifecycle tests the full CRUD lifecycle of a prompt.
func TestPromptLifecycle(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// 1. Create a prompt
	w := postJSON(r, "/api/prompts", map[string]interface{}{
		"title":       "Integration Test Prompt",
		"content":     "Hello {{name}}, welcome to {{place}}",
		"description": "A test prompt for integration testing",
		"category":    "testing",
		"tags":        []string{"test", "integration"},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create prompt: status %d", w.Code)
	}
	var createResp struct {
		Success bool `json:"success"`
		Data    struct {
			ID uint `json:"id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	if !createResp.Success {
		t.Fatal("expected success response")
	}
	promptID := createResp.Data.ID

	// 2. Get the prompt
	w = getJSON(r, "/api/prompts/"+itoa(int(promptID)))
	if w.Code != http.StatusOK {
		t.Fatalf("get prompt: status %d", w.Code)
	}
	var getResp struct {
		Success bool `json:"success"`
		Data    struct {
			Title    string `json:"title"`
			Category string `json:"category"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &getResp)
	if getResp.Data.Title != "Integration Test Prompt" {
		t.Errorf("title mismatch: got %s", getResp.Data.Title)
	}
	if getResp.Data.Category != "testing" {
		t.Errorf("category mismatch: got %s", getResp.Data.Category)
	}

	// 3. Update the prompt
	w = putJSON(r, "/api/prompts/"+itoa(int(promptID)), map[string]interface{}{
		"content": "Hello {{name}}, welcome to the {{place}}!",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update prompt: status %d", w.Code)
	}

	// 4. List prompts and verify
	w = getJSON(r, "/api/prompts")
	var listResp struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if !listResp.Success || len(listResp.Data) != 1 {
		t.Fatalf("list: success=%v count=%d", listResp.Success, len(listResp.Data))
	}

	// 5. Delete the prompt
	w = deleteReq(r, "/api/prompts/"+itoa(int(promptID)))
	if w.Code != http.StatusOK {
		t.Fatalf("delete prompt: status %d", w.Code)
	}

	// 6. Verify deletion
	w = getJSON(r, "/api/prompts/"+itoa(int(promptID)))
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

// TestVersionLifecycle tests auto-versioning when prompt content changes.
func TestVersionLifecycle(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create prompt (auto-creates version 1)
	w := postJSON(r, "/api/prompts", map[string]interface{}{
		"title":   "Version Test",
		"content": "Original content",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status %d", w.Code)
	}

	// Verify initial version
	w = getJSON(r, "/api/prompts/1/versions")
	var verList struct {
		Success bool `json:"success"`
		Data    []struct {
			Version int    `json:"version"`
			Content string `json:"content"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &verList)
	if len(verList.Data) != 1 || verList.Data[0].Version != 1 {
		t.Errorf("expected 1 version, got %d", len(verList.Data))
	}

	// Update content (should auto-create version 2)
	w = putJSON(r, "/api/prompts/1", map[string]interface{}{
		"content": "Updated content v2",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("update: status %d", w.Code)
	}

	// Verify version 2 was created
	w = getJSON(r, "/api/prompts/1/versions")
	json.Unmarshal(w.Body.Bytes(), &verList)
	if len(verList.Data) != 2 {
		t.Errorf("expected 2 versions after update, got %d", len(verList.Data))
	}

	// Update without content change (should NOT create version)
	w = putJSON(r, "/api/prompts/1", map[string]interface{}{
		"title": "New Title Only",
	})
	json.Unmarshal(w.Body.Bytes(), &verList)
	// Still 2 versions
	if len(verList.Data) != 2 {
		t.Errorf("expected 2 versions (no content change), got %d", len(verList.Data))
	}
}

// TestTestWorkflow tests the full testing workflow: create prompt -> run test -> view analytics.
func TestTestWorkflow(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create a prompt first
	postJSON(r, "/api/prompts", map[string]interface{}{
		"title":   "Test Workflow",
		"content": "Say hello to {{name}}",
	})

	// Run a test (mock AI since no API key)
	w := postJSON(r, "/api/prompts/1/test", map[string]interface{}{
		"content": "Say hello to Alice",
		"model":   "gpt-4",
		"provider": "openai",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("test: status %d", w.Code)
	}
	var testResp struct {
		Success bool `json:"success"`
		Data    struct {
			Response     string `json:"response"`
			TokensUsed   int    `json:"tokens_used"`
			Provider     string `json:"provider"`
			TestRecordID uint   `json:"test_record_id"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &testResp)
	if !testResp.Success {
		t.Fatal("expected success")
	}
	if testResp.Data.TokensUsed == 0 {
		t.Error("expected tokens_used > 0 (mock response)")
	}
	if testResp.Data.TestRecordID == 0 {
		t.Error("expected test_record_id")
	}

	// List tests for the prompt
	w = getJSON(r, "/api/prompts/1/tests")
	var listTests struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listTests)
	if !listTests.Success || len(listTests.Data) != 1 {
		t.Errorf("expected 1 test record, got %d", len(listTests.Data))
	}

	// Run another test
	postJSON(r, "/api/prompts/1/test", map[string]interface{}{
		"content": "Say hello to Bob",
		"model":   "gpt-4",
	})

	// View analytics
	w = getJSON(r, "/api/prompts/1/analytics")
	var analytics struct {
		Success bool `json:"success"`
		Data    struct {
			TotalTests int64 `json:"total_tests"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &analytics)
	if analytics.Data.TotalTests != 2 {
		t.Errorf("total_tests: got %d, want 2", analytics.Data.TotalTests)
	}

	// Compare tests
	w = getJSON(r, "/api/prompts/1/test-compare")
	if w.Code != http.StatusOK {
		t.Fatalf("compare: status %d", w.Code)
	}
}

// TestOptimizeWorkflow tests: create prompt -> optimize -> verify.
func TestOptimizeWorkflow(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create a prompt
	postJSON(r, "/api/prompts", map[string]interface{}{
		"title":   "To Optimize",
		"content": "Say hello",
	})

	// Optimize prompt
	w := postJSON(r, "/api/prompts/1/optimize", map[string]interface{}{
		"content":  "Say hello",
		"mode":     "improve",
		"provider": "openai",
		"model":    "gpt-4",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("optimize: status %d", w.Code)
	}
	var optResp struct {
		Success bool `json:"success"`
		Data    struct {
			Optimized string `json:"optimized"`
			Provider  string `json:"provider"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &optResp)
	if !optResp.Success {
		t.Fatal("expected success")
	}
	if optResp.Data.Optimized == "" {
		t.Error("expected optimized content")
	}
}

// TestCloneWorkflow tests: create -> clone -> verify all data copied.
func TestCloneWorkflow(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create a prompt with version
	w := postJSON(r, "/api/prompts", map[string]interface{}{
		"title":    "Original",
		"content":  "Original content",
		"category": "test",
		"tags":     []string{"original"},
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status %d", w.Code)
	}

	// Clone the prompt
	w = postJSON(r, "/api/prompts/1/clone", nil)
	if w.Code != http.StatusCreated {
		t.Fatalf("clone: status %d", w.Code)
	}
	var cloneResp struct {
		Success bool `json:"success"`
		Data    struct {
			ID       uint   `json:"id"`
			Title    string `json:"title"`
			Category string `json:"category"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &cloneResp)
	if cloneResp.Data.Title != "Original (Copy)" {
		t.Errorf("cloned title: got %s", cloneResp.Data.Title)
	}
	if cloneResp.Data.Category != "test" {
		t.Errorf("cloned category: got %s", cloneResp.Data.Category)
	}

	// Verify cloned prompt has version
	w = getJSON(r, "/api/prompts/"+itoa(int(cloneResp.Data.ID))+"/versions")
	var verResp struct {
		Success bool `json:"success"`
		Data    []struct {
			Version int `json:"version"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &verResp)
	if len(verResp.Data) != 1 {
		t.Errorf("cloned prompt should have 1 version, got %d", len(verResp.Data))
	}
}

// TestSkillAgentLifecycle tests skill and agent CRUD together.
func TestSkillAgentLifecycle(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create skill
	w := postJSON(r, "/api/skills", map[string]interface{}{
		"name":        "/test-skill",
		"description": "A test skill",
		"content":     "Skill content",
		"category":    "testing",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create skill: status %d", w.Code)
	}

	// Create agent
	w = postJSON(r, "/api/agents", map[string]interface{}{
		"name":         "test-agent",
		"role":         "Test Agent",
		"content":      "Agent content",
		"capabilities": "Testing",
		"category":     "testing",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create agent: status %d", w.Code)
	}

	// Verify stats
	w = getJSON(r, "/api/stats")
	var stats struct {
		Success bool `json:"success"`
		Data    struct {
			Skills  int64 `json:"skills"`
			Agents  int64 `json:"agents"`
			Prompts int64 `json:"prompts"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &stats)
	if stats.Data.Skills != 1 {
		t.Errorf("skills count: got %d, want 1", stats.Data.Skills)
	}
	if stats.Data.Agents != 1 {
		t.Errorf("agents count: got %d, want 1", stats.Data.Agents)
	}

	// Clone skill
	w = postJSON(r, "/api/skills/1/clone", nil)
	if w.Code != http.StatusCreated {
		t.Fatalf("clone skill: status %d", w.Code)
	}

	// Clone agent
	w = postJSON(r, "/api/agents/1/clone", nil)
	if w.Code != http.StatusCreated {
		t.Fatalf("clone agent: status %d", w.Code)
	}

	// Verify counts after cloning
	w = getJSON(r, "/api/stats")
	json.Unmarshal(w.Body.Bytes(), &stats)
	if stats.Data.Skills != 2 {
		t.Errorf("skills after clone: got %d, want 2", stats.Data.Skills)
	}
	if stats.Data.Agents != 2 {
		t.Errorf("agents after clone: got %d, want 2", stats.Data.Agents)
	}

	// Delete cloned skill
	w = deleteReq(r, "/api/skills/2")
	if w.Code != http.StatusOK {
		t.Fatalf("delete skill: status %d", w.Code)
	}

	// Delete cloned agent
	w = deleteReq(r, "/api/agents/2")
	if w.Code != http.StatusOK {
		t.Fatalf("delete agent: status %d", w.Code)
	}
}

// TestImportExportWorkflow tests: export -> delete -> import.
func TestImportExportWorkflow(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create some prompts
	postJSON(r, "/api/prompts", map[string]interface{}{"title": "P1", "content": "C1"})
	postJSON(r, "/api/prompts", map[string]interface{}{"title": "P2", "content": "C2"})

	// Export
	w := getJSON(r, "/api/prompts/export")
	var export struct {
		Success bool `json:"success"`
		Data    struct {
			Prompts []json.RawMessage `json:"prompts"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &export)
	if len(export.Data.Prompts) != 2 {
		t.Fatalf("export count: got %d, want 2", len(export.Data.Prompts))
	}

	// Delete original prompts
	deleteReq(r, "/api/prompts/1")
	deleteReq(r, "/api/prompts/2")

	// Import back
	importPayload := map[string]interface{}{
		"prompts": []map[string]interface{}{
			{"title": "P1", "content": "C1"},
			{"title": "P2", "content": "C2"},
		},
	}
	w = postJSON(r, "/api/prompts/import", importPayload)
	var importResult struct {
		Success  bool `json:"success"`
		Imported int  `json:"imported"`
	}
	json.Unmarshal(w.Body.Bytes(), &importResult)
	if importResult.Imported != 2 {
		t.Errorf("imported: got %d, want 2", importResult.Imported)
	}

	// Verify stats
	w = getJSON(r, "/api/stats")
	var stats struct {
		Success bool `json:"success"`
		Data    struct {
			Prompts int64 `json:"prompts"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &stats)
	if stats.Data.Prompts != 2 {
		t.Errorf("prompts after import: got %d, want 2", stats.Data.Prompts)
	}
}

// TestFullExportImportCycle tests the full system export and import.
func TestFullExportImportCycle(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create data in all modules
	postJSON(r, "/api/prompts", map[string]interface{}{"title": "P1", "content": "C1"})
	postJSON(r, "/api/skills", map[string]interface{}{"name": "/s1", "content": "C1"})
	postJSON(r, "/api/agents", map[string]interface{}{"name": "a1", "content": "C1"})

	// Full export
	w := getJSON(r, "/api/export")
	var fullExport struct {
		Success bool `json:"success"`
		Data    struct {
			Version string `json:"version"`
			Prompts []models.Prompt `json:"prompts"`
			Skills  []models.Skill `json:"skills"`
			Agents  []models.Agent `json:"agents"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &fullExport)
	if !fullExport.Success {
		t.Fatal("export failed")
	}
	if len(fullExport.Data.Prompts) != 1 {
		t.Errorf("prompts: got %d, want 1", len(fullExport.Data.Prompts))
	}
	if len(fullExport.Data.Skills) != 1 {
		t.Errorf("skills: got %d, want 1", len(fullExport.Data.Skills))
	}
	if len(fullExport.Data.Agents) != 1 {
		t.Errorf("agents: got %d, want 1", len(fullExport.Data.Agents))
	}

	// Delete all
	deleteReq(r, "/api/prompts/1")
	deleteReq(r, "/api/skills/1")
	deleteReq(r, "/api/agents/1")

	// Import prompts
	w = postJSON(r, "/api/prompts/import", map[string]interface{}{
		"prompts": []map[string]interface{}{
			{"title": "P1", "content": "C1"},
		},
	})

	// Import skills
	w = postJSON(r, "/api/skills/import", map[string]interface{}{
		"skills": []map[string]interface{}{
			{"name": "/s1", "content": "C1"},
		},
	})

	// Import agents
	w = postJSON(r, "/api/agents/import", map[string]interface{}{
		"agents": []map[string]interface{}{
			{"name": "a1", "content": "C1"},
		},
	})

	// Verify all restored
	w = getJSON(r, "/api/stats")
	var stats struct {
		Success bool `json:"success"`
		Data    struct {
			Prompts int64 `json:"prompts"`
			Skills  int64 `json:"skills"`
			Agents  int64 `json:"agents"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &stats)
	if stats.Data.Prompts != 1 || stats.Data.Skills != 1 || stats.Data.Agents != 1 {
		t.Errorf("after restore: prompts=%d skills=%d agents=%d",
			stats.Data.Prompts, stats.Data.Skills, stats.Data.Agents)
	}
}

// TestActivityLogIntegration verifies activity logging works across operations.
func TestActivityLogIntegration(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create prompt (should log activity)
	postJSON(r, "/api/prompts", map[string]interface{}{
		"title":   "Activity Test",
		"content": "Content",
	})

	// Update prompt (should log activity)
	putJSON(r, "/api/prompts/1", map[string]interface{}{
		"title": "Updated Title",
	})

	// Clone prompt (should log activity)
	postJSON(r, "/api/prompts/1/clone", nil)

	// Create skill (should log activity)
	postJSON(r, "/api/skills", map[string]interface{}{
		"name":    "/activity",
		"content": "Content",
	})

	// Check activity logs
	w := getJSON(r, "/api/activity-logs")
	var logs struct {
		Success bool `json:"success"`
		Data    []struct {
			EntityType string `json:"entity_type"`
			Action     string `json:"action"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &logs)
	if !logs.Success {
		t.Fatal("activity logs failed")
	}
	// Should have at least 4 logs: prompt create, prompt update, prompt clone, skill create
	if len(logs.Data) < 4 {
		t.Errorf("expected at least 4 activity logs, got %d", len(logs.Data))
	}
}

// TestSettingsLifecycle tests encrypted settings CRUD.
func TestSettingsLifecycle(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Set a setting
	w := putJSON(r, "/api/settings/test_key", map[string]interface{}{
		"value": "test_value",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("set setting: status %d", w.Code)
	}
	var setResp struct {
		Success bool `json:"success"`
	}
	json.Unmarshal(w.Body.Bytes(), &setResp)
	if !setResp.Success {
		t.Error("expected success")
	}

	// Get the setting
	w = getJSON(r, "/api/settings/test_key")
	var getResp struct {
		Success bool `json:"success"`
		Data    struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &getResp)
	if getResp.Data.Value != "test_value" {
		t.Errorf("value: got %s, want test_value", getResp.Data.Value)
	}

	// List all settings
	w = getJSON(r, "/api/settings")
	var listResp struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if !listResp.Success || len(listResp.Data) == 0 {
		t.Error("list settings failed")
	}

	// Delete the setting
	w = deleteReq(r, "/api/settings/test_key")
	if w.Code != http.StatusOK {
		t.Fatalf("delete setting: status %d", w.Code)
	}

	// Verify deletion
	w = getJSON(r, "/api/settings/test_key")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after delete, got %d", w.Code)
	}
}

// TestTranslationWorkflow tests entity translation.
func TestTranslationWorkflow(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create a prompt with Chinese field
	w := postJSON(r, "/api/prompts", map[string]interface{}{
		"title":       "English Title",
		"content":     "Hello {{name}}",
		"description": "English description",
	})
	if w.Code != http.StatusCreated {
		t.Fatalf("create: status %d", w.Code)
	}

	// Translate entity (mock translation - no API key)
	w = postJSON(r, "/api/translate/prompt/1", map[string]interface{}{
		"target_lang": "zh",
	})
	// May fail without API key, but should not panic
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("translate: unexpected status %d", w.Code)
	}

	// Free text translation
	w = postJSON(r, "/api/translate", map[string]interface{}{
		"text":      "Hello world",
		"target_lang": "zh",
	})
	// Without API key, may return mock or error
	if w.Code != http.StatusOK && w.Code != http.StatusInternalServerError {
		t.Errorf("free translate: unexpected status %d", w.Code)
	}
}

// TestPaginationIntegration tests pagination across multiple endpoints.
func TestPaginationIntegration(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Seed 25 prompts
	for i := 1; i <= 25; i++ {
		postJSON(r, "/api/prompts", map[string]interface{}{
			"title":   "Prompt " + itoa(i),
			"content": "Content " + itoa(i),
		})
	}

	// Test prompt pagination
	w := getJSON(r, "/api/prompts?page=1&limit=10")
	var page1 struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
		Meta    models.PaginationMeta `json:"meta"`
	}
	json.Unmarshal(w.Body.Bytes(), &page1)
	if len(page1.Data) != 10 {
		t.Errorf("page 1: got %d, want 10", len(page1.Data))
	}
	if page1.Meta.TotalPages != 3 {
		t.Errorf("total pages: got %d, want 3", page1.Meta.TotalPages)
	}

	w = getJSON(r, "/api/prompts?page=3&limit=10")
	var page3 struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
		Meta    models.PaginationMeta `json:"meta"`
	}
	json.Unmarshal(w.Body.Bytes(), &page3)
	if len(page3.Data) != 5 {
		t.Errorf("page 3: got %d, want 5", len(page3.Data))
	}

	// Test activity log pagination
	w = getJSON(r, "/api/activity-logs?page=1&limit=5")
	var actPage struct {
		Success bool `json:"success"`
		Meta    models.PaginationMeta `json:"meta"`
	}
	json.Unmarshal(w.Body.Bytes(), &actPage)
	if actPage.Meta.TotalPages < 1 {
		t.Error("expected at least 1 activity page")
	}
}

// TestCompareTestsWorkflow tests version comparison with tests.
func TestCompareTestsWorkflow(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create prompt
	postJSON(r, "/api/prompts", map[string]interface{}{
		"title":   "Compare Test",
		"content": "Version 1",
	})

	// Create version 2
	postJSON(r, "/api/prompts/1/versions", map[string]interface{}{
		"content": "Version 2",
	})

	// Test version 1
	postJSON(r, "/api/prompts/1/test", map[string]interface{}{
		"content": "Test V1",
		"model":   "gpt-4",
	})

	// Test version 2
	postJSON(r, "/api/prompts/1/test", map[string]interface{}{
		"content": "Test V2",
		"model":   "gpt-4",
	})

	// Compare
	w := getJSON(r, "/api/prompts/1/test-compare")
	var compare struct {
		Success bool `json:"success"`
		Data    []struct {
			VersionID uint `json:"version_id"`
			Version   int  `json:"version"`
			Tests     []struct {
				Model string `json:"model"`
			} `json:"tests"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &compare)
	if !compare.Success {
		t.Fatal("compare failed")
	}
	if len(compare.Data) != 2 {
		t.Errorf("expected 2 versions in compare, got %d", len(compare.Data))
	}
}

// TestFavoritePinnedWorkflow tests favorite and pinned functionality.
func TestFavoritePinnedWorkflow(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create prompts
	postJSON(r, "/api/prompts", map[string]interface{}{"title": "P1", "content": "C"})
	postJSON(r, "/api/prompts", map[string]interface{}{"title": "P2", "content": "C"})
	postJSON(r, "/api/prompts", map[string]interface{}{"title": "P3", "content": "C"})

	// Favorite and pin prompts
	putJSON(r, "/api/prompts/1", map[string]interface{}{"is_favorite": true})
	putJSON(r, "/api/prompts/2", map[string]interface{}{"is_pinned": true})
	putJSON(r, "/api/prompts/3", map[string]interface{}{"is_favorite": true, "is_pinned": true})

	// List favorites
	w := getJSON(r, "/api/prompts?favorite=true")
	var favResp struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &favResp)
	if len(favResp.Data) != 2 {
		t.Errorf("favorites: got %d, want 2", len(favResp.Data))
	}

	// Verify order: pinned first
	w = getJSON(r, "/api/prompts")
	var listResp struct {
		Success bool `json:"success"`
		Data    []struct {
			Title string `json:"title"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if len(listResp.Data) < 3 {
		t.Fatalf("list: got %d, want at least 3", len(listResp.Data))
	}
	// P3 should be first because it's pinned
	if listResp.Data[0].Title != "P3" {
		t.Errorf("first item should be P3 (pinned), got %s", listResp.Data[0].Title)
	}
}

// TestErrorHandlingIntegration tests error scenarios across modules.
func TestErrorHandlingIntegration(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	tests := []struct {
		name       string
		method     string
		path       string
		body       interface{}
		wantStatus int
	}{
		// Prompt errors
		{"get non-existent prompt", "GET", "/api/prompts/999", nil, http.StatusNotFound},
		{"update non-existent prompt", "PUT", "/api/prompts/999", map[string]interface{}{"title": "X"}, http.StatusNotFound},
		{"delete non-existent prompt", "DELETE", "/api/prompts/999", nil, http.StatusNotFound},
		{"create prompt missing title", "POST", "/api/prompts", map[string]interface{}{"content": "C"}, http.StatusBadRequest},
		{"create prompt missing content", "POST", "/api/prompts", map[string]interface{}{"title": "T"}, http.StatusBadRequest},

		// Skill errors
		{"get non-existent skill", "GET", "/api/skills/999", nil, http.StatusNotFound},
		{"update non-existent skill", "PUT", "/api/skills/999", map[string]interface{}{"name": "X"}, http.StatusNotFound},

		// Agent errors
		{"get non-existent agent", "GET", "/api/agents/999", nil, http.StatusNotFound},
		{"update non-existent agent", "PUT", "/api/agents/999", map[string]interface{}{"name": "X"}, http.StatusNotFound},

		// Version errors
		{"get non-existent version", "GET", "/api/versions/999", nil, http.StatusNotFound},

		// Settings errors
		{"get non-existent setting", "GET", "/api/settings/nonexistent", nil, http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var w *httptest.ResponseRecorder
			switch tt.method {
			case "GET":
				w = getJSON(r, tt.path)
			case "POST":
				w = postJSON(r, tt.path, tt.body)
			case "PUT":
				w = putJSON(r, tt.path, tt.body)
			case "DELETE":
				w = deleteReq(r, tt.path)
			}
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

// TestSearchAndFilterIntegration tests search and filter across endpoints.
func TestSearchAndFilterIntegration(t *testing.T) {
	db := setupTestDB(t)
	r := setupRouter(db)

	// Create prompts with different categories
	postJSON(r, "/api/prompts", map[string]interface{}{"title": "Go Tutorial", "content": "Learn Go", "category": "coding"})
	postJSON(r, "/api/prompts", map[string]interface{}{"title": "Python Tutorial", "content": "Learn Python", "category": "coding"})
	postJSON(r, "/api/prompts", map[string]interface{}{"title": "Write Story", "content": "Creative writing", "category": "writing"})
	postJSON(r, "/api/prompts", map[string]interface{}{"title": "Fav Item", "content": "Content"})
	// Set Fav Item as favorite via update (is_favorite is not a create field)
	putJSON(r, "/api/prompts/4", map[string]interface{}{"is_favorite": true})

	// Search prompts
	w := getJSON(r, "/api/prompts?search=Tutorial")
	var searchResp struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &searchResp)
	if len(searchResp.Data) != 2 {
		t.Errorf("search Tutorial: got %d, want 2", len(searchResp.Data))
	}

	// Filter by category
	w = getJSON(r, "/api/prompts?category=coding")
	json.Unmarshal(w.Body.Bytes(), &searchResp)
	if len(searchResp.Data) != 2 {
		t.Errorf("category coding: got %d, want 2", len(searchResp.Data))
	}

	// Filter favorites
	w = getJSON(r, "/api/prompts?favorite=true")
	json.Unmarshal(w.Body.Bytes(), &searchResp)
	if len(searchResp.Data) != 1 {
		t.Errorf("favorites: got %d, want 1", len(searchResp.Data))
	}

	// Combined search and filter
	w = getJSON(r, "/api/prompts?search=Go&category=coding")
	json.Unmarshal(w.Body.Bytes(), &searchResp)
	if len(searchResp.Data) != 1 {
		t.Errorf("search+filter: got %d, want 1", len(searchResp.Data))
	}

	// List categories
	w = getJSON(r, "/api/prompts/categories")
	var catResp struct {
		Success    bool     `json:"success"`
		Categories []string `json:"categories"`
	}
	json.Unmarshal(w.Body.Bytes(), &catResp)
	if len(catResp.Categories) != 2 {
		t.Errorf("categories: got %v, want 2 categories", catResp.Categories)
	}
}

// ----- Helper -----

func itoa(i int) string {
	return fmt.Sprintf("%d", i)
}
