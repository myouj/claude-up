package worker

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"gorm.io/gorm"

	"prompt-vault/models"
)

type CacheService struct {
	db *gorm.DB
}

func NewCacheService(db *gorm.DB) *CacheService {
	return &CacheService{db: db}
}

type CacheEntry struct {
	Hash        string    `gorm:"primaryKey" json:"hash"`
	Provider    string    `gorm:"size:50;not null" json:"provider"`
	Model       string    `gorm:"size:100" json:"model"`
	RequestHash string    `gorm:"size:64;not null" json:"request_hash"`
	Response    string    `gorm:"type:text" json:"response"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

func (s *CacheService) TableName() string {
	return "response_cache"
}

func (s *CacheService) Get(provider, model, request string) (string, bool, error) {
	hash := s.hashRequest(provider, model, request)
	var entry models.ResponseCache
	if err := s.db.Where("hash = ? AND expires_at > ?", hash, time.Now()).First(&entry).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", false, nil
		}
		return "", false, err
	}
	return entry.Response, true, nil
}

func (s *CacheService) Set(provider, model, request, response string, ttl time.Duration) error {
	hash := s.hashRequest(provider, model, request)
	entry := &models.ResponseCache{
		Hash:        hash,
		Provider:    provider,
		Model:       model,
		RequestHash: hash,
		Response:    response,
		CreatedAt:  time.Now(),
		ExpiresAt:   time.Now().Add(ttl),
	}
	return s.db.Save(entry).Error
}

func (s *CacheService) hashRequest(provider, model, request string) string {
	data := provider + ":" + model + ":" + request
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func (s *CacheService) Cleanup() error {
	return s.db.Where("expires_at < ?", time.Now()).Delete(&models.ResponseCache{}).Error
}
