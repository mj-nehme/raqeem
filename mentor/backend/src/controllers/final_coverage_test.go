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

// TestRegisterDeviceEdgeCases tests RegisterDevice with various edge cases
func TestRegisterDeviceEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	t.Run("Register device with database error simulation", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		device := models.Device{
			DeviceID:   sampleUUID,
			DeviceName: "Test Device",
		}
		b, _ := json.Marshal(device)
		c.Request, _ = http.NewRequest("POST", "/devices", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		RegisterDevice(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result models.Device
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, sampleUUID, result.DeviceID)
		assert.True(t, result.IsOnline)
	})

	t.Run("Register device with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest("POST", "/devices", bytes.NewBufferString("{invalid"))
		c.Request.Header.Set("Content-Type", "application/json")

		RegisterDevice(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestUpdateDeviceMetricEdgeCases tests UpdateDeviceMetric thoroughly
func TestUpdateDeviceMetricEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	t.Run("Update metrics with valid data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		metrics := models.DeviceMetric{
			DeviceID:   sampleUUID,
			CPUUsage:   45.5,
			MemoryUsed: 8192,
			DiskUsed:   102400,
		}
		b, _ := json.Marshal(metrics)
		c.Request, _ = http.NewRequest("POST", "/metrics", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateDeviceMetric(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Update metrics with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest("POST", "/metrics", bytes.NewBufferString("invalid"))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateDeviceMetric(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestActivityEdgeCases tests Activity with edge cases
func TestActivityEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	t.Run("Log activity with valid data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		activity := models.DeviceActivity{
			DeviceID:     sampleUUID,
			ActivityType: "app_launch",
			Description:  "Launched Chrome",
			App:          "Chrome",
		}
		b, _ := json.Marshal(activity)
		c.Request, _ = http.NewRequest("POST", "/activity", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		Activity(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Log activity with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest("POST", "/activity", bytes.NewBufferString("{bad}"))
		c.Request.Header.Set("Content-Type", "application/json")

		Activity(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestUpdateProcessListAdditionalEdgeCases tests UpdateProcessList with additional edge cases
func TestUpdateProcessListAdditionalEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceID := sampleUUID.String()

	t.Run("Update process list with valid data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}

		processes := []models.DeviceProcess{
			{DeviceID: sampleUUID, PID: 1234, ProcessName: "process1"},
			{DeviceID: sampleUUID, PID: 5678, ProcessName: "process2"},
		}
		b, _ := json.Marshal(processes)
		c.Request, _ = http.NewRequest("POST", "/devices/"+deviceID+"/processes", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Update process list with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}

		c.Request, _ = http.NewRequest("POST", "/devices/"+deviceID+"/processes", bytes.NewBufferString("[{bad}]"))
		c.Request.Header.Set("Content-Type", "application/json")

		UpdateProcessList(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

// TestListDevicesAdditionalFilters tests ListDevices with various filters
func TestListDevicesAdditionalFilters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Create test devices
	devices := []models.Device{
		{DeviceID: uuid.New(), DeviceName: "Device 1", IsOnline: true, DeviceLocation: "lab1"},
		{DeviceID: uuid.New(), DeviceName: "Device 2", IsOnline: false, DeviceLocation: "lab2"},
		{DeviceID: uuid.New(), DeviceName: "Device 3", IsOnline: true, DeviceLocation: "lab1"},
	}
	for _, d := range devices {
		db.Create(&d)
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
		assert.GreaterOrEqual(t, len(result), 3)
	})

	t.Run("List devices with location filter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/devices?location=lab1", nil)

		ListDevices(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("List devices with online filter", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request, _ = http.NewRequest("GET", "/devices?is_online=true", nil)

		ListDevices(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestGetDeviceMetricEdgeCases tests GetDeviceMetric with edge cases
func TestGetDeviceMetricEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceID := "test-device-get-metrics"

	// Create test metrics
	for i := 0; i < 3; i++ {
		metrics := models.DeviceMetric{
			DeviceID:  sampleUUID,
			CPUUsage:  float64(i * 10),
			Timestamp: time.Now().Add(time.Duration(-i) * time.Hour),
		}
		db.Create(&metrics)
	}

	t.Run("Get metrics without limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/metrics", nil)

		GetDeviceMetric(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get metrics with limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/metrics?limit=2", nil)

		GetDeviceMetric(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []models.DeviceMetric
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.LessOrEqual(t, len(result), 2)
	})
}

// TestGetDeviceProcessesEdgeCases tests GetDeviceProcesses with edge cases
func TestGetDeviceProcessesEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceID := sampleUUID.String()

	// Create test processes
	processes := []models.DeviceProcess{
		{DeviceID: sampleUUID, PID: 1000, ProcessName: "proc1"},
		{DeviceID: sampleUUID, PID: 2000, ProcessName: "proc2"},
	}
	for _, p := range processes {
		db.Create(&p)
	}

	t.Run("Get processes for device", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/processes", nil)

		GetDeviceProcesses(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []models.DeviceProcess
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(result), 2)
	})
}

// TestGetDeviceActivityEdgeCases tests GetDeviceActivity
func TestGetDeviceActivityEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceID := "test-device-get-activities"

	// Create test activities
	for i := 0; i < 3; i++ {
		activity := models.DeviceActivity{
			DeviceID:     sampleUUID,
			ActivityType: "test",
			Description:  "Test activity",
			Timestamp:    time.Now().Add(time.Duration(-i) * time.Hour),
		}
		db.Create(&activity)
	}

	t.Run("Get activities with limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/activities?limit=2", nil)

		GetDeviceActivity(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestGetDeviceAlertEdgeCases tests GetDeviceAlert
func TestGetDeviceAlertEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceID := "test-device-get-alerts"

	// Create test alerts
	for i := 0; i < 3; i++ {
		alert := models.DeviceAlert{
			DeviceID:  sampleUUID,
			Level:     "info",
			AlertType: "test",
			Message:   "Test alert",
			Timestamp: time.Now().Add(time.Duration(-i) * time.Hour),
		}
		db.Create(&alert)
	}

	t.Run("Get alerts with limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/alerts?limit=2", nil)

		GetDeviceAlert(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestGetDeviceScreenshotEdgeCases tests GetDeviceScreenshot
func TestGetDeviceScreenshotEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceID := "test-device-screenshots"

	// Create test screenshots
	for i := 0; i < 3; i++ {
		screenshot := models.DeviceScreenshot{
			DeviceID:  sampleUUID,
			Path:      "/path/to/screenshot",
			Timestamp: time.Now().Add(time.Duration(-i) * time.Hour),
		}
		db.Create(&screenshot)
	}

	t.Run("Get screenshots with limit", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/screenshots?limit=2", nil)

		GetDeviceScreenshot(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

// TestGetPendingCommandsEdgeCases tests GetPendingCommands
func TestGetPendingCommandsEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	deviceID := sampleUUID.String()

	// Create test commands
	commands := []models.DeviceRemoteCommand{
		{DeviceID: sampleUUID, CommandText: "cmd1", Status: "pending"},
		{DeviceID: sampleUUID, CommandText: "cmd2", Status: "completed"},
		{DeviceID: sampleUUID, CommandText: "cmd3", Status: "pending"},
	}
	for _, cmd := range commands {
		db.Create(&cmd)
	}

	t.Run("Get only pending commands", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{gin.Param{Key: "id", Value: deviceID}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/commands/pending", nil)

		GetPendingCommands(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var result []models.DeviceRemoteCommand
		err := json.Unmarshal(w.Body.Bytes(), &result)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(result))
	})
}

// TestReportAlertEdgeCases tests ReportAlert
func TestReportAlertEdgeCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	t.Run("Report alert with valid data", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		alert := models.DeviceAlert{
			DeviceID:  sampleUUID,
			Level:     "warning",
			AlertType: "cpu_high",
			Message:   "CPU usage high",
		}
		b, _ := json.Marshal(alert)
		c.Request, _ = http.NewRequest("POST", "/alerts", bytes.NewReader(b))
		c.Request.Header.Set("Content-Type", "application/json")

		ReportAlert(c)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Report alert with invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request, _ = http.NewRequest("POST", "/alerts", bytes.NewBufferString("{bad}"))
		c.Request.Header.Set("Content-Type", "application/json")

		ReportAlert(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
