package logging

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates request ID when not provided", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestIDMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		requestID := w.Header().Get(RequestIDHeader)
		if requestID == "" {
			t.Error("Request ID should be generated when not provided")
		}
	})

	t.Run("uses existing request ID from header", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestIDMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		existingID := "test-request-id-123"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(RequestIDHeader, existingID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		requestID := w.Header().Get(RequestIDHeader)
		if requestID != existingID {
			t.Errorf("Expected request ID %s, got %s", existingID, requestID)
		}
	})

	t.Run("stores request ID in context", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestIDMiddleware())

		var contextRequestID string
		router.GET("/test", func(c *gin.Context) {
			if val, exists := c.Get(RequestIDKey); exists {
				contextRequestID = val.(string)
			}
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if contextRequestID == "" {
			t.Error("Request ID should be stored in context")
		}
	})
}

func TestRequestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("logs successful request", func(t *testing.T) {
		var buf bytes.Buffer
		oldLogger := defaultLogger
		defer func() { defaultLogger = oldLogger }()

		defaultLogger = NewLogger(Config{
			Level:      "INFO",
			JSONFormat: false,
			Output:     &buf,
		})

		router := gin.New()
		router.Use(RequestIDMiddleware())
		router.Use(RequestLoggingMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buf.String()
		if !strings.Contains(output, "Request started") {
			t.Error("Should log request start")
		}
		if !strings.Contains(output, "Request completed") {
			t.Error("Should log request completion")
		}
		if !strings.Contains(output, "GET") {
			t.Error("Should log HTTP method")
		}
		if !strings.Contains(output, "/test") {
			t.Error("Should log request path")
		}
	})

	t.Run("logs 4xx client error", func(t *testing.T) {
		var buf bytes.Buffer
		oldLogger := defaultLogger
		defer func() { defaultLogger = oldLogger }()

		defaultLogger = NewLogger(Config{
			Level:      "WARNING",
			JSONFormat: false,
			Output:     &buf,
		})

		router := gin.New()
		router.Use(RequestIDMiddleware())
		router.Use(RequestLoggingMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(404, "Not Found")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buf.String()
		if !strings.Contains(output, "Request failed with client error") {
			t.Error("Should log client error for 4xx status")
		}
		if !strings.Contains(output, "404") {
			t.Error("Should include status code")
		}
	})

	t.Run("logs 5xx server error", func(t *testing.T) {
		var buf bytes.Buffer
		oldLogger := defaultLogger
		defer func() { defaultLogger = oldLogger }()

		defaultLogger = NewLogger(Config{
			Level:      "ERROR",
			JSONFormat: false,
			Output:     &buf,
		})

		router := gin.New()
		router.Use(RequestIDMiddleware())
		router.Use(RequestLoggingMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(500, "Internal Server Error")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buf.String()
		if !strings.Contains(output, "Request failed with server error") {
			t.Error("Should log server error for 5xx status")
		}
		if !strings.Contains(output, "500") {
			t.Error("Should include status code")
		}
	})

	t.Run("logs with errors from context", func(t *testing.T) {
		var buf bytes.Buffer
		oldLogger := defaultLogger
		defer func() { defaultLogger = oldLogger }()

		defaultLogger = NewLogger(Config{
			Level:      "WARNING",
			JSONFormat: false,
			Output:     &buf,
		})

		router := gin.New()
		router.Use(RequestIDMiddleware())
		router.Use(RequestLoggingMiddleware())
		router.GET("/test", func(c *gin.Context) {
			_ = c.Error(http.ErrAbortHandler)
			c.String(400, "Bad Request")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buf.String()
		if !strings.Contains(output, "Request failed with client error") {
			t.Error("Should log request with errors")
		}
	})

	t.Run("includes duration in log", func(t *testing.T) {
		var buf bytes.Buffer
		oldLogger := defaultLogger
		defer func() { defaultLogger = oldLogger }()

		defaultLogger = NewLogger(Config{
			Level:      "INFO",
			JSONFormat: false,
			Output:     &buf,
		})

		router := gin.New()
		router.Use(RequestIDMiddleware())
		router.Use(RequestLoggingMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		output := buf.String()
		if !strings.Contains(output, "duration_ms") {
			t.Error("Should log request duration")
		}
	})
}

func TestGetRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns request ID from context", func(t *testing.T) {
		router := gin.New()
		router.Use(RequestIDMiddleware())

		var retrievedID string
		router.GET("/test", func(c *gin.Context) {
			retrievedID = GetRequestID(c)
			c.String(200, "OK")
		})

		expectedID := "test-request-id-456"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(RequestIDHeader, expectedID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if retrievedID != expectedID {
			t.Errorf("Expected request ID %s, got %s", expectedID, retrievedID)
		}
	})

	t.Run("returns 'unknown' when request ID not in context", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		requestID := GetRequestID(c)
		if requestID != "unknown" {
			t.Errorf("Expected 'unknown', got %s", requestID)
		}
	})

	t.Run("returns 'unknown' for invalid request ID type", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(RequestIDKey, 12345) // Set non-string value

		requestID := GetRequestID(c)
		if requestID != "unknown" {
			t.Errorf("Expected 'unknown' for invalid type, got %s", requestID)
		}
	})
}

func TestMiddlewareIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("request ID flows through both middlewares", func(t *testing.T) {
		var buf bytes.Buffer
		oldLogger := defaultLogger
		defer func() { defaultLogger = oldLogger }()

		defaultLogger = NewLogger(Config{
			Level:      "INFO",
			JSONFormat: false,
			Output:     &buf,
		})

		router := gin.New()
		router.Use(RequestIDMiddleware())
		router.Use(RequestLoggingMiddleware())

		var handlerRequestID string
		router.GET("/test", func(c *gin.Context) {
			handlerRequestID = GetRequestID(c)
			c.String(200, "OK")
		})

		expectedID := "integration-test-id"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set(RequestIDHeader, expectedID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Verify request ID in handler
		if handlerRequestID != expectedID {
			t.Errorf("Expected handler request ID %s, got %s", expectedID, handlerRequestID)
		}

		// Verify request ID in response header
		responseRequestID := w.Header().Get(RequestIDHeader)
		if responseRequestID != expectedID {
			t.Errorf("Expected response request ID %s, got %s", expectedID, responseRequestID)
		}

		// Verify request ID appears in logs
		output := buf.String()
		if !strings.Contains(output, expectedID) {
			t.Error("Request ID should appear in logs")
		}
	})
}
