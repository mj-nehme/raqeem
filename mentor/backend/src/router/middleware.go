package router

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CorrelationIDMiddleware adds a correlation ID to each request for tracing
func CorrelationIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if correlation ID already exists in headers
		correlationID := c.GetHeader("X-Correlation-ID")
		if correlationID == "" {
			// Generate new correlation ID if not present
			correlationID = uuid.New().String()
		}

		// Set correlation ID in context and response header
		c.Set("correlation_id", correlationID)
		c.Header("X-Correlation-ID", correlationID)

		c.Next()
	}
}

// RequestLoggerMiddleware logs incoming requests with correlation ID
func RequestLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Get correlation ID
		correlationID, _ := c.Get("correlation_id")

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Log request details
		log.Printf("[%s] %s %s - Status: %d - Latency: %v - IP: %s",
			correlationID,
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			latency,
			c.ClientIP(),
		)
	}
}

// ErrorHandlerMiddleware handles errors and provides structured error responses
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			correlationID, _ := c.Get("correlation_id")

			// Log all errors
			for _, err := range c.Errors {
				log.Printf("[%s] Error: %v", correlationID, err.Err)
			}

			// Return the first error to the client
			err := c.Errors[0]
			c.JSON(-1, gin.H{
				"error":          err.Error(),
				"correlation_id": correlationID,
			})
		}
	}
}

// RecoveryMiddleware recovers from panics and returns a 500 error
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				correlationID, _ := c.Get("correlation_id")
				log.Printf("[%s] PANIC RECOVERED: %v", correlationID, err)

				c.JSON(500, gin.H{
					"error":          "Internal server error",
					"correlation_id": correlationID,
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
