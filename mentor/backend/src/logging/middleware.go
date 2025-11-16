// Package logging provides request ID middleware and request logging
package logging

import (
"time"

"github.com/gin-gonic/gin"
"github.com/google/uuid"
)

const (
// RequestIDHeader is the HTTP header name for request ID
RequestIDHeader = "X-Request-ID"
// RequestIDKey is the context key for request ID
RequestIDKey = "request_id"
)

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() gin.HandlerFunc {
return func(c *gin.Context) {
// Extract or generate request ID
requestID := c.GetHeader(RequestIDHeader)
if requestID == "" {
requestID = uuid.New().String()
}

// Store in context
c.Set(RequestIDKey, requestID)

// Add to response headers
c.Header(RequestIDHeader, requestID)

c.Next()
}
}

// RequestLoggingMiddleware logs request details with structured logging
func RequestLoggingMiddleware() gin.HandlerFunc {
return func(c *gin.Context) {
// Start timer
startTime := time.Now()

// Get request ID
requestID, _ := c.Get(RequestIDKey)

// Log incoming request
Info("Request started", map[string]interface{}{
"request_id":  requestID,
"method":      c.Request.Method,
"path":        c.Request.URL.Path,
"client_ip":   c.ClientIP(),
"user_agent":  c.Request.UserAgent(),
})

// Process request
c.Next()

// Calculate duration
duration := time.Since(startTime)

// Determine log level based on status code
statusCode := c.Writer.Status()
logFields := map[string]interface{}{
"request_id":  requestID,
"method":      c.Request.Method,
"path":        c.Request.URL.Path,
"status_code": statusCode,
"duration_ms": duration.Milliseconds(),
}

// Add error if present
if len(c.Errors) > 0 {
logFields["errors"] = c.Errors.String()
}

// Log based on status code
if statusCode >= 500 {
Error("Request failed with server error", logFields)
} else if statusCode >= 400 {
Warning("Request failed with client error", logFields)
} else {
Info("Request completed", logFields)
}
}
}

// GetRequestID retrieves the request ID from the Gin context
func GetRequestID(c *gin.Context) string {
if requestID, exists := c.Get(RequestIDKey); exists {
if id, ok := requestID.(string); ok {
return id
}
}
return "unknown"
}
