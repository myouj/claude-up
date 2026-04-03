package models

import (
	"time"
)

// AICallLog records all AI API calls for cost analysis and monitoring.
type AICallLog struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Provider     string    `gorm:"size:50;not null" json:"provider"`
	Model        string    `gorm:"size:100" json:"model"`
	InputTokens  int       `json:"input_tokens"`
	OutputTokens int       `json:"output_tokens"`
	LatencyMs    int       `json:"latency_ms"`
	Cost         float64   `json:"cost"`
	TraceID      string    `gorm:"size:100" json:"trace_id"`
	PromptID     uint      `json:"prompt_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
