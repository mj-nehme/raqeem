package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestUpdateProcessListInvalidDeviceID tests UpdateProcessList with an invalid UUID
func TestUpdateProcessListInvalidDeviceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Valid JSON but invalid device ID in URL parameter
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "not-a-valid-uuid"}}

	processes := []models.DeviceProcess{
		{PID: 1234, ProcessName: "test-process", CPU: 10.5, Memory: 1024},
	}
	b, _ := json.Marshal(processes)
	c.Request, _ = http.NewRequest("POST", "/devices/not-a-valid-uuid/processes", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateProcessList(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid device ID")
}

// TestUpdateProcessListNoIDEmptyList tests UpdateProcessList with no ID and empty list
func TestUpdateProcessListNoIDEmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Empty process list without device ID - should return success
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// No "id" param set

	processes := []models.DeviceProcess{}
	b, _ := json.Marshal(processes)
	c.Request, _ = http.NewRequest("POST", "/devices/processes", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateProcessList(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

// TestUpdateProcessListDeviceIDFromProcess tests UpdateProcessList getting device ID from first process
func TestUpdateProcessListDeviceIDFromProcess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	testUUID := uuid.New()

	// Process list with device ID in the process, no URL param
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	// No "id" param set - will use deviceID from first process

	processes := []models.DeviceProcess{
		{DeviceID: testUUID, PID: 1234, ProcessName: "test-process", CPU: 10.5, Memory: 1024},
		{DeviceID: testUUID, PID: 5678, ProcessName: "another-process", CPU: 5.0, Memory: 512},
	}
	b, _ := json.Marshal(processes)
	c.Request, _ = http.NewRequest("POST", "/devices/processes", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateProcessList(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestUpdateCommandStatusMissingCommandID tests UpdateCommandStatus without commandid
func TestUpdateCommandStatusMissingCommandID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Status update without commandid should fail
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Empty commandid (nil UUID)
	updateCmd := map[string]interface{}{
		"status": "completed",
		"result": "success",
	}
	b, _ := json.Marshal(updateCmd)
	c.Request, _ = http.NewRequest("POST", "/commands/status", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateCommandStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "commandid required")
}

// TestUpdateCommandStatusFallbackToDeviceID tests the fallback logic in UpdateCommandStatus
func TestUpdateCommandStatusFallbackToDeviceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceUUID := uuid.New()

	// Create a command for this device
	cmd := models.DeviceRemoteCommand{
		DeviceID:    deviceUUID,
		CommandText: "get_info",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	db.Create(&cmd)

	// Update using commandid that doesn't match (but deviceid does) - triggers fallback
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Use a non-existent commandid but include deviceid for fallback
	updateCmd := models.DeviceRemoteCommand{
		CommandID: uuid.New(), // Doesn't exist
		DeviceID:  deviceUUID,
		Status:    "completed",
		Result:    "success via fallback",
	}
	b, _ := json.Marshal(updateCmd)
	c.Request, _ = http.NewRequest("POST", "/commands/status", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateCommandStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the command was updated via fallback
	var updatedCmd models.DeviceRemoteCommand
	db.Where("deviceid = ?", deviceUUID).First(&updatedCmd)
	assert.Equal(t, "completed", updatedCmd.Status)
}

// TestCreateRemoteCommandEmptyCommand tests CreateRemoteCommand with empty command text
func TestCreateRemoteCommandEmptyCommand(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Command with empty command_text
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	cmd := models.DeviceRemoteCommand{
		DeviceID:    sampleUUID,
		CommandText: "", // Empty - should be rejected
	}
	b, _ := json.Marshal(cmd)
	c.Request, _ = http.NewRequest("POST", "/devices/commands", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateRemoteCommand(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "deviceid and command_text are required")
}

// TestCreateRemoteCommandMissingDeviceID tests CreateRemoteCommand without deviceid
func TestCreateRemoteCommandMissingDeviceID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Command without deviceid
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	cmd := map[string]interface{}{
		"command_text": "get_info",
	}
	b, _ := json.Marshal(cmd)
	c.Request, _ = http.NewRequest("POST", "/devices/commands", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateRemoteCommand(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "deviceid and command_text are required")
}

// TestGetDeviceMetricZeroLimit tests GetDeviceMetric with zero limit
func TestGetDeviceMetricZeroLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: sampleUUID.String()}}
	c.Request, _ = http.NewRequest("GET", "/devices/"+sampleUUID.String()+"/metrics?limit=0", nil)

	GetDeviceMetric(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "limit must be greater than 0")
}

// TestGetDeviceProcessesZeroLimit tests GetDeviceProcesses with zero limit
func TestGetDeviceProcessesZeroLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: sampleUUID.String()}}
	c.Request, _ = http.NewRequest("GET", "/devices/"+sampleUUID.String()+"/processes?limit=0", nil)

	GetDeviceProcesses(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "limit must be greater than 0")
}

// TestGetDeviceMetricNegativeLimit tests GetDeviceMetric with negative limit
func TestGetDeviceMetricNegativeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: sampleUUID.String()}}
	c.Request, _ = http.NewRequest("GET", "/devices/"+sampleUUID.String()+"/metrics?limit=-5", nil)

	GetDeviceMetric(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "limit must be greater than 0")
}

// TestGetDeviceProcessesNegativeLimit tests GetDeviceProcesses with negative limit
func TestGetDeviceProcessesNegativeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: sampleUUID.String()}}
	c.Request, _ = http.NewRequest("GET", "/devices/"+sampleUUID.String()+"/processes?limit=-5", nil)

	GetDeviceProcesses(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "limit must be greater than 0")
}

// TestRegisterDeviceWithNewUUID tests RegisterDevice generating a new UUID
func TestRegisterDeviceWithNewUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Device without UUID - should have one generated
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	device := map[string]interface{}{
		"device_name": "Test Device Without UUID",
		"device_type": "laptop",
	}
	b, _ := json.Marshal(device)
	c.Request, _ = http.NewRequest("POST", "/devices", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	RegisterDevice(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.Device
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.DeviceID)
	assert.True(t, result.IsOnline)
}

// TestUpdateDeviceMetricWithNewUUID tests UpdateDeviceMetric generating a new UUID
func TestUpdateDeviceMetricWithNewUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Metric without MetricID - should have one generated
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	metric := models.DeviceMetric{
		DeviceID:   sampleUUID,
		CPUUsage:   75.5,
		MemoryUsed: 8192,
	}
	b, _ := json.Marshal(metric)
	c.Request, _ = http.NewRequest("POST", "/metrics", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateDeviceMetric(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.DeviceMetric
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.MetricID)
	assert.NotZero(t, result.Timestamp)
}

// TestActivityWithNewUUID tests Activity generating a new UUID
func TestActivityWithNewUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Activity without ActivityID - should have one generated
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	activity := models.DeviceActivity{
		DeviceID:     sampleUUID,
		ActivityType: "app_launch",
		Description:  "Launched test app",
	}
	b, _ := json.Marshal(activity)
	c.Request, _ = http.NewRequest("POST", "/activity", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	Activity(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.DeviceActivity
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.ActivityID)
	assert.NotZero(t, result.Timestamp)
}

// TestReportAlertWithNewUUID tests ReportAlert generating a new UUID
func TestReportAlertWithNewUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Alert without AlertID - should have one generated
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	alert := models.DeviceAlert{
		DeviceID:  sampleUUID,
		AlertType: "cpu_high",
		Level:     "warning",
		Message:   "High CPU usage",
		Value:     90.0,
		Threshold: 80.0,
	}
	b, _ := json.Marshal(alert)
	c.Request, _ = http.NewRequest("POST", "/alerts", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	ReportAlert(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.DeviceAlert
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.AlertID)
	assert.NotZero(t, result.Timestamp)
}

// TestStoreScreenshotWithNewUUID tests StoreScreenshot generating a new UUID
func TestStoreScreenshotWithNewUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Screenshot without ScreenshotID - should have one generated
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	screenshot := models.DeviceScreenshot{
		DeviceID:   sampleUUID,
		Path:       "s3://bucket/screenshots/test.png",
		Resolution: "1920x1080",
		Size:       1024000,
	}
	b, _ := json.Marshal(screenshot)
	c.Request, _ = http.NewRequest("POST", "/screenshots", bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	StoreScreenshot(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.DeviceScreenshot
	err = json.Unmarshal(w.Body.Bytes(), &result)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.ScreenshotID)
}
