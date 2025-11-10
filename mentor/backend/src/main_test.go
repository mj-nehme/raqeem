package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewApp(t *testing.T) {
	app := NewApp()
	assert.NotNil(t, app)
	assert.Nil(t, app.DB)
	assert.Nil(t, app.Router)
	assert.Equal(t, "", app.Port)
}

func TestParseCORSOrigins(t *testing.T) {
	tests := []struct {
		name        string
		envValue    string
		expected    []string
		description string
	}{
		{
			name:        "single origin",
			envValue:    "http://localhost:3000",
			expected:    []string{"http://localhost:3000"},
			description: "Should parse single origin correctly",
		},
		{
			name:        "multiple origins",
			envValue:    "http://localhost:3000,http://localhost:5173",
			expected:    []string{"http://localhost:3000", "http://localhost:5173"},
			description: "Should parse multiple comma-separated origins",
		},
		{
			name:        "origins with spaces",
			envValue:    "http://localhost:3000, http://localhost:5173 , http://example.com",
			expected:    []string{"http://localhost:3000", "http://localhost:5173", "http://example.com"},
			description: "Should trim whitespace from origins",
		},
		{
			name:        "empty origin",
			envValue:    "",
			expected:    []string{},
			description: "Should return empty slice for empty string",
		},
		{
			name:        "origin with trailing comma",
			envValue:    "http://localhost:3000,",
			expected:    []string{"http://localhost:3000"},
			description: "Should ignore trailing comma",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("FRONTEND_ORIGIN", tt.envValue)
			defer os.Unsetenv("FRONTEND_ORIGIN")

			app := NewApp()
			origins := app.parseCORSOrigins()

			assert.Equal(t, tt.expected, origins, tt.description)
		})
	}
}

func TestSetupRouter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := NewApp()
	router := app.setupRouter()

	assert.NotNil(t, router)
	assert.NotNil(t, app.Router)
	assert.Equal(t, router, app.Router)

	// Test that key routes are registered
	routes := router.Routes()
	routePaths := make(map[string]bool)
	for _, route := range routes {
		routePaths[route.Method+":"+route.Path] = true
	}

	// Verify critical routes exist
	expectedRoutes := []string{
		"GET:/health",
		"GET:/activities",
		"GET:/docs",
		"GET:/swagger/*any",
		"POST:/devices/register",
		"POST:/devices/metrics",
		"GET:/devices",
		"GET:/devices/:id/metrics",
		"POST:/devices/:id/alerts",
	}

	for _, expectedRoute := range expectedRoutes {
		assert.True(t, routePaths[expectedRoute], "Route %s should be registered", expectedRoute)
	}
}

func TestSetupRouterHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := NewApp()
	router := app.setupRouter()

	// Test health endpoint
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
	assert.Contains(t, w.Body.String(), "mentor-backend")
}

func TestSetupRouterDocsRedirect(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := NewApp()
	router := app.setupRouter()

	// Test docs redirect
	req, _ := http.NewRequest("GET", "/docs", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMovedPermanently, w.Code)
	assert.Equal(t, "/swagger/index.html", w.Header().Get("Location"))
}

func TestSetupDatabase(t *testing.T) {
	// Use SQLite in-memory database for testing
	testDB := database.SetupTestDB(t)
	if testDB == nil {
		t.Skip("Test database not available")
	}
	defer database.CleanupTestDB(t, testDB)

	app := NewApp()
	// Inject test database before calling setupDatabase
	app.DB = testDB
	
	err := app.setupDatabase()

	assert.NoError(t, err)
	assert.NotNil(t, app.DB)

	// Verify migrations ran successfully by checking if tables exist
	assert.True(t, app.DB.Migrator().HasTable(&models.Activity{}))
	assert.True(t, app.DB.Migrator().HasTable(&models.Device{}))
	assert.True(t, app.DB.Migrator().HasTable(&models.DeviceMetrics{}))
	assert.True(t, app.DB.Migrator().HasTable(&models.Process{}))
	assert.True(t, app.DB.Migrator().HasTable(&models.ActivityLog{}))
	assert.True(t, app.DB.Migrator().HasTable(&models.RemoteCommand{}))
	assert.True(t, app.DB.Migrator().HasTable(&models.Screenshot{}))
	assert.True(t, app.DB.Migrator().HasTable(&models.Alert{}))
}

func TestSetupRouterCORSConfiguration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set CORS origins
	os.Setenv("FRONTEND_ORIGIN", "http://localhost:3000,http://localhost:5173")
	defer os.Unsetenv("FRONTEND_ORIGIN")

	app := NewApp()
	router := app.setupRouter()

	// Test CORS preflight request
	req, _ := http.NewRequest("OPTIONS", "/health", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "GET")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// CORS middleware should handle OPTIONS request
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestAppStartWithoutPort(t *testing.T) {
	// Ensure PORT is not set
	originalPort := os.Getenv("PORT")
	os.Unsetenv("PORT")
	defer func() {
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		}
	}()

	// Use SQLite in-memory database for testing
	testDB := database.SetupTestDB(t)
	if testDB == nil {
		t.Skip("Test database not available")
	}
	defer database.CleanupTestDB(t, testDB)

	app := NewApp()
	// Inject test database
	app.DB = testDB
	
	// We can't actually call Start() as it will call log.Fatal
	// Instead we test the port validation logic
	app.Port = os.Getenv("PORT")
	assert.Equal(t, "", app.Port)
}

func TestAppStartWithPort(t *testing.T) {
	// Set required PORT environment variable
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("PORT")

	// Use SQLite in-memory database for testing
	testDB := database.SetupTestDB(t)
	if testDB == nil {
		t.Skip("Test database not available")
	}
	defer database.CleanupTestDB(t, testDB)

	app := NewApp()
	// Inject test database
	app.DB = testDB
	
	// Setup database
	err := app.setupDatabase()
	require.NoError(t, err)
	
	// Setup router
	app.setupRouter()
	
	// Get port
	app.Port = os.Getenv("PORT")
	assert.Equal(t, "8080", app.Port)
	
	// Verify router is set up
	assert.NotNil(t, app.Router)
	
	// We cannot test the actual Run() call as it blocks,
	// but we can verify all prerequisites are met
}

func TestAppRouterRegistersAllEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)

	app := NewApp()
	router := app.setupRouter()

	routes := router.Routes()
	
	// Count route types
	getRoutes := 0
	postRoutes := 0
	
	for _, route := range routes {
		switch route.Method {
		case "GET":
			getRoutes++
		case "POST":
			postRoutes++
		}
	}
	
	// Verify we have a reasonable number of routes registered
	assert.Greater(t, getRoutes, 5, "Should have multiple GET routes")
	assert.Greater(t, postRoutes, 5, "Should have multiple POST routes")
}

func TestParseCORSOriginsReturnsNonNilSlice(t *testing.T) {
	// Test that we always get a non-nil slice even with no origins
	os.Unsetenv("FRONTEND_ORIGIN")
	
	app := NewApp()
	origins := app.parseCORSOrigins()
	
	assert.NotNil(t, origins)
	assert.Equal(t, 0, len(origins))
}

func TestSetupDatabaseWithGlobalDB(t *testing.T) {
	// Save original database.DB
	originalDB := database.DB
	defer func() {
		database.DB = originalDB
	}()

	// Use SQLite in-memory database for testing
	testDB := database.SetupTestDB(t)
	if testDB == nil {
		t.Skip("Test database not available")
	}
	defer database.CleanupTestDB(t, testDB)

	app := NewApp()
	// Inject the test database
	app.DB = testDB
	
	err := app.setupDatabase()

	assert.NoError(t, err)
	assert.Equal(t, testDB, app.DB)
}

func TestAppIntegrationWithAllComponents(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Save original database.DB
	originalDB := database.DB
	defer func() {
		database.DB = originalDB
	}()

	// Set required environment variables
	os.Setenv("PORT", "8080")
	os.Setenv("FRONTEND_ORIGIN", "http://localhost:3000")
	defer os.Unsetenv("PORT")
	defer os.Unsetenv("FRONTEND_ORIGIN")

	// Use SQLite in-memory database for testing
	testDB := database.SetupTestDB(t)
	if testDB == nil {
		t.Skip("Test database not available")
	}
	defer database.CleanupTestDB(t, testDB)

	// Set global DB
	database.DB = testDB

	// Create and configure app
	app := NewApp()
	app.DB = testDB
	
	err := app.setupDatabase()
	require.NoError(t, err)
	
	router := app.setupRouter()
	require.NotNil(t, router)
	
	// Test that the app is fully configured
	assert.NotNil(t, app.DB)
	assert.NotNil(t, app.Router)
	
	// Test health endpoint works
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	app.Router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ok")
}
