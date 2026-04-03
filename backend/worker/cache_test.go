package worker

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"prompt-vault/models"
)

func newCacheTestDB(t *testing.T) *gorm.DB {
	tmpDir := t.TempDir()
	db, err := gorm.Open(sqlite.Open(tmpDir+"/cache_test.db?_busy_timeout=10000"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(&models.ResponseCache{})
	if err != nil {
		t.Fatalf("failed to migrate test db: %v", err)
	}
	return db
}

func TestCacheService_SetAndGet(t *testing.T) {
	db := newCacheTestDB(t)
	svc := NewCacheService(db)

	err := svc.Set("openai", "gpt-4o", "Hello, world!", "Hi there!", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error on set: %v", err)
	}

	response, found, err := svc.Get("openai", "gpt-4o", "Hello, world!")
	if err != nil {
		t.Fatalf("unexpected error on get: %v", err)
	}
	if !found {
		t.Error("expected to find cached response")
	}
	if response != "Hi there!" {
		t.Errorf("response: got %s, want %s", response, "Hi there!")
	}
}

func TestCacheService_Get_NotFound(t *testing.T) {
	db := newCacheTestDB(t)
	svc := NewCacheService(db)

	_, found, err := svc.Get("openai", "gpt-4o", "Non-existent request")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected not to find non-existent request")
	}
}

func TestCacheService_Get_Expired(t *testing.T) {
	db := newCacheTestDB(t)
	svc := NewCacheService(db)

	hash := svc.hashRequest("openai", "gpt-4o", "Expired request")
	entry := &models.ResponseCache{
		Hash:        hash,
		Provider:    "openai",
		Model:       "gpt-4o",
		RequestHash: hash,
		Response:    "Expired response",
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}
	db.Create(entry)

	_, found, err := svc.Get("openai", "gpt-4o", "Expired request")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Error("expected not to find expired entry")
	}
}

func TestCacheService_Cleanup(t *testing.T) {
	db := newCacheTestDB(t)
	svc := NewCacheService(db)

	hash := svc.hashRequest("openai", "gpt-4o", "Cleanup test")
	entry := &models.ResponseCache{
		Hash:        hash,
		Provider:    "openai",
		Model:       "gpt-4o",
		RequestHash: hash,
		Response:    "Cleanup response",
		CreatedAt:   time.Now().Add(-2 * time.Hour),
		ExpiresAt:   time.Now().Add(-1 * time.Hour),
	}
	db.Create(entry)

	err := svc.Cleanup()
	if err != nil {
		t.Fatalf("unexpected error on cleanup: %v", err)
	}

	var count int64
	db.Model(&models.ResponseCache{}).Count(&count)
	if count != 0 {
		t.Errorf("expected 0 entries after cleanup, got %d", count)
	}
}

func TestCacheService_HashRequest(t *testing.T) {
	db := newCacheTestDB(t)
	svc := NewCacheService(db)

	hash1 := svc.hashRequest("openai", "gpt-4o", "test request")
	hash2 := svc.hashRequest("openai", "gpt-4o", "test request")
	hash3 := svc.hashRequest("openai", "gpt-4o", "different request")

	if hash1 != hash2 {
		t.Error("same inputs should produce same hash")
	}
	if hash1 == hash3 {
		t.Error("different inputs should produce different hash")
	}
	if len(hash1) != 64 {
		t.Errorf("expected 64 character hash, got %d", len(hash1))
	}
}

func TestCacheService_DifferentProviderModel(t *testing.T) {
	db := newCacheTestDB(t)
	svc := NewCacheService(db)

	err := svc.Set("anthropic", "claude-3", "Hello!", "Hi from Claude!", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error on set: %v", err)
	}

	response, found, err := svc.Get("anthropic", "claude-3", "Hello!")
	if err != nil {
		t.Fatalf("unexpected error on get: %v", err)
	}
	if !found {
		t.Error("expected to find cached response for claude")
	}
	if response != "Hi from Claude!" {
		t.Errorf("response: got %s, want %s", response, "Hi from Claude!")
	}

	_, found, err = svc.Get("openai", "gpt-4o", "Hello!")
	if err != nil {
		t.Fatalf("unexpected error on get: %v", err)
	}
	if found {
		t.Error("should not find response for different provider/model")
	}
}
