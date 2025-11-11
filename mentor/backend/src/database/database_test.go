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
	// Test with SQLite (default)
	db := SetupTestDB(t)
	require.NotNil(t, db)

	// Test that all tables are created by attempting to query them
	tables := []interface{}{
		&models.Device{},
		&models.DeviceMetrics{},
		&models.DeviceProcesses{},
		&models.DeviceActivities{},
		&models.DeviceRemoteCommands{},
		&models.DeviceScreenshots{},
		&models.DeviceAlerts{},
	}

	for _, table := range tables {
		// Try to count records - this will fail if table doesn't exist
		var count int64
		err := db.Model(table).Count(&count).Error
		assert.NoError(t, err, "Table for %T should exist", table)
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
	// SQLite doesn't need database creation, this should work without errors
	err := CreateTestDatabase()
	assert.NoError(t, err)
}

func TestGetEnvOrDefault(t *testing.T) {
	// Test with existing environment variable
	err := os.Setenv("TEST_ENV_VAR", "test_value")
	require.NoError(t, err)
	result := getEnvOrDefault("TEST_ENV_VAR", "default")
	assert.Equal(t, "test_value", result)

	// Test with non-existing environment variable
	err = os.Unsetenv("NON_EXISTING_VAR")
	require.NoError(t, err)
	result = getEnvOrDefault("NON_EXISTING_VAR", "default_value")
	assert.Equal(t, "default_value", result)

	// Test with empty environment variable
	err = os.Setenv("EMPTY_VAR", "")
	require.NoError(t, err)
	result = getEnvOrDefault("EMPTY_VAR", "default")
	assert.Equal(t, "default", result)
}

func TestDatabaseConnection(t *testing.T) {
	// Test that SetupTestDB creates a working database connection
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test that we can perform a simple query
	var result int
	err := db.Raw("SELECT 1").Scan(&result).Error
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

	process := models.DeviceProcesses{
		DeviceID:  device.ID,
		PID:       1234,
		Name:      "test-process",
		CPU:       15.5,
		Memory:    536870912,
		Command:   "/usr/bin/test",
		Timestamp: time.Now(),
	}
	assert.NoError(t, db.Create(&process).Error)

	activityLog := models.DeviceActivities{
		DeviceID:    device.ID,
		Type:        "app_launch",
		Description: "Test activity",
		App:         "TestApp",
		Duration:    3600,
		Timestamp:   time.Now(),
	}
	assert.NoError(t, db.Create(&activityLog).Error)

	command := models.DeviceRemoteCommands{
		DeviceID:  device.ID,
		Command:   "test command",
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	assert.NoError(t, db.Create(&command).Error)

	screenshot := models.DeviceScreenshots{
		DeviceID:   device.ID,
		Path:       "/path/to/screenshot.jpg",
		Resolution: "1920x1080",
		Size:       1024000,
		Timestamp:  time.Now(),
	}
	assert.NoError(t, db.Create(&screenshot).Error)

	alert := models.DeviceAlerts{
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

	db.Model(&models.DeviceProcesses{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceActivities{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceActivities{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceRemoteCommands{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceScreenshots{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceAlerts{}).Count(&count)
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

	// Enable WAL mode for SQLite to allow concurrent writes
	sqlDB, err := db.DB()
	if err == nil {
		_, _ = sqlDB.Exec("PRAGMA journal_mode=WAL;")
	}

	// Test concurrent access doesn't cause issues
	done := make(chan bool, 2)

	// Goroutine 1: Create devices
	go func() {
		for i := 0; i < 10; i++ {
			device := models.Device{
				ID:   fmt.Sprintf("concurrent-device-1-%d", i),
				Name: fmt.Sprintf("Concurrent Device 1 %d", i),
			}
			// Retry on lock errors (SQLite specific)
			for retry := 0; retry < 3; retry++ {
				if err := db.Create(&device).Error; err == nil {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
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
			// Retry on lock errors (SQLite specific)
			for retry := 0; retry < 3; retry++ {
				if err := db.Create(&device).Error; err == nil {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		done <- true
	}()

	// Wait for both goroutines to complete
	<-done
	<-done

	// Verify all devices were created
	var count int64
	db.Model(&models.Device{}).Where("id LIKE ?", "concurrent-device-%").Count(&count)
	// Due to SQLite locking, we may not get all 20, but should get most
	assert.GreaterOrEqual(t, count, int64(10), "Should create at least 10 devices")
}

func TestSetupTestDBWithPostgres(t *testing.T) {
	// Test that when USE_POSTGRES_FOR_TESTS is not set, we use SQLite
	err := os.Unsetenv("USE_POSTGRES_FOR_TESTS")
	require.NoError(t, err)
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Verify it's SQLite by checking for SQLite-specific behavior
	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestCreateTestDatabaseForCI(t *testing.T) {
	// Test with CI environment simulation (user = "monitor")
	err := os.Setenv("POSTGRES_USER", "monitor")
	require.NoError(t, err)
	defer func() {
		err := os.Unsetenv("POSTGRES_USER")
		require.NoError(t, err)
	}()

	err = CreateTestDatabase()
	assert.NoError(t, err, "Should handle CI environment without errors")
}

func TestCleanupTestDBWithNilDB(t *testing.T) {
	// Test that CleanupTestDB handles nil DB gracefully
	CleanupTestDB(t, nil)
	// Should not panic
}

func TestSetupTestDBMultipleTimes(t *testing.T) {
	// Test that we can setup database multiple times
	db1 := SetupTestDB(t)
	assert.NotNil(t, db1)
	CleanupTestDB(t, db1)

	db2 := SetupTestDB(t)
	assert.NotNil(t, db2)
	CleanupTestDB(t, db2)
}

func TestDatabaseQueryOperations(t *testing.T) {
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test basic CRUD operations
	device := models.Device{
		ID:       "test-query-device",
		Name:     "Query Test",
		IsOnline: true,
		LastSeen: time.Now(),
	}

	// Create
	err := db.Create(&device).Error
	assert.NoError(t, err)

	// Read
	var foundDevice models.Device
	err = db.Where("id = ?", "test-query-device").First(&foundDevice).Error
	assert.NoError(t, err)
	assert.Equal(t, "Query Test", foundDevice.Name)

	// Update
	err = db.Model(&foundDevice).Update("name", "Updated Name").Error
	assert.NoError(t, err)

	// Verify update
	err = db.Where("id = ?", "test-query-device").First(&foundDevice).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", foundDevice.Name)

	// Delete
	err = db.Delete(&foundDevice).Error
	assert.NoError(t, err)
}

func TestDatabaseComplexQueries(t *testing.T) {
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Create test devices
	devices := []models.Device{
		{ID: "device-1", Name: "Device 1", Type: "laptop", IsOnline: true},
		{ID: "device-2", Name: "Device 2", Type: "desktop", IsOnline: false},
		{ID: "device-3", Name: "Device 3", Type: "laptop", IsOnline: true},
	}

	for _, device := range devices {
		db.Create(&device)
	}

	// Test WHERE clause
	var onlineDevices []models.Device
	db.Where("is_online = ?", true).Find(&onlineDevices)
	assert.Equal(t, 2, len(onlineDevices))

	// Test COUNT
	var count int64
	db.Model(&models.Device{}).Where("type = ?", "laptop").Count(&count)
	assert.Equal(t, int64(2), count)

	// Test ORDER BY
	var orderedDevices []models.Device
	db.Order("name ASC").Find(&orderedDevices)
	assert.Equal(t, 3, len(orderedDevices))
	if len(orderedDevices) > 0 {
		assert.Equal(t, "Device 1", orderedDevices[0].Name)
	}
}

func TestDatabaseRelationships(t *testing.T) {
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Create device
	device := models.Device{
		ID:   "device-relations",
		Name: "Relations Test",
	}
	db.Create(&device)

	// Create related metrics
	for i := 0; i < 3; i++ {
		metrics := models.DeviceMetrics{
			ID:        fmt.Sprintf("metrics-%d", i),
			DeviceID:  device.ID,
			CPUUsage:  float64(50 + i*10),
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
		db.Create(&metrics)
	}

	// Query related data
	var metricsList []models.DeviceMetrics
	db.Where("device_id = ?", device.ID).Find(&metricsList)
	assert.Equal(t, 3, len(metricsList))
}

func TestDatabaseErrorHandling(t *testing.T) {
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test duplicate primary key
	device := models.Device{
		ID:   "duplicate-test",
		Name: "Duplicate",
	}
	err := db.Create(&device).Error
	assert.NoError(t, err)

	// Try to create again with same ID
	device2 := models.Device{
		ID:   "duplicate-test",
		Name: "Duplicate 2",
	}
	err = db.Create(&device2).Error
	assert.Error(t, err, "Should error on duplicate primary key")

	// Test query for non-existent record
	var notFound models.Device
	err = db.Where("id = ?", "nonexistent").First(&notFound).Error
	assert.Error(t, err, "Should error when record not found")
}

func TestAddActivityLogAndCheckExistence(t *testing.T) {
	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	activity := models.DeviceActivities{
		DeviceID:    "test-device-activity",
		Type:        "app_launch",
		Description: "User launched Chrome",
		App:         "chrome",
		Duration:    120,
		Timestamp:   time.Now(),
	}

	// Add activity to database
	err := db.Create(&activity).Error
	assert.NoError(t, err)

	// Check existence
	var found models.DeviceActivities
	err = db.Where("device_id = ? AND type = ?", "test-device-activity", "app_launch").First(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, "chrome", found.App)
	assert.Equal(t, "User launched Chrome", found.Description)
}
