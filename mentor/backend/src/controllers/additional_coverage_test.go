package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var sampleUUID, _ = uuid.Parse("550e8400-e29b-41d4-a716-446655440000")

// TestCreateRemoteCommandWithForwarding tests CreateRemoteCommand with DEVICES_API_URL set
func TestCreateRemoteCommandWithForwarding(t *testing.T) {

	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	if db == nil {
		t.Fatal("Test database is nil")
	}
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Create a mock devices backend server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer mockServer.Close()

	// Set DEVICES_API_URL to mock server
	originalURL := os.Getenv("DEVICES_API_URL")
	_ = os.Setenv("DEVICES_API_URL", mockServer.URL)
	defer func() { _ = os.Setenv("DEVICES_API_URL", originalURL) }()

	t.Run("CreateCommand with forwarding to devices backend", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "test-device-forward"}}

		cmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "get_info",
			Status:      "pending",
		}
		b, _ := json.Marshal(cmd)
		c.Request, _ = http.NewRequest("POST", "/devices/test-device-forward/commands", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateRemoteCommand(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "pending", result.Status)
		assert.Equal(t, "get_info", result.CommandText)

		// Note: Goroutine forwarding is fire-and-forget, no synchronization needed
		// The command is already saved successfully to the database
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
		_ = os.Setenv("DEVICES_API_URL", "http://localhost:99999")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "test-device-fail"}}

		cmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "get_info",
			Status:      "pending",
		}
		b, _ := json.Marshal(cmd)
		c.Request, _ = http.NewRequest("POST", "/devices/test-device-fail/commands", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateRemoteCommand(c)

		// Should still return 200 even if forwarding fails
		assert.Equal(t, http.StatusOK, w.Code)

		// Note: Goroutine forwarding errors are logged but don't affect response
		// No synchronization needed as test verifies successful command creation

		// Restore
		_ = os.Setenv("DEVICES_API_URL", mockServer.URL)
	})
}

// TestGetDeviceCommandsWithLimit tests GetDeviceCommands with various limit values
func TestGetDeviceCommandsWithLimit(t *testing.T) {

	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	if db == nil {
		t.Fatal("Test database is nil")
	}
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceID := sampleUUID.String()

	// Create multiple commands
	for i := 0; i < 5; i++ {
		cmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "test_cmd",
			Status:      "pending",
			CreatedAt:   time.Now(),
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

		var commands []models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &commands)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(commands), 5)
	})
}

// TestStoreScreenshotComprehensive tests StoreScreenshot thoroughly
func TestStoreScreenshotComprehensive(t *testing.T) {

	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	if db == nil {
		t.Fatal("Test database is nil")
	}
	defer database.CleanupTestDB(t, db)
	database.DB = db

	t.Run("Store screenshot with valid data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		screenshot := models.DeviceScreenshot{
			DeviceID: sampleUUID,
			Path:     "https://example.com/screenshot.png",
		}
		b, _ := json.Marshal(screenshot)
		c.Request, _ = http.NewRequest("POST", "/screenshots", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		StoreScreenshot(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DeviceScreenshot
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, sampleUUID, result.DeviceID)
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

		screenshot := models.DeviceScreenshot{
			DeviceID: sampleUUID,
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

	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	if db == nil {
		t.Fatal("Test database is nil")
	}
	defer database.CleanupTestDB(t, db)
	database.DB = db

	t.Run("UpdateProcessList with empty processes array", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: sampleUUID.String()}}

		processes := []models.DeviceProcess{}
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
		c.Params = gin.Params{gin.Param{Key: "id", Value: sampleUUID.String()}}

		processes := make([]models.DeviceProcess, 50)
		for i := 0; i < 50; i++ {
			processes[i] = models.DeviceProcess{
				DeviceID:    sampleUUID,
				ProcessName: fmt.Sprintf("process%d", i),
				PID:         i + 1,
				CPU:         1.0,
				Memory:      1024,
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

	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}
	if db == nil {
		t.Fatal("Test database is nil")
	}
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Create test devices
	uuid1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440011")
	uuid2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440012")
	uuid3 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440013")

	devices := []models.Device{
		{DeviceID: uuid1, DeviceName: "Device 1", DeviceType: "laptop", IsOnline: true},
		{DeviceID: uuid2, DeviceName: "Device 2", DeviceType: "desktop", IsOnline: false},
		{DeviceID: uuid3, DeviceName: "Device 3", DeviceType: "server", IsOnline: true},
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
