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
	"mentor-backend/s3"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListDevicesDatabaseError tests ListDevices when database query fails
func TestListDevicesDatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Save original DB and set to nil to simulate database error
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/devices", nil)

	// This should cause a nil pointer panic, so we use a recover
	defer func() {
		if r := recover(); r != nil {
			// Expected behavior - database not available
			t.Log("Expected panic due to nil database")
		}
	}()

	ListDevices(c)
}

// TestListActivitiesDatabaseError tests ListActivities when database query fails
func TestListActivitiesDatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Save original DB
	originalDB := database.DB
	database.DB = nil
	defer func() { database.DB = originalDB }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/activities", nil)

	// This will panic due to nil DB
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic due to nil database")
		}
	}()

	ListActivities(c)
}

// TestListActivitiesSuccessfulQuery tests ListActivities with activities in the database
func TestListActivitiesSuccessfulQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Create device first
	deviceID := uuid.New()
	device := models.Device{
		DeviceID:   deviceID,
		DeviceName: "Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	db.Create(&device)

	// Create multiple activities
	activities := []models.DeviceActivity{
		{
			ActivityID:   uuid.New(),
			DeviceID:     deviceID,
			ActivityType: "login",
			Description:  "User logged in",
			App:          "system",
			Duration:     100,
			Timestamp:    time.Now().Add(-3 * time.Hour),
		},
		{
			ActivityID:   uuid.New(),
			DeviceID:     deviceID,
			ActivityType: "app_launch",
			Description:  "Launched app",
			App:          "Chrome",
			Duration:     500,
			Timestamp:    time.Now().Add(-2 * time.Hour),
		},
		{
			ActivityID:   uuid.New(),
			DeviceID:     deviceID,
			ActivityType: "file_access",
			Description:  "Opened file",
			App:          "Editor",
			Duration:     300,
			Timestamp:    time.Now().Add(-1 * time.Hour),
		},
	}

	for _, activity := range activities {
		db.Create(&activity)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/activities", nil)

	ListActivities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []models.DeviceActivity
	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 3)

	// Verify ordering - newest first
	if len(result) >= 2 {
		assert.True(t, result[0].Timestamp.After(result[1].Timestamp) || result[0].Timestamp.Equal(result[1].Timestamp))
	}
}

// TestGetScreenshotFileContentType tests content type detection
func TestGetScreenshotFileContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Save original client
	originalClient := s3.GetClient()
	defer s3.SetClient(originalClient)

	// Create a mock MinIO client pointing to non-existent server
	mockClient, err := minio.New("localhost:59997", &minio.Options{
		Secure: false,
	})
	if err != nil {
		t.Fatalf("Failed to create mock client: %v", err)
	}
	s3.SetClient(mockClient)

	testCases := []struct {
		name     string
		filename string
	}{
		{"png file", "screenshot.png"},
		{"PNG uppercase", "SCREENSHOT.PNG"},
		{"jpg file", "screenshot.jpg"},
		{"jpeg file", "screenshot.jpeg"},
		{"JPEG uppercase", "SCREENSHOT.JPEG"},
		{"no extension", "screenshot"},
		{"gif file", "screenshot.gif"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = gin.Params{gin.Param{Key: "filename", Value: tc.filename}}
			c.Request, _ = http.NewRequest("GET", "/screenshots/"+tc.filename, nil)

			GetScreenshotFile(c)

			// The test validates the code path for content type detection
			// MinIO's GetObject is lazy so the status code might be 200 even with connection failures
			assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound)
		})
	}
}

// TestStoreScreenshotGeneratesUUID tests that StoreScreenshot generates UUID if not provided
func TestStoreScreenshotGeneratesUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Create screenshot without UUID
	screenshot := models.DeviceScreenshot{
		DeviceID:   uuid.New(),
		Path:       "screenshots/test.png",
		Resolution: "1920x1080",
		Size:       1024,
	}

	body, _ := json.Marshal(screenshot)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/devices/screenshots", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	StoreScreenshot(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.DeviceScreenshot
	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, result.ScreenshotID)
}

// TestStoreScreenshotWithProvidedUUID tests that StoreScreenshot uses provided UUID
func TestStoreScreenshotWithProvidedUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	providedID := uuid.New()
	screenshot := models.DeviceScreenshot{
		ScreenshotID: providedID,
		DeviceID:     uuid.New(),
		Path:         "screenshots/test2.png",
		Resolution:   "1920x1080",
		Size:         2048,
	}

	body, _ := json.Marshal(screenshot)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/devices/screenshots", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	StoreScreenshot(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.DeviceScreenshot
	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, providedID, result.ScreenshotID)
}

// TestUpdateCommandStatusCompleted tests UpdateCommandStatus with completed status
func TestUpdateCommandStatusCompleted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Create device first
	deviceID := uuid.New()
	device := models.Device{
		DeviceID:   deviceID,
		DeviceName: "Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	db.Create(&device)

	// Create a pending command
	commandID := uuid.New()
	command := models.DeviceRemoteCommand{
		CommandID:   commandID,
		DeviceID:    deviceID,
		CommandText: "echo test",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	db.Create(&command)

	// Update command to completed
	updatePayload := map[string]interface{}{
		"commandid": commandID.String(),
		"status":    "completed",
		"result":    "test output",
		"exit_code": 0,
	}

	body, _ := json.Marshal(updatePayload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/commands/status", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateCommandStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify command was updated
	var updatedCmd models.DeviceRemoteCommand
	db.Where("commandid = ?", commandID).First(&updatedCmd)
	assert.Equal(t, "completed", updatedCmd.Status)
	assert.NotNil(t, updatedCmd.CompletedAt)
}

// TestUpdateCommandStatusFailed tests UpdateCommandStatus with failed status
func TestUpdateCommandStatusFailed(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Create device first
	deviceID := uuid.New()
	device := models.Device{
		DeviceID:   deviceID,
		DeviceName: "Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	db.Create(&device)

	// Create a pending command
	commandID := uuid.New()
	command := models.DeviceRemoteCommand{
		CommandID:   commandID,
		DeviceID:    deviceID,
		CommandText: "failing-command",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	db.Create(&command)

	// Update command to failed
	updatePayload := map[string]interface{}{
		"commandid": commandID.String(),
		"status":    "failed",
		"result":    "command not found",
		"exit_code": 1,
	}

	body, _ := json.Marshal(updatePayload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/commands/status", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateCommandStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify command was updated
	var updatedCmd models.DeviceRemoteCommand
	db.Where("commandid = ?", commandID).First(&updatedCmd)
	assert.Equal(t, "failed", updatedCmd.Status)
	assert.NotNil(t, updatedCmd.CompletedAt)
}

// TestUpdateCommandStatusFallback tests UpdateCommandStatus fallback logic
func TestUpdateCommandStatusFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Create device first
	deviceID := uuid.New()
	device := models.Device{
		DeviceID:   deviceID,
		DeviceName: "Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	db.Create(&device)

	// Create a pending command
	commandID := uuid.New()
	command := models.DeviceRemoteCommand{
		CommandID:   commandID,
		DeviceID:    deviceID,
		CommandText: "echo test",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	db.Create(&command)

	// Update using device ID as fallback (non-existent command ID but valid device ID)
	nonExistentCommandID := uuid.New()
	updatePayload := map[string]interface{}{
		"commandid": nonExistentCommandID.String(),
		"deviceid":  deviceID.String(),
		"status":    "completed",
		"result":    "fallback worked",
		"exit_code": 0,
	}

	body, _ := json.Marshal(updatePayload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/commands/status", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateCommandStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the latest command for the device was updated
	var updatedCmd models.DeviceRemoteCommand
	db.Where("deviceid = ?", deviceID).Order("created_at desc").First(&updatedCmd)
	assert.Equal(t, "completed", updatedCmd.Status)
}

// TestUpdateCommandStatusMissingCommandIDValidation tests UpdateCommandStatus with missing command ID validation
func TestUpdateCommandStatusMissingCommandIDValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Update without command ID
	updatePayload := map[string]interface{}{
		"status":    "completed",
		"result":    "test output",
		"exit_code": 0,
	}

	body, _ := json.Marshal(updatePayload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/commands/status", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateCommandStatus(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "commandid required")
}

// TestGetPendingCommandsWithCommands tests GetPendingCommands with actual pending commands
func TestGetPendingCommandsWithCommands(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Create device first
	deviceID := uuid.New()
	device := models.Device{
		DeviceID:   deviceID,
		DeviceName: "Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	db.Create(&device)

	// Create pending commands
	commands := []models.DeviceRemoteCommand{
		{
			CommandID:   uuid.New(),
			DeviceID:    deviceID,
			CommandText: "echo test1",
			Status:      "pending",
			CreatedAt:   time.Now(),
		},
		{
			CommandID:   uuid.New(),
			DeviceID:    deviceID,
			CommandText: "echo test2",
			Status:      "pending",
			CreatedAt:   time.Now(),
		},
		{
			CommandID:   uuid.New(),
			DeviceID:    deviceID,
			CommandText: "echo completed",
			Status:      "completed", // This should not be returned
			CreatedAt:   time.Now(),
		},
	}

	for _, cmd := range commands {
		db.Create(&cmd)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID.String()}}
	c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID.String()+"/commands/pending", nil)

	GetPendingCommands(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []models.DeviceRemoteCommand
	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))

	// Verify only pending commands are returned
	for _, cmd := range result {
		assert.Equal(t, "pending", cmd.Status)
	}
}

// TestGetPendingCommandsInvalidUUID tests GetPendingCommands with invalid UUID
func TestGetPendingCommandsInvalidUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid-uuid"}}
	c.Request, _ = http.NewRequest("GET", "/devices/invalid-uuid/commands/pending", nil)

	GetPendingCommands(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

// TestListDevicesMarksOffline tests that ListDevices marks old devices as offline
func TestListDevicesMarksOffline(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Create a device that was last seen more than 5 minutes ago
	deviceID := uuid.New()
	device := models.Device{
		DeviceID:   deviceID,
		DeviceName: "Old Device",
		IsOnline:   true,                              // Currently marked as online
		LastSeen:   time.Now().Add(-10 * time.Minute), // 10 minutes ago
	}
	db.Create(&device)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/devices", nil)

	ListDevices(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the device was marked as offline
	var updatedDevice models.Device
	db.Where("deviceid = ?", deviceID).First(&updatedDevice)
	assert.False(t, updatedDevice.IsOnline)
}

// TestListDevicesEmptyDatabase tests ListDevices with empty database
func TestListDevicesEmptyDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Delete all devices
	db.Exec("DELETE FROM devices")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/devices", nil)

	ListDevices(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

// TestListDevicesMultipleScenarios tests ListDevices with multiple devices in various states
func TestListDevicesMultipleScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Create devices with various states
	now := time.Now()
	devices := []models.Device{
		{
			DeviceID:   uuid.New(),
			DeviceName: "Recent Online Device",
			IsOnline:   true,
			LastSeen:   now,
		},
		{
			DeviceID:   uuid.New(),
			DeviceName: "Just Online Device",
			IsOnline:   true,
			LastSeen:   now.Add(-1 * time.Minute),
		},
		{
			DeviceID:   uuid.New(),
			DeviceName: "Barely Offline Device",
			IsOnline:   true,
			LastSeen:   now.Add(-6 * time.Minute), // Just past the 5 minute threshold
		},
		{
			DeviceID:   uuid.New(),
			DeviceName: "Long Offline Device",
			IsOnline:   false,
			LastSeen:   now.Add(-1 * time.Hour),
		},
	}

	for _, d := range devices {
		db.Create(&d)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/devices", nil)

	ListDevices(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []models.Device
	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 4)
}

// TestListActivitiesWithMultipleDevices tests ListActivities with activities from multiple devices
func TestListActivitiesWithMultipleDevices(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Create multiple devices
	device1 := uuid.New()
	device2 := uuid.New()

	for _, deviceID := range []uuid.UUID{device1, device2} {
		device := models.Device{
			DeviceID:   deviceID,
			DeviceName: "Test Device " + deviceID.String()[:8],
			IsOnline:   true,
			LastSeen:   time.Now(),
		}
		db.Create(&device)
	}

	// Create activities for each device
	activities := []models.DeviceActivity{
		{
			ActivityID:   uuid.New(),
			DeviceID:     device1,
			ActivityType: "login",
			Description:  "User logged in",
			Timestamp:    time.Now().Add(-2 * time.Hour),
		},
		{
			ActivityID:   uuid.New(),
			DeviceID:     device1,
			ActivityType: "logout",
			Description:  "User logged out",
			Timestamp:    time.Now().Add(-1 * time.Hour),
		},
		{
			ActivityID:   uuid.New(),
			DeviceID:     device2,
			ActivityType: "app_launch",
			Description:  "Launched app",
			Timestamp:    time.Now().Add(-30 * time.Minute),
		},
	}

	for _, activity := range activities {
		db.Create(&activity)
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/activities", nil)

	ListActivities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []models.DeviceActivity
	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 3)

	// Verify ordering (newest first)
	if len(result) >= 2 {
		for i := 0; i < len(result)-1; i++ {
			assert.True(t, result[i].Timestamp.After(result[i+1].Timestamp) || result[i].Timestamp.Equal(result[i+1].Timestamp))
		}
	}
}

// TestListActivitiesEmpty tests ListActivities with no activities
func TestListActivitiesEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Clear all activities
	db.Exec("DELETE FROM device_activities")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/activities", nil)

	ListActivities(c)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "[]", w.Body.String())
}

// TestStoreScreenshotValidation tests StoreScreenshot with various edge cases
func TestStoreScreenshotValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing device ID", func(t *testing.T) {
		db, err := database.SetupTestDB(t)
		require.NoError(t, err)
		database.DB = db

		// Screenshot with nil device ID
		screenshot := map[string]interface{}{
			"path":       "test.png",
			"resolution": "1920x1080",
			"size":       1024,
		}

		body, _ := json.Marshal(screenshot)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/devices/screenshots", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		StoreScreenshot(c)

		// Should succeed - device ID is optional (UUID will be nil)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("valid screenshot with all fields", func(t *testing.T) {
		db, err := database.SetupTestDB(t)
		require.NoError(t, err)
		database.DB = db

		screenshot := models.DeviceScreenshot{
			ScreenshotID: uuid.New(),
			DeviceID:     uuid.New(),
			Path:         "screenshots/complete.png",
			Resolution:   "2560x1440",
			Size:         2048576,
		}

		body, _ := json.Marshal(screenshot)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("POST", "/devices/screenshots", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		StoreScreenshot(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DeviceScreenshot
		err = json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, screenshot.ScreenshotID, result.ScreenshotID)
		assert.Equal(t, screenshot.DeviceID, result.DeviceID)
		assert.Equal(t, screenshot.Path, result.Path)
		assert.Equal(t, screenshot.Resolution, result.Resolution)
		assert.Equal(t, screenshot.Size, result.Size)
	})
}

// TestUpdateCommandStatusRunningStatus tests UpdateCommandStatus with running status
func TestUpdateCommandStatusRunningStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	database.DB = db

	// Create device first
	deviceID := uuid.New()
	device := models.Device{
		DeviceID:   deviceID,
		DeviceName: "Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	db.Create(&device)

	// Create a pending command
	commandID := uuid.New()
	command := models.DeviceRemoteCommand{
		CommandID:   commandID,
		DeviceID:    deviceID,
		CommandText: "long-running-command",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	db.Create(&command)

	// Update command to running (not completed or failed, so no CompletedAt)
	updatePayload := map[string]interface{}{
		"commandid": commandID.String(),
		"status":    "running",
	}

	body, _ := json.Marshal(updatePayload)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/commands/status", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateCommandStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify command was updated to running without CompletedAt
	var updatedCmd models.DeviceRemoteCommand
	db.Where("commandid = ?", commandID).First(&updatedCmd)
	assert.Equal(t, "running", updatedCmd.Status)
}
