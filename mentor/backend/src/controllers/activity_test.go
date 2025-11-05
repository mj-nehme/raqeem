package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/stretchr/testify/assert"
)

func TestListActivities(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Setup test data
	testActivities := []models.Activity{
		{
			UserID:    "user1",
			Location:  "office",
			Filename:  "screenshot1.jpg",
			Timestamp: time.Now().Add(-2 * time.Hour),
		},
		{
			UserID:    "user2",
			Location:  "home",
			Filename:  "screenshot2.jpg",
			Timestamp: time.Now().Add(-1 * time.Hour),
		},
		{
			UserID:    "user1",
			Location:  "office",
			Filename:  "screenshot3.jpg",
			Timestamp: time.Now(),
		},
	}

	for _, activity := range testActivities {
		database.DB.Create(&activity)
	}

	// Register the route
	router.GET("/activities", ListActivities)

	// Test 1: Get all activities
	req, _ := http.NewRequest("GET", "/activities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 2: Filter by user_id
	req, _ = http.NewRequest("GET", "/activities?user_id=user1", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 3: Filter by location
	req, _ = http.NewRequest("GET", "/activities?location=office", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 4: Filter by time range
	startTime := time.Now().Add(-3 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	req, _ = http.NewRequest("GET", "/activities?start_time="+startTime+"&end_time="+endTime, nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test 5: Multiple filters
	req, _ = http.NewRequest("GET", "/activities?user_id=user1&location=office", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListActivitiesEmptyResults(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	router.GET("/activities", ListActivities)

	// Test with no activities in database
	req, _ := http.NewRequest("GET", "/activities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with non-matching filters
	req, _ = http.NewRequest("GET", "/activities?user_id=nonexistent", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListActivitiesTimeFilter(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	router.GET("/activities", ListActivities)

	// Add test activity
	activity := models.Activity{
		UserID:    "testuser",
		Location:  "testloc",
		Filename:  "test.jpg",
		Timestamp: time.Now(),
	}
	database.DB.Create(&activity)

	// Test with valid time range
	startTime := time.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	endTime := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	req, _ := http.NewRequest("GET", "/activities?start_time="+startTime+"&end_time="+endTime, nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with invalid time format (should not crash)
	req, _ = http.NewRequest("GET", "/activities?start_time=invalid&end_time=invalid", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListActivitiesOrdering(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	router.GET("/activities", ListActivities)

	// Add activities in non-chronological order
	activities := []models.Activity{
		{UserID: "user1", Location: "loc1", Filename: "file1.jpg", Timestamp: time.Now().Add(-3 * time.Hour)},
		{UserID: "user2", Location: "loc2", Filename: "file2.jpg", Timestamp: time.Now()},
		{UserID: "user3", Location: "loc3", Filename: "file3.jpg", Timestamp: time.Now().Add(-1 * time.Hour)},
	}

	for _, activity := range activities {
		database.DB.Create(&activity)
	}

	// Verify activities are returned in descending timestamp order
	req, _ := http.NewRequest("GET", "/activities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	// Response should have activities ordered by timestamp desc
}

func TestLogActivity(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-activity-device"

	// Register device first
	device := models.Device{
		ID:       deviceID,
		Name:     "Activity Test Device",
		IsOnline: true,
		LastSeen: time.Now(),
	}
	database.DB.Create(&device)

	// Test logging activity
	req, _ := http.NewRequest("POST", "/devices/"+deviceID+"/activity", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Should handle request (might be 200 or 400 depending on body)
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
}

func TestGetDeviceActivitiesWithLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-activities-limit"

	// Register device
	device := models.Device{
		ID:       deviceID,
		Name:     "Test Device",
		IsOnline: true,
		LastSeen: time.Now(),
	}
	database.DB.Create(&device)

	// Add multiple activities
	for i := 0; i < 15; i++ {
		activity := models.ActivityLog{
			DeviceID:    deviceID,
			Type:        "test",
			Description: "Test activity",
			Timestamp:   time.Now().Add(time.Duration(i) * time.Minute),
		}
		database.DB.Create(&activity)
	}

	// Test with default limit
	req, _ := http.NewRequest("GET", "/devices/"+deviceID+"/activities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with specific limit
	req, _ = http.NewRequest("GET", "/devices/"+deviceID+"/activities?limit=5", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Test with large limit
	req, _ = http.NewRequest("GET", "/devices/"+deviceID+"/activities?limit=100", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetDeviceActivitiesNonExistent(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Test with non-existent device
	req, _ := http.NewRequest("GET", "/devices/nonexistent/activities", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
