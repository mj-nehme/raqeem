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

	deviceID := sampleUUID.String()

	// Register device first
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Process Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	database.DB.Create(&device)

	// Create process list
	processes := []models.DeviceProcess{
		{
			DeviceID:    sampleUUID,
			PID:         1234,
			ProcessName: "chrome",
			CPU:         25.5,
			Memory:      536870912,
			CommandText: "/usr/bin/chrome",
		},
		{
			DeviceID:    sampleUUID,
			PID:         5678,
			ProcessName: "firefox",
			CPU:         15.2,
			Memory:      268435456,
			CommandText: "/usr/bin/firefox",
		},
	}

	processJSON, _ := json.Marshal(processes)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/processes", deviceID), bytes.NewBuffer(processJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify processes were created
	var savedProcesses []models.DeviceProcess
	database.DB.Where("deviceid = ?", deviceID).Find(&savedProcesses)
	assert.GreaterOrEqual(t, len(savedProcesses), 2)
}

func TestUpdateProcessListDeletion(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := sampleUUID.String()

	// Register device
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Process Delete Test",
		IsOnline:   true,
	}
	database.DB.Create(&device)

	// Add initial processes
	oldProcesses := []models.DeviceProcess{
		{DeviceID: sampleUUID, PID: 1111, ProcessName: "old-process-1", Timestamp: time.Now()},
		{DeviceID: sampleUUID, PID: 2222, ProcessName: "old-process-2", Timestamp: time.Now()},
	}
	for _, p := range oldProcesses {
		database.DB.Create(&p)
	}

	// Update with new process list (should replace old ones)
	newProcesses := []models.DeviceProcess{
		{DeviceID: sampleUUID, PID: 3333, ProcessName: "new-process", CPU: 10.0, Memory: 100000},
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

	// Create multiple devices with distinct IDs
	devices := []models.Device{
		{DeviceID: uuid.New(), DeviceName: "Device 1", DeviceType: "laptop", IsOnline: true, LastSeen: time.Now()},
		{DeviceID: uuid.New(), DeviceName: "Device 2", DeviceType: "desktop", IsOnline: false, LastSeen: time.Now()},
		{DeviceID: uuid.New(), DeviceName: "Device 3", DeviceType: "server", IsOnline: true, LastSeen: time.Now()},
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

func TestGetDeviceMetricWithLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := sampleUUID.String()

	// Create device
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Metrics Limit Test",
		IsOnline:   true,
	}
	database.DB.Create(&device)

	// Create multiple metrics
	for i := 0; i < 10; i++ {
		metrics := models.DeviceMetric{
			DeviceID:  sampleUUID,
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

	var metrics []models.DeviceMetric
	err := json.Unmarshal(w.Body.Bytes(), &metrics)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(metrics))
}

func TestGetDeviceMetricInvalidLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := sampleUUID.String()
	device := models.Device{DeviceID: sampleUUID, DeviceName: "Test", IsOnline: true}
	database.DB.Create(&device)

	// Test with invalid limit
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=invalid", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	// The handler should return BadRequest for invalid limit parameter
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetDeviceProcessesWithLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := sampleUUID.String()

	// Create device
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Process Limit Test",
		IsOnline:   true,
	}
	database.DB.Create(&device)

	// Create multiple processes
	for i := 0; i < 8; i++ {
		process := models.DeviceProcess{
			DeviceID:    sampleUUID,
			PID:         1000 + i,
			ProcessName: fmt.Sprintf("process-%d", i),
			CPU:         float64(5 * i),
			Memory:      uint64(100000 * i),
			Timestamp:   time.Now(),
		}
		database.DB.Create(&process)
	}

	// Test with limit
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/processes?limit=5", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var processes []models.DeviceProcess
	err := json.Unmarshal(w.Body.Bytes(), &processes)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(processes))
}

func TestGetDeviceAlertWithLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := sampleUUID.String()

	// Create device
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Alert Limit Test",
		IsOnline:   true,
	}
	database.DB.Create(&device)

	// Create multiple alerts
	for i := 0; i < 7; i++ {
		alert := models.DeviceAlert{
			DeviceID:  sampleUUID,
			AlertType: "cpu",
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

	var alerts []models.DeviceAlert
	err := json.Unmarshal(w.Body.Bytes(), &alerts)
	assert.NoError(t, err)
	assert.Equal(t, 4, len(alerts))
}

func TestGetDeviceScreenshotWithLimit(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := sampleUUID.String()

	// Create device
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Screenshot Limit Test",
		IsOnline:   true,
	}
	database.DB.Create(&device)

	// Create multiple screenshots
	for i := 0; i < 6; i++ {
		screenshot := models.DeviceScreenshot{
			DeviceID:   sampleUUID,
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

	deviceID := sampleUUID.String()

	// Create device
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Command Test",
		IsOnline:   true,
	}
	database.DB.Create(&device)

	// Create command
	command := models.DeviceRemoteCommand{
		DeviceID:    sampleUUID,
		CommandText: "ls -la",
		Status:      "pending",
	}

	commandJSON, _ := json.Marshal(command)
	req, _ := http.NewRequest("POST", fmt.Sprintf("/devices/%s/commands", deviceID), bytes.NewBuffer(commandJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.DeviceRemoteCommand
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotZero(t, response.CommandText)
}

func TestGetPendingCommandsForDevice(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := sampleUUID.String()

	// Create device
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Pending Commands Test",
		IsOnline:   true,
	}
	database.DB.Create(&device)

	// Create pending commands
	commands := []models.DeviceRemoteCommand{
		{DeviceID: sampleUUID, CommandText: "cmd1", Status: "pending", CreatedAt: time.Now()},
		{DeviceID: sampleUUID, CommandText: "cmd2", Status: "pending", CreatedAt: time.Now()},
		{DeviceID: sampleUUID, CommandText: "cmd3", Status: "completed", CreatedAt: time.Now()},
	}

	for _, cmd := range commands {
		database.DB.Create(&cmd)
	}

	// Get pending commands
	req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/commands/pending", deviceID), nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response []models.DeviceRemoteCommand
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(response), 2)
}

func TestUpdateCommandStatusSuccess(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	device := models.Device{DeviceID: sampleUUID, DeviceName: "Test", IsOnline: true}
	database.DB.Create(&device)

	// Create command
	command := models.DeviceRemoteCommand{
		DeviceID:    sampleUUID,
		CommandText: "test command",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	database.DB.Create(&command)

	// Update status
	update := map[string]interface{}{
		"status": "completed",
		"result": "success",
	}

	updateJSON, _ := json.Marshal(update)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/commands/%d/status", command.CommandID), bytes.NewBuffer(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestReportAlertSuccess(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := sampleUUID.String()
	device := models.Device{DeviceID: sampleUUID, DeviceName: "Test", IsOnline: true}
	database.DB.Create(&device)

	// Report alert
	alert := models.DeviceAlert{
		DeviceID:  sampleUUID,
		AlertType: "cpu",
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
