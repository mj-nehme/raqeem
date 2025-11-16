package controllers_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mentor-backend/controllers"
	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestListDevicesEdgeCases tests edge cases in ListDevices
func TestListDevicesEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("lists devices when none exist", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/devices", nil)

		controllers.ListDevices(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var devices []models.Device
		err := json.Unmarshal(w.Body.Bytes(), &devices)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(devices))
	})

	t.Run("marks offline devices correctly", func(t *testing.T) {
		// Create a device that was last seen more than 5 minutes ago
		deviceID := uuid.New()
		oldDevice := models.Device{
			DeviceID:   deviceID,
			DeviceName: "Old Device",
			IsOnline:   true,
			LastSeen:   time.Now().Add(-10 * time.Minute),
		}
		database.DB.Create(&oldDevice)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/devices", nil)

		controllers.ListDevices(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Verify the device was marked offline
		var updated models.Device
		database.DB.Where("deviceid = ?", deviceID).First(&updated)
		assert.False(t, updated.IsOnline)
	})
}

// TestGetDeviceMetricEdgeCases tests edge cases in GetDeviceMetric
func TestGetDeviceMetricEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns 400 for invalid limit parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/devices/test-id/metrics?limit=invalid", nil)
		c.Params = gin.Params{{Key: "id", Value: "test-id"}}

		controllers.GetDeviceMetric(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid limit parameter")
	})

	t.Run("returns 400 for negative limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/devices/test-id/metrics?limit=-10", nil)
		c.Params = gin.Params{{Key: "id", Value: "test-id"}}

		controllers.GetDeviceMetric(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("handles very large limit values", func(t *testing.T) {
		// Create test device
		deviceID := uuid.New()
		device := models.Device{
			DeviceID:   deviceID,
			DeviceName: "Test Device",
			IsOnline:   true,
			LastSeen:   time.Now(),
		}
		database.DB.Create(&device)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=999999", deviceID), nil)
		c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}

		controllers.GetDeviceMetric(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestGetDeviceProcessesEdgeCases tests edge cases in GetDeviceProcesses
func TestGetDeviceProcessesEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns 400 for invalid limit parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/devices/test-id/processes?limit=abc", nil)
		c.Params = gin.Params{{Key: "id", Value: "test-id"}}

		controllers.GetDeviceProcesses(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid limit parameter")
	})
}

// TestGetDeviceActivityEdgeCases tests edge cases in GetDeviceActivity
func TestGetDeviceActivityEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns 400 for invalid limit parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/devices/test-id/activity?limit=xyz", nil)
		c.Params = gin.Params{{Key: "id", Value: "test-id"}}

		controllers.GetDeviceActivity(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid limit parameter")
	})
}

// TestGetDeviceAlertEdgeCases tests edge cases in GetDeviceAlert
func TestGetDeviceAlertEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns 400 for invalid limit parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/devices/test-id/alerts?limit=bad", nil)
		c.Params = gin.Params{{Key: "id", Value: "test-id"}}

		controllers.GetDeviceAlert(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid limit parameter")
	})
}

// TestGetDeviceScreenshotEdgeCases tests edge cases in GetDeviceScreenshot
func TestGetDeviceScreenshotEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns 400 for invalid limit parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/devices/test-id/screenshots?limit=invalid", nil)
		c.Params = gin.Params{{Key: "id", Value: "test-id"}}

		controllers.GetDeviceScreenshot(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid limit parameter")
	})

	t.Run("handles device with no screenshots", func(t *testing.T) {
		// Create test device without screenshots
		deviceID := uuid.New()
		device := models.Device{
			DeviceID:   deviceID,
			DeviceName: "Test Device",
			IsOnline:   true,
			LastSeen:   time.Now(),
		}
		database.DB.Create(&device)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/devices/%s/screenshots", deviceID), nil)
		c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}

		controllers.GetDeviceScreenshot(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var screenshots []models.DeviceScreenshot
		err := json.Unmarshal(w.Body.Bytes(), &screenshots)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(screenshots))
	})
}

// TestCreateRemoteCommandEdgeCases tests edge cases in CreateRemoteCommand
func TestCreateRemoteCommandEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns 400 for malformed JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/devices/test-id/commands", strings.NewReader("{invalid json}"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: "test-id"}}

		controllers.CreateRemoteCommand(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 400 for validation error", func(t *testing.T) {
		// Create device first
		deviceID := uuid.New()
		device := models.Device{
			DeviceID:   deviceID,
			DeviceName: "Test Device",
			IsOnline:   true,
			LastSeen:   time.Now(),
		}
		database.DB.Create(&device)

		// Invalid command - empty type
		payload := fmt.Sprintf(`{"deviceid":"%s","type":"","command":"test"}`, deviceID)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", fmt.Sprintf("/devices/%s/commands", deviceID), strings.NewReader(payload))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}

		controllers.CreateRemoteCommand(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestGetPendingCommandsEdgeCases tests edge cases in GetPendingCommands
func TestGetPendingCommandsEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns empty array for device with no pending commands", func(t *testing.T) {
		// Create device
		deviceID := uuid.New()
		device := models.Device{
			DeviceID:   deviceID,
			DeviceName: "Test Device",
			IsOnline:   true,
			LastSeen:   time.Now(),
		}
		database.DB.Create(&device)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", fmt.Sprintf("/devices/%s/commands/pending", deviceID), nil)
		c.Params = gin.Params{{Key: "id", Value: deviceID.String()}}

		controllers.GetPendingCommands(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var commands []models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &commands)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(commands))
	})
}

// TestGetDeviceCommandsEdgeCases tests edge cases in GetDeviceCommands
func TestGetDeviceCommandsEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns 400 for invalid limit parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/devices/test-id/commands?limit=notanumber", nil)
		c.Params = gin.Params{{Key: "id", Value: "test-id"}}

		controllers.GetDeviceCommands(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "invalid limit parameter")
	})
}

// TestUpdateCommandStatusEdgeCases tests edge cases in UpdateCommandStatus
func TestUpdateCommandStatusEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns 400 for malformed JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/commands/123/result", strings.NewReader("{bad json}"))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "command_id", Value: "123"}}

		controllers.UpdateCommandStatus(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("returns 404 for nonexistent command", func(t *testing.T) {
		payload := `{"status":"completed","output":"test output"}`

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/commands/99999/result", strings.NewReader(payload))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "command_id", Value: "99999"}}

		controllers.UpdateCommandStatus(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// TestReportAlertEdgeCases tests edge cases in ReportAlert
func TestReportAlertEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns 400 for malformed JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/alerts", strings.NewReader("{invalid}"))
		c.Request.Header.Set("Content-Type", "application/json")

		controllers.ReportAlert(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestStoreScreenshotEdgeCases tests edge cases in StoreScreenshot
func TestStoreScreenshotEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns 400 for malformed JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/screenshots", strings.NewReader("{bad}"))
		c.Request.Header.Set("Content-Type", "application/json")

		controllers.StoreScreenshot(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestListActivitiesEdgeCases tests edge cases in ListActivities
func TestListActivitiesEdgeCases(t *testing.T) {
	setupTestDB(t)
	defer teardownTestDB(t)

	gin.SetMode(gin.TestMode)

	t.Run("returns empty array when no activities exist", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activities", nil)

		controllers.ListActivities(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var activities []models.DeviceActivity
		err := json.Unmarshal(w.Body.Bytes(), &activities)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(activities))
	})

	t.Run("orders activities by timestamp descending", func(t *testing.T) {
		// Create test device
		deviceID := uuid.New()
		device := models.Device{
			DeviceID:   deviceID,
			DeviceName: "Test Device",
			IsOnline:   true,
			LastSeen:   time.Now(),
		}
		database.DB.Create(&device)

		// Create activities with different timestamps
		older := models.DeviceActivity{
			DeviceID:     deviceID,
			ActivityType: "test1",
			Description:  "older",
			Timestamp:    time.Now().Add(-1 * time.Hour),
		}
		newer := models.DeviceActivity{
			DeviceID:     deviceID,
			ActivityType: "test2",
			Description:  "newer",
			Timestamp:    time.Now(),
		}
		database.DB.Create(&older)
		database.DB.Create(&newer)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/activities", nil)

		controllers.ListActivities(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var activities []models.DeviceActivity
		err := json.Unmarshal(w.Body.Bytes(), &activities)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(activities), 2)
		// First activity should be the newer one
		assert.Equal(t, "test2", activities[0].ActivityType)
	})
}

// Helper functions
func setupTestDB(t *testing.T) {
	db, err := database.SetupTestDB(t)
	if err != nil {
		t.Skipf("Skipping test - database not available: %v", err)
		return
	}
	database.DB = db
	
	// Clean up tables (no-op since SetupTestDB uses transactions that auto-rollback)
	database.DB.Exec("DELETE FROM remote_commands")
	database.DB.Exec("DELETE FROM device_screenshots")
	database.DB.Exec("DELETE FROM device_alerts")
	database.DB.Exec("DELETE FROM device_activities")
	database.DB.Exec("DELETE FROM device_processes")
	database.DB.Exec("DELETE FROM device_metrics")
	database.DB.Exec("DELETE FROM devices")
}

func teardownTestDB(t *testing.T) {
	// Clean up after tests
	database.DB.Exec("DELETE FROM device_remote_commands")
	database.DB.Exec("DELETE FROM device_screenshots")
	database.DB.Exec("DELETE FROM device_alerts")
	database.DB.Exec("DELETE FROM device_activities")
	database.DB.Exec("DELETE FROM device_processes")
	database.DB.Exec("DELETE FROM device_metrics")
	database.DB.Exec("DELETE FROM devices")
}
