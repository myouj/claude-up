package middleware

import (
	"net/http"
	"strconv"

	"prompt-vault/service"

	"github.com/gin-gonic/gin"
)

// ContextKeyQuotaService is the context key for the QuotaService.
const ContextKeyQuotaService = "quota_service"

// QuotaMiddleware checks if the request's AI provider has sufficient quota.
// It reads the X-AI-Provider header and calls QuotaService.Check.
func QuotaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		provider := c.GetHeader(HeaderAIProvider)
		if provider == "" {
			// No provider specified, skip quota check
			c.Next()
			return
		}

		// Get QuotaService from context (set during initialization)
		qsVal, exists := c.Get(ContextKeyQuotaService)
		if !exists {
			// QuotaService not configured, skip check
			c.Next()
			return
		}
		qs, ok := qsVal.(*service.QuotaService)
		if !ok {
			// Invalid QuotaService type, skip check
			c.Next()
			return
		}

		// Default cost is 1, but can be overridden by header
		cost := 1
		if costStr := c.GetHeader("X-AI-Cost"); costStr != "" {
			if parsedCost, err := strconv.Atoi(costStr); err == nil && parsedCost > 0 {
				cost = parsedCost
			}
		}

		allowed, err := qs.Check(provider, cost)
		if err != nil {
			GetTraceLogger(c).Error("quota check failed", map[string]interface{}{
				"provider": provider,
				"error":    err.Error(),
			})
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to check quota",
			})
			c.Abort()
			return
		}

		if !allowed {
			usage, _ := qs.GetUsage(provider)
			GetTraceLogger(c).Warn("quota exceeded", map[string]interface{}{
				"provider": provider,
				"usage":    usage,
			})
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Quota exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ConsumeQuotaMiddleware consumes quota after a successful API call.
// This should be used after the handler processes the request.
func ConsumeQuotaMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request first
		c.Next()

		// After request processing, consume quota if successful
		if c.Writer.Status() >= 200 && c.Writer.Status() < 300 {
			provider := c.GetHeader(HeaderAIProvider)
			if provider == "" {
				return
			}

			qsVal, exists := c.Get(ContextKeyQuotaService)
			if !exists {
				return
			}
			qs, ok := qsVal.(*service.QuotaService)
			if !ok {
				return
			}

			cost := 1
			if costStr := c.GetHeader("X-AI-Cost"); costStr != "" {
				if parsedCost, err := strconv.Atoi(costStr); err == nil && parsedCost > 0 {
					cost = parsedCost
				}
			}

			if err := qs.Consume(provider, cost); err != nil {
				GetTraceLogger(c).Error("failed to consume quota", map[string]interface{}{
					"provider": provider,
					"cost":     cost,
					"error":    err.Error(),
				})
			}
		}
	}
}
