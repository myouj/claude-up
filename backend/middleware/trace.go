package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"prompt-vault/utils"
)

// TraceMiddleware generates a unique trace ID for each request
// and injects a trace-aware logger into the context.
func TraceMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use existing trace ID from header if present, otherwise generate one.
		traceID := c.GetHeader(HeaderTraceID)
		if traceID == "" {
			traceID = uuid.New().String()
		}

		// Store in context for downstream use.
		c.Set(TraceIDKey, traceID)

		// Inject trace-aware logger.
		c.Set(TraceLoggerKey, utils.DefaultLogger.WithTraceID(traceID))

		// Set trace ID in response header for client correlation.
		c.Header(HeaderTraceID, traceID)

		c.Next()
	}
}

// GetTraceID retrieves the trace ID from the Gin context.
func GetTraceID(c *gin.Context) string {
	if v, exists := c.Get(TraceIDKey); exists {
		if id, ok := v.(string); ok {
			return id
		}
	}
	return ""
}

// GetTraceLogger retrieves the trace-aware logger from the Gin context.
func GetTraceLogger(c *gin.Context) *utils.LoggerWithTrace {
	if v, exists := c.Get(TraceLoggerKey); exists {
		if l, ok := v.(*utils.LoggerWithTrace); ok {
			return l
		}
	}
	// Fallback: return a no-op logger if somehow not set.
	return utils.DefaultLogger.WithTraceID("")
}
