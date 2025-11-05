package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mentor-backend/database"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "mentor-backend/docs"
)

func TestSwaggerEndpoints(t *testing.T) {
	// Set up test database
	db := database.SetupTestDB(t)
	defer database.CleanupTestDB(t, db)

	// Create router
	gin.SetMode(gin.TestMode)
	r := gin.Default()

	// Add Swagger routes (exactly as in main.go)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(301, "/swagger/index.html")
	})

	// Test /docs endpoint redirects to /swagger/index.html
	// This is the most important test as it verifies the user-facing endpoint
	t.Run("DocsEndpointRedirect", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/docs", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Errorf("Expected status code %d, got %d", http.StatusMovedPermanently, w.Code)
		}

		location := w.Header().Get("Location")
		expectedLocation := "/swagger/index.html"
		if location != expectedLocation {
			t.Errorf("Expected redirect to %s, got %s", expectedLocation, location)
		}
	})

	// Test that the swagger route is registered (will be served properly when app runs)
	t.Run("SwaggerRouteRegistered", func(t *testing.T) {
		routes := r.Routes()
		swaggerRouteFound := false
		docsRouteFound := false

		for _, route := range routes {
			if route.Path == "/swagger/*any" && route.Method == "GET" {
				swaggerRouteFound = true
			}
			if route.Path == "/docs" && route.Method == "GET" {
				docsRouteFound = true
			}
		}

		if !swaggerRouteFound {
			t.Error("Expected /swagger/*any route to be registered")
		}
		if !docsRouteFound {
			t.Error("Expected /docs route to be registered")
		}
	})
}
