package service

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func newRegressionTestDB(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(tmpDir+"/regression_test.db?_busy_timeout=10000"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(
		&models.Prompt{},
		&models.ResponseCache{},
	)
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

func TestRegressionService_Detect(t *testing.T) {
	db := newRegressionTestDB(t)
	scoringSvc := NewScoringService()
	svc := NewRegressionService(db, scoringSvc)

	prompt := models.Prompt{ID: 1, Title: "Test Prompt", Content: "Test content"}
	db.Create(&prompt)

	report, err := svc.Detect(1, "gpt-4o", "gpt-4o-mini")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.PromptID != 1 {
		t.Errorf("PromptID: got %d, want 1", report.PromptID)
	}
	if report.PromptTitle != "Test Prompt" {
		t.Errorf("PromptTitle: got %s, want %s", report.PromptTitle, "Test Prompt")
	}
	if report.OldModel != "gpt-4o" {
		t.Errorf("OldModel: got %s, want %s", report.OldModel, "gpt-4o")
	}
	if report.NewModel != "gpt-4o-mini" {
		t.Errorf("NewModel: got %s, want %s", report.NewModel, "gpt-4o-mini")
	}
	if report.OldScore < 0 || report.OldScore > 1 {
		t.Errorf("OldScore: got %f, expected between 0 and 1", report.OldScore)
	}
	if report.NewScore < 0 || report.NewScore > 1 {
		t.Errorf("NewScore: got %f, expected between 0 and 1", report.NewScore)
	}
	if report.ScoreDelta != report.NewScore-report.OldScore {
		t.Errorf("ScoreDelta: got %f, want %f", report.ScoreDelta, report.NewScore-report.OldScore)
	}
	if report.OldScore == report.NewScore && !report.HasRegression && report.ScoreDelta != 0 {
		t.Error("ScoreDelta calculation is inconsistent with HasRegression")
	}
	if report.GeneratedAt.IsZero() {
		t.Error("GeneratedAt should not be zero")
	}
}

func TestRegressionService_Detect_PromptNotFound(t *testing.T) {
	db := newRegressionTestDB(t)
	scoringSvc := NewScoringService()
	svc := NewRegressionService(db, scoringSvc)

	_, err := svc.Detect(999, "gpt-4o", "gpt-4o-mini")
	if err == nil {
		t.Error("expected error for non-existent prompt")
	}
}

func TestRegressionService_GetReport(t *testing.T) {
	db := newRegressionTestDB(t)
	scoringSvc := NewScoringService()
	svc := NewRegressionService(db, scoringSvc)

	prompt := models.Prompt{ID: 2, Title: "Another Test", Content: "Another content"}
	db.Create(&prompt)

	report, err := svc.GetReport(2, "claude-3", "claude-3-5")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if report.PromptID != 2 {
		t.Errorf("PromptID: got %d, want 2", report.PromptID)
	}
	if report.PromptTitle != "Another Test" {
		t.Errorf("PromptTitle: got %s, want %s", report.PromptTitle, "Another Test")
	}
	if report.OldModel != "claude-3" {
		t.Errorf("OldModel: got %s, want %s", report.OldModel, "claude-3")
	}
	if report.NewModel != "claude-3-5" {
		t.Errorf("NewModel: got %s, want %s", report.NewModel, "claude-3-5")
	}
}
