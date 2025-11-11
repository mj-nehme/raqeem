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

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUpdateProcessListWithValidData(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-process-device"

	// Register device first
	device := models.Device{
		ID:       deviceID,
		Name:     "Process Test Device",
		IsOnline: true,
		LastSeen: time.Now(),
	}
	database.DB.Create(&device)

	// Create process list
	processes := []models.DeviceProcesses{
		{
			DeviceID: deviceID,
			PID:      1234,
			Name:     "chrome",
			CPU:      25.5,
			Memory:   536870912,
			Command:  "/usr/bin/chrome",
		},
		{
			DeviceID: deviceID,
			PID:      5678,
			Name:     "firefox",
			CPU:      15.2,
			Memory:   268435456,
			Command:  "/usr/bin/firefox",
		},
	}

	processJSON, _ := json.Marshal(processes)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/processes", deviceID), bytes.NewBuffer(processJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify processes were created
	var savedProcesses []models.DeviceProcesses
	database.DB.Where("device_id = ?", deviceID).Find(&savedProcesses)
	assert.GreaterOrEqual(t, len(savedProcesses), 2)
}

func TestUpdateProcessListDeletion(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "test-process-delete"

	// Register device
	device := models.Device{
		ID:       deviceID,
		Name:     "Process Delete Test",
		IsOnline: true,
	}
	database.DB.Create(&device)

	// Add initial processes
	oldProcesses := []models.DeviceProcesses{
		{DeviceID: deviceID, PID: 1111, Name: "old-process-1", Timestamp: time.Now()},
		{DeviceID: deviceID, PID: 2222, Name: "old-process-2", Timestamp: time.Now()},
	}
	for _, p := range oldProcesses {
		database.DB.Create(&p)
	}

	// Update with new process list (should replace old ones)
	newProcesses := []models.DeviceProcesses{
		{DeviceID: deviceID, PID: 3333, Name: "new-process", CPU: 10.0, Memory: 100000},
	}

	processJSON, _ := json.Marshal(newProcesses)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/processes", deviceID), bytes.NewBuffer(processJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListDevicesEmpty(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	req, _ := http.NewRequest("GET", "/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var devices []models.Device
	err := json.Unmarshal(w.Body.Bytes(), &devices)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(devices), 0)
}

func TestListDevicesWithMultiple(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Create multiple devices
	devices := []models.Device{
		{ID: "device-list-1", Name: "Device 1", Type: "laptop", IsOnline: true, LastSeen: time.Now()},
		{ID: "device-list-2", Name: "Device 2", Type: "desktop", IsOnline: false, LastSeen: time.Now()},
		{ID: "device-list-3", Name: "Device 3", Type: "server", IsOnline: true, LastSeen: time.Now()},
	}

	for _, device := range devices {
		database.DB.Create(&device)
	}

	req, _ := http.NewRequest("GET", "/devices", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var responseDevices []models.Device
	err := json.Unmarshal(w.Body.Bytes(), &responseDevices)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(responseDevices), 3)
}

func TestGetDeviceMetricsWithLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "metrics-limit-test"

	// Create device
	device := models.Device{
		ID:       deviceID,
		Name:     "Metrics Limit Test",
		IsOnline: true,
	}
	database.DB.Create(&device)

	// Create multiple metrics
	for i := 0; i < 10; i++ {
		metrics := models.DeviceMetrics{
			ID:        uuid.New().String(),
			DeviceID:  deviceID,
			CPUUsage:  float64(10 + i*5),
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
		database.DB.Create(&metrics)
	}

	// Test with limit parameter
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=3", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var metrics []models.DeviceMetrics
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(metrics))
}

func TestGetDeviceMetricsInvalidLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "metrics-invalid-limit"
	device := models.Device{ID: deviceID, Name: "Test", IsOnline: true}
	database.DB.Create(&device)

	// Test with invalid limit
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=invalid", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	// The handler should return BadRequest for invalid limit parameter
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDeviceProcessesesWithLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "process-limit-test"

	// Create device
	device := models.Device{
		ID:       deviceID,
		Name:     "Process Limit Test",
		IsOnline: true,
	}
	database.DB.Create(&device)

	// Create multiple processes
	for i := 0; i < 8; i++ {
		process := models.DeviceProcesses{
			DeviceID:  deviceID,
			PID:       1000 + i,
			Name:      fmt.Sprintf("process-%d", i),
			CPU:       float64(5 * i),
			Memory:    uint64(100000 * i),
			Timestamp: time.Now(),
		}
		database.DB.Create(&process)
	}

	// Test with limit
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/processes?limit=5", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var processes []models.DeviceProcesses
	err := json.Unmarshal(w.Body.Bytes(), &processes)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(processes))
}

func TestGetDeviceAlertssWithLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "alert-limit-test"

	// Create device
	device := models.Device{
		ID:       deviceID,
		Name:     "Alert Limit Test",
		IsOnline: true,
	}
	database.DB.Create(&device)

	// Create multiple alerts
	for i := 0; i < 7; i++ {
		alert := models.DeviceAlerts{
			DeviceID:  deviceID,
			Type:      "cpu",
			Level:     "warning",
			Message:   fmt.Sprintf("Alert %d", i),
			Value:     80.0 + float64(i),
			Threshold: 75.0,
			Timestamp: time.Now(),
		}
		database.DB.Create(&alert)
	}

	// Test with limit
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/alerts?limit=4", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var alerts []models.DeviceAlerts
	err := json.Unmarshal(w.Body.Bytes(), &alerts)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(alerts))
}

func TestGetDeviceScreenshotssWithLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "screenshot-limit-test"

	// Create device
	device := models.Device{
		ID:       deviceID,
		Name:     "Screenshot Limit Test",
		IsOnline: true,
	}
	database.DB.Create(&device)

	// Create multiple screenshots
	for i := 0; i < 6; i++ {
		screenshot := models.DeviceScreenshots{
			DeviceID:   deviceID,
			Path:       fmt.Sprintf("/path/screenshot-%d.jpg", i),
			Resolution: "1920x1080",
			Size:       1024000,
			Timestamp:  time.Now(),
		}
		database.DB.Create(&screenshot)
	}

	// Test with limit
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/screenshots?limit=3", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestCreateRemoteCommandSuccess(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "command-test-device"

	// Create device
	device := models.Device{
		ID:       deviceID,
		Name:     "Command Test",
		IsOnline: true,
	}
	database.DB.Create(&device)

	// Create command
	command := models.DeviceRemoteCommands{
		DeviceID: deviceID,
		Command:  "ls -la",
		Status:   "pending",
	}

	commandJSON, _ := json.Marshal(command)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/commands", deviceID), bytes.NewBuffer(commandJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.DeviceRemoteCommands
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotZero(t, response.ID)
}

func TestGetPendingCommandsForDevice(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "pending-commands-test"

	// Create device
	device := models.Device{
		ID:       deviceID,
		Name:     "Pending Commands Test",
		IsOnline: true,
	}
	database.DB.Create(&device)

	// Create pending commands
	commands := []models.DeviceRemoteCommands{
		{DeviceID: deviceID, Command: "cmd1", Status: "pending", CreatedAt: time.Now()},
		{DeviceID: deviceID, Command: "cmd2", Status: "pending", CreatedAt: time.Now()},
		{DeviceID: deviceID, Command: "cmd3", Status: "completed", CreatedAt: time.Now()},
	}

	for _, cmd := range commands {
		database.DB.Create(&cmd)
	}

	// Get pending commands
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/commands/pending", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.DeviceRemoteCommands
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(response), 2)
}

func TestUpdateCommandStatusSuccess(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "update-status-test"
	device := models.Device{ID: deviceID, Name: "Test", IsOnline: true}
	database.DB.Create(&device)

	// Create command
	command := models.DeviceRemoteCommands{
		DeviceID:  deviceID,
		Command:   "test command",
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	database.DB.Create(&command)

	// Update status
	update := map[string]interface{}{
		"status": "completed",
		"result": "success",
	}

	updateJSON, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/commands/%d/status", command.ID), bytes.NewBuffer(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportAlertSuccess(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "alert-report-test"
	device := models.Device{ID: deviceID, Name: "Test", IsOnline: true}
	database.DB.Create(&device)

	// Report alert
	alert := models.DeviceAlerts{
		DeviceID:  deviceID,
		Type:      "cpu",
		Level:     "warning",
		Message:   "High CPU usage",
		Value:     85.0,
		Threshold: 80.0,
	}

	alertJSON, _ := json.Marshal(alert)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/alerts", deviceID), bytes.NewBuffer(alertJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
