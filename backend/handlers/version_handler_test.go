package handlers

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func setupTestRouterWithVersionHandler(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewVersionHandler(db)

	api := r.Group("/api")
	api.GET("/prompts/:id/versions", h.List)
	api.POST("/prompts/:id/versions", h.Create)
	api.GET("/versions/:id", h.Get)

	return r
}

// ----- VersionHandler.List Tests -----

func TestVersionHandler_List(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	prompt := models.Prompt{Title: "Test", Content: "Content"}
	db.Create(&prompt)

	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 1, Content: "v1"})
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 2, Content: "v2"})
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 3, Content: "v3"})

	w := getJSON(router, "/api/prompts/1/versions")
	if w.Code != http.StatusOK {
		t.Fatalf("status: got %d", w.Code)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    []struct {
			ID      uint   `json:"id"`
			Version int    `json:"version"`
			Content string `json:"content"`
			Comment string `json:"comment"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Fatal("expected success")
	}
	if len(resp.Data) != 3 {
		t.Errorf("count: got %d, want 3", len(resp.Data))
	}
	// Should be ordered by version DESC
	if resp.Data[0].Version != 3 || resp.Data[1].Version != 2 || resp.Data[2].Version != 1 {
		t.Errorf("order wrong: %v", resp.Data)
	}
}

func TestVersionHandler_List_InvalidPromptID(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	w := getJSON(router, "/api/prompts/abc/versions")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestVersionHandler_List_Empty(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	db.Create(&models.Prompt{Title: "Empty", Content: "No versions"})

	w := getJSON(router, "/api/prompts/1/versions")
	var resp struct {
		Success bool `json:"success"`
		Data    []json.RawMessage `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Fatal("expected success")
	}
	if len(resp.Data) != 0 {
		t.Errorf("expected 0, got %d", len(resp.Data))
	}
}

func TestVersionHandler_List_Pagination(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	prompt := models.Prompt{Title: "Test", Content: "Content"}
	db.Create(&prompt)

	for i := 1; i <= 10; i++ {
		db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: i, Content: "v" + itoa(i)})
	}

	tests := []struct {
		query       string
		wantCount   int
		wantTotal   int64
		wantPages   int
	}{
		{"?page=1&limit=3", 3, 10, 4},
		{"?page=2&limit=3", 3, 10, 4},
		{"?page=3&limit=3", 3, 10, 4},
		{"?page=4&limit=3", 1, 10, 4},
		{"?page=1&limit=10", 10, 10, 1},
		{"?page=1&limit=100", 10, 10, 1},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			w := getJSON(router, "/api/prompts/1/versions"+tt.query)
			var resp struct {
				Success bool `json:"success"`
				Data    []json.RawMessage `json:"data"`
				Meta    models.PaginationMeta `json:"meta"`
			}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if len(resp.Data) != tt.wantCount {
				t.Errorf("count: got %d, want %d", len(resp.Data), tt.wantCount)
			}
			if resp.Meta.Total != tt.wantTotal {
				t.Errorf("total: got %d, want %d", resp.Meta.Total, tt.wantTotal)
			}
			if resp.Meta.TotalPages != tt.wantPages {
				t.Errorf("totalPages: got %d, want %d", resp.Meta.TotalPages, tt.wantPages)
			}
		})
	}
}

// ----- VersionHandler.Create Tests -----

func TestVersionHandler_Create(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	prompt := models.Prompt{Title: "Test", Content: "Original"}
	db.Create(&prompt)

	w := postJSON(router, "/api/prompts/1/versions", map[string]interface{}{
		"content": "Version 1 content",
		"comment": "Initial version",
	})
	if w.Code != http.StatusCreated {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusCreated)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			ID        uint   `json:"id"`
			PromptID  uint   `json:"prompt_id"`
			Version   int    `json:"version"`
			Content   string `json:"content"`
			Comment   string `json:"comment"`
			CreatedAt string `json:"created_at"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Fatal("expected success")
	}
	if resp.Data.Version != 1 {
		t.Errorf("version: got %d, want 1", resp.Data.Version)
	}
	if resp.Data.Content != "Version 1 content" {
		t.Errorf("content: got %s", resp.Data.Content)
	}
	if resp.Data.Comment != "Initial version" {
		t.Errorf("comment: got %s", resp.Data.Comment)
	}

	// Verify prompt content was updated
	var updatedPrompt models.Prompt
	db.First(&updatedPrompt, 1)
	if updatedPrompt.Content != "Version 1 content" {
		t.Errorf("prompt content not updated: got %s", updatedPrompt.Content)
	}
}

func TestVersionHandler_Create_SecondVersion(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	prompt := models.Prompt{Title: "Test", Content: "Original"}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: prompt.ID, Version: 1, Content: "First"})

	w := postJSON(router, "/api/prompts/1/versions", map[string]interface{}{
		"content": "Second version content",
	})
	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Version int `json:"version"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Fatal("expected success")
	}
	if resp.Data.Version != 2 {
		t.Errorf("version: got %d, want 2", resp.Data.Version)
	}
}

func TestVersionHandler_Create_InvalidPromptID(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	w := postJSON(router, "/api/prompts/abc/versions", map[string]interface{}{
		"content": "x",
	})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestVersionHandler_Create_PromptNotFound(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	w := postJSON(router, "/api/prompts/999/versions", map[string]interface{}{
		"content": "x",
	})
	if w.Code != http.StatusNotFound {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestVersionHandler_Create_MissingContent(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	db.Create(&models.Prompt{Title: "Test", Content: "Content"})

	w := postJSON(router, "/api/prompts/1/versions", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestVersionHandler_Create_WithoutComment(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	db.Create(&models.Prompt{Title: "Test", Content: "Content"})

	w := postJSON(router, "/api/prompts/1/versions", map[string]interface{}{
		"content": "Some content",
	})
	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Comment string `json:"comment"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Fatal("expected success")
	}
	// Comment should be empty string (zero value)
	if resp.Data.Comment != "" {
		t.Errorf("comment: got %q, want empty", resp.Data.Comment)
	}
}

// ----- VersionHandler.Get Tests -----

func TestVersionHandler_Get(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	prompt := models.Prompt{Title: "Test", Content: "Content"}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{
		PromptID: prompt.ID,
		Version:  1,
		Content:  "Version 1 content",
		Comment:  "Test comment",
	})

	w := getJSON(router, "/api/versions/1")
	if w.Code != http.StatusOK {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			ID        uint   `json:"id"`
			PromptID  uint   `json:"prompt_id"`
			Version   int    `json:"version"`
			Content   string `json:"content"`
			Comment   string `json:"comment"`
			CreatedAt string `json:"created_at"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Fatal("expected success")
	}
	if resp.Data.ID != 1 {
		t.Errorf("id: got %d, want 1", resp.Data.ID)
	}
	if resp.Data.PromptID != 1 {
		t.Errorf("promptID: got %d, want 1", resp.Data.PromptID)
	}
	if resp.Data.Version != 1 {
		t.Errorf("version: got %d, want 1", resp.Data.Version)
	}
	if resp.Data.Content != "Version 1 content" {
		t.Errorf("content: got %s", resp.Data.Content)
	}
	if resp.Data.Comment != "Test comment" {
		t.Errorf("comment: got %s", resp.Data.Comment)
	}
	if resp.Data.CreatedAt == "" {
		t.Error("created_at should not be empty")
	}
}

func TestVersionHandler_Get_InvalidID(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	w := getJSON(router, "/api/versions/abc")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status: got %d, want %d", w.Code, http.StatusBadRequest)
	}
}


func TestVersionHandler_Get_ResponseFormat(t *testing.T) {
	db := newTestDB(t)
	router := setupTestRouterWithVersionHandler(db)

	db.Create(&models.Prompt{Title: "Test", Content: "C"})
	db.Create(&models.PromptVersion{PromptID: 1, Version: 1, Content: "c"})

	w := getJSON(router, "/api/versions/1")
	var resp struct {
		Success bool `json:"success"`
		Error   string `json:"error,omitempty"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("expected success=true")
	}
	if resp.Error != "" {
		t.Errorf("expected no error field, got %s", resp.Error)
	}
}
