package models

import (
	"time"
)

type TestRecord struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	PromptID   uint      `gorm:"index;not null" json:"prompt_id"`
	VersionID  uint      `gorm:"index" json:"version_id"`
	Model      string    `gorm:"size:50;not null" json:"model"`
	Provider   string    `gorm:"size:20" json:"provider"` // openai, claude, gemini, minimax
	PromptText string    `gorm:"type:text" json:"prompt_text"`
	Response   string    `gorm:"type:text" json:"response"`
	TokensUsed int       `json:"tokens_used"`
	LatencyMs  int64     `json:"latency_ms"`
	CreatedAt  time.Time `json:"created_at"`
}

type TestRequest struct {
	Content  string    `json:"content" binding:"required"`
	Model    string    `json:"model" binding:"required"`
	Provider string    `json:"provider"` // defaults to "openai"
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OptimizeRequest struct {
	Content  string `json:"content" binding:"required"`
	Mode     string `json:"mode"` // "improve", "structure", "style", "suggest"
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

type TestComparison struct {
	Records     []TestRecord `json:"records"`
	VersionInfo []VersionInfo `json:"version_info"`
}

type VersionInfo struct {
	VersionID    uint   `json:"version_id"`
	Version      int    `json:"version"`
	PromptContent string `json:"prompt_content"`
	Comment      string `json:"comment"`
	CreatedAt    string `json:"created_at"`
}

type DailyStats struct {
	Date      string  `json:"date"`
	Count     int     `json:"count"`
	AvgTokens float64 `json:"avg_tokens"`
}

type PromptAnalytics struct {
	TotalTests  int64             `json:"total_tests"`
	AvgTokens   float64           `json:"avg_tokens"`
	AvgLatency  float64           `json:"avg_latency_ms"`
	SuccessRate float64           `json:"success_rate"`
	ByModel     map[string]int    `json:"by_model"`
	ByDate      []DailyStats      `json:"by_date"`
}
