package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func newServiceDB(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(tmpDir+"/service_test.db?_busy_timeout=10000"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(
		&models.Prompt{},
		&models.PromptVersion{},
		&models.TestRecord{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

// ----- EnsureVersion -----

func TestEnsureVersion_NoExistingVersions(t *testing.T) {
	db := newServiceDB(t)
	svc := NewPromptService(db)

	prompt := models.Prompt{ID: 1, Title: "Test", Content: "Content"}
	db.Create(&prompt)

	created, version, err := svc.EnsureVersion(1, "New Content", "comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Error("expected version to be created")
	}
	if version != 1 {
		t.Errorf("version: got %d, want 1", version)
	}

	var v models.PromptVersion
	db.Where("prompt_id = ?", 1).First(&v)
	if v.Content != "New Content" {
		t.Errorf("version content: got %s, want %s", v.Content, "New Content")
	}
	if v.Comment != "comment" {
		t.Errorf("version comment: got %s, want %s", v.Comment, "comment")
	}
}

func TestEnsureVersion_ContentUnchanged(t *testing.T) {
	db := newServiceDB(t)
	svc := NewPromptService(db)

	prompt := models.Prompt{ID: 1, Title: "Test", Content: "Same Content"}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: 1, Version: 1, Content: "Same Content"})

	created, version, err := svc.EnsureVersion(1, "Same Content", "comment")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if created {
		t.Error("expected no version to be created when content unchanged")
	}
	if version != 1 {
		t.Errorf("version: got %d, want 1", version)
	}

	var count int64
	db.Model(&models.PromptVersion{}).Count(&count)
	if count != 1 {
		t.Errorf("version count: got %d, want 1", count)
	}
}

func TestEnsureVersion_ContentChanged(t *testing.T) {
	db := newServiceDB(t)
	svc := NewPromptService(db)

	prompt := models.Prompt{ID: 1, Title: "Test", Content: "Old Content"}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: 1, Version: 1, Content: "Old Content"})

	created, version, err := svc.EnsureVersion(1, "New Content", "updated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !created {
		t.Error("expected version to be created")
	}
	if version != 2 {
		t.Errorf("version: got %d, want 2", version)
	}

	var count int64
	db.Model(&models.PromptVersion{}).Count(&count)
	if count != 2 {
		t.Errorf("version count: got %d, want 2", count)
	}
}

func TestEnsureVersion_PromptNotFound(t *testing.T) {
	db := newServiceDB(t)
	svc := NewPromptService(db)

	_, _, err := svc.EnsureVersion(999, "Content", "")
	if err == nil {
		t.Error("expected error for non-existent prompt")
	}
}

// ----- DeleteWithVersionsAndTests -----

func TestDeleteWithVersionsAndTests(t *testing.T) {
	db := newServiceDB(t)
	svc := NewPromptService(db)

	prompt := models.Prompt{ID: 1, Title: "Test", Content: "Content"}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: 1, Version: 1, Content: "Content"})
	db.Create(&models.PromptVersion{PromptID: 1, Version: 2, Content: "Content2"})
	db.Create(&models.TestRecord{PromptID: 1, Model: "gpt-4o", PromptText: "in", Response: "out", TokensUsed: 100})

	err := svc.DeleteWithVersionsAndTests(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var promptCount int64
	db.Model(&models.Prompt{}).Count(&promptCount)
	if promptCount != 0 {
		t.Errorf("prompt count: got %d, want 0", promptCount)
	}

	var versionCount int64
	db.Model(&models.PromptVersion{}).Count(&versionCount)
	if versionCount != 0 {
		t.Errorf("version count: got %d, want 0", versionCount)
	}

	var testCount int64
	db.Model(&models.TestRecord{}).Count(&testCount)
	if testCount != 0 {
		t.Errorf("test count: got %d, want 0", testCount)
	}
}

func TestDeleteWithVersionsAndTests_NotFound(t *testing.T) {
	db := newServiceDB(t)
	svc := NewPromptService(db)

	err := svc.DeleteWithVersionsAndTests(999)
	if err == nil {
		t.Error("expected error for non-existent prompt")
	}
}

// ----- CountVersions -----

func TestCountVersions(t *testing.T) {
	db := newServiceDB(t)
	svc := NewPromptService(db)

	prompt := models.Prompt{ID: 1, Title: "Test", Content: "Content"}
	db.Create(&prompt)
	db.Create(&models.PromptVersion{PromptID: 1, Version: 1, Content: "C1"})
	db.Create(&models.PromptVersion{PromptID: 1, Version: 2, Content: "C2"})
	db.Create(&models.PromptVersion{PromptID: 1, Version: 3, Content: "C3"})

	count, err := svc.CountVersions(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("count: got %d, want 3", count)
	}
}

func TestCountVersions_NoVersions(t *testing.T) {
	db := newServiceDB(t)
	svc := NewPromptService(db)

	prompt := models.Prompt{ID: 1, Title: "Test", Content: "Content"}
	db.Create(&prompt)

	count, err := svc.CountVersions(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("count: got %d, want 0", count)
	}
}

func TestCountVersions_NonExistent(t *testing.T) {
	db := newServiceDB(t)
	svc := NewPromptService(db)

	count, err := svc.CountVersions(999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("count: got %d, want 0", count)
	}
}
