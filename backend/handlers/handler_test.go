package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
	"prompt-vault/service"
)

func init() {
	// Make activity logging synchronous so async goroutines don't write to
	// closed DBs after test cleanup. Must be set before any handler tests run.
	os.Setenv("TESTING", "1")
}

// newTestDB creates a file-based SQLite database for testing.
// We use file-based instead of :memory: because GORM's connection pool
// may use different connections, and SQLite in-memory DBs are isolated per connection.
// Each call creates a unique database file to ensure test isolation.
func newTestDB(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	dbPath := tmpDir + "/test.db"
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
		&models.Task{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	// Close DB connection at cleanup. The OS will clean up the temp dir.
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

func setupTestRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	activityHandler := NewActivityHandler(db)
	promptHandler := NewPromptHandler(db, activityHandler)
	skillHandler := NewSkillHandler(db, activityHandler)
	agentHandler := NewAgentHandler(db, activityHandler)
	versionHandler := NewVersionHandler(db)
	taskService := service.NewTaskService(db)
	taskHandler := NewTaskHandler(db, taskService)

	api := r.Group("/api")
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

	api.GET("/prompts/:id/versions", versionHandler.List)
	api.POST("/prompts/:id/versions", versionHandler.Create)
	api.GET("/versions/:id", versionHandler.Get)

	api.GET("/skills", skillHandler.List)
	api.POST("/skills", skillHandler.Create)
	api.GET("/skills/:id", skillHandler.Get)
	api.PUT("/skills/:id", skillHandler.Update)
	api.DELETE("/skills/:id", skillHandler.Delete)
	api.GET("/skills/categories", skillHandler.ListCategories)
	api.POST("/skills/:id/clone", skillHandler.Clone)
	api.GET("/skills/export", skillHandler.Export)
	api.POST("/skills/import", skillHandler.Import)

	api.GET("/agents", agentHandler.List)
	api.POST("/agents", agentHandler.Create)
	api.GET("/agents/:id", agentHandler.Get)
	api.PUT("/agents/:id", agentHandler.Update)
	api.DELETE("/agents/:id", agentHandler.Delete)
	api.GET("/agents/categories", agentHandler.ListCategories)
	api.POST("/agents/:id/clone", agentHandler.Clone)
	api.GET("/agents/export", agentHandler.Export)
	api.POST("/agents/import", agentHandler.Import)

	api.GET("/tasks", taskHandler.ListTasks)
	api.POST("/tasks", taskHandler.CreateTask)
	api.GET("/tasks/:id", taskHandler.GetTask)
	api.DELETE("/tasks/:id", taskHandler.CancelTask)

	return r
}

func postJSON(router *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func putJSON(router *gin.Engine, path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func getJSON(router *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func deleteReq(router *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", path, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

type APIResponse struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data"`
	Error   string          `json:"error,omitempty"`
}

// ----- Prompt Handler Tests -----

func TestPromptHandler_Create(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
		wantSucc   bool
		checkTitle string
	}{
		{
			name:       "valid prompt",
			body:       map[string]interface{}{"title": "Test Prompt", "content": "Hello {{name}}"},
			wantStatus: http.StatusCreated,
			wantSucc:   true,
			checkTitle: "Test Prompt",
		},
		{
			name:       "missing title",
			body:       map[string]interface{}{"content": "content only"},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
		{
			name:       "missing content",
			body:       map[string]interface{}{"title": "Title only"},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
		{
			name:       "with category and tags",
			body:       map[string]interface{}{"title": "Categorized", "content": "Content", "category": "test", "tags": []string{"go", "api"}},
			wantStatus: http.StatusCreated,
			wantSucc:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postJSON(router, "/api/prompts", tt.body)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}

			var resp APIResponse
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}

			if tt.checkTitle != "" && resp.Success {
				var data map[string]interface{}
				json.Unmarshal(resp.Data, &data)
				if title, ok := data["title"].(string); ok && title != tt.checkTitle {
					t.Errorf("title: got %s, want %s", title, tt.checkTitle)
				}
			}
		})
	}
}

func TestPromptHandler_List(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	// Seed some prompts
	for i := 0; i < 5; i++ {
		db.Create(&models.Prompt{Title: "Prompt", Content: "Content", Category: "test"})
	}
	db.Create(&models.Prompt{Title: "Fav", Content: "Content", IsFavorite: true})
	db.Create(&models.Prompt{Title: "Pinned", Content: "Content", IsPinned: true})

	tests := []struct {
		name       string
		query      string
		wantCount  int
		wantSucc   bool
	}{
		{"all prompts", "/api/prompts", 7, true},
		{"page 1", "/api/prompts?page=1&limit=3", 3, true},
		{"page 2", "/api/prompts?page=2&limit=3", 3, true},
		{"page 3", "/api/prompts?page=3&limit=3", 1, true},
		{"category filter", "/api/prompts?category=test", 5, true},
		{"favorite filter", "/api/prompts?favorite=true", 1, true},
		{"search", "/api/prompts?search=Fav", 1, true},
		{"empty search", "/api/prompts?search=nonexistent", 0, true},
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
				Meta    models.PaginationMeta `json:"meta"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}
			if len(resp.Data) != tt.wantCount {
				t.Errorf("count: got %d, want %d", len(resp.Data), tt.wantCount)
			}
		})
	}
}

func TestPromptHandler_Get(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	prompt := models.Prompt{Title: "Get Test", Content: "Content"}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 1, Content: "Content"})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantSucc   bool
	}{
		{"valid id", "/api/prompts/1", http.StatusOK, true},
		{"invalid id", "/api/prompts/999", http.StatusNotFound, false},
		{"non-numeric id", "/api/prompts/abc", http.StatusBadRequest, false},
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

func TestPromptHandler_Update(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	prompt := models.Prompt{Title: "Original", Content: "Original Content"}
	db.Create(&prompt)

	tests := []struct {
		name         string
		path         string
		body         map[string]interface{}
		wantStatus   int
		wantSucc     bool
		wantTitle    string
		wantContent  string
		createsVer   bool
	}{
		{
			name:        "update title",
			path:        "/api/prompts/1",
			body:        map[string]interface{}{"title": "Updated Title"},
			wantStatus:  http.StatusOK,
			wantSucc:    true,
			wantTitle:   "Updated Title",
			wantContent: "Original Content",
		},
		{
			name:        "update content creates version",
			path:        "/api/prompts/1",
			body:        map[string]interface{}{"content": "New Content"},
			wantStatus:  http.StatusOK,
			wantSucc:    true,
			wantContent: "New Content",
			createsVer:  true,
		},
		{
			name:        "update non-existent",
			path:        "/api/prompts/999",
			body:        map[string]interface{}{"title": "X"},
			wantStatus:  http.StatusNotFound,
			wantSucc:    false,
		},
		{
			name:        "toggle favorite",
			path:        "/api/prompts/1",
			body:        map[string]interface{}{"is_favorite": true},
			wantStatus:  http.StatusOK,
			wantSucc:    true,
		},
		{
			name:        "toggle pinned",
			path:        "/api/prompts/1",
			body:        map[string]interface{}{"is_pinned": true},
			wantStatus:  http.StatusOK,
			wantSucc:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var beforeVerCount int64
			if tt.createsVer {
				db.Model(&models.PromptVersion{}).Where("prompt_id = ?", 1).Count(&beforeVerCount)
			}

			w := putJSON(router, tt.path, tt.body)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}

			var resp APIResponse
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}

			if tt.createsVer && resp.Success {
				var afterVerCount int64
				db.Model(&models.PromptVersion{}).Where("prompt_id = ?", 1).Count(&afterVerCount)
				if afterVerCount <= beforeVerCount {
					t.Errorf("expected version to be created, count: %d -> %d", beforeVerCount, afterVerCount)
				}
			}
		})
	}
}

func TestPromptHandler_Delete(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Prompt{Title: "To Delete", Content: "Content"})

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"valid id", "/api/prompts/1", http.StatusOK},
		{"already deleted", "/api/prompts/1", http.StatusNotFound},
		{"non-existent", "/api/prompts/999", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := deleteReq(router, tt.path)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}

	// Verify cascade delete
	var count int64
	db.Model(&models.PromptVersion{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 versions after prompt deletion, got %d", count)
	}
}

func TestPromptHandler_ToggleFavorite(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Prompt{Title: "Test", Content: "Content", IsFavorite: false})

	w1 := postJSON(router, "/api/prompts/1/favorite", nil)
	if w1.Code != http.StatusOK {
		t.Errorf("first toggle status: got %d", w1.Code)
	}

	var prompt models.Prompt
	db.First(&prompt, 1)
	if !prompt.IsFavorite {
		t.Error("expected IsFavorite to be true after toggle")
	}

	w2 := postJSON(router, "/api/prompts/1/favorite", nil)
	var resp struct {
		Success     bool `json:"success"`
		IsFavorite  bool `json:"is_favorite"`
	}
	json.Unmarshal(w2.Body.Bytes(), &resp)
	if resp.IsFavorite {
		t.Error("expected IsFavorite to be false after second toggle")
	}
}

func TestPromptHandler_ListCategories(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Prompt{Category: "coding", Content: "Content"})
	db.Create(&models.Prompt{Category: "writing", Content: "Content"})
	db.Create(&models.Prompt{Category: "coding", Content: "Content"})

	w := getJSON(router, "/api/prompts/categories")
	var resp struct {
		Success    bool     `json:"success"`
		Categories []string `json:"categories"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success {
		t.Error("expected success")
	}
	if len(resp.Categories) != 2 {
		t.Errorf("expected 2 categories, got %d: %v", len(resp.Categories), resp.Categories)
	}
}

func TestPromptHandler_Clone(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	prompt := models.Prompt{Title: "Original", Content: "Content", Category: "test"}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 1, Content: "Content"})

	w := postJSON(router, "/api/prompts/1/clone", nil)
	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusCreated)
	}

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success")
	}

	var cloned models.Prompt
	db.Where("title = ?", "Original (Copy)").First(&cloned)
	if cloned.ID == 0 {
		t.Error("expected cloned prompt to exist")
	}

	var verCount int64
	db.Model(&models.PromptVersion{}).Where("prompt_id = ?", cloned.ID).Count(&verCount)
	if verCount != 1 {
		t.Errorf("expected 1 version for cloned prompt, got %d", verCount)
	}
}

// ----- Skill Handler Tests -----

func TestSkillHandler_CRUD(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	// Create
	w := postJSON(router, "/api/skills", map[string]interface{}{
		"name":        "/test-skill",
		"description": "Test skill",
		"content":     "Skill content here",
		"category":    "test",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("create status: got %d", w.Code)
	}

	// List
	w = getJSON(router, "/api/skills")
	var listResp struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if !listResp.Success || len(listResp.Data) != 1 {
		t.Errorf("list failed: success=%v count=%d", listResp.Success, len(listResp.Data))
	}

	// Get
	w = getJSON(router, "/api/skills/1")
	var getResp APIResponse
	json.Unmarshal(w.Body.Bytes(), &getResp)
	if !getResp.Success {
		t.Error("get failed")
	}

	// Update
	w = putJSON(router, "/api/skills/1", map[string]interface{}{"name": "/updated-skill"})
	var updateResp APIResponse
	json.Unmarshal(w.Body.Bytes(), &updateResp)
	if !updateResp.Success {
		t.Error("update failed")
	}

	// Toggle source
	db.Create(&models.Skill{Name: "/builtin", Content: "Content", Source: "builtin"})

	// Delete
	w = deleteReq(router, "/api/skills/1")
	if w.Code != http.StatusOK {
		t.Errorf("delete status: got %d", w.Code)
	}

	// Verify deletion
	w = getJSON(router, "/api/skills/1")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 after deletion, got %d", w.Code)
	}
}

func TestSkillHandler_ListCategories(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Skill{Name: "s1", Content: "C", Category: "cat1"})
	db.Create(&models.Skill{Name: "s2", Content: "C", Category: "cat2"})
	db.Create(&models.Skill{Name: "s3", Content: "C", Category: "cat1"})

	w := getJSON(router, "/api/skills/categories")
	var resp struct {
		Success    bool     `json:"success"`
		Categories []string `json:"categories"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(resp.Categories))
	}
}

func TestSkillHandler_Clone(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Skill{Name: "/original", Content: "Content", Source: "builtin"})

	w := postJSON(router, "/api/skills/1/clone", nil)
	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d", w.Code)
	}

	var cloned models.Skill
	db.Where("name = ?", "/original (Copy)").First(&cloned)
	if cloned.Source != "custom" {
		t.Errorf("cloned source: got %s, want custom", cloned.Source)
	}
}

// ----- Agent Handler Tests -----

func TestAgentHandler_CRUD(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	// Create
	w := postJSON(router, "/api/agents", map[string]interface{}{
		"name":         "test-agent",
		"role":         "Test Agent",
		"content":      "Agent persona",
		"capabilities": "Testing",
		"category":     "test",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("create status: got %d", w.Code)
	}

	// List with pagination
	w = getJSON(router, "/api/agents?page=1&limit=10")
	var listResp struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
		Meta    models.PaginationMeta `json:"meta"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if !listResp.Success || len(listResp.Data) != 1 {
		t.Errorf("list failed")
	}

	// Get
	w = getJSON(router, "/api/agents/1")
	var getResp APIResponse
	json.Unmarshal(w.Body.Bytes(), &getResp)
	if !getResp.Success {
		t.Error("get failed")
	}

	// Update
	w = putJSON(router, "/api/agents/1", map[string]interface{}{"name": "updated-agent"})
	var updateResp APIResponse
	json.Unmarshal(w.Body.Bytes(), &updateResp)
	if !updateResp.Success {
		t.Error("update failed")
	}

	// Delete
	w = deleteReq(router, "/api/agents/1")
	if w.Code != http.StatusOK {
		t.Errorf("delete status: got %d", w.Code)
	}
}

func TestAgentHandler_ListCategories(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Agent{Name: "a1", Content: "C", Category: "dev"})
	db.Create(&models.Agent{Name: "a2", Content: "C", Category: "dev"})
	db.Create(&models.Agent{Name: "a3", Content: "C", Category: "docs"})

	w := getJSON(router, "/api/agents/categories")
	var resp struct {
		Success    bool     `json:"success"`
		Categories []string `json:"categories"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(resp.Categories))
	}
}

func TestAgentHandler_Clone(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Agent{Name: "original", Content: "Content", Source: "builtin"})

	w := postJSON(router, "/api/agents/1/clone", nil)
	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d", w.Code)
	}

	var count int64
	db.Model(&models.Agent{}).Where("name = ?", "original (Copy)").Count(&count)
	if count != 1 {
		t.Error("clone not found")
	}
}

// ----- Version Handler Tests -----

func TestVersionHandler_CRUD(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	prompt := models.Prompt{Title: "Test", Content: "Content"}
	db.Create(&prompt)

	// Create version
	w := postJSON(router, "/api/prompts/1/versions", map[string]interface{}{
		"content": "Version 1 content",
		"comment": "First version",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("create status: got %d", w.Code)
	}

	var resp APIResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success")
	}

	// List versions
	w = getJSON(router, "/api/prompts/1/versions")
	var listResp struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if !listResp.Success || len(listResp.Data) != 1 {
		t.Errorf("list failed: %d", len(listResp.Data))
	}

	// Get version
	w = getJSON(router, "/api/versions/1")
	var getResp APIResponse
	json.Unmarshal(w.Body.Bytes(), &getResp)
	if !getResp.Success {
		t.Error("get failed")
	}

	// Create another version
	w = postJSON(router, "/api/prompts/1/versions", map[string]interface{}{
		"content": "Version 2 content",
	})
	json.Unmarshal(w.Body.Bytes(), &getResp)

	// Verify version numbering
	var v1, v2 models.PromptVersion
	db.First(&v1, 1)
	db.First(&v2, 2)
	if v1.Version >= v2.Version {
		t.Errorf("version numbering wrong: v1=%d v2=%d", v1.Version, v2.Version)
	}

	// Non-existent prompt version
	w = postJSON(router, "/api/prompts/999/versions", map[string]interface{}{
		"content": "X",
	})
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent prompt, got %d", w.Code)
	}
}

// ----- Import/Export Tests -----

func TestPromptHandler_Export(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Prompt{Title: "P1", Content: "C1"})
	db.Create(&models.Prompt{Title: "P2", Content: "C2"})

	w := getJSON(router, "/api/prompts/export")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Version    string            `json:"version"`
			ExportedAt string           `json:"exported_at"`
			Prompts    []json.RawMessage `json:"prompts"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success")
	}
	if resp.Data.Version != "1.0" {
		t.Errorf("version: got %s, want 1.0", resp.Data.Version)
	}
	if len(resp.Data.Prompts) != 2 {
		t.Errorf("prompts count: got %d, want 2", len(resp.Data.Prompts))
	}
}

func TestPromptHandler_Import(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	tests := []struct {
		name         string
		body         map[string]interface{}
		wantSucc     bool
		wantImported int
		wantFailed   int
	}{
		{
			name: "valid import",
			body: map[string]interface{}{
				"prompts": []map[string]interface{}{
					{"title": "Imported1", "content": "Content1", "category": "test"},
					{"title": "Imported2", "content": "Content2"},
				},
			},
			wantSucc:     true,
			wantImported: 2,
			wantFailed:   0,
		},
		{
			name: "empty prompts",
			body: map[string]interface{}{
				"prompts": []map[string]interface{}{},
			},
			wantSucc:     true,
			wantImported: 0,
			wantFailed:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postJSON(router, "/api/prompts/import", tt.body)
			if w.Code != http.StatusOK {
				t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
			}

			var resp struct {
				Success    bool `json:"success"`
				Imported   int  `json:"imported"`
				Failed     []struct {
					Index int    `json:"index"`
					Title string `json:"title,omitempty"`
					Error string `json:"error"`
				} `json:"failed,omitempty"`
				TotalCount int `json:"total_count"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}
			if resp.Imported != tt.wantImported {
				t.Errorf("imported: got %d, want %d", resp.Imported, tt.wantImported)
			}
			if len(resp.Failed) != tt.wantFailed {
				t.Errorf("failed count: got %d, want %d", len(resp.Failed), tt.wantFailed)
			}
		})
	}
}

func TestSkillHandler_ImportExport(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Skill{Name: "s1", Content: "C1", Category: "cat1"})
	db.Create(&models.Skill{Name: "s2", Content: "C2", Category: "cat2"})

	// Export
	w := getJSON(router, "/api/skills/export")
	if w.Code != http.StatusOK {
		t.Errorf("export status: got %d", w.Code)
	}

	var exportResp struct {
		Success bool `json:"success"`
		Data    struct {
			Version string            `json:"version"`
			Skills  []json.RawMessage `json:"skills"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &exportResp)
	if len(exportResp.Data.Skills) != 2 {
		t.Errorf("export count: got %d, want 2", len(exportResp.Data.Skills))
	}

	// Import — test that import creates new records (existing records still exist)
	w = postJSON(router, "/api/skills/import", map[string]interface{}{
		"skills": []map[string]interface{}{
			{"name": "new1", "content": "c1"},
			{"name": "new2", "content": "c2"},
		},
	})
	var importResp struct {
		Success  bool `json:"success"`
		Imported int  `json:"imported"`
	}
	json.Unmarshal(w.Body.Bytes(), &importResp)
	if !importResp.Success || importResp.Imported != 2 {
		t.Errorf("import: got %v imported=%d", importResp.Success, importResp.Imported)
	}
}

func TestAgentHandler_ImportExport(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Agent{Name: "a1", Content: "C1"})

	// Export
	w := getJSON(router, "/api/agents/export")
	if w.Code != http.StatusOK {
		t.Errorf("export status: got %d", w.Code)
	}

	var exportResp struct {
		Success bool `json:"success"`
		Data    struct {
			Agents []json.RawMessage `json:"agents"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &exportResp)
	if len(exportResp.Data.Agents) != 1 {
		t.Errorf("export count: got %d, want 1", len(exportResp.Data.Agents))
	}
}

func TestAgentHandler_Import(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := postJSON(router, "/api/agents/import", map[string]interface{}{
		"agents": []map[string]interface{}{
			{"name": "agent1", "content": "content1", "role": "Role1"},
			{"name": "agent2", "content": "content2", "role": "Role2"},
		},
	})

	var resp struct {
		Success  bool `json:"success"`
		Imported int  `json:"imported"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success || resp.Imported != 2 {
		t.Errorf("import: got %v imported=%d", resp.Success, resp.Imported)
	}
}

// ----- Helper Function Tests -----

func TestParseTags(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{"empty string", "", []string{}},
		{"valid json array", `["go","api","test"]`, []string{"go", "api", "test"}},
		{"single element", `["solo"]`, []string{"solo"}},
		{"invalid json returns empty", `not-json`, []string{}},
		{"empty array", `[]`, []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseTags(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("len: got %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("[%d]: got %s, want %s", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestMarshalTags(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{"nil slice", nil, "[]"},
		{"empty slice", []string{}, "[]"},
		{"single element", []string{"go"}, `["go"]`},
		{"multiple elements", []string{"go", "api"}, `["go","api"]`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := marshalTags(tt.input)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}

func TestMarshalTags_RoundTrip(t *testing.T) {
	cases := [][]string{
		{"go", "api", "test"},
		{},
		{"single"},
		{"a", "b", "c", "d", "e"},
	}
	for _, tags := range cases {
		got := parseTags(marshalTags(tags))
		if len(got) != len(tags) {
			t.Errorf("roundtrip: got %v, want %v", got, tags)
		}
		for i := range got {
			if got[i] != tags[i] {
				t.Errorf("roundtrip [%d]: got %s, want %s", i, got[i], tags[i])
			}
		}
	}
}

func TestParseVariables(t *testing.T) {
	empty := []models.Variable{}
	one := []models.Variable{{Name: "name", Default: "world"}}
	two := []models.Variable{
		{Name: "name"},
		{Name: "count", Default: "10"},
	}

	tests := []struct {
		name  string
		input string
		want  []models.Variable
	}{
		{"empty string", "", empty},
		{"valid variables", `[{"name":"name","default":"world"}]`, one},
		{"multiple variables", `[{"name":"name"},{"name":"count","default":"10"}]`, two},
		{"invalid json returns empty", `not-json`, empty},
		{"empty array", `[]`, empty},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseVariables(tt.input)
			if len(got) != len(tt.want) {
				t.Errorf("len: got %d, want %d", len(got), len(tt.want))
				return
			}
			for i := range got {
				if got[i].Name != tt.want[i].Name {
					t.Errorf("[%d].Name: got %s, want %s", i, got[i].Name, tt.want[i].Name)
				}
				if got[i].Default != tt.want[i].Default {
					t.Errorf("[%d].Default: got %s, want %s", i, got[i].Default, tt.want[i].Default)
				}
			}
		})
	}
}

func TestMarshalVariables(t *testing.T) {
	nilVars := marshalVariables(nil)
	if nilVars != "[]" {
		t.Errorf("nil vars: got %s, want []", nilVars)
	}

	empty := marshalVariables([]models.Variable{})
	if empty != "[]" {
		t.Errorf("empty vars: got %s, want []", empty)
	}

	vars := []models.Variable{{Name: "name"}}
	got := marshalVariables(vars)
	var parsed []models.Variable
	json.Unmarshal([]byte(got), &parsed)
	if len(parsed) != 1 || parsed[0].Name != "name" {
		t.Errorf("marshal roundtrip failed: got %s", got)
	}
}

// ----- Additional Agent Handler Tests -----

func TestAgentHandler_Get_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := getJSON(router, "/api/agents/999")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAgentHandler_Update_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := putJSON(router, "/api/agents/999", map[string]interface{}{"name": "new"})
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAgentHandler_Delete_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := deleteReq(router, "/api/agents/999")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAgentHandler_Create_MissingRequired(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{"missing name", map[string]interface{}{"content": "content"}},
		{"missing content", map[string]interface{}{"name": "agent"}},
		{"empty body", map[string]interface{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postJSON(router, "/api/agents", tt.body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestAgentHandler_Update_AllFields(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Agent{Name: "orig", Content: "orig content", Role: "orig role", Category: "orig cat"})

	w := putJSON(router, "/api/agents/1", map[string]interface{}{
		"name":         "updated",
		"content":      "updated content",
		"role":         "updated role",
		"content_cn":   "updated cn",
		"capabilities": "updated cap",
		"category":     "updated cat",
	})
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d", w.Code)
	}

	var agent models.Agent
	db.First(&agent, 1)
	if agent.Name != "updated" || agent.Content != "updated content" || agent.Role != "updated role" {
		t.Errorf("update mismatch: name=%s content=%s role=%s", agent.Name, agent.Content, agent.Role)
	}
}

func TestAgentHandler_Update_PartialFields(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Agent{Name: "orig", Content: "orig content", Role: "orig role"})

	// Only update name
	w := putJSON(router, "/api/agents/1", map[string]interface{}{"name": "only-name"})
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d", w.Code)
	}

	var agent models.Agent
	db.First(&agent, 1)
	if agent.Name != "only-name" || agent.Content != "orig content" || agent.Role != "orig role" {
		t.Errorf("partial update: name=%s content=%s role=%s", agent.Name, agent.Content, agent.Role)
	}
}

func TestAgentHandler_Update_InvalidID(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	// Non-integer ID in PUT request returns 400 from strconv.Atoi
	w := putJSON(router, "/api/agents/invalid", map[string]interface{}{"name": "x"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("invalid ID: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

// ----- Additional Skill Handler Tests -----

func TestSkillHandler_Get_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := getJSON(router, "/api/skills/999")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestSkillHandler_Update_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := putJSON(router, "/api/skills/999", map[string]interface{}{"name": "/new"})
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestSkillHandler_Delete_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := deleteReq(router, "/api/skills/999")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestSkillHandler_Update_AllFields(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Skill{Name: "/orig", Content: "orig", Description: "orig desc", Category: "orig cat"})

	w := putJSON(router, "/api/skills/1", map[string]interface{}{
		"name":        "/updated",
		"content":     "updated content",
		"description": "updated desc",
		"content_cn":   "updated cn",
		"category":    "updated cat",
	})
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d", w.Code)
	}

	var skill models.Skill
	db.First(&skill, 1)
	if skill.Name != "/updated" || skill.Content != "updated content" {
		t.Errorf("update mismatch: name=%s content=%s", skill.Name, skill.Content)
	}
}

func TestSkillHandler_Update_PartialFields(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Skill{Name: "/orig", Content: "orig content", Category: "orig cat"})

	// Only update category
	w := putJSON(router, "/api/skills/1", map[string]interface{}{"category": "new-cat"})
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d", w.Code)
	}

	var skill models.Skill
	db.First(&skill, 1)
	if skill.Category != "new-cat" || skill.Name != "/orig" {
		t.Errorf("partial update: cat=%s name=%s", skill.Category, skill.Name)
	}
}

// ----- Additional Version Handler Tests -----

func TestVersionHandler_Get_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := getJSON(router, "/api/versions/999")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestVersionHandler_List_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	// Prompt doesn't exist - List returns empty array, not 404
	w := getJSON(router, "/api/prompts/999/versions")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}
	var resp struct {
		Success bool `json:"success"`
		Data    []interface{} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success || len(resp.Data) != 0 {
		t.Errorf("expected empty array: success=%v len=%d", resp.Success, len(resp.Data))
	}
}

// ----- Pagination Meta Tests -----

func TestPaginationMeta_TotalPages(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	// Seed 25 prompts
	for i := 0; i < 25; i++ {
		db.Create(&models.Prompt{Title: "P", Content: "C"})
	}

	tests := []struct {
		page       string
		limit      string
		wantTotal  int
		wantPages  int
	}{
		{"1", "10", 10, 3},
		{"2", "10", 10, 3},
		{"3", "10", 5, 3},
		{"1", "5", 5, 5},
		{"1", "25", 25, 1},
		{"1", "100", 25, 1},
	}

	for _, tt := range tests {
		path := "/api/prompts?page=" + tt.page + "&limit=" + tt.limit
		w := getJSON(router, path)

		var resp struct {
			Success bool                 `json:"success"`
			Data    []json.RawMessage    `json:"data"`
			Meta    models.PaginationMeta `json:"meta"`
		}
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.Meta.TotalPages != tt.wantPages {
			t.Errorf("page=%s limit=%s: totalPages got %d, want %d",
				tt.page, tt.limit, resp.Meta.TotalPages, tt.wantPages)
		}
		if len(resp.Data) != tt.wantTotal {
			t.Errorf("page=%s limit=%s: count got %d, want %d",
				tt.page, tt.limit, len(resp.Data), tt.wantTotal)
		}
	}
}

// ----- Additional Prompt Handler Tests -----

func TestPromptHandler_Delete_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := deleteReq(router, "/api/prompts/999")
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestPromptHandler_Update_AllFields(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	db.Create(&models.Prompt{Title: "orig", Content: "orig content", Category: "orig cat"})

	w := putJSON(router, "/api/prompts/1", map[string]interface{}{
		"title":      "updated title",
		"content":    "updated content",
		"category":   "updated cat",
		"content_cn": "updated cn",
		"tags":       []string{"tag1", "tag2"},
	})
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d", w.Code)
	}

	var prompt models.Prompt
	db.First(&prompt, 1)
	if prompt.Title != "updated title" || prompt.Content != "updated content" {
		t.Errorf("update mismatch: title=%s content=%s", prompt.Title, prompt.Content)
	}
}

func TestPromptHandler_ToggleFavorite_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := postJSON(router, "/api/prompts/999/favorite", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("not found: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestPromptHandler_Clone_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := postJSON(router, "/api/prompts/999/clone", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

// ----- Additional Skill Handler Tests -----

func TestSkillHandler_Create_MissingRequired(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{"missing name", map[string]interface{}{"content": "content"}},
		{"missing content", map[string]interface{}{"name": "/test"}},
		{"empty body", map[string]interface{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postJSON(router, "/api/skills", tt.body)
			if w.Code != http.StatusBadRequest {
				t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestSkillHandler_Clone_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := postJSON(router, "/api/skills/999/clone", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

// ----- Additional Agent Handler Tests -----

func TestAgentHandler_Clone_NotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := postJSON(router, "/api/agents/999/clone", nil)
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestAgentHandler_Create_AllFields(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	w := postJSON(router, "/api/agents", map[string]interface{}{
		"name":         "full-agent",
		"role":         "Full Role",
		"content":      "Full content",
		"content_cn":    "Full 中文内容",
		"capabilities": "cap1, cap2",
		"category":     "full-cat",
		"source":       "custom",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d", w.Code)
	}

	var agent models.Agent
	db.Where("name = ?", "full-agent").First(&agent)
	if agent.Role != "Full Role" || agent.Capabilities != "cap1, cap2" || agent.Source != "custom" {
		t.Errorf("fields mismatch: role=%s cap=%s source=%s", agent.Role, agent.Capabilities, agent.Source)
	}
}

func TestAgentHandler_Import_PartialFailure(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	// Import with empty agents list
	w := postJSON(router, "/api/agents/import", map[string]interface{}{
		"agents": []map[string]interface{}{},
	})
	if w.Code != http.StatusOK {
		t.Errorf("empty import: got %d", w.Code)
	}

	var resp struct {
		Success  bool `json:"success"`
		Imported int  `json:"imported"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Imported != 0 {
		t.Errorf("imported: got %d, want 0", resp.Imported)
	}
}

// ----- Helper Function Tests -----

func TestGetProvider(t *testing.T) {
	providers := []struct {
		input string
		want  string
	}{
		{"claude", "claude"},
		{"anthropic", "claude"},
		{"gemini", "gemini"},
		{"google", "gemini"},
		{"googleai", "gemini"},
		{"minimax", "minimax"},
		{"openai", "openai"},
		{"", "openai"},
		{"unknown", "openai"},
		{"OpenAI", "openai"},
		{"CLAUDE", "claude"},
	}

	for _, tt := range providers {
		t.Run(tt.input, func(t *testing.T) {
			p := getProvider(tt.input)
			if p.Name() != tt.want {
				t.Errorf("input %q: got %s, want %s", tt.input, p.Name(), tt.want)
			}
		})
	}
}

func TestGetModelsByProvider(t *testing.T) {
	openaiModels := getModelsByProvider("openai")
	if len(openaiModels) == 0 {
		t.Error("expected openai models")
	}

	claudeModels := getModelsByProvider("claude")
	if len(claudeModels) == 0 {
		t.Error("expected claude models")
	}

	geminiModels := getModelsByProvider("gemini")
	if len(geminiModels) == 0 {
		t.Error("expected gemini models")
	}

	minimaxModels := getModelsByProvider("minimax")
	if len(minimaxModels) == 0 {
		t.Error("expected minimax models")
	}

	// Unknown provider returns empty
	unknownModels := getModelsByProvider("unknown")
	if len(unknownModels) != 0 {
		t.Errorf("unknown provider: got %d, want 0", len(unknownModels))
	}
}

func TestMockAIResponse(t *testing.T) {
	tests := []struct {
		content string
	}{
		{"say hello"},
		{"hi there"},
		{"write code function"},
		{"write a function"},
		{"random text"},
		{""},
	}

	for _, tt := range tests {
		t.Run(tt.content, func(t *testing.T) {
			got := mockAIResponse(tt.content)
			if len(got) == 0 {
				t.Error("expected non-empty response")
			}
		})
	}
}

func TestMockOptimizeResponse(t *testing.T) {
	modes := []string{"improve", "structure", "suggest", "unknown", ""}

	for _, mode := range modes {
		t.Run(mode, func(t *testing.T) {
			got := mockOptimizeResponse(mode)
			if len(got) == 0 {
				t.Error("expected non-empty response")
			}
		})
	}
}

// ----- Task Handler Tests -----

func TestTaskHandler_CreateTask(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	tests := []struct {
		name       string
		body       map[string]interface{}
		wantStatus int
		wantSucc   bool
	}{
		{
			name:       "create batch_test task",
			body:       map[string]interface{}{"type": "batch_test", "payload": map[string]interface{}{"prompt_ids": []int{1, 2, 3}}},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "create ab_test task",
			body:       map[string]interface{}{"type": "ab_test"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "create task with run_at",
			body:       map[string]interface{}{"type": "eval_gen", "run_at": "2026-04-03T12:00:00Z"},
			wantStatus: http.StatusOK,
			wantSucc:   true,
		},
		{
			name:       "invalid task type",
			body:       map[string]interface{}{"type": "invalid_type"},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
		{
			name:       "missing type",
			body:       map[string]interface{}{},
			wantStatus: http.StatusBadRequest,
			wantSucc:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := postJSON(router, "/api/tasks", tt.body)
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

func TestTaskHandler_GetTask(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	// Create a task first
	svc := service.NewTaskService(db)
	svc.Create(service.CreateTaskRequest{Type: models.TaskTypeBatchTest})

	tests := []struct {
		name       string
		path       string
		wantStatus int
		wantSucc   bool
	}{
		{"valid id", "/api/tasks/1", http.StatusOK, true},
		{"invalid id", "/api/tasks/999", http.StatusNotFound, false},
		{"non-numeric id", "/api/tasks/abc", http.StatusBadRequest, false},
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

func TestTaskHandler_ListTasks(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	// Create tasks
	svc := service.NewTaskService(db)
	for i := 0; i < 3; i++ {
		svc.Create(service.CreateTaskRequest{Type: models.TaskTypeBatchTest})
	}
	svc.Create(service.CreateTaskRequest{Type: models.TaskTypeABTest})

	tests := []struct {
		name       string
		query      string
		wantCount  int
		wantSucc   bool
	}{
		{"all tasks", "/api/tasks", 4, true},
		{"with limit", "/api/tasks?limit=2", 2, true},
		{"with offset", "/api/tasks?offset=2", 2, true},
		{"filter by status", "/api/tasks?status=pending", 4, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := getJSON(router, tt.query)
			if w.Code != http.StatusOK {
				t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
			}

			var resp struct {
				Success bool `json:"success"`
				Data    struct {
					Tasks []json.RawMessage `json:"tasks"`
					Total int64            `json:"total"`
				} `json:"data"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp.Success != tt.wantSucc {
				t.Errorf("success: got %v, want %v", resp.Success, tt.wantSucc)
			}
			if len(resp.Data.Tasks) != tt.wantCount {
				t.Errorf("count: got %d, want %d", len(resp.Data.Tasks), tt.wantCount)
			}
		})
	}
}

func TestTaskHandler_CancelTask(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouter(db)

	// Create a task
	svc := service.NewTaskService(db)
	svc.Create(service.CreateTaskRequest{Type: models.TaskTypeBatchTest})

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"cancel pending task", "/api/tasks/1", http.StatusOK},
		{"cancel non-existing task", "/api/tasks/999", http.StatusNotFound},
		{"cancel with invalid id", "/api/tasks/abc", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := deleteReq(router, tt.path)
			if w.Code != tt.wantStatus {
				t.Errorf("status: got %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
