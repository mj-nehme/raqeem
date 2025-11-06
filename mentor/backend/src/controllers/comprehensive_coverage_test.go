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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestListDevicesFullScenarios tests ListDevices with various scenarios
func TestListDevicesFullScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := database.SetupTestDB(t)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Create devices with different online statuses and timestamps
	now := time.Now()
	devices := []models.Device{
		{
			ID:       "device-online-1",
			Name:     "Online Device 1",
			IsOnline: true,
			LastSeen: now,
		},
		{
			ID:       "device-online-2",
			Name:     "Online Device 2",
			IsOnline: true,
			LastSeen: now.Add(-2 * time.Minute), // Recently seen
		},
		{
			ID:       "device-offline-1",
			Name:     "Offline Device 1",
			IsOnline: true, // Will be marked offline
			LastSeen: now.Add(-10 * time.Minute), // Seen more than 5 minutes ago
		},
		{
			ID:       "device-offline-2",
			Name:     "Offline Device 2",
			IsOnline: false,
			LastSeen: now.Add(-30 * time.Minute),
		},
	}

	for _, device := range devices {
		db.Create(&device)
	}

	t.Run("List all devices", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/devices", nil)

		ListDevices(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []models.Device
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 4)

		// Verify that devices with old last_seen are marked offline
		for _, d := range result {
			if d.ID == "device-offline-1" || d.ID == "device-offline-2" {
				// These should be offline now
				// Note: The update happens but we need to check the database
			}
		}
	})

	t.Run("Verify offline marking", func(t *testing.T) {
		// Give the update some time to process
		time.Sleep(10 * time.Millisecond)

		var offlineDevice models.Device
		db.Where("id = ?", "device-offline-1").First(&offlineDevice)
		// Should be marked offline due to old last_seen
		assert.False(t, offlineDevice.IsOnline)
	})
}

// TestUpdateProcessListFullScenarios tests UpdateProcessList with various scenarios
func TestUpdateProcessListFullScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := database.SetupTestDB(t)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceID := "test-device-process-update"

	t.Run("Update with new processes", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}

		processes := []models.Process{
			{DeviceID: deviceID, PID: 100, Name: "process-1", CPU: 10.5},
			{DeviceID: deviceID, PID: 200, Name: "process-2", CPU: 20.5},
			{DeviceID: deviceID, PID: 300, Name: "process-3", CPU: 30.5},
		}
		b, _ := json.Marshal(processes)
		c.Request, _ = http.NewRequest("POST", "/devices/"+deviceID+"/processes", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify processes were created
		var count int64
		db.Model(&models.Process{}).Where("device_id = ?", deviceID).Count(&count)
		assert.Equal(t, int64(3), count)
	})

	t.Run("Update replaces old processes", func(t *testing.T) {
		// First update
		w1 := httptest.NewRecorder()
		c1, _ := gin.CreateTestContext(w1)
		c1.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}

		processes1 := []models.Process{
			{DeviceID: deviceID, PID: 400, Name: "new-process-1"},
		}
		b1, _ := json.Marshal(processes1)
		c1.Request, _ = http.NewRequest("POST", "/devices/"+deviceID+"/processes", bytes.NewReader(b1))
		c1.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c1)
		assert.Equal(t, http.StatusOK, w1.Code)

		// Verify old processes are gone
		var count int64
		db.Model(&models.Process{}).Where("device_id = ?", deviceID).Count(&count)
		assert.Equal(t, int64(1), count)

		// Verify new process exists
		var newProcess models.Process
		db.Where("device_id = ? AND pid = ?", deviceID, 400).First(&newProcess)
		assert.Equal(t, "new-process-1", newProcess.Name)
	})

	t.Run("Update with empty list", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}

		processes := []models.Process{}
		b, _ := json.Marshal(processes)
		c.Request, _ = http.NewRequest("POST", "/devices/"+deviceID+"/processes", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestGetPendingCommandsAdvanced tests GetPendingCommands thoroughly
func TestGetPendingCommandsAdvanced(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := database.SetupTestDB(t)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceID := "test-device-pending-cmds"

	// Create commands with various statuses
	commands := []models.RemoteCommand{
		{DeviceID: deviceID, Command: "cmd1", Status: "pending", CreatedAt: time.Now()},
		{DeviceID: deviceID, Command: "cmd2", Status: "completed", CreatedAt: time.Now()},
		{DeviceID: deviceID, Command: "cmd3", Status: "pending", CreatedAt: time.Now()},
		{DeviceID: deviceID, Command: "cmd4", Status: "failed", CreatedAt: time.Now()},
		{DeviceID: deviceID, Command: "cmd5", Status: "pending", CreatedAt: time.Now()},
		{DeviceID: "other-device", Command: "cmd6", Status: "pending", CreatedAt: time.Now()},
	}

	for _, cmd := range commands {
		db.Create(&cmd)
	}

	t.Run("Get pending commands for device", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands/pending", nil)

		GetPendingCommands(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []models.RemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(result)) // Only 3 pending commands for this device

		// Verify all are pending
		for _, cmd := range result {
			assert.Equal(t, "pending", cmd.Status)
			assert.Equal(t, deviceID, cmd.DeviceID)
		}
	})

	t.Run("Get pending commands for device with no pending", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "device-no-commands"}}
		c.Request, _ = http.NewRequest("GET", "/devices/device-no-commands/commands/pending", nil)

		GetPendingCommands(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []models.RemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(result))
	})
}

// TestStoreScreenshotFullScenarios tests StoreScreenshot with various scenarios
func TestStoreScreenshotFullScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := database.SetupTestDB(t)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	t.Run("Store screenshot with full data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		screenshot := models.Screenshot{
			DeviceID:  "test-device-screenshot",
			Path:      "s3://bucket/screenshots/test.png",
			Timestamp: time.Now(),
		}
		b, _ := json.Marshal(screenshot)
		c.Request, _ = http.NewRequest("POST", "/screenshots", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		StoreScreenshot(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.Screenshot
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "test-device-screenshot", result.DeviceID)
		assert.NotZero(t, result.ID)
	})

	t.Run("Store multiple screenshots", func(t *testing.T) {
		for i := 0; i < 3; i++ {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			screenshot := models.Screenshot{
				DeviceID: "test-device-multi",
				Path:     "s3://bucket/screenshots/test-" + string(rune(i)) + ".png",
			}
			b, _ := json.Marshal(screenshot)
			c.Request, _ = http.NewRequest("POST", "/screenshots", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")

			StoreScreenshot(c)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Verify all screenshots were stored
		var count int64
		db.Model(&models.Screenshot{}).Where("device_id = ?", "test-device-multi").Count(&count)
		assert.Equal(t, int64(3), count)
	})
}

// TestLogActivityFullScenarios tests LogActivity with various scenarios
func TestLogActivityFullScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := database.SetupTestDB(t)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	t.Run("Log activity with all fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		activity := models.ActivityLog{
			DeviceID:    "test-device-activity",
			Type:        "app_launch",
			Description: "Launched application",
			App:         "Chrome",
			Duration:    120,
			Timestamp:   time.Now(),
		}
		b, _ := json.Marshal(activity)
		c.Request, _ = http.NewRequest("POST", "/activity", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		LogActivity(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify activity was stored
		var storedActivity models.ActivityLog
		db.Where("device_id = ?", "test-device-activity").First(&storedActivity)
		assert.Equal(t, "app_launch", storedActivity.Type)
		assert.Equal(t, "Chrome", storedActivity.App)
	})

	t.Run("Log multiple activities", func(t *testing.T) {
		deviceID := "test-device-multi-activity"

		activityTypes := []string{"app_launch", "file_access", "browser"}
		for _, actType := range activityTypes {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			activity := models.ActivityLog{
				DeviceID:    deviceID,
				Type:        actType,
				Description: "Activity of type " + actType,
			}
			b, _ := json.Marshal(activity)
			c.Request, _ = http.NewRequest("POST", "/activity", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")

			LogActivity(c)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Verify all activities were stored
		var count int64
		db.Model(&models.ActivityLog{}).Where("device_id = ?", deviceID).Count(&count)
		assert.Equal(t, int64(3), count)
	})
}

// TestReportAlertFullScenarios tests ReportAlert with various scenarios
func TestReportAlertFullScenarios(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := database.SetupTestDB(t)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	t.Run("Report alert with all fields", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		alert := models.Alert{
			DeviceID:  "test-device-alert",
			Level:     "critical",
			Type:      "disk_full",
			Message:   "Disk usage at 98%",
			Value:     98.0,
			Threshold: 90.0,
		}
		b, _ := json.Marshal(alert)
		c.Request, _ = http.NewRequest("POST", "/alerts", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		ReportAlert(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify alert was stored
		var storedAlert models.Alert
		db.Where("device_id = ? AND type = ?", "test-device-alert", "disk_full").First(&storedAlert)
		assert.Equal(t, "critical", storedAlert.Level)
		assert.Equal(t, 98.0, storedAlert.Value)
	})

	t.Run("Report multiple alerts", func(t *testing.T) {
		deviceID := "test-device-multi-alert"

		alertTypes := []string{"cpu_high", "memory_high", "disk_full"}
		for i, alertType := range alertTypes {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			alert := models.Alert{
				DeviceID: deviceID,
				Level:    "warning",
				Type:     alertType,
				Message:  "Alert for " + alertType,
				Value:    float64(80 + i),
			}
			b, _ := json.Marshal(alert)
			c.Request, _ = http.NewRequest("POST", "/alerts", bytes.NewReader(b))
			c.Request.Header.Set("Content-Type", "application/json")

			ReportAlert(c)

			assert.Equal(t, http.StatusOK, w.Code)
		}

		// Verify all alerts were stored
		var count int64
		db.Model(&models.Alert{}).Where("device_id = ?", deviceID).Count(&count)
		assert.Equal(t, int64(3), count)
	})
}
