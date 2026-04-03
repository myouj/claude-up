package models

import (
	"time"
)

// ABTest represents an A/B test configuration for prompts.
type ABTest struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PromptID  uint      `gorm:"not null" json:"prompt_id"`
	Name      string    `gorm:"size:200" json:"name"`
	Config    string    `gorm:"type:text" json:"config"` // JSON: {variant_a, variant_b, config}
	Status    string    `gorm:"size:20;default:'pending'" json:"status"` // pending | running | completed | stopped
	Result    string    `gorm:"type:text" json:"result,omitempty"` // JSON: SPRT result
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ABTestStatus constants
const (
	ABTestStatusPending   = "pending"
	ABTestStatusRunning   = "running"
	ABTestStatusCompleted = "completed"
	ABTestStatusStopped   = "stopped"
)

// ABTestResult represents a single result record in an A/B test.
type ABTestResult struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	ABTestID  uint      `gorm:"not null" json:"ab_test_id"`
	RunIndex  int       `json:"run_index"`
	Variant   string    `gorm:"size:10" json:"variant"` // A or B
	Score     float64   `json:"score"`
	LatencyMs int       `json:"latency_ms"`
	CreatedAt time.Time `json:"created_at"`
}

// ABTestConfig represents the JSON config stored in ABTest.Config
type ABTestConfig struct {
	VariantA     string  `json:"variant_a"`      // Prompt content for variant A
	VariantB     string  `json:"variant_b"`      // Prompt content for variant B
	Model        string  `json:"model"`          // AI model to use
	MaxRuns      int     `json:"max_runs"`       // Maximum number of runs
	Alpha        float64 `json:"alpha"`          // Significance level (default 0.05)
	MinRuns      int     `json:"min_runs"`       // Minimum runs before checking significance
	EarlyStop    bool    `json:"early_stop"`     // Whether to stop early when significant
	VariableJSON string  `json:"variable_json"`   // JSON string of variables to substitute
}

// SPRTResult represents the Sequential Probability Ratio Test result
type SPRTResult struct {
	Winner          string  `json:"winner"`           // "A", "B", or ""
	IsSignificant   bool    `json:"is_significant"`    // Whether test reached significance
	PA              float64 `json:"p_a"`              // Probability A is better
	PB              float64 `json:"p_b"`              // Probability B is better
	TotalRunsA      int     `json:"total_runs_a"`     // Total runs for variant A
	TotalRunsB      int     `json:"total_runs_b"`     // Total runs for variant B
	LambdaA         float64 `json:"lambda_a"`         // Likelihood ratio for A
	LambdaB         float64 `json:"lambda_b"`         // Likelihood ratio for B
	AverageScoreA   float64 `json:"average_score_a"` // Average score for A
	AverageScoreB   float64 `json:"average_score_b"` // Average score for B
	AverageLatencyA float64 `json:"average_latency_a"` // Average latency for A
	AverageLatencyB float64 `json:"average_latency_b"` // Average latency for B
}

// ABTestResponse is the API response for ABTest
type ABTestResponse struct {
	ID        uint      `json:"id"`
	PromptID  uint      `json:"prompt_id"`
	Name      string    `json:"name"`
	Config    string    `json:"config"`
	Status    string    `json:"status"`
	Result    string    `json:"result,omitempty"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

// ABTestResultResponse is the API response for ABTestResult
type ABTestResultResponse struct {
	ID        uint      `json:"id"`
	ABTestID  uint      `json:"ab_test_id"`
	RunIndex  int       `json:"run_index"`
	Variant   string    `json:"variant"`
	Score     float64   `json:"score"`
	LatencyMs int       `json:"latency_ms"`
	CreatedAt string    `json:"created_at"`
}

// ABTestListResponse is the paginated list response
type ABTestListResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Meta    PaginationMeta `json:"meta,omitempty"`
}
