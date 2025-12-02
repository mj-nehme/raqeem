package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mentor-backend/database"
	"mentor-backend/s3"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHealthCheckReadyAllHealthy tests HealthCheckReady when all services are healthy
func TestHealthCheckReadyAllHealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database - this will skip if DB not available
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Fail if MinIO client not available
	if s3.GetClient() == nil {
		t.Fatal("Test failed - MinIO client not available")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health/ready", nil)

	HealthCheckReady(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
	assert.Equal(t, "mentor-backend", response["service"])
}

// TestHealthCheckReadyDatabaseUnavailable tests HealthCheckReady when database is unavailable
func TestHealthCheckReadyDatabaseUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Save original DB and set to nil to simulate unavailable database
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	// Create a mock MinIO client to avoid MinIO errors
	// We can't easily mock MinIO, so we'll skip if it's not available
	if s3.GetClient() == nil {
		// Skip the S3 part of the test, but we can still test database unavailability
		t.Log("MinIO not available, testing only database unavailability scenario")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health/ready", nil)

	HealthCheckReady(c)

	// When DB is nil, HealthCheck returns error, so we expect 503
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "degraded", response["status"])
}

// TestHealthCheckReadyMinIOUnavailable tests HealthCheckReady when MinIO is unavailable
func TestHealthCheckReadyMinIOUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Save original MinIO client and set to nil
	originalClient := s3.GetClient()
	s3.SetClient(nil)
	defer func() { s3.SetClient(originalClient) }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health/ready", nil)

	HealthCheckReady(c)

	// When MinIO is nil, HealthCheck returns error, so we expect 503
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "degraded", response["status"])
}

// TestHealthCheckReadyBothUnavailable tests HealthCheckReady when both services are unavailable
func TestHealthCheckReadyBothUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Save original values
	originalDB := database.DB
	originalClient := s3.GetClient()

	// Set both to nil
	database.DB = nil
	s3.SetClient(nil)

	defer func() {
		database.DB = originalDB
		s3.SetClient(originalClient)
	}()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health/ready", nil)

	HealthCheckReady(c)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, "degraded", response["status"])

	// Verify both checks are in the response
	checks, ok := response["checks"].(map[string]interface{})
	require.True(t, ok, "checks should be present in response")

	dbCheck, ok := checks["database"].(map[string]interface{})
	require.True(t, ok, "database check should be present")
	assert.Equal(t, "unhealthy", dbCheck["status"])

	minioCheck, ok := checks["minio"].(map[string]interface{})
	require.True(t, ok, "minio check should be present")
	assert.Equal(t, "unhealthy", minioCheck["status"])
}

// TestHealthCheckReadyWithMockClient tests with a mock MinIO client that can be set
func TestHealthCheckReadyWithMockClient(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Create a MinIO client pointing to a non-existent server
	// This will cause the health check to fail
	client, err := minio.New("localhost:99999", &minio.Options{
		Secure: false,
	})
	if err != nil {
		t.Fatalf("Failed to create mock MinIO client: %v", err)
	}

	// Save original and set mock
	originalClient := s3.GetClient()
	s3.SetClient(client)
	defer func() { s3.SetClient(originalClient) }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/health/ready", nil)

	HealthCheckReady(c)

	// The health check should fail because the mock client can't connect
	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}
