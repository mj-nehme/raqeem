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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterDevice_ErrorCases(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Test invalid JSON
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBufferString(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test empty request body
	req, _ = http.NewRequest("POST", "/devices", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test missing required fields
	device := map[string]interface{}{
		"devicename": "Test Device",
		// Missing ID
	}
	deviceJSON, _ := json.Marshal(device)
	req, _ = http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	// Should still succeed but with empty ID
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateDeviceMetric_ErrorCases(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := sampleUUID.String()

	// Test invalid JSON
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/metrics", deviceID), bytes.NewBufferString(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test empty request body
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/metrics", deviceID), bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test metrics with invalid data types
	invalidMetrics := map[string]interface{}{
		"device_id":    deviceID,
		"cpu_usage":    "not_a_number", // Should be float64
		"memory_total": true,           // Should be uint64
	}
	metricsJSON, _ := json.Marshal(invalidMetrics)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/metrics", deviceID), bytes.NewBuffer(metricsJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestActivity_ErrorCases(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-activity-error"

	// Test invalid JSON
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/activity", deviceID), bytes.NewBufferString(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test empty request body
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/activity", deviceID), bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateProcessList_ErrorCases(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-process-error"

	// Test invalid JSON
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/processes", deviceID), bytes.NewBufferString(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test empty request body
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/processes", deviceID), bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test empty process list (should succeed)
	emptyProcesses := []models.DeviceProcess{}
	processesJSON, _ := json.Marshal(emptyProcesses)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/processes", deviceID), bytes.NewBuffer(processesJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateRemoteCommand_ErrorCases(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-command-error"

	// Test invalid JSON
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/commands", deviceID), bytes.NewBufferString(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test empty request body
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/commands", deviceID), bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test command with empty command string
	command := models.DeviceRemoteCommand{
		DeviceID:    sampleUUID,
		CommandText: "", // Empty command
	}
	commandJSON, _ := json.Marshal(command)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/commands", deviceID), bytes.NewBuffer(commandJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code) // Should still succeed with empty command
}

func TestUpdateCommandStatus_ErrorCases(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Test invalid JSON
	req, _ := http.NewRequest("PUT", "/commands/999/status", bytes.NewBufferString(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test empty request body
	req, _ = http.NewRequest("PUT", "/commands/999/status", bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test updating non-existent command
	statusUpdate := models.DeviceRemoteCommand{
		CommandID: sampleUUID, // Non-existent ID
		Status:    "completed",
	}
	statusJSON, _ := json.Marshal(statusUpdate)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/commands/%s/status", statusUpdate.CommandID), bytes.NewBuffer(statusJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code) // Should succeed even if command doesn't exist
}

func TestReportAlert_ErrorCases(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-alert-error"

	// Test invalid JSON
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/alerts", deviceID), bytes.NewBufferString(`{"invalid": json}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test empty request body
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/alerts", deviceID), bytes.NewBuffer(nil))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDeviceMetric_QueryParameters(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-metrics-query"

	// Register device and add some metrics first
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	database.DB.Create(&device)

	// Add multiple metrics
	for i := 0; i < 5; i++ {
		metrics := models.DeviceMetric{
			DeviceID:  sampleUUID,
			CPUUsage:  float64(10 + i*10),
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
		database.DB.Create(&metrics)
	}

	// Test invalid limit parameter
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=invalid", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test valid limit parameter
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=3", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var metrics []models.DeviceMetric
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	require.NoError(t, err)
	assert.LessOrEqual(t, len(metrics), 3)

	// Test no limit parameter (should use default)
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetDeviceProcesses_QueryParameters(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-processes-query"

	// Test invalid limit parameter
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/processes?limit=invalid", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test valid limit parameter
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/processes?limit=5", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetDeviceActivity_QueryParameters(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-activities-query"

	// Test invalid limit parameter
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/activities?limit=invalid", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test valid limit parameter
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/activities?limit=5", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetDeviceAlert_QueryParameters(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-alerts-query"

	// Test invalid limit parameter
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/alerts?limit=invalid", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test valid limit parameter
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/alerts?limit=5", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetDeviceScreenshot_QueryParameters(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-screenshots-query"

	// Test invalid limit parameter
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/screenshots?limit=invalid", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test valid limit parameter
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/screenshots?limit=10", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Response should be an array
	var screenshots []interface{}
	err := json.Unmarshal(w.Body.Bytes(), &screenshots)
	require.NoError(t, err)
}

func TestGetPendingCommands_EdgeCases(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-pending-commands"

	// Test device with no commands
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/commands/pending", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var commands []models.DeviceRemoteCommand
	err := json.Unmarshal(w.Body.Bytes(), &commands)
	require.NoError(t, err)
	assert.Equal(t, 0, len(commands))

	// Create some commands with different statuses
	pendingCmd := models.DeviceRemoteCommand{
		DeviceID:    sampleUUID,
		CommandText: "echo pending",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	database.DB.Create(&pendingCmd)

	now := time.Now()
	completedCmd := models.DeviceRemoteCommand{
		DeviceID:    sampleUUID,
		CommandText: "echo completed",
		Status:      "completed",
		CreatedAt:   now,
		CompletedAt: now,
	}
	database.DB.Create(&completedCmd)

	// Test that only pending commands are returned
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/commands/pending", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &commands)
	require.NoError(t, err)
	assert.Equal(t, 1, len(commands))
	assert.Equal(t, "pending", commands[0].Status)
}

func TestDeviceOnlineStatusUpdate(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Create devices with different last seen times
	oldDevice := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Old Device",
		IsOnline:   true,
		LastSeen:   time.Now().Add(-10 * time.Minute), // 10 minutes ago
	}
	database.DB.Create(&oldDevice)

	recentDevice := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Recent Device",
		IsOnline:   true,
		LastSeen:   time.Now().Add(-2 * time.Minute), // 2 minutes ago
	}
	database.DB.Create(&recentDevice)

	// List devices should mark old device as offline
	req, _ := http.NewRequest("GET", "/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var devices []models.Device
	err := json.Unmarshal(w.Body.Bytes(), &devices)
	require.NoError(t, err)

	// Find our test devices in the response
	var oldDeviceResponse, recentDeviceResponse *models.Device
	for i := range devices {
		if devices[i].DeviceID == sampleUUID {
			oldDeviceResponse = &devices[i]
		}
		if devices[i].DeviceID == sampleUUID {
			recentDeviceResponse = &devices[i]
		}
	}

	require.NotNil(t, oldDeviceResponse)
	require.NotNil(t, recentDeviceResponse)

	// Old device should be marked as offline
	assert.False(t, oldDeviceResponse.IsOnline)
	// Recent device should still be online
	assert.True(t, recentDeviceResponse.IsOnline)
}

func TestDeviceMetricTimestampHandling(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-timestamp"

	// Register device first
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Timestamp Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	database.DB.Create(&device)

	// Test metrics without timestamp (should be auto-set)
	metrics := models.DeviceMetric{
		DeviceID: sampleUUID,
		CPUUsage: 50.0,
		// Timestamp will be auto-set by controller
	}

	metricsJSON, _ := json.Marshal(metrics)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/metrics", deviceID), bytes.NewBuffer(metricsJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.DeviceMetric
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotZero(t, response.Timestamp)
	assert.WithinDuration(t, time.Now(), response.Timestamp, 5*time.Second)
}

func TestActivityLogTimestampHandling(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-activity-timestamp"

	// Test activity without timestamp (should be auto-set)
	activity := models.DeviceActivity{
		DeviceID:    sampleUUID,
		Type:        "test",
		Description: "Test activity",
		// Timestamp will be auto-set by controller
	}

	activityJSON, _ := json.Marshal(activity)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/activity", deviceID), bytes.NewBuffer(activityJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.DeviceActivity
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotZero(t, response.Timestamp)
	assert.WithinDuration(t, time.Now(), response.Timestamp, 5*time.Second)
}

func TestProcessListTransaction(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-process-transaction"

	// Create initial processes
	initialProcesses := []models.DeviceProcess{
		{
			DeviceID:    sampleUUID,
			PID:         100,
			ProcessName: "initial-process",
			CPU:         10.0,
			Memory:      1000,
			CommandText: "initial",
		},
	}

	processesJSON, _ := json.Marshal(initialProcesses)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/processes", deviceID), bytes.NewBuffer(processesJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update with new processes (should replace old ones)
	newProcesses := []models.DeviceProcess{
		{
			DeviceID:    sampleUUID,
			PID:         200,
			ProcessName: "new-process-1",
			CPU:         20.0,
			Memory:      2000,
			CommandText: "new1",
		},
		{
			DeviceID:    sampleUUID,
			PID:         300,
			ProcessName: "new-process-2",
			CPU:         30.0,
			Memory:      3000,
			CommandText: "new2",
		},
	}

	processesJSON, _ = json.Marshal(newProcesses)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/processes", deviceID), bytes.NewBuffer(processesJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify old processes are gone and new ones exist
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/processes", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var retrievedProcesses []models.DeviceProcess
	err := json.Unmarshal(w.Body.Bytes(), &retrievedProcesses)
	require.NoError(t, err)

	// Should only have the new processes
	assert.Equal(t, 2, len(retrievedProcesses))

	// Check that initial process is not in the results
	for _, p := range retrievedProcesses {
		assert.NotEqual(t, 100, p.PID)
		assert.NotEqual(t, "initial-process", p.ProcessName)
	}
}

func TestRemoteCommandStatusTransitions(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Create a command
	command := models.DeviceRemoteCommand{
		DeviceID:    sampleUUID,
		CommandText: "test command",
	}

	commandJSON, _ := json.Marshal(command)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/commands", sampleUUID.String()), bytes.NewBuffer(commandJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var createdCommand models.DeviceRemoteCommand
	err := json.Unmarshal(w.Body.Bytes(), &createdCommand)
	require.NoError(t, err)
	assert.Equal(t, "pending", createdCommand.Status)
	assert.NotZero(t, createdCommand.CommandID)

	commandID := createdCommand.CommandID

	// Test status update to "completed"
	statusUpdate := models.DeviceRemoteCommand{
		CommandID: commandID,
		Status:    "completed",
		Result:    "Command executed successfully",
	}

	statusJSON, _ := json.Marshal(statusUpdate)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/commands/%d/status", commandID), bytes.NewBuffer(statusJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var updatedCommand models.DeviceRemoteCommand
	err = json.Unmarshal(w.Body.Bytes(), &updatedCommand)
	require.NoError(t, err)
	assert.Equal(t, "completed", updatedCommand.Status)
	assert.Equal(t, "Command executed successfully", updatedCommand.Result)

	// Test status update to "failed"
	failedUpdate := models.DeviceRemoteCommand{
		CommandID: commandID,
		Status:    "failed",
		Result:    "Command failed",
		ExitCode:  1,
	}

	failedJSON, _ := json.Marshal(failedUpdate)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/commands/%d/status", commandID), bytes.NewBuffer(failedJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	err = json.Unmarshal(w.Body.Bytes(), &updatedCommand)
	require.NoError(t, err)
	assert.Equal(t, "failed", updatedCommand.Status)
	assert.Equal(t, "Command failed", updatedCommand.Result)
	assert.Equal(t, 1, updatedCommand.ExitCode)
}

func TestAlertTimestampHandling(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-alert-timestamp"

	// Test alert without timestamp (should be auto-set)
	alert := models.DeviceAlert{
		DeviceID:  sampleUUID,
		Type:      "test",
		Level:     "info",
		Message:   "Test alert",
		Value:     50.0,
		Threshold: 75.0,
		// Timestamp will be auto-set by controller
	}

	alertJSON, _ := json.Marshal(alert)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/alerts", deviceID), bytes.NewBuffer(alertJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.DeviceAlert
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.NotZero(t, response.Timestamp)
	assert.WithinDuration(t, time.Now(), response.Timestamp, 5*time.Second)
}

func TestDeviceLastSeenUpdate(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-lastseen"

	// Register device
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Last Seen Test Device",
		IsOnline:   true,
		LastSeen:   time.Now().Add(-1 * time.Hour), // 1 hour ago
	}
	database.DB.Create(&device)

	// Update metrics (should update last seen)
	metrics := models.DeviceMetric{
		DeviceID: sampleUUID,
		CPUUsage: 25.0,
	}

	beforeUpdate := time.Now()
	metricsJSON, _ := json.Marshal(metrics)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/metrics", deviceID), bytes.NewBuffer(metricsJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that device's last seen was updated
	var updatedDevice models.Device
	database.DB.Where("id = ?", deviceID).First(&updatedDevice)
	assert.True(t, updatedDevice.LastSeen.After(beforeUpdate))
	assert.True(t, updatedDevice.IsOnline)
}

func TestLargeDataHandling(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-device-large-data"

	// Test with large process list
	largeProcessList := make([]models.DeviceProcess, 100)
	for i := 0; i < 100; i++ {
		largeProcessList[i] = models.DeviceProcess{
			DeviceID:    sampleUUID,
			PID:         1000 + i,
			ProcessName: fmt.Sprintf("process-%d", i),
			CPU:         float64(i % 100),
			Memory:      uint64(1000000 + i*1000),
			CommandText: fmt.Sprintf("/usr/bin/process-%d --option=%d", i, i),
		}
	}

	processesJSON, _ := json.Marshal(largeProcessList)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/processes", deviceID), bytes.NewBuffer(processesJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Retrieve and verify large process list
	req, _ = http.NewRequest("GET", fmt.Sprintf("/devices/%s/processes?limit=150", deviceID), nil)
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var retrievedProcesses []models.DeviceProcess
	err := json.Unmarshal(w.Body.Bytes(), &retrievedProcesses)
	require.NoError(t, err)
	assert.Equal(t, 100, len(retrievedProcesses))
}
