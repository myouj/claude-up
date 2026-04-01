package middleware

import (
	"bytes"
	"io"
	"time"

	"github.com/gin-gonic/gin"

	"prompt-vault/utils"
)

// responseWriter wraps gin.ResponseWriter to capture the status code.
type responseWriter struct {
	gin.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = 200
	}
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

// RequestLoggerMiddleware logs every incoming request and its response.
// Logs: method, path, status, latency, client IP, request size, response size, trace ID.
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log := GetTraceLogger(c)

		// Read and restore request body (for logging purposes).
		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		start := time.Now()
		rw := &responseWriter{ResponseWriter: c.Writer, status: 0}
		c.Writer = rw

		// Log request entry.
		fields := map[string]interface{}{
			"method":      c.Request.Method,
			"path":        c.Request.URL.Path,
			"query":       c.Request.URL.RawQuery,
			"client_ip":   c.ClientIP(),
			"user_agent":  c.Request.UserAgent(),
			"request_uri": c.Request.RequestURI,
		}
		if len(bodyBytes) > 0 && len(bodyBytes) <= 4096 {
			fields["request_body"] = string(bodyBytes)
		}
		if len(bodyBytes) > 4096 {
			fields["request_body_size"] = len(bodyBytes)
		}

		log.Info("request started", fields)

		// Process request.
		c.Next()

		// Compute latency.
		latency := time.Since(start)

		// Log response.
		respFields := map[string]interface{}{
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"status":       rw.status,
			"latency_ms":   latency.Milliseconds(),
			"latency":      latency.String(),
			"client_ip":    c.ClientIP(),
			"response_size": rw.size,
		}

		msg := "request completed"
		if rw.status >= 500 {
			msg = "request failed"
			// Log panic info if available.
			if errMsg := c.GetString("panic_error"); errMsg != "" {
				respFields["panic_error"] = errMsg
				respFields["panic_stack"] = c.GetString("panic_stack")
			}
		} else if rw.status >= 400 {
			msg = "request error"
		}

		log.Info(msg, respFields)
	}
}

// RecoveryLoggerMiddleware recovers from panics and logs them with stack trace.
// This replaces the default gin Recovery middleware.
func RecoveryLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log := GetTraceLogger(c)
				stack := utils.StackTrace(3)
				errMsg := formatPanic(r)

				// Store in context for request logger to pick up.
				c.Set("panic_error", errMsg)
				c.Set("panic_stack", stack)

				log.Error("panic recovered", map[string]interface{}{
					"error": errMsg,
					"stack": stack,
					"path":  c.Request.URL.Path,
					"method": c.Request.Method,
				})

				c.AbortWithStatusJSON(500, gin.H{
					"success":  false,
					"error":    "Internal server error",
					"trace_id": GetTraceID(c),
				})
			}
		}()
		c.Next()
	}
}

func formatPanic(r interface{}) string {
	switch v := r.(type) {
	case error:
		return v.Error()
	case string:
		return v
	default:
		return "unknown panic"
	}
}
