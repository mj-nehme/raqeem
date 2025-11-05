package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestCreateRemoteCommandWithForwarding tests CreateRemoteCommand with DEVICES_API_URL set
func TestCreateRemoteCommandWithForwarding(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.RemoteCommand{}); err != nil {
		t.Fatalf("AutoMigrate RemoteCommand failed: %v", err)
	}

	// Create a mock devices backend server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer mockServer.Close()

	// Set DEVICES_API_URL to mock server
	originalURL := os.Getenv("DEVICES_API_URL")
	os.Setenv("DEVICES_API_URL", mockServer.URL)
	defer os.Setenv("DEVICES_API_URL", originalURL)

	t.Run("CreateCommand with forwarding to devices backend", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "test-device-forward"}}

		cmd := models.RemoteCommand{
			DeviceID: "test-device-forward",
			Command:  "get_info",
		}
		b, _ := json.Marshal(cmd)
		c.Request, _ = http.NewRequest("POST", "/devices/test-device-forward/commands", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateRemoteCommand(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.RemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "pending", result.Status)
		assert.Equal(t, "get_info", result.Command)

		// Give goroutine time to complete
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("CreateCommand with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "test-device"}}

		c.Request, _ = http.NewRequest("POST", "/devices/test-device/commands", bytes.NewBufferString("{invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateRemoteCommand(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("CreateCommand forwarding with failed backend", func(t *testing.T) {
		// Set DEVICES_API_URL to non-existent server
		os.Setenv("DEVICES_API_URL", "http://localhost:99999")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "test-device-fail"}}

		cmd := models.RemoteCommand{
			DeviceID: "test-device-fail",
			Command:  "get_info",
		}
		b, _ := json.Marshal(cmd)
		c.Request, _ = http.NewRequest("POST", "/devices/test-device-fail/commands", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateRemoteCommand(c)

		// Should still return 200 even if forwarding fails
		assert.Equal(t, http.StatusOK, w.Code)

		// Give goroutine time to fail
		time.Sleep(100 * time.Millisecond)

		// Restore
		os.Setenv("DEVICES_API_URL", mockServer.URL)
	})
}

// TestGetDeviceCommandsWithLimit tests GetDeviceCommands with various limit values
func TestGetDeviceCommandsWithLimit(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.RemoteCommand{}); err != nil {
		t.Fatalf("AutoMigrate RemoteCommand failed: %v", err)
	}

	deviceID := "test-device-limit-" + time.Now().Format("20060102150405")

	// Clean up any existing commands for this device
	database.DB.Where("device_id = ?", deviceID).Delete(&models.RemoteCommand{})

	// Create multiple commands
	for i := 0; i < 5; i++ {
		cmd := models.RemoteCommand{
			DeviceID:  deviceID,
			Command:   "test_cmd",
			Status:    "pending",
			CreatedAt: time.Now(),
		}
		database.DB.Create(&cmd)
	}

	t.Run("GetCommands with invalid limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands?limit=abc", nil)

		GetDeviceCommands(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetCommands with zero limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands?limit=0", nil)

		GetDeviceCommands(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("GetCommands with large limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands?limit=1000", nil)

		GetDeviceCommands(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var commands []models.RemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &commands)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(commands), 5)
	})
}

// TestStoreScreenshotComprehensive tests StoreScreenshot thoroughly
func TestStoreScreenshotComprehensive(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.Screenshot{}); err != nil {
		t.Fatalf("AutoMigrate Screenshot failed: %v", err)
	}

	t.Run("Store screenshot with valid data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		screenshot := models.Screenshot{
			DeviceID: "test-device-screenshot",
			Path:     "https://example.com/screenshot.png",
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
		assert.Equal(t, "https://example.com/screenshot.png", result.Path)
	})

	t.Run("Store screenshot with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest("POST", "/screenshots", bytes.NewBufferString("{invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		StoreScreenshot(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Store screenshot with empty body", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest("POST", "/screenshots", bytes.NewBuffer(nil))
		c.Request.Header.Set("Content-Type", "application/json")

		StoreScreenshot(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Store screenshot with minimal data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		screenshot := models.Screenshot{
			DeviceID: "minimal-device",
		}
		b, _ := json.Marshal(screenshot)
		c.Request, _ = http.NewRequest("POST", "/screenshots", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		StoreScreenshot(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestUpdateProcessListEdgeCases tests UpdateProcessList with edge cases
func TestUpdateProcessListEdgeCases(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.Process{}); err != nil {
		t.Fatalf("AutoMigrate Process failed: %v", err)
	}

	t.Run("UpdateProcessList with empty processes array", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "test-device-empty"}}

		processes := []models.Process{}
		b, _ := json.Marshal(processes)
		c.Request, _ = http.NewRequest("POST", "/devices/test-device-empty/processes", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("UpdateProcessList with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "test-device"}}

		c.Request, _ = http.NewRequest("POST", "/devices/test-device/processes", bytes.NewBufferString("{invalid"))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("UpdateProcessList with many processes", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "test-device-many"}}

		processes := make([]models.Process, 50)
		for i := 0; i < 50; i++ {
			processes[i] = models.Process{
				DeviceID: "test-device-many",
				Name:     "process" + string(rune(i)),
				PID:      i + 1,
				CPU:      1.0,
				Memory:   1024,
			}
		}

		b, _ := json.Marshal(processes)
		c.Request, _ = http.NewRequest("POST", "/devices/test-device-many/processes", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestListDevicesWithQuery tests ListDevices with various query parameters
func TestListDevicesWithQuery(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.Device{}); err != nil {
		t.Fatalf("AutoMigrate Device failed: %v", err)
	}

	// Create test devices
	devices := []models.Device{
		{ID: "device1", Name: "Device 1", Type: "laptop", IsOnline: true},
		{ID: "device2", Name: "Device 2", Type: "desktop", IsOnline: false},
		{ID: "device3", Name: "Device 3", Type: "server", IsOnline: true},
	}
	for _, d := range devices {
		database.DB.Create(&d)
	}

	t.Run("ListDevices with status online", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/devices?status=online", nil)

		ListDevices(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []models.Device
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
	})

	t.Run("ListDevices with status offline", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/devices?status=offline", nil)

		ListDevices(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []models.Device
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
	})

	t.Run("ListDevices with status all", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/devices?status=all", nil)

		ListDevices(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []models.Device
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 3)
	})
}
