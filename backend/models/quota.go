package models

import (
	"time"
)

// Quota represents a provider's monthly API quota.
type Quota struct {
	ID       uint      `gorm:"primaryKey" json:"id"`
	Provider string    `gorm:"size:50;not null;index" json:"provider"` // openai | claude | gemini | minimax
	Model    string    `gorm:"size:100" json:"model"`                  // optional model-specific quota
	Limit    int       `gorm:"not null" json:"limit"`                  // monthly limit
	Usage    int       `gorm:"default:0" json:"usage"`                 // current month usage
	ResetAt  time.Time `json:"reset_at"`                               // reset timestamp
}

// QuotaResponse is the API response format for quota data.
type QuotaResponse struct {
	ID       uint   `json:"id"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
	Limit    int    `json:"limit"`
	Usage    int    `json:"usage"`
	ResetAt  string `json:"reset_at"`
}

// ToResponse converts a Quota model to its API response format.
func (q *Quota) ToResponse() QuotaResponse {
	return QuotaResponse{
		ID:       q.ID,
		Provider: q.Provider,
		Model:    q.Model,
		Limit:    q.Limit,
		Usage:    q.Usage,
		ResetAt:  q.ResetAt.Format(time.RFC3339),
	}
}
