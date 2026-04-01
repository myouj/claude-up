package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func newSkillDB(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(tmpDir+"/skill_test.db?_busy_timeout=10000"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&models.Skill{})
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

// ----- Clone -----

func TestSkillService_Clone(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	db.Create(&models.Skill{
		Name:        "/test",
		Description: "Test skill",
		Content:     "Skill content",
		ContentCN:   "内容",
		Category:    "test",
		Source:      "builtin",
	})

	clone, err := svc.Clone(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clone.Name != "/test (Copy)" {
		t.Errorf("name: got %s, want /test (Copy)", clone.Name)
	}
	if clone.Source != "custom" {
		t.Errorf("source: got %s, want custom", clone.Source)
	}
	if clone.Content != "Skill content" {
		t.Errorf("content: got %s, want %s", clone.Content, "Skill content")
	}
	if clone.Description != "Test skill" {
		t.Errorf("description: got %s", clone.Description)
	}
}

func TestSkillService_CloneWithActivity(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	db.Create(&models.Skill{Name: "/test", Content: "Content", Source: "builtin"})

	clone, details, err := svc.CloneWithActivity(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if details != `{"from_id": 1}` {
		t.Errorf("details: got %s, want %s", details, `{"from_id": 1}`)
	}
	if clone.Name != "/test (Copy)" {
		t.Errorf("name: got %s", clone.Name)
	}
}

func TestSkillService_Clone_NotFound(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	_, err := svc.Clone(999)
	if err == nil {
		t.Error("expected error for non-existent skill")
	}
}

// ----- Delete -----

func TestSkillService_Delete(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	db.Create(&models.Skill{Name: "/test", Content: "Content"})

	err := svc.Delete(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int64
	db.Model(&models.Skill{}).Count(&count)
	if count != 0 {
		t.Errorf("count after delete: got %d, want 0", count)
	}
}

func TestSkillService_Delete_NotFound(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	err := svc.Delete(999)
	if err == nil {
		t.Error("expected error for non-existent skill")
	}
}

// ----- Count -----

func TestSkillService_Count(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	db.Create(&models.Skill{Name: "/s1", Content: "C1"})
	db.Create(&models.Skill{Name: "/s2", Content: "C2"})

	count, err := svc.Count()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("count: got %d, want 2", count)
	}
}

// ----- GetByID -----

func TestSkillService_GetByID(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	db.Create(&models.Skill{Name: "/test", Content: "Content", Category: "cat1"})

	skill, err := svc.GetByID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if skill.Name != "/test" {
		t.Errorf("name: got %s", skill.Name)
	}
	if skill.Category != "cat1" {
		t.Errorf("category: got %s", skill.Category)
	}
}

func TestSkillService_GetByID_NotFound(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	_, err := svc.GetByID(999)
	if err == nil {
		t.Error("expected error for non-existent skill")
	}
}

// ----- BatchClone -----

func TestSkillService_BatchClone(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	db.Create(&models.Skill{Name: "/s1", Content: "C1", Category: "cat1", Source: "builtin"})
	db.Create(&models.Skill{Name: "/s2", Content: "C2", Category: "cat2", Source: "builtin"})
	db.Create(&models.Skill{Name: "/s3", Content: "C3", Category: "cat3", Source: "builtin"})

	// Full success: all 3 cloned
	cloned, batchErrs, err := svc.BatchClone([]uint{1, 2, 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cloned) != 3 {
		t.Errorf("cloned count: got %d, want 3", len(cloned))
	}
	if len(batchErrs) != 0 {
		t.Errorf("batch errors: got %d, want 0", len(batchErrs))
	}
	for _, c := range cloned {
		if c.Source != "custom" {
			t.Errorf("source: got %s, want custom", c.Source)
		}
	}
}

func TestSkillService_BatchClone_Partial(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	db.Create(&models.Skill{Name: "/exists", Content: "Content"})

	cloned, batchErrs, err := svc.BatchClone([]uint{1, 999, 888})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cloned) != 1 {
		t.Errorf("cloned count: got %d, want 1", len(cloned))
	}
	if len(batchErrs) != 2 {
		t.Errorf("batch errors: got %d, want 2", len(batchErrs))
	}
}

func TestSkillService_BatchClone_AllNotFound(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	cloned, batchErrs, err := svc.BatchClone([]uint{999, 998})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cloned) != 0 {
		t.Errorf("cloned count: got %d, want 0", len(cloned))
	}
	if len(batchErrs) != 2 {
		t.Errorf("batch errors: got %d, want 2", len(batchErrs))
	}
}

func TestSkillService_BatchClone_EmptySlice(t *testing.T) {
	db := newSkillDB(t)
	svc := NewSkillService(db)

	cloned, batchErrs, err := svc.BatchClone([]uint{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cloned) != 0 {
		t.Errorf("cloned count: got %d, want 0", len(cloned))
	}
	if len(batchErrs) != 0 {
		t.Errorf("batch errors: got %d, want 0", len(batchErrs))
	}
}
