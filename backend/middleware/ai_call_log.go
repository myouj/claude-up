package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"prompt-vault/models"
)

// Context keys for AI call details
const (
	HeaderAIProvider = "X-AI-Provider"
	ContextAICallLog = "ai_call_log"
)

// AICallLogMiddleware records all AI API calls for cost analysis.
// It intercepts requests with X-AI-Provider header and records
// provider, model, tokens, latency, and cost after the request completes.
func AICallLogMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if this is an AI call request
		provider := c.GetHeader(HeaderAIProvider)
		isAICall := provider != ""

		// Track start time for latency calculation
		startTime := time.Now()

		// Process request
		c.Next()

		// Only record if this was an AI call
		if !isAICall {
			return
		}

		// Get trace ID for correlation
		traceID := GetTraceID(c)

		// Get AI call details from context (set by handler)
		var inputTokens, outputTokens int
		var cost float64
		var model string
		var promptID uint

		if aiLog, exists := c.Get(ContextAICallLog); exists {
			if log, ok := aiLog.(*models.AICallLog); ok {
				inputTokens = log.InputTokens
				outputTokens = log.OutputTokens
				cost = log.Cost
				model = log.Model
				promptID = log.PromptID
			}
		}

		// Calculate latency
		latencyMs := int(time.Since(startTime).Milliseconds())

		// Create log entry
		aiCallLog := &models.AICallLog{
			Provider:     provider,
			Model:        model,
			InputTokens:  inputTokens,
			OutputTokens: outputTokens,
			LatencyMs:    latencyMs,
			Cost:         cost,
			TraceID:      traceID,
			PromptID:     promptID,
			CreatedAt:    time.Now(),
		}

		// Record to database
		if err := db.Create(aiCallLog).Error; err != nil {
			// Log error but don't fail the request
			GetTraceLogger(c).Error("failed to record AI call log", map[string]interface{}{
				"error":   err.Error(),
				"trace_id": traceID,
			})
		}
	}
}

// SetAICallLog sets the AI call details in the context for the middleware to record.
func SetAICallLog(c *gin.Context, log *models.AICallLog) {
	c.Set(ContextAICallLog, log)
}

// GetAICallLog retrieves the AI call details from the context.
func GetAICallLog(c *gin.Context) *models.AICallLog {
	if v, exists := c.Get(ContextAICallLog); exists {
		if log, ok := v.(*models.AICallLog); ok {
			return log
		}
	}
	return nil
}
