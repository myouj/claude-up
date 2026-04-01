package models

import (
	"time"
)

// Setting stores key-value configuration, including encrypted API keys.
type Setting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Key       string    `gorm:"size:100;uniqueIndex" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	IsSecret  bool      `gorm:"default:false" json:"is_secret"` // if true, value is encrypted
	UpdatedAt time.Time `json:"updated_at"`
}

// Known setting keys:
// "openai_api_key", "anthropic_api_key", "gemini_api_key", "minimax_api_key"
// "default_provider", "default_model", "theme", "language"
