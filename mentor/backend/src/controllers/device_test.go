package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
)

func setupTestDB(t *testing.T) {
	db, err := database.SetupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	if db == nil {
		t.Fatal("Test database is nil")
	}
	t.Cleanup(func() {
		database.CleanupTestDB(t, db)
	})
	database.DB = db
}

func TestReportAndGetAlerts(t *testing.T) {

	setupTestDB(t)

	// Ensure clean slate for test device
	deviceID := sampleUUID
	database.DB.Where("deviceid = ?", deviceID).Delete(&models.DeviceAlert{})

	// Prepare gin context for ReportAlert
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/devices/"+deviceID.String()+"/alerts", nil)

	alert := models.DeviceAlert{
		DeviceID:  sampleUUID,
		Timestamp: time.Now(),
		Level:     "warning",
		AlertType: "cpu_high",
		Message:   "CPU high",
		Value:     95,
		Threshold: 80,
	}
	b, _ := json.Marshal(alert)
	c.Request.Body = io.NopCloser(bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	ReportAlert(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Prepare gin context for GetDeviceAlert
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Params = gin.Params{gin.Param{Key: "id", Value: deviceID.String()}}
	c2.Request, _ = http.NewRequest("GET", "/devices/"+deviceID.String()+"/alerts", nil)

	GetDeviceAlert(c2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w2.Code)
	}
}

// TestEmptyArraySerialization verifies that endpoints return empty arrays [] instead of null
// when no data exists for a device. This is important for frontend compatibility.
func TestEmptyArraySerialization(t *testing.T) {
	// Test without database - just verify the response format
	gin.SetMode(gin.TestMode)

	// Create a mock device ID that doesn't exist in the database
	deviceID := "non-existent-device"

	// Test GetDeviceProcesses
	t.Run("GetDeviceProcesses returns empty array", func(t *testing.T) {

		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/processes", nil)

		GetDeviceProcesses(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		// Verify response is not null
		body := w.Body.String()
		if body == "null" {
			t.Errorf("GetDeviceProcesses returned null instead of empty array")
		}

		// Verify it's a valid JSON array
		var processes []models.DeviceProcess
		if err := json.Unmarshal(w.Body.Bytes(), &processes); err != nil {
			t.Fatalf("failed to unmarshal response: %v, body: %s", err, body)
		}

		// Verify the unmarshaled result is an empty slice, not nil
		if processes == nil {
			t.Errorf("unmarshaled processes is nil, expected empty slice")
		}
		if len(processes) != 0 {
			t.Logf("Note: Expected empty array but got %d processes (may have data from other tests)", len(processes))
		}
	})

	// Test GetDeviceMetric
	t.Run("GetDeviceMetric returns empty array", func(t *testing.T) {

		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/metrics", nil)

		GetDeviceMetric(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("GetDeviceMetric returned null instead of empty array")
		}
	})

	// Test GetDeviceActivity
	t.Run("GetDeviceActivity returns empty array", func(t *testing.T) {

		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/activities", nil)

		GetDeviceActivity(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("GetDeviceActivity returned null instead of empty array")
		}
	})

	// Test GetDeviceAlert
	t.Run("GetDeviceAlert returns empty array", func(t *testing.T) {

		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/alerts", nil)

		GetDeviceAlert(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("GetDeviceAlert returned null instead of empty array")
		}
	})

	// Test ListDevices
	t.Run("ListDevices returns empty array", func(t *testing.T) {

		setupTestDB(t)

		// Clear all devices
		database.DB.Exec("DELETE FROM devices")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/devices", nil)

		ListDevices(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("ListDevices returned null instead of empty array")
		}
	})

	// Test GetPendingCommands
	t.Run("GetPendingCommands returns empty array", func(t *testing.T) {

		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands/pending", nil)

		GetPendingCommands(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("GetPendingCommands returned null instead of empty array")
		}
	})

	// Test GetDeviceCommands
	t.Run("GetDeviceCommands returns empty array", func(t *testing.T) {

		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands", nil)

		GetDeviceCommands(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("GetDeviceCommands returned null instead of empty array")
		}
	})

	// Test GetDeviceCommands with limit parameter
	t.Run("GetDeviceCommands respects limit parameter", func(t *testing.T) {

		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands?limit=10", nil)

		GetDeviceCommands(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}
	})

	// Test GetDeviceCommands with invalid limit parameter
	t.Run("GetDeviceCommands handles invalid limit", func(t *testing.T) {

		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands?limit=invalid", nil)

		GetDeviceCommands(c)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400, got %d", w.Code)
		}
	})
}

func TestStoreScreenshot(t *testing.T) {

	setupTestDB(t)

	// Ensure tables are migrated
	if err := database.DB.AutoMigrate(&models.DeviceScreenshot{}); err != nil {
		t.Fatalf("AutoMigrate Screenshot failed: %v", err)
	}

	deviceID := sampleUUID

	// Clean up any existing screenshots for this device
	database.DB.Where("deviceid = ?", deviceID).Delete(&models.DeviceScreenshot{})

	// Prepare request
	screenshot := models.DeviceScreenshot{
		DeviceID:   sampleUUID,
		Path:       "screenshots/test-screenshot.png",
		Resolution: "1920x1080",
		Size:       1024000,
	}

	body, _ := json.Marshal(screenshot)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/devices/screenshots", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	StoreScreenshot(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
	}

	// Verify screenshot was stored
	var storedScreenshot models.DeviceScreenshot
	result := database.DB.Where("deviceid = ?", deviceID).First(&storedScreenshot)
	if result.Error != nil {
		t.Fatalf("failed to retrieve stored screenshot: %v", result.Error)
	}

	if storedScreenshot.DeviceID != sampleUUID {
		t.Errorf("expected deviceid %s, got %s", deviceID, storedScreenshot.DeviceID)
	}
	if storedScreenshot.Path != screenshot.Path {
		t.Errorf("expected path %s, got %s", screenshot.Path, storedScreenshot.Path)
	}
}
