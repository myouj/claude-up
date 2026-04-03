package models

import (
	"encoding/json"
	"time"
)

// EvalSet represents an evaluation set for prompt testing
type EvalSet struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PromptID  uint      `gorm:"not null" json:"prompt_id"`
	Name      string    `gorm:"size:200" json:"name"`
	Cases     string    `gorm:"type:text" json:"cases"`     // JSON array of test cases
	Weights   string    `gorm:"type:text" json:"weights"`   // JSON object of dimension weights
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// EvalCase represents a single test case in an evaluation set
type EvalCase struct {
	Input    string            `json:"input"`
	Expected string            `json:"expected,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

// EvalWeights represents the weights for each scoring dimension
type EvalWeights struct {
	Clarity     float64 `json:"clarity"`
	Completeness float64 `json:"completeness"`
	Example     float64 `json:"example"`
	Role        float64 `json:"role"`
}

// EvalSetResponse is the API response for an EvalSet
type EvalSetResponse struct {
	ID        uint        `json:"id"`
	PromptID  uint        `json:"prompt_id"`
	Name      string      `json:"name"`
	Cases     []EvalCase  `json:"cases"`
	Weights   EvalWeights `json:"weights"`
	CreatedAt string      `json:"created_at"`
	UpdatedAt string      `json:"updated_at"`
}

// ToResponse converts an EvalSet to EvalSetResponse
func (e *EvalSet) ToResponse() EvalSetResponse {
	var cases []EvalCase
	var weights EvalWeights

	if e.Cases != "" {
		json.Unmarshal([]byte(e.Cases), &cases)
	}
	if e.Weights != "" {
		json.Unmarshal([]byte(e.Weights), &weights)
	}

	return EvalSetResponse{
		ID:        e.ID,
		PromptID:  e.PromptID,
		Name:      e.Name,
		Cases:     cases,
		Weights:   weights,
		CreatedAt: e.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: e.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// DefaultEvalWeights returns the default weights for evaluation
func DefaultEvalWeights() EvalWeights {
	return EvalWeights{
		Clarity:     0.30,
		Completeness: 0.30,
		Example:     0.25,
		Role:        0.15,
	}
}
