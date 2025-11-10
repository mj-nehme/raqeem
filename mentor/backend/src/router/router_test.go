package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	router := New()
	assert.NotNil(t, router)
	assert.NotNil(t, router.engine)
}

func TestNewWithEngine(t *testing.T) {
	engine := gin.New()
	router := NewWithEngine(engine)
	assert.NotNil(t, router)
	assert.Equal(t, engine, router.engine)
}

func TestGetEngine(t *testing.T) {
	engine := gin.New()
	router := NewWithEngine(engine)
	assert.Equal(t, engine, router.GetEngine())
}

func TestSetupHealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := New()
	router.SetupHealthCheck()

	// Test health endpoint
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
	assert.Contains(t, w.Body.String(), "mentor-backend")

	// Test ping endpoint
	req, _ = http.NewRequest("GET", "/ping", nil)
	w = httptest.NewRecorder()
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "pong")
}

func TestSetupSwagger(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := New()
	router.SetupSwagger()

	// Test /docs redirect
	req, _ := http.NewRequest("GET", "/docs", nil)
	w := httptest.NewRecorder()
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "/swagger/index.html", w.Header().Get("Location"))

	// Test that swagger route is registered
	routes := router.GetRoutes()
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

	assert.True(t, swaggerRouteFound, "Swagger route should be registered")
	assert.True(t, docsRouteFound, "Docs route should be registered")
}

func TestSetupCORS(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := New()
	router.SetupCORS()

	// Add a simple test route
	router.engine.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "test"})
	})

	// Test CORS headers
	req, _ := http.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()
	router.engine.ServeHTTP(w, req)

	// CORS should add appropriate headers
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestSetupDeviceRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := New()
	router.SetupDeviceRoutes()

	routes := router.GetRoutes()

	expectedRoutes := map[string]string{
		"POST": "/devices/register",
		"GET":  "/devices",
	}

	routeMap := make(map[string][]string)
	for _, route := range routes {
		routeMap[route.Method] = append(routeMap[route.Method], route.Path)
	}

	for method, path := range expectedRoutes {
		found := false
		for _, registeredPath := range routeMap[method] {
			if registeredPath == path {
				found = true
				break
			}
		}
		assert.True(t, found, "Route %s %s should be registered", method, path)
	}
}

func TestSetupActivityRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := New()
	router.SetupActivityRoutes()

	routes := router.GetRoutes()
	activitiesRouteFound := false

	for _, route := range routes {
		if route.Path == "/activities" && route.Method == "GET" {
			activitiesRouteFound = true
			break
		}
	}

	assert.True(t, activitiesRouteFound, "Activities route should be registered")
}

func TestSetupAllRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := New()
	router.SetupAllRoutes()

	routes := router.GetRoutes()
	assert.Greater(t, len(routes), 0, "Routes should be registered")

	// Check that key routes exist
	routePaths := make([]string, len(routes))
	for i, route := range routes {
		routePaths[i] = route.Path
	}

	// Health check routes
	assert.Contains(t, routePaths, "/health")
	assert.Contains(t, routePaths, "/ping")

	// Swagger routes
	assert.Contains(t, routePaths, "/docs")
	assert.Contains(t, routePaths, "/swagger/*any")

	// Device routes
	assert.Contains(t, routePaths, "/devices")
	assert.Contains(t, routePaths, "/devices/register")

	// Activity routes
	assert.Contains(t, routePaths, "/activities")
}

func TestGetRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := New()
	router.SetupHealthCheck()

	routes := router.GetRoutes()
	assert.NotNil(t, routes)
	assert.Greater(t, len(routes), 0)
}

func TestRouterIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := New()
	router.SetupAllRoutes()

	// Test that the router can handle multiple requests
	tests := []struct {
		method     string
		path       string
		statusCode int
	}{
		{"GET", "/health", http.StatusOK},
		{"GET", "/ping", http.StatusOK},
		{"GET", "/docs", http.StatusMovedPermanently},
	}

	for _, test := range tests {
		t.Run(test.method+"_"+test.path, func(t *testing.T) {
			req, _ := http.NewRequest(test.method, test.path, nil)
			w := httptest.NewRecorder()
			router.engine.ServeHTTP(w, req)
			assert.Equal(t, test.statusCode, w.Code)
		})
	}
}

func TestRouterMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := New()

	// Add a custom middleware
	middleware := func() gin.HandlerFunc {
		return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
			return ""
		})
	}

	router.engine.Use(middleware())
	router.SetupHealthCheck()

	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.engine.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
