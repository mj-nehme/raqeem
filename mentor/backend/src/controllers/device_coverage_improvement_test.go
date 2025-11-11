package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
)

// TestCreateRemoteCommandComprehensive tests CreateRemoteCommand with various scenarios
func TestCreateRemoteCommandComprehensive(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.DeviceRemoteCommands{}); err != nil {
		t.Fatalf("AutoMigrate RemoteCommand failed: %v", err)
	}

	t.Run("Create command with valid payload", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "test-device-cmd"}}

		cmd := models.DeviceRemoteCommands{
			Command: "get_info",
		}
		b, _ := json.Marshal(cmd)
		c.Request, _ = http.NewRequest("POST", "/devices/test-device-cmd/commands", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateRemoteCommand(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}

		var result models.DeviceRemoteCommands
		if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if result.DeviceID != "test-device-cmd" {
			t.Errorf("expected device_id test-device-cmd, got %s", result.DeviceID)
		}
		if result.Command != "get_info" {
			t.Errorf("expected command get_info, got %s", result.Command)
		}
		if result.Status != "pending" {
			t.Errorf("expected status pending, got %s", result.Status)
		}
	})

	t.Run("Create command with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "test-device-cmd"}}

		c.Request, _ = http.NewRequest("POST", "/devices/test-device-cmd/commands", bytes.NewReader([]byte("invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")

		CreateRemoteCommand(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

// TestUpdateProcessListComprehensive tests UpdateProcessList with various scenarios
func TestUpdateProcessListComprehensive(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.DeviceProcesses{}); err != nil {
		t.Fatalf("AutoMigrate Process failed: %v", err)
	}

	t.Run("Update process list with valid processes", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		processes := []models.DeviceProcesses{
			{DeviceID: "test-device-proc", PID: 1234, Name: "test-process", CPU: 10.5, Memory: 1024},
			{DeviceID: "test-device-proc", PID: 5678, Name: "another-process", CPU: 5.2, Memory: 2048},
		}
		b, _ := json.Marshal(processes)
		c.Request, _ = http.NewRequest("POST", "/devices/processes", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Update process list with empty array", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		processes := []models.DeviceProcesses{}
		b, _ := json.Marshal(processes)
		c.Request, _ = http.NewRequest("POST", "/devices/processes", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("Update process list with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest("POST", "/devices/processes", bytes.NewReader([]byte("not json")))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

// TestListDevicesComprehensive tests ListDevices with various scenarios
func TestListDevicesComprehensive(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.Device{}); err != nil {
		t.Fatalf("AutoMigrate Device failed: %v", err)
	}

	// Clean up test devices first
	database.DB.Where("id LIKE ?", "test-list-device-%").Delete(&models.Device{})

	t.Run("List devices returns array", func(t *testing.T) {
		// Create a test device
		device := models.Device{
			ID:   "test-list-device-1",
			Name: "Test Device 1",
			Type: "laptop",
			OS:   "Linux",
		}
		database.DB.Create(&device)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/devices", nil)

		ListDevices(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var devices []models.Device
		if err := json.Unmarshal(w.Body.Bytes(), &devices); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		// Should have at least our test device
		found := false
		for _, d := range devices {
			if d.ID == "test-list-device-1" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("test device not found in list")
		}
	})

	t.Run("List devices with pagination", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/devices?page=1&limit=10", nil)

		ListDevices(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})
}

// TestGetPendingCommandsComprehensive tests GetPendingCommands with various scenarios
func TestGetPendingCommandsComprehensive(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.DeviceRemoteCommands{}); err != nil {
		t.Fatalf("AutoMigrate RemoteCommand failed: %v", err)
	}

	deviceID := "test-pending-device"

	// Clean up first
	database.DB.Where("device_id = ?", deviceID).Delete(&models.DeviceRemoteCommands{})

	t.Run("Get pending commands returns array", func(t *testing.T) {
		// Create a pending command
		cmd := models.DeviceRemoteCommands{
			DeviceID: deviceID,
			Command:  "get_info",
			Status:   "pending",
		}
		database.DB.Create(&cmd)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands/pending", nil)

		GetPendingCommands(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var commands []models.DeviceRemoteCommands
		if err := json.Unmarshal(w.Body.Bytes(), &commands); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(commands) == 0 {
			t.Errorf("expected at least one pending command")
		}
	})

	t.Run("Get pending commands for device with no commands", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: "device-with-no-commands"}}
		c.Request, _ = http.NewRequest("GET", "/devices/device-with-no-commands/commands/pending", nil)

		GetPendingCommands(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		body := w.Body.String()
		if body == "null" {
			t.Errorf("expected empty array, got null")
		}
	})
}

// TestUpdateCommandStatusComprehensive tests UpdateCommandStatus with various scenarios
func TestUpdateCommandStatusComprehensive(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.DeviceRemoteCommands{}); err != nil {
		t.Fatalf("AutoMigrate RemoteCommand failed: %v", err)
	}

	t.Run("Update command status with valid data", func(t *testing.T) {
		// Create a command first
		cmd := models.DeviceRemoteCommands{
			DeviceID: "test-status-device",
			Command:  "get_info",
			Status:   "pending",
		}
		database.DB.Create(&cmd)

		// Update its status
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		updateCmd := models.DeviceRemoteCommands{
			ID:       cmd.ID,
			Status:   "completed",
			Result:   "Command executed successfully",
			ExitCode: 0,
		}
		b, _ := json.Marshal(updateCmd)
		c.Request, _ = http.NewRequest("POST", "/commands/status", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateCommandStatus(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d, body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("Update command status with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest("POST", "/commands/status", bytes.NewReader([]byte("{invalid")))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateCommandStatus(c)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})
}

// TestGetDeviceCommandsComprehensive tests GetDeviceCommands with edge cases
func TestGetDeviceCommandsComprehensive(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}

	gin.SetMode(gin.TestMode)
	database.Connect()
	if err := database.DB.AutoMigrate(&models.DeviceRemoteCommands{}); err != nil {
		t.Fatalf("AutoMigrate RemoteCommand failed: %v", err)
	}

	deviceID := "test-cmd-history-device"

	// Clean up first
	database.DB.Where("device_id = ?", deviceID).Delete(&models.DeviceRemoteCommands{})

	t.Run("Get commands without limit parameter", func(t *testing.T) {
		// Create some commands
		for i := 0; i < 3; i++ {
			cmd := models.DeviceRemoteCommands{
				DeviceID: deviceID,
				Command:  "test command",
				Status:   "completed",
			}
			database.DB.Create(&cmd)
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands", nil)

		GetDeviceCommands(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var commands []models.DeviceRemoteCommands
		if err := json.Unmarshal(w.Body.Bytes(), &commands); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(commands) < 3 {
			t.Errorf("expected at least 3 commands, got %d", len(commands))
		}
	})

	t.Run("Get commands with valid limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands?limit=1", nil)

		GetDeviceCommands(c)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var commands []models.DeviceRemoteCommands
		if err := json.Unmarshal(w.Body.Bytes(), &commands); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(commands) > 1 {
			t.Errorf("expected at most 1 command due to limit, got %d", len(commands))
		}
	})
}
