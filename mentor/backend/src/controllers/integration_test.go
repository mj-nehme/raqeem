package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRouterWithDB sets up test router with real database
func setupTestRouterWithDB(t *testing.T) (*gin.Engine, func()) {
	gin.SetMode(gin.TestMode)

	// Setup test database
	testDB := database.SetupTestDB(t)
	if testDB == nil {
		t.Skip("Test database not available")
	}

	// Set the global DB variable for controllers to use
	database.DB = testDB

	router := gin.New()
	router.Use(gin.Recovery())

	// Register all routes
	router.POST("/devices", RegisterDevice)
	router.POST("/devices/:id/metrics", UpdateDeviceMetrics)
	router.POST("/devices/:id/activity", LogActivity)
	router.POST("/devices/:id/processes", UpdateProcessList)
	router.GET("/devices", ListDevices)
	router.GET("/devices/:id/metrics", GetDeviceMetrics)
	router.GET("/devices/:id/processes", GetDeviceProcesses)
	router.GET("/devices/:id/activities", GetDeviceActivities)
	router.GET("/devices/:id/alerts", GetDeviceAlerts)
	router.GET("/devices/:id/screenshots", GetDeviceScreenshots)
	router.POST("/devices/:id/commands", CreateRemoteCommand)
	router.GET("/devices/:id/commands/pending", GetPendingCommands)
	router.PUT("/commands/:id/status", UpdateCommandStatus)
	router.POST("/devices/:id/alerts", ReportAlert)

	// Return cleanup function
	cleanup := func() {
		database.CleanupTestDB(t, testDB)
	}

	return router, cleanup
}

func TestDeviceLifecycleIntegration(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "integration-test-device-001"

	// 1. Register a new device
	device := models.Device{
		ID:          deviceID,
		Name:        "Integration Test Device",
		Type:        "laptop",
		OS:          "Ubuntu",
		IPAddress:   "192.168.1.100",
		MacAddress:  "aa:bb:cc:dd:ee:ff",
		Location:    "Test Lab",
		IsOnline:    true,
		CurrentUser: "testuser",
		LastSeen:    time.Now(),
	}

	deviceJSON, _ := json.Marshal(device)
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var registerResponse models.Device
	err := json.Unmarshal(w.Body.Bytes(), &registerResponse)
	require.NoError(t, err)
	assert.Equal(t, deviceID, registerResponse.ID)

	// 2. Update device metrics
	metrics := models.DeviceMetrics{
		DeviceID:    deviceID,
		CPUUsage:    75.5,
		CPUTemp:     65.0,
		MemoryTotal: 16777216000, // 16GB
		MemoryUsed:  10737418240, // 10GB
		SwapUsed:    0,
		DiskTotal:   1000000000000, // 1TB
		DiskUsed:    500000000000,  // 500GB
		NetBytesIn:  1024,
		NetBytesOut: 512,
		Timestamp:   time.Now(),
	}

	metricsJSON, _ := json.Marshal(metrics)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/metrics", deviceID), bytes.NewBuffer(metricsJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 3. Log activity
	activity := models.ActivityLog{
		DeviceID:    deviceID,
		Type:        "app_launch",
		Description: "User opened Firefox browser",
		App:         "Firefox",
		Duration:    3600,
		Timestamp:   time.Now(),
	}

	activityJSON, _ := json.Marshal(activity)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/activity", deviceID), bytes.NewBuffer(activityJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 4. Update process list
	processes := []models.Process{
		{
			DeviceID:  deviceID,
			PID:       1234,
			Name:      "firefox",
			CPU:       15.5,
			Memory:    536870912, // 512MB in bytes
			Command:   "/usr/bin/firefox",
			Timestamp: time.Now(),
		},
		{
			DeviceID:  deviceID,
			PID:       5678,
			Name:      "code",
			CPU:       8.2,
			Memory:    268435456, // 256MB in bytes
			Command:   "/usr/bin/code",
			Timestamp: time.Now(),
		},
	}

	processesJSON, _ := json.Marshal(processes)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/processes", deviceID), bytes.NewBuffer(processesJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 5. Create a remote command
	command := models.RemoteCommand{
		DeviceID:  deviceID,
		Command:   "ls -la /home",
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	commandJSON, _ := json.Marshal(command)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/commands", deviceID), bytes.NewBuffer(commandJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var commandResponse models.RemoteCommand
	err = json.Unmarshal(w.Body.Bytes(), &commandResponse)
	require.NoError(t, err)
	assert.NotZero(t, commandResponse.ID)
	commandID := commandResponse.ID

	// 6. Report an alert
	alert := models.Alert{
		DeviceID:  deviceID,
		Type:      "security",
		Level:     "critical",
		Message:   "Suspicious network activity detected",
		Value:     95.0,
		Threshold: 90.0,
		Timestamp: time.Now(),
	}

	alertJSON, _ := json.Marshal(alert)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/alerts", deviceID), bytes.NewBuffer(alertJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 7. List all devices
	req, _ = http.NewRequest("GET", "/devices", nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var devicesResponse []models.Device
	err = json.Unmarshal(w.Body.Bytes(), &devicesResponse)
	require.NoError(t, err)

	assert.GreaterOrEqual(t, len(devicesResponse), 1)

	// 8. Get device metrics
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=10", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var metricsResponse []models.DeviceMetrics
	err = json.Unmarshal(w.Body.Bytes(), &metricsResponse)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(metricsResponse), 1)

	// 9. Get device processes
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/processes?limit=10", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var processesResponse []models.Process
	err = json.Unmarshal(w.Body.Bytes(), &processesResponse)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(processesResponse), 2)

	// 10. Get device activities
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/activities?limit=10", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var activitiesResponse []models.ActivityLog
	err = json.Unmarshal(w.Body.Bytes(), &activitiesResponse)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(activitiesResponse), 1)

	// 11. Get device alerts
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/alerts?limit=10", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var alertsResponse []models.Alert
	err = json.Unmarshal(w.Body.Bytes(), &alertsResponse)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(alertsResponse), 1)

	// 12. Get pending commands
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/commands/pending", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var commandsResponse []models.RemoteCommand
	err = json.Unmarshal(w.Body.Bytes(), &commandsResponse)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(commandsResponse), 1)

	// 13. Update command status
	statusUpdate := models.RemoteCommand{
		ID:     commandID,
		Status: "completed",
		Result: "total 24\ndrwxr-xr-x 3 user user 4096 Nov  3 10:00 .",
	}

	statusJSON, _ := json.Marshal(statusUpdate)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/commands/%d/status", commandID), bytes.NewBuffer(statusJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var updateResponse models.RemoteCommand
	err = json.Unmarshal(w.Body.Bytes(), &updateResponse)
	require.NoError(t, err)
	assert.Equal(t, "completed", updateResponse.Status)
}

func TestDeviceMetricsIntegration(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "metrics-test-device"

	// Register device first
	device := models.Device{
		ID:       deviceID,
		Name:     "Metrics Test Device",
		Type:     "server",
		OS:       "Linux",
		IsOnline: true,
		LastSeen: time.Now(),
	}

	deviceJSON, _ := json.Marshal(device)
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Add multiple metrics entries
	for i := 0; i < 5; i++ {
		metrics := models.DeviceMetrics{
			DeviceID:    deviceID,
			CPUUsage:    float64(50 + i*10),
			CPUTemp:     float64(55 + i*5),
			MemoryTotal: 16777216000,
			MemoryUsed:  uint64(8000000000 + i*1000000000), // Increasing memory usage
			DiskTotal:   1000000000000,
			DiskUsed:    uint64(300000000000 + i*50000000000), // Increasing disk usage
			NetBytesIn:  uint64(100 + i*50),
			NetBytesOut: uint64(50 + i*25),
			Timestamp:   time.Now().Add(time.Duration(i) * time.Minute),
		}

		metricsJSON, _ := json.Marshal(metrics)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/metrics", deviceID), bytes.NewBuffer(metricsJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Test retrieving metrics with different limits
	testCases := []struct {
		limit    string
		expected int
	}{
		{"1", 1},
		{"3", 3},
		{"10", 5}, // Should return all 5 metrics
	}

	for _, tc := range testCases {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=%s", deviceID, tc.limit), nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var metricsResponse []models.DeviceMetrics
		err := json.Unmarshal(w.Body.Bytes(), &metricsResponse)
		require.NoError(t, err)
		assert.Equal(t, tc.expected, len(metricsResponse))
	}
}

func TestAlertFlowIntegration(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "alert-test-device"

	// Register device
	device := models.Device{
		ID:       deviceID,
		Name:     "Alert Test Device",
		Type:     "workstation",
		OS:       "Windows",
		IsOnline: true,
		LastSeen: time.Now(),
	}

	deviceJSON, _ := json.Marshal(device)
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	// Create alerts of different types and levels
	alertTests := []struct {
		alertType string
		level     string
		message   string
	}{
		{"security", "critical", "Malware detected"},
		{"performance", "warning", "High CPU usage"},
		{"security", "high", "Failed login attempts"},
		{"system", "info", "System update available"},
	}

	for _, alertTest := range alertTests {
		alert := models.Alert{
			DeviceID:  deviceID,
			Type:      alertTest.alertType,
			Level:     alertTest.level,
			Message:   alertTest.message,
			Value:     85.0,
			Threshold: 80.0,
			Timestamp: time.Now(),
		}

		alertJSON, _ := json.Marshal(alert)
		req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/alerts", deviceID), bytes.NewBuffer(alertJSON))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Retrieve and verify alerts
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/alerts", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var alertsResponse []models.Alert
	err := json.Unmarshal(w.Body.Bytes(), &alertsResponse)
	require.NoError(t, err)
	assert.Equal(t, len(alertTests), len(alertsResponse))

	// Verify alert levels are correctly stored
	levelCounts := make(map[string]int)
	for _, alert := range alertsResponse {
		levelCounts[alert.Level]++
	}

	assert.Equal(t, 1, levelCounts["critical"])
	assert.Equal(t, 1, levelCounts["warning"])
	assert.Equal(t, 1, levelCounts["high"])
	assert.Equal(t, 1, levelCounts["info"])
}

func TestErrorHandlingIntegration(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Test duplicate device registration
	device := models.Device{
		ID:   "duplicate-test",
		Name: "Duplicate Test Device",
		Type: "laptop",
		OS:   "macOS",
	}

	deviceJSON, _ := json.Marshal(device)

	// First registration should succeed
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second registration should handle duplicate appropriately
	req, _ = http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// This might be 200 (updated) or 409 (conflict) depending on implementation
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusConflict)

	// Test invalid data handling
	invalidJSON := `{"invalid": "json structure"`
	req, _ = http.NewRequest("POST", "/devices", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test non-existent device operations
	req, _ = http.NewRequest("GET", "/devices/non-existent-device/metrics", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code) // Should return empty array

	var metricsResponse []models.DeviceMetrics
	err := json.Unmarshal(w.Body.Bytes(), &metricsResponse)
	require.NoError(t, err)
	assert.Equal(t, 0, len(metricsResponse))
}
