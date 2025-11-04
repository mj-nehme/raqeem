package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
)

func setupTestDB(t *testing.T) {
	os.Setenv("POSTGRES_USER", os.Getenv("POSTGRES_USER"))
	os.Setenv("POSTGRES_PASSWORD", os.Getenv("POSTGRES_PASSWORD"))
	os.Setenv("POSTGRES_DB", os.Getenv("POSTGRES_DB"))
	os.Setenv("POSTGRES_HOST", os.Getenv("POSTGRES_HOST"))
	os.Setenv("POSTGRES_PORT", os.Getenv("POSTGRES_PORT"))
	database.Connect()
	// Auto-migrate tables
	if err := database.DB.AutoMigrate(&models.Alert{}); err != nil {
		t.Fatalf("AutoMigrate Alert failed: %v", err)
	}
}

func TestReportAndGetAlerts(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}
	setupTestDB(t)

	// Ensure clean slate for test device
	deviceID := "test-device-go"
	database.DB.Where("device_id = ?", deviceID).Delete(&models.Alert{})

	// Prepare gin context for ReportAlert
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/devices/"+deviceID+"/alerts", nil)

	alert := models.Alert{
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Level:     "warning",
		Type:      "cpu_high",
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

	// Prepare gin context for GetDeviceAlerts
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
	c2.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/alerts", nil)

	GetDeviceAlerts(c2)
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
		if os.Getenv("POSTGRES_HOST") == "" {
			t.Skip("POSTGRES_* env vars not set; skipping integration test")
		}
		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/processes", nil)

		GetDeviceProcesses(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("GetDeviceProcesses returned null instead of empty array")
		}
		if body != "[]" {
			t.Logf("Response body: %s", body)
		}
		// Verify it's a valid JSON array
		var processes []models.Process
		if err := json.Unmarshal(w.Body.Bytes(), &processes); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if processes == nil {
			t.Errorf("unmarshaled processes is nil, expected empty slice")
		}
	})

	// Test GetDeviceMetrics
	t.Run("GetDeviceMetrics returns empty array", func(t *testing.T) {
		if os.Getenv("POSTGRES_HOST") == "" {
			t.Skip("POSTGRES_* env vars not set; skipping integration test")
		}
		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/metrics", nil)

		GetDeviceMetrics(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("GetDeviceMetrics returned null instead of empty array")
		}
	})

	// Test GetDeviceActivities
	t.Run("GetDeviceActivities returns empty array", func(t *testing.T) {
		if os.Getenv("POSTGRES_HOST") == "" {
			t.Skip("POSTGRES_* env vars not set; skipping integration test")
		}
		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/activities", nil)

		GetDeviceActivities(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("GetDeviceActivities returned null instead of empty array")
		}
	})

	// Test GetDeviceAlerts
	t.Run("GetDeviceAlerts returns empty array", func(t *testing.T) {
		if os.Getenv("POSTGRES_HOST") == "" {
			t.Skip("POSTGRES_* env vars not set; skipping integration test")
		}
		setupTestDB(t)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/alerts", nil)

		GetDeviceAlerts(c)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("GetDeviceAlerts returned null instead of empty array")
		}
	})

	// Test ListDevices
	t.Run("ListDevices returns empty array", func(t *testing.T) {
		if os.Getenv("POSTGRES_HOST") == "" {
			t.Skip("POSTGRES_* env vars not set; skipping integration test")
		}
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
		if os.Getenv("POSTGRES_HOST") == "" {
			t.Skip("POSTGRES_* env vars not set; skipping integration test")
		}
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
}
