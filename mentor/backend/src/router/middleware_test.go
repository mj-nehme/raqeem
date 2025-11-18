package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCorrelationIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("generates correlation ID when not provided", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		correlationID := w.Header().Get("X-Correlation-ID")
		assert.NotEmpty(t, correlationID, "Should generate correlation ID")
	})

	t.Run("uses existing correlation ID from header", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		existingID := "test-correlation-id-123"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Correlation-ID", existingID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		correlationID := w.Header().Get("X-Correlation-ID")
		assert.Equal(t, existingID, correlationID, "Should use existing correlation ID")
	})

	t.Run("stores correlation ID in context", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())

		var contextCorrelationID string
		router.GET("/test", func(c *gin.Context) {
			if val, exists := c.Get("correlation_id"); exists {
				contextCorrelationID = val.(string)
			}
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.NotEmpty(t, contextCorrelationID, "Should store correlation ID in context")
	})
}

func TestRequestLoggerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("logs request details", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(RequestLoggerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("logs request with correlation ID", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(RequestLoggerMiddleware())

		var loggedCorrelationID interface{}
		router.GET("/test", func(c *gin.Context) {
			loggedCorrelationID, _ = c.Get("correlation_id")
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.NotNil(t, loggedCorrelationID)
	})

	t.Run("logs latency", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(RequestLoggerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})
}

func TestErrorHandlerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles no errors gracefully", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(ErrorHandlerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("handles errors and returns JSON", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(ErrorHandlerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			_ = c.Error(http.ErrAbortHandler)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, w.Body.String(), "error")
		assert.Contains(t, w.Body.String(), "correlation_id")
	})

	t.Run("logs multiple errors", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(ErrorHandlerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			_ = c.Error(http.ErrAbortHandler)
			_ = c.Error(http.ErrBodyNotAllowed)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		// Should return the first error
		assert.Contains(t, w.Body.String(), "error")
	})

	t.Run("includes correlation ID in error response", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(ErrorHandlerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			_ = c.Error(http.ErrAbortHandler)
		})

		existingID := "test-error-correlation-id"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Correlation-ID", existingID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, w.Body.String(), existingID)
	})
}

func TestRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("handles normal execution", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(RecoveryMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	t.Run("recovers from panic", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(RecoveryMiddleware())
		router.GET("/test", func(c *gin.Context) {
			panic("test panic")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
	})

	t.Run("includes correlation ID in panic response", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(RecoveryMiddleware())
		router.GET("/test", func(c *gin.Context) {
			panic("test panic with correlation")
		})

		existingID := "test-panic-correlation-id"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Correlation-ID", existingID)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
		assert.Contains(t, w.Body.String(), existingID)
	})

	t.Run("aborts request after panic", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(RecoveryMiddleware())

		handlerCalled := false
		router.GET("/test", func(c *gin.Context) {
			panic("test panic abort")
		})
		router.Use(func(c *gin.Context) {
			handlerCalled = true
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.False(t, handlerCalled, "Should not call handlers after panic")
	})
}

func TestMiddlewareChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("all middlewares work together", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(RequestLoggerMiddleware())
		router.Use(ErrorHandlerMiddleware())
		router.Use(RecoveryMiddleware())
		router.GET("/test", func(c *gin.Context) {
			c.String(200, "OK")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Correlation-ID"))
	})

	t.Run("middlewares handle errors in chain", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(RequestLoggerMiddleware())
		router.Use(ErrorHandlerMiddleware())
		router.GET("/test", func(c *gin.Context) {
			_ = c.Error(http.ErrAbortHandler)
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Contains(t, w.Body.String(), "error")
		assert.Contains(t, w.Body.String(), "correlation_id")
	})

	t.Run("middlewares handle panics in chain", func(t *testing.T) {
		router := gin.New()
		router.Use(CorrelationIDMiddleware())
		router.Use(RequestLoggerMiddleware())
		router.Use(RecoveryMiddleware())
		router.GET("/test", func(c *gin.Context) {
			panic("test chain panic")
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, 500, w.Code)
	})
}
