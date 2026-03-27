package models

import (
	"time"
)

type TestRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PromptID   uint      `gorm:"index;not null" json:"prompt_id"`
	VersionID  uint      `gorm:"index" json:"version_id"`
	Model      string    `gorm:"size:50;not null" json:"model"`
	PromptText string    `gorm:"type:text" json:"prompt_text"`
	Response   string    `gorm:"type:text" json:"response"`
	TokensUsed int       `json:"tokens_used"`
	CreatedAt  time.Time `json:"created_at"`
}

type TestRequest struct {
	Content  string `json:"content" binding:"required"`
	Model    string `json:"model" binding:"required"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OptimizeRequest struct {
	Content  string `json:"content" binding:"required"`
	Mode     string `json:"mode"` // "improve", "structure", "style", "suggest"
}
