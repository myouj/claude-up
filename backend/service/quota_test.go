package service

import (
	"testing"
	"time"

	"prompt-vault/models"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupQuotaDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&models.Quota{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestQuotaService_Check(t *testing.T) {
	db := setupQuotaDB(t)
	qs := NewQuotaService(db)

	// Create a quota entry with limit 10, usage 0
	nextMonth := time.Date(time.Now().Year(), time.Now().Month()+1, 1, 0, 0, 0, 0, time.UTC)
	quota := &models.Quota{
		Provider: "openai",
		Model:    "",
		Limit:    10,
		Usage:    0,
		ResetAt:  nextMonth,
	}
	if err := db.Create(quota).Error; err != nil {
		t.Fatalf("failed to create quota: %v", err)
	}

	tests := []struct {
		name      string
		provider  string
		cost      int
		wantAllow bool
	}{
		{
			name:      "sufficient quota",
			provider:  "openai",
			cost:      5,
			wantAllow: true,
		},
		{
			name:      "exact limit",
			provider:  "openai",
			cost:      10,
			wantAllow: true,
		},
		{
			name:      "exceeds limit",
			provider:  "openai",
			cost:      11,
			wantAllow: false,
		},
		{
			name:      "no quota configured",
			provider:  "claude",
			cost:      1,
			wantAllow: true, // unlimited if no quota found
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			allowed, err := qs.Check(tt.provider, tt.cost)
			if err != nil {
				t.Errorf("Check() error = %v", err)
				return
			}
			if allowed != tt.wantAllow {
				t.Errorf("Check() = %v, want %v", allowed, tt.wantAllow)
			}
		})
	}
}

func TestQuotaService_Consume(t *testing.T) {
	db := setupQuotaDB(t)
	qs := NewQuotaService(db)

	nextMonth := time.Date(time.Now().Year(), time.Now().Month()+1, 1, 0, 0, 0, 0, time.UTC)
	quota := &models.Quota{
		Provider: "openai",
		Model:    "",
		Limit:    10,
		Usage:    0,
		ResetAt:  nextMonth,
	}
	if err := db.Create(quota).Error; err != nil {
		t.Fatalf("failed to create quota: %v", err)
	}

	tests := []struct {
		name    string
		provide string
		cost    int
		wantErr bool
	}{
		{
			name:    "consume within limit",
			provide: "openai",
			cost:    3,
			wantErr: false,
		},
		{
			name:    "consume to exact limit",
			provide: "openai",
			cost:    7,
			wantErr: false,
		},
		{
			name:    "consume exceeds limit",
			provide: "openai",
			cost:    1,
			wantErr: true,
		},
		{
			name:    "consume no quota configured",
			provide: "claude",
			cost:    1,
			wantErr: false, // unlimited
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := qs.Consume(tt.provide, tt.cost)
			if (err != nil) != tt.wantErr {
				t.Errorf("Consume() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Verify final usage
	usage, err := qs.GetUsage("openai")
	if err != nil {
		t.Errorf("GetUsage() error = %v", err)
	}
	if usage != 10 {
		t.Errorf("GetUsage() = %d, want 10", usage)
	}
}

func TestQuotaService_GetUsage(t *testing.T) {
	db := setupQuotaDB(t)
	qs := NewQuotaService(db)

	nextMonth := time.Date(time.Now().Year(), time.Now().Month()+1, 1, 0, 0, 0, 0, time.UTC)
	quota := &models.Quota{
		Provider: "openai",
		Model:    "",
		Limit:    10,
		Usage:    5,
		ResetAt:  nextMonth,
	}
	if err := db.Create(quota).Error; err != nil {
		t.Fatalf("failed to create quota: %v", err)
	}

	tests := []struct {
		name     string
		provider string
		want     int
		wantErr  bool
	}{
		{
			name:     "existing provider",
			provider: "openai",
			want:     5,
			wantErr:  false,
		},
		{
			name:     "non-existing provider",
			provider: "claude",
			want:     0,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := qs.GetUsage(tt.provider)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetUsage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetUsage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuotaService_CreateOrUpdate(t *testing.T) {
	db := setupQuotaDB(t)
	qs := NewQuotaService(db)

	// Create new quota
	quota, err := qs.CreateOrUpdate("openai", "", 100)
	if err != nil {
		t.Fatalf("CreateOrUpdate() error = %v", err)
	}
	if quota.Limit != 100 {
		t.Errorf("CreateOrUpdate() limit = %d, want 100", quota.Limit)
	}
	if quota.Usage != 0 {
		t.Errorf("CreateOrUpdate() usage = %d, want 0", quota.Usage)
	}

	// Update existing quota
	quota, err = qs.CreateOrUpdate("openai", "", 200)
	if err != nil {
		t.Fatalf("CreateOrUpdate() error = %v", err)
	}
	if quota.Limit != 200 {
		t.Errorf("CreateOrUpdate() limit = %d, want 200", quota.Limit)
	}
}

func TestQuotaService_ResetUsage(t *testing.T) {
	db := setupQuotaDB(t)
	qs := NewQuotaService(db)

	nextMonth := time.Date(time.Now().Year(), time.Now().Month()+1, 1, 0, 0, 0, 0, time.UTC)
	quota := &models.Quota{
		Provider: "openai",
		Model:    "",
		Limit:    10,
		Usage:    5,
		ResetAt:  nextMonth,
	}
	if err := db.Create(quota).Error; err != nil {
		t.Fatalf("failed to create quota: %v", err)
	}

	if err := qs.ResetUsage("openai"); err != nil {
		t.Fatalf("ResetUsage() error = %v", err)
	}

	usage, err := qs.GetUsage("openai")
	if err != nil {
		t.Fatalf("GetUsage() error = %v", err)
	}
	if usage != 0 {
		t.Errorf("GetUsage() after reset = %d, want 0", usage)
	}
}

func TestNextMonth(t *testing.T) {
	// This test verifies the nextMonth function logic
	next := nextMonth()
	now := time.Now()
	if next.Month() != now.Month()+1 && !(now.Month() == 12 && next.Month() != 1) {
		t.Errorf("nextMonth() = %v, expected next month", next)
	}
	if next.Day() != 1 {
		t.Errorf("nextMonth() day = %d, want 1", next.Day())
	}
	if next.Hour() != 0 || next.Minute() != 0 || next.Second() != 0 {
		t.Errorf("nextMonth() time not at midnight")
	}
}
