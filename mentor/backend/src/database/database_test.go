package database

import (
	"fmt"
	"os"
	"testing"
	"time"

	"mentor-backend/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupTestDB(t *testing.T) {
	// Test with valid PostgreSQL connection
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "postgres")
	os.Setenv("POSTGRES_PASSWORD", "postgres")
	os.Setenv("POSTGRES_TEST_DB", "raqeem_test")

	db := SetupTestDB(t)
	require.NotNil(t, db)

	// Test that all tables are created
	tables := []string{
		"devices", "device_metrics", "processes", "activities",
		"activity_logs", "remote_commands", "screenshots", "alerts",
	}

	for _, table := range tables {
		var exists bool
		err := db.Raw("SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = ?)", table).Scan(&exists).Error
		assert.NoError(t, err)
		assert.True(t, exists, "Table %s should exist", table)
	}

	CleanupTestDB(t, db)
}

func TestCleanupTestDB(t *testing.T) {
	db := SetupTestDB(t)
	require.NotNil(t, db)

	// Insert test data
	device := models.Device{
		ID:       "test-cleanup-device",
		Name:     "Test Device",
		IsOnline: true,
		LastSeen: time.Now(),
	}
	db.Create(&device)

	metrics := models.DeviceMetrics{
		DeviceID:  "test-cleanup-device",
		CPUUsage:  50.0,
		Timestamp: time.Now(),
	}
	db.Create(&metrics)

	// Verify data exists
	var deviceCount int64
	db.Model(&models.Device{}).Count(&deviceCount)
	assert.Greater(t, deviceCount, int64(0))

	var metricsCount int64
	db.Model(&models.DeviceMetrics{}).Count(&metricsCount)
	assert.Greater(t, metricsCount, int64(0))

	// Cleanup
	CleanupTestDB(t, db)

	// Verify data is cleaned
	db.Model(&models.Device{}).Count(&deviceCount)
	assert.Equal(t, int64(0), deviceCount)

	db.Model(&models.DeviceMetrics{}).Count(&metricsCount)
	assert.Equal(t, int64(0), metricsCount)
}

func TestCreateTestDatabase(t *testing.T) {
	// Set test environment
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "postgres")
	os.Setenv("POSTGRES_PASSWORD", "postgres")
	os.Setenv("POSTGRES_TEST_DB", "raqeem_test_creation")

	err := CreateTestDatabase()
	assert.NoError(t, err)

	// Verify database was created by connecting to it
	db := SetupTestDB(t)
	assert.NotNil(t, db)

	CleanupTestDB(t, db)
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with existing environment variable
	os.Setenv("TEST_ENV_VAR", "test_value")
	result := getEnvOrDefault("TEST_ENV_VAR", "default")
	assert.Equal(t, "test_value", result)

	// Test with non-existing environment variable
	os.Unsetenv("NON_EXISTING_VAR")
	result = getEnvOrDefault("NON_EXISTING_VAR", "default_value")
	assert.Equal(t, "default_value", result)

	// Test with empty environment variable
	os.Setenv("EMPTY_VAR", "")
	result = getEnvOrDefault("EMPTY_VAR", "default")
	assert.Equal(t, "default", result)
}

func TestDatabaseConnection(t *testing.T) {
	// Test database connection with proper environment variables
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_USER", "postgres")
	os.Setenv("POSTGRES_PASSWORD", "postgres")
	os.Setenv("POSTGRES_DB", "raqeem_test")

	// Store original DB value
	originalDB := DB
	defer func() { DB = originalDB }()

	// Test connection
	Connect()
	assert.NotNil(t, DB)

	// Test that we can perform a simple query
	var result int
	err := DB.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestDatabaseMigrationIntegrity(t *testing.T) {
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test that we can create records of all model types
	device := models.Device{
		ID:          "test-migration-device",
		Name:        "Migration Test Device",
		Type:        "laptop",
		OS:          "Linux",
		IPAddress:   "192.168.1.100",
		MacAddress:  "aa:bb:cc:dd:ee:ff",
		Location:    "Test Lab",
		IsOnline:    true,
		CurrentUser: "testuser",
		LastSeen:    time.Now(),
	}
	assert.NoError(t, db.Create(&device).Error)

	metrics := models.DeviceMetrics{
		DeviceID:    device.ID,
		CPUUsage:    75.5,
		CPUTemp:     65.0,
		MemoryTotal: 16777216000,
		MemoryUsed:  10737418240,
		SwapUsed:    0,
		DiskTotal:   1000000000000,
		DiskUsed:    500000000000,
		NetBytesIn:  1024,
		NetBytesOut: 512,
		Timestamp:   time.Now(),
	}
	assert.NoError(t, db.Create(&metrics).Error)

	process := models.Process{
		DeviceID:  device.ID,
		PID:       1234,
		Name:      "test-process",
		CPU:       15.5,
		Memory:    536870912,
		Command:   "/usr/bin/test",
		Timestamp: time.Now(),
	}
	assert.NoError(t, db.Create(&process).Error)

	activity := models.Activity{
		UserID:    "testuser",
		Location:  "office",
		Filename:  "test.jpg",
		Timestamp: time.Now(),
	}
	assert.NoError(t, db.Create(&activity).Error)

	activityLog := models.ActivityLog{
		DeviceID:    device.ID,
		Type:        "app_launch",
		Description: "Test activity",
		App:         "TestApp",
		Duration:    3600,
		Timestamp:   time.Now(),
	}
	assert.NoError(t, db.Create(&activityLog).Error)

	command := models.RemoteCommand{
		DeviceID:  device.ID,
		Command:   "test command",
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	assert.NoError(t, db.Create(&command).Error)

	screenshot := models.Screenshot{
		DeviceID:   device.ID,
		Path:       "/path/to/screenshot.jpg",
		Resolution: "1920x1080",
		Size:       1024000,
		Timestamp:  time.Now(),
	}
	assert.NoError(t, db.Create(&screenshot).Error)

	alert := models.Alert{
		DeviceID:  device.ID,
		Type:      "security",
		Level:     "critical",
		Message:   "Test alert",
		Value:     95.0,
		Threshold: 90.0,
		Timestamp: time.Now(),
	}
	assert.NoError(t, db.Create(&alert).Error)

	// Verify all records were created
	var count int64

	db.Model(&models.Device{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceMetrics{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.Process{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.Activity{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.ActivityLog{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.RemoteCommand{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.Screenshot{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.Alert{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))
}

func TestDatabaseTransactionRollback(t *testing.T) {
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test transaction rollback
	tx := db.Begin()

	device := models.Device{
		ID:   "test-rollback-device",
		Name: "Rollback Test Device",
	}

	err := tx.Create(&device).Error
	assert.NoError(t, err)

	// Rollback transaction
	tx.Rollback()

	// Verify device was not actually saved
	var count int64
	db.Model(&models.Device{}).Where("id = ?", "test-rollback-device").Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestConcurrentDatabaseAccess(t *testing.T) {
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test concurrent access doesn't cause issues
	done := make(chan bool, 2)

	// Goroutine 1: Create devices
	go func() {
		for i := 0; i < 10; i++ {
			device := models.Device{
				ID:   fmt.Sprintf("concurrent-device-1-%d", i),
				Name: fmt.Sprintf("Concurrent Device 1 %d", i),
			}
			db.Create(&device)
		}
		done <- true
	}()

	// Goroutine 2: Create devices with different prefix
	go func() {
		for i := 0; i < 10; i++ {
			device := models.Device{
				ID:   fmt.Sprintf("concurrent-device-2-%d", i),
				Name: fmt.Sprintf("Concurrent Device 2 %d", i),
			}
			db.Create(&device)
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify all devices were created
	var count int64
	db.Model(&models.Device{}).Where("id LIKE ?", "concurrent-device-%").Count(&count)
	assert.Equal(t, int64(20), count)
}
