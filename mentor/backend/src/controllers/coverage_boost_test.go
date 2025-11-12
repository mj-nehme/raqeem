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
	"github.com/stretchr/testify/require"
)

// TestGetDeviceCommandsWithSQLite tests GetDeviceCommands with SQLite
func TestGetDeviceCommandsWithSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := database.SetupTestDB(t)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)

	// Set the global DB for controllers
	database.DB = db

	deviceID := "test-device-commands-sqlite"

	// Clean up any existing commands for this device
	db.Where("device_id = ?", deviceID).Delete(&models.DeviceRemoteCommand{})

	// Create multiple commands
	for i := 0; i < 5; i++ {
		cmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "test_cmd",
			Status:      "completed",
			CreatedAt:   time.Now().Add(time.Duration(-i) * time.Hour),
		}
		db.Create(&cmd)
	}

	t.Run("GetCommands without limit parameter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands", nil)

		GetDeviceCommands(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var commands []models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &commands)
		assert.NoError(t, err)
		assert.Equal(t, 5, len(commands))
	})

	t.Run("GetCommands with valid limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands?limit=3", nil)

		GetDeviceCommands(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var commands []models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &commands)
		assert.NoError(t, err)
		assert.Equal(t, 3, len(commands))
	})

	t.Run("GetCommands with invalid limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands?limit=invalid", nil)

		GetDeviceCommands(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("GetCommands for non-existent device", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "non-existent-device"}}
		c.Request, _ = http.NewRequest("GET", "/devices/non-existent-device/commands", nil)

		GetDeviceCommands(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var commands []models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &commands)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(commands))
	})
}

// TestStoreScreenshotWithSQLite tests StoreScreenshot thoroughly with SQLite
func TestStoreScreenshotWithSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := database.SetupTestDB(t)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)

	// Set the global DB for controllers
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
		assert.Equal(t, "test-device-screenshot", result.DeviceID)
		assert.Equal(t, "https://example.com/screenshot.png", result.Path)
		assert.NotZero(t, result.DeviceID)
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

		c.Request, _ = http.NewRequest("POST", "/screenshots", bytes.NewBuffer([]byte("{}")))
		c.Request.Header.Set("Content-Type", "application/json")

		StoreScreenshot(c)

		// Empty body should still work as long as it's valid JSON
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Store screenshot with full data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		screenshot := models.DeviceScreenshot{
			DeviceID: sampleUUID,
			Path:     "s3://bucket/screenshots/test.png",
		}
		b, _ := json.Marshal(screenshot)
		c.Request, _ = http.NewRequest("POST", "/screenshots", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		StoreScreenshot(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DeviceScreenshot
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "test-device-full", result.DeviceID)
		assert.Equal(t, "s3://bucket/screenshots/test.png", result.Path)
	})
}

// TestCreateRemoteCommandWithForwardingSQLite tests CreateRemoteCommand with SQLite
func TestCreateRemoteCommandWithForwardingSQLite(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := database.SetupTestDB(t)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)

	// Set the global DB for controllers
	database.DB = db

	// Create a mock devices backend server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer mockServer.Close()

	// Save and restore original DEVICES_API_URL
	originalURL := os.Getenv("DEVICES_API_URL")
	defer func() {
		if originalURL != "" {
			_ = os.Setenv("DEVICES_API_URL", originalURL)
		} else {
			_ = os.Unsetenv("DEVICES_API_URL")
		}
	}()

	t.Run("CreateCommand with forwarding enabled", func(t *testing.T) {
		_ = os.Setenv("DEVICES_API_URL", mockServer.URL)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		cmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "get_info",
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

		// Give goroutine time to forward the command
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("CreateCommand without forwarding", func(t *testing.T) {
		_ = os.Unsetenv("DEVICES_API_URL")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		cmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "get_status",
		}
		b, _ := json.Marshal(cmd)
		c.Request, _ = http.NewRequest("POST", "/devices/test-device-no-forward/commands", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateRemoteCommand(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "pending", result.Status)
	})

	t.Run("CreateCommand with failed backend", func(t *testing.T) {
		_ = os.Setenv("DEVICES_API_URL", "http://localhost:99999")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		cmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "get_info",
		}
		b, _ := json.Marshal(cmd)
		c.Request, _ = http.NewRequest("POST", "/devices/test-device-fail/commands", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateRemoteCommand(c)

		// Should still return 200 even if forwarding fails
		assert.Equal(t, http.StatusOK, w.Code)

		// Give goroutine time to attempt forwarding
		time.Sleep(100 * time.Millisecond)
	})

	t.Run("CreateCommand with backend returning error", func(t *testing.T) {
		errorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer errorServer.Close()

		_ = os.Setenv("DEVICES_API_URL", errorServer.URL)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		cmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "get_info",
		}
		b, _ := json.Marshal(cmd)
		c.Request, _ = http.NewRequest("POST", "/devices/test-device-error/commands", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateRemoteCommand(c)

		assert.Equal(t, http.StatusOK, w.Code)

		// Give goroutine time to process
		time.Sleep(100 * time.Millisecond)
	})
}

// TestUpdateCommandStatusEdgeCases tests UpdateCommandStatus with various edge cases
func TestUpdateCommandStatusEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db := database.SetupTestDB(t)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)

	// Set the global DB for controllers
	database.DB = db

	// Create a test command
	testCmd := models.DeviceRemoteCommand{
		DeviceID:    sampleUUID,
		CommandText: "test_command",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	db.Create(&testCmd)

	t.Run("Update status to completed", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		updateCmd := models.DeviceRemoteCommand{
			CommandID: sampleUUID,
			Status:    "completed",
			Result:    "success",
			ExitCode:  0,
		}
		b, _ := json.Marshal(updateCmd)
		c.Request, _ = http.NewRequest("PUT", "/commands/status", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateCommandStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "completed", result.Status)
		assert.NotZero(t, result.CompletedAt)
	})

	t.Run("Update status to failed", func(t *testing.T) {
		// Create another test command
		failCmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "fail_command",
			Status:      "pending",
			CreatedAt:   time.Now(),
		}
		db.Create(&failCmd)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		updateCmd := models.DeviceRemoteCommand{
			CommandID: sampleUUID,
			Status:    "failed",
			Result:    "error occurred",
			ExitCode:  1,
		}
		b, _ := json.Marshal(updateCmd)
		c.Request, _ = http.NewRequest("PUT", "/commands/status", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateCommandStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "failed", result.Status)
		assert.NotZero(t, result.CompletedAt)
	})

	t.Run("Update status to running (no CompletedAt)", func(t *testing.T) {
		// Create another test command
		runCmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "run_command",
			Status:      "pending",
			CreatedAt:   time.Now(),
		}
		db.Create(&runCmd)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		updateCmd := models.DeviceRemoteCommand{
			CommandID: runCmd.CommandID,
			Status:    "running",
		}
		b, _ := json.Marshal(updateCmd)
		c.Request, _ = http.NewRequest("PUT", "/commands/status", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateCommandStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, "running", result.Status)
	})
}
