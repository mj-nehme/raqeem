package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "mentor-backend/docs"

	"github.com/gin-gonic/gin"
)

func TestNew(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New()

	if r == nil {
		t.Fatal("Expected router to be created, got nil")
	}

	if r.engine == nil {
		t.Error("Expected router.engine to be initialized, got nil")
	}
}

func TestEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New()
	engine := r.Engine()

	if engine == nil {
		t.Error("Expected Engine() to return non-nil gin.Engine")
	}

	if engine != r.engine {
		t.Error("Expected Engine() to return the same engine instance")
	}
}

func TestSetupSwagger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New()
	r.setupSwagger()

	t.Run("DocsEndpointRedirect", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/docs", nil)
		w := httptest.NewRecorder()
		r.engine.ServeHTTP(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Errorf("Expected status code %d, got %d", http.StatusMovedPermanently, w.Code)
		}

		location := w.Header().Get("Location")
		expectedLocation := "/swagger/index.html"
		if location != expectedLocation {
			t.Errorf("Expected redirect to %s, got %s", expectedLocation, location)
		}
	})

	t.Run("SwaggerRouteRegistered", func(t *testing.T) {
		routes := r.engine.Routes()
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

func TestSetupHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New()
	r.setupHealthCheck()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	// Check that health endpoint is registered
	routes := r.engine.Routes()
	healthRouteFound := false
	for _, route := range routes {
		if route.Path == "/health" && route.Method == "GET" {
			healthRouteFound = true
			break
		}
	}

	if !healthRouteFound {
		t.Error("Expected /health route to be registered")
	}
}

func TestSetupCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("WithSingleOrigin", func(t *testing.T) {
		_ = os.Setenv("FRONTEND_ORIGIN", "http://localhost:3000")
		defer func() { _ = os.Unsetenv("FRONTEND_ORIGIN") }()

		r := New()
		r.setupCORS()

		// CORS middleware is added, verify by making a request
		req, _ := http.NewRequest("OPTIONS", "/health", nil)
		req.Header.Set("Origin", "http://localhost:3000")
		w := httptest.NewRecorder()
		r.engine.ServeHTTP(w, req)

		// CORS middleware should handle OPTIONS requests
		// Note: Just verify setup doesn't panic
	})

	t.Run("WithMultipleOrigins", func(t *testing.T) {
		_ = os.Setenv("FRONTEND_ORIGIN", "http://localhost:3000,http://localhost:5173")
		defer func() { _ = os.Unsetenv("FRONTEND_ORIGIN") }()

		r := New()
		r.setupCORS()

		// CORS middleware is added, verify setup doesn't panic
		req, _ := http.NewRequest("OPTIONS", "/health", nil)
		w := httptest.NewRecorder()
		r.engine.ServeHTTP(w, req)
	})

	t.Run("WithEmptyOrigin", func(t *testing.T) {
		_ = os.Setenv("FRONTEND_ORIGIN", "")
		defer func() { _ = os.Unsetenv("FRONTEND_ORIGIN") }()

		r := New()
		r.setupCORS()

		// CORS middleware should still be added even with empty origins
		req, _ := http.NewRequest("OPTIONS", "/health", nil)
		w := httptest.NewRecorder()
		r.engine.ServeHTTP(w, req)
	})

	t.Run("WithWhitespaceOrigins", func(t *testing.T) {
		_ = os.Setenv("FRONTEND_ORIGIN", "http://localhost:3000,  , http://localhost:5173,  ")
		defer func() { _ = os.Unsetenv("FRONTEND_ORIGIN") }()

		r := New()
		r.setupCORS()

		// CORS middleware should handle whitespace correctly
		req, _ := http.NewRequest("OPTIONS", "/health", nil)
		w := httptest.NewRecorder()
		r.engine.ServeHTTP(w, req)
	})

	t.Run("WithOriginRegex", func(t *testing.T) {
		// Test CORS with regex pattern for dynamic port ranges
		_ = os.Setenv("FRONTEND_ORIGIN_REGEX", "^http://localhost:(4000|4001|4002|4003|4004)$")
		defer func() { _ = os.Unsetenv("FRONTEND_ORIGIN_REGEX") }()

		r := New()
		r.setupCORS()
		r.setupHealthCheck()

		// Test with matching origin
		req, _ := http.NewRequest("OPTIONS", "/health", nil)
		req.Header.Set("Origin", "http://localhost:4002")
		w := httptest.NewRecorder()
		r.engine.ServeHTTP(w, req)

		// Should allow the matching origin
		if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
			t.Errorf("Expected CORS preflight to succeed, got status %d", w.Code)
		}
	})
}

func TestSetupActivityRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New()
	r.setupActivityRoutes()

	routes := r.engine.Routes()
	activitiesRouteFound := false

	for _, route := range routes {
		if route.Path == "/activities" && route.Method == "GET" {
			activitiesRouteFound = true
			break
		}
	}

	if !activitiesRouteFound {
		t.Error("Expected /activities route to be registered")
	}
}

func TestSetupDeviceRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := New()
	r.setupDeviceRoutes()

	// Mentor backend is read-focused with command creation
	expectedRoutes := map[string]string{
		// POST routes (only command creation for mentor)
		"POST /devices/commands": "POST",

		// GET routes (read-only for dashboards)
		"GET /devices":                      "GET",
		"GET /devices/:id/metrics":          "GET",
		"GET /devices/:id/processes":        "GET",
		"GET /devices/:id/activities":       "GET",
		"GET /devices/:id/alerts":           "GET",
		"GET /devices/:id/screenshots":      "GET",
		"GET /screenshots/:filename":        "GET",
		"GET /devices/:id/commands/pending": "GET",
		"GET /devices/:id/commands":         "GET",
	}

	// Routes that should NOT exist on mentor (device-only write operations)
	removedRoutes := []string{
		"POST /devices/register",
		"POST /devices/metrics",
		"POST /devices/processes",
		"POST /devices/activity",
		"POST /devices/screenshots",
		"POST /commands/status",
		"POST /devices/:id/alerts",
	}

	routes := r.engine.Routes()
	routeMap := make(map[string]bool)
	for _, route := range routes {
		key := route.Method + " " + route.Path
		routeMap[key] = true
	}

	for expectedRoute, method := range expectedRoutes {
		if !routeMap[expectedRoute] {
			t.Errorf("Expected route %s to be registered", expectedRoute)
		}
		_ = method // Use variable to avoid unused warning
	}

	// Verify removed routes are not present
	for _, removedRoute := range removedRoutes {
		if routeMap[removedRoute] {
			t.Errorf("Route %s should NOT be registered on mentor backend (device-only)", removedRoute)
		}
	}
}

func TestSetupAllRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = os.Setenv("FRONTEND_ORIGIN", "http://localhost:3000")
	defer func() { _ = os.Unsetenv("FRONTEND_ORIGIN") }()

	r := New()
	r.SetupAllRoutes()

	routes := r.engine.Routes()
	if len(routes) == 0 {
		t.Error("Expected routes to be registered")
	}

	// Verify key routes are present (mentor is read-focused with command creation)
	keyRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/health"},
		{"GET", "/docs"},
		{"GET", "/swagger/*any"},
		{"GET", "/activities"},
		{"GET", "/devices"},
		{"POST", "/devices/commands"},
	}

	routeMap := make(map[string]bool)
	for _, route := range routes {
		key := route.Method + " " + route.Path
		routeMap[key] = true
	}

	for _, kr := range keyRoutes {
		key := kr.method + " " + kr.path
		if !routeMap[key] {
			t.Errorf("Expected key route %s to be registered", key)
		}
	}

	// Verify CORS middleware is present by testing OPTIONS request
	req, _ := http.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	r.engine.ServeHTTP(w, req)
	// Just verify setup works without panicking
}

func TestRun(t *testing.T) {
	// Note: We can't easily test Run() as it starts a blocking server
	// This test just verifies the method exists and has the correct signature
	gin.SetMode(gin.TestMode)
	r := New()

	// Test that Run exists and returns an error
	// We'll test with an invalid address to ensure it returns quickly
	err := r.Run("invalid:address:format")
	if err == nil {
		t.Error("Expected Run() to return error for invalid address")
	}
}

func TestRouterIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_ = os.Setenv("FRONTEND_ORIGIN", "http://localhost:3000")
	defer func() { _ = os.Unsetenv("FRONTEND_ORIGIN") }()

	// Create router and setup all routes
	r := New()
	r.SetupAllRoutes()

	// Test health endpoint
	t.Run("HealthEndpoint", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		r.engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
		}
	})

	// Test docs redirect
	t.Run("DocsRedirect", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/docs", nil)
		w := httptest.NewRecorder()
		r.engine.ServeHTTP(w, req)

		if w.Code != http.StatusMovedPermanently {
			t.Errorf("Expected status code %d, got %d", http.StatusMovedPermanently, w.Code)
		}
	})

	// Verify total route count is reasonable (mentor has fewer routes now - read-focused)
	routes := r.engine.Routes()
	if len(routes) < 12 {
		t.Errorf("Expected at least 12 routes, got %d", len(routes))
	}
}
