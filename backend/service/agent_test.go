package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func newAgentDB(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(tmpDir+"/agent_test.db?_busy_timeout=10000"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&models.Agent{})
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

// ----- Clone -----

func TestAgentService_Clone(t *testing.T) {
	db := newAgentDB(t)
	svc := NewAgentService(db)

	db.Create(&models.Agent{
		Name:         "test-agent",
		Role:         "Tester",
		Content:      "Agent content",
		ContentCN:    "内容",
		Capabilities: "Testing",
		Category:    "qa",
		Source:      "builtin",
	})

	clone, err := svc.Clone(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clone.Name != "test-agent (Copy)" {
		t.Errorf("name: got %s, want test-agent (Copy)", clone.Name)
	}
	if clone.Source != "custom" {
		t.Errorf("source: got %s, want custom", clone.Source)
	}
	if clone.Role != "Tester" {
		t.Errorf("role: got %s", clone.Role)
	}
	if clone.Capabilities != "Testing" {
		t.Errorf("capabilities: got %s", clone.Capabilities)
	}
}

func TestAgentService_CloneWithActivity(t *testing.T) {
	db := newAgentDB(t)
	svc := NewAgentService(db)

	db.Create(&models.Agent{Name: "agent", Content: "Content", Source: "builtin"})

	clone, details, err := svc.CloneWithActivity(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details != `{"from_id": 1}` {
		t.Errorf("details: got %s", details)
	}
	if clone.Name != "agent (Copy)" {
		t.Errorf("name: got %s", clone.Name)
	}
}

func TestAgentService_Clone_NotFound(t *testing.T) {
	db := newAgentDB(t)
	svc := NewAgentService(db)

	_, err := svc.Clone(999)
	if err == nil {
		t.Error("expected error for non-existent agent")
	}
}

// ----- Delete -----

func TestAgentService_Delete(t *testing.T) {
	db := newAgentDB(t)
	svc := NewAgentService(db)

	db.Create(&models.Agent{Name: "test", Content: "Content"})

	err := svc.Delete(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int64
	db.Model(&models.Agent{}).Count(&count)
	if count != 0 {
		t.Errorf("count after delete: got %d, want 0", count)
	}
}

func TestAgentService_Delete_NotFound(t *testing.T) {
	db := newAgentDB(t)
	svc := NewAgentService(db)

	err := svc.Delete(999)
	if err == nil {
		t.Error("expected error for non-existent agent")
	}
}

// ----- Count -----

func TestAgentService_Count(t *testing.T) {
	db := newAgentDB(t)
	svc := NewAgentService(db)

	db.Create(&models.Agent{Name: "a1", Content: "C1"})
	db.Create(&models.Agent{Name: "a2", Content: "C2"})
	db.Create(&models.Agent{Name: "a3", Content: "C3"})

	count, err := svc.Count()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("count: got %d, want 3", count)
	}
}

// ----- GetByID -----

func TestAgentService_GetByID(t *testing.T) {
	db := newAgentDB(t)
	svc := NewAgentService(db)

	db.Create(&models.Agent{Name: "test", Content: "Content", Role: "Tester"})

	agent, err := svc.GetByID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.Name != "test" {
		t.Errorf("name: got %s", agent.Name)
	}
	if agent.Role != "Tester" {
		t.Errorf("role: got %s", agent.Role)
	}
}

func TestAgentService_GetByID_NotFound(t *testing.T) {
	db := newAgentDB(t)
	svc := NewAgentService(db)

	_, err := svc.GetByID(999)
	if err == nil {
		t.Error("expected error for non-existent agent")
	}
}
