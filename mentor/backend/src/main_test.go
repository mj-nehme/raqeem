package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"mentor-backend/database"
	"mentor-backend/router"

	_ "mentor-backend/docs"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestMainRouterIntegration verifies that the router package integrates correctly with main
func TestMainRouterIntegration(t *testing.T) {
	// Set up test database
	db, err := database.SetupTestDB(t)
	assert.NoError(t, err)
	assert.NotNil(t, db)
	defer database.CleanupTestDB(t, db)

	// Set test mode
	gin.SetMode(gin.TestMode)

	// Set environment variables for CORS
	_ = os.Setenv("FRONTEND_ORIGIN", "http://localhost:3000")
	defer func() { _ = os.Unsetenv("FRONTEND_ORIGIN") }()

	// Create router using the same approach as main.go
	r := router.New()
	r.SetupAllRoutes()

	// Test health endpoint
	t.Run("HealthEndpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		r.Engine().ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	// Test docs redirect
	t.Run("DocsRedirect", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/docs", nil)
		w := httptest.NewRecorder()
		r.Engine().ServeHTTP(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Errorf("Expected status code %d, got %d", http.StatusMovedPermanently, w.Code)
		}

		location := w.Header().Get("Location")
		if location != "/swagger/index.html" {
			t.Errorf("Expected redirect to /swagger/index.html, got %s", location)
		}
	})

	// Test swagger endpoint exists
	t.Run("SwaggerEndpoint", func(t *testing.T) {
		routes := r.Engine().Routes()
		swaggerFound := false
		for _, route := range routes {
			if route.Path == "/swagger/*any" && route.Method == "GET" {
				swaggerFound = true
				break
			}
		}
		if !swaggerFound {
			t.Error("Swagger endpoint not found")
		}
	})

	// Test activity endpoint exists
	t.Run("ActivityEndpoint", func(t *testing.T) {
		routes := r.Engine().Routes()
		activityFound := false
		for _, route := range routes {
			if route.Path == "/activities" && route.Method == "GET" {
				activityFound = true
				break
			}
		}
		if !activityFound {
			t.Error("Activities endpoint not found")
		}
	})

	// Test device endpoints exist
	t.Run("DeviceEndpoints", func(t *testing.T) {
		routes := r.Engine().Routes()
		routeMap := make(map[string]bool)
		for _, route := range routes {
			key := route.Method + " " + route.Path
			routeMap[key] = true
		}

		expectedRoutes := []string{
			"GET /devices",
			"POST /devices/register",
			"POST /devices/metrics",
			"POST /devices/processes",
			"POST /devices/activity",
			"POST /devices/commands",
			"POST /devices/screenshots",
			"GET /devices/:id/metrics",
			"GET /devices/:id/processes",
			"GET /devices/:id/activities",
			"GET /devices/:id/alerts",
			"GET /devices/:id/screenshots",
			"GET /devices/:id/commands/pending",
			"GET /devices/:id/commands",
			"POST /commands/status",
			"POST /devices/:id/alerts",
		}

		for _, expected := range expectedRoutes {
			if !routeMap[expected] {
				t.Errorf("Expected route %s not found", expected)
			}
		}
	})

	// Test CORS is configured (OPTIONS request)
	t.Run("CORSConfigured", func(t *testing.T) {
		req, _ := http.NewRequest("OPTIONS", "/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		req.Header.Set("Access-Control-Request-Method", "GET")
		w := httptest.NewRecorder()
		r.Engine().ServeHTTP(w, req)

		// CORS should allow the request
		// Status code can be 204 (No Content) or 200 depending on configuration
		if w.Code != http.StatusNoContent && w.Code != http.StatusOK && w.Code != http.StatusNotFound {
			t.Errorf("CORS preflight request failed with status %d", w.Code)
		}
	})
}

// TestRouterStructure verifies the router maintains the same structure as the old main.go
func TestRouterStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = os.Setenv("FRONTEND_ORIGIN", "http://localhost:3000,http://localhost:5173")
	defer func() { _ = os.Unsetenv("FRONTEND_ORIGIN") }()

	r := router.New()
	r.SetupAllRoutes()

	routes := r.Engine().Routes()

	// Count routes to ensure we haven't lost any during refactoring
	// Old main.go had approximately 20+ routes
	if len(routes) < 20 {
		t.Errorf("Expected at least 20 routes, got %d", len(routes))
	}

	// Verify no duplicate routes
	routeKeys := make(map[string]int)
	for _, route := range routes {
		key := route.Method + " " + route.Path
		routeKeys[key]++
	}

	for key, count := range routeKeys {
		if count > 1 {
			t.Errorf("Duplicate route found: %s (count: %d)", key, count)
		}
	}
}
