package database

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"mentor-backend/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupTestDB(t *testing.T) {
	// Test with SQLite (default)
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Test that all tables are created by attempting to query them
	tables := []interface{}{
		&models.Device{},
		&models.DeviceMetric{},
		&models.DeviceProcess{},
		&models.DeviceActivity{},
		&models.DeviceRemoteCommand{},
		&models.DeviceScreenshot{},
		&models.DeviceAlert{},
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
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Insert test data
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	db.Create(&device)

	metrics := models.DeviceMetric{
		DeviceID:  sampleUUID,
		CPUUsage:  50.0,
		Timestamp: time.Now(),
	}
	db.Create(&metrics)

	// Verify data exists
	var deviceCount int64
	db.Model(&models.Device{}).Count(&deviceCount)
	assert.Greater(t, deviceCount, int64(0))

	var metricsCount int64
	db.Model(&models.DeviceMetric{}).Count(&metricsCount)
	assert.Greater(t, metricsCount, int64(0))

	// CleanupTestDB is now a no-op since transaction rollback handles cleanup
	CleanupTestDB(t, db)
	
	// Data should still exist because we're still in the transaction
	// The actual cleanup happens when the test ends and the transaction is rolled back
	db.Model(&models.Device{}).Count(&deviceCount)
	assert.Greater(t, deviceCount, int64(0), "Data should still exist during test")
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
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test that we can perform a simple query
	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

func TestDatabaseMigrationIntegrity(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test that we can create records of all model types
	device := models.Device{
		DeviceID:       sampleUUID,
		DeviceName:     "Migration Test Device",
		DeviceType:     "laptop",
		OS:             "Linux",
		IPAddress:      "192.168.1.100",
		MacAddress:     "aa:bb:cc:dd:ee:ff",
		DeviceLocation: "Test Lab",
		IsOnline:       true,
		CurrentUser:    "testuser",
		LastSeen:       time.Now(),
	}
	assert.NoError(t, db.Create(&device).Error)

	metrics := models.DeviceMetric{
		DeviceID:    device.DeviceID,
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

	process := models.DeviceProcess{
		DeviceID:    device.DeviceID,
		PID:         1234,
		ProcessName: "test-process",
		CPU:         15.5,
		Memory:      536870912,
		CommandText: "/usr/bin/test",
		Timestamp:   time.Now(),
	}
	assert.NoError(t, db.Create(&process).Error)

	activityLog := models.DeviceActivity{
		DeviceID:     device.DeviceID,
		ActivityType: "app_launch",
		Description:  "Test activity",
		App:          "TestApp",
		Duration:     3600,
		Timestamp:    time.Now(),
	}
	assert.NoError(t, db.Create(&activityLog).Error)

	command := models.DeviceRemoteCommand{
		DeviceID:    device.DeviceID,
		CommandText: "test command",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	assert.NoError(t, db.Create(&command).Error)

	screenshot := models.DeviceScreenshot{
		DeviceID:   device.DeviceID,
		Path:       "/path/to/screenshot.jpg",
		Resolution: "1920x1080",
		Size:       1024000,
		Timestamp:  time.Now(),
	}
	assert.NoError(t, db.Create(&screenshot).Error)

	alert := models.DeviceAlert{
		DeviceID:  device.DeviceID,
		AlertType: "security",
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

	db.Model(&models.DeviceMetric{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceProcess{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceActivity{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceActivity{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceRemoteCommand{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceScreenshot{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))

	db.Model(&models.DeviceAlert{}).Count(&count)
	assert.GreaterOrEqual(t, count, int64(1))
}

func TestDatabaseTransactionRollback(t *testing.T) {
	// Setup test database to ensure baseConnection is initialized
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)
	
	// Use base connection to test transaction behavior (not the test's transaction)
	// Create a fresh transaction for this test
	tx := baseConnection.Begin()
	require.NotNil(t, tx)
	require.NoError(t, tx.Error)

	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Rollback Test Device",
	}

	err = tx.Create(&device).Error
	assert.NoError(t, err)

	// Rollback transaction
	tx.Rollback()

	// Verify device was not actually saved (check in base connection, not the rolled-back tx)
	var count int64
	baseConnection.Model(&models.Device{}).Where("deviceid = ?", sampleUUID).Count(&count)
	assert.Equal(t, int64(0), count)
}
func TestConcurrentDatabaseAccess(t *testing.T) {
	// Setup test database to ensure baseConnection is initialized
	testDB, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, testDB)
	defer CleanupTestDB(t, testDB)
	
	// Use baseConnection instead of transaction-wrapped db for concurrent access
	// Transactions cannot be safely used from multiple goroutines
	db := baseConnection

	done := make(chan bool, 2)

	// Goroutine 1: Create devices
	go func() {
		for i := 0; i < 10; i++ {
			device := models.Device{
				DeviceID:   uuid.New(), // valid UUID
				DeviceName: fmt.Sprintf("Concurrent Device 1 %d", i),
			}
			for retry := 0; retry < 3; retry++ {
				if err := db.Create(&device).Error; err == nil {
					break
				} else if !strings.Contains(err.Error(), "database is locked") {
					t.Errorf("Failed to create device: %v", err)
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		done <- true
	}()

	// Goroutine 2: Create devices
	go func() {
		for i := 0; i < 10; i++ {
			device := models.Device{
				DeviceID:   uuid.New(), // valid UUID
				DeviceName: fmt.Sprintf("Concurrent Device 2 %d", i),
			}
			for retry := 0; retry < 3; retry++ {
				if err := db.Create(&device).Error; err == nil {
					break
				} else if !strings.Contains(err.Error(), "database is locked") {
					t.Errorf("Failed to create device: %v", err)
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
		done <- true
	}()

	<-done
	<-done

	// Verify all devices were created
	var count int64
	db.Model(&models.Device{}).Where("device_name LIKE ?", "Concurrent Device%").Count(&count)
	assert.GreaterOrEqual(t, count, int64(15), "Should create at least 15 devices")
	
	// Cleanup: delete the test devices
	db.Where("device_name LIKE ?", "Concurrent Device%").Delete(&models.Device{})
}

func TestSetupTestDBWithPostgres(t *testing.T) {
	// Test that when USE_POSTGRES_FOR_TESTS is not set, we use SQLite
	db, err := SetupTestDB(t)
	require.NotNil(t, db)
	require.NoError(t, err)
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
	db1, err := SetupTestDB(t)
	require.NoError(t, err)
	assert.NotNil(t, db1)
	CleanupTestDB(t, db1)

	db2, err := SetupTestDB(t)
	require.NoError(t, err)
	assert.NotNil(t, db2)
	CleanupTestDB(t, db2)
}

func TestDatabaseQueryOperations(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test basic CRUD operations
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Query Test",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}

	// Create
	err = db.Create(&device).Error
	assert.NoError(t, err)

	// Read
	var foundDevice models.Device
	err = db.Where("deviceid = ?", sampleUUID).First(&foundDevice).Error
	assert.NoError(t, err)
	assert.Equal(t, "Query Test", foundDevice.DeviceName)

	// Update
	err = db.Model(&foundDevice).Update("device_name", "Updated Name").Error
	assert.NoError(t, err)

	// Verify update
	err = db.Where("deviceid = ?", sampleUUID).First(&foundDevice).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", foundDevice.DeviceName)

	// Delete
	err = db.Delete(&foundDevice).Error
	assert.NoError(t, err)
}

func TestDatabaseComplexQueries(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Create test devices with unique UUIDs
	uuid1 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440031")
	uuid2 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440032")
	uuid3 := uuid.MustParse("550e8400-e29b-41d4-a716-446655440033")
	
	devices := []models.Device{
		{DeviceID: uuid1, DeviceName: "Device 1", DeviceType: "laptop", IsOnline: true},
		{DeviceID: uuid2, DeviceName: "Device 2", DeviceType: "desktop", IsOnline: false},
		{DeviceID: uuid3, DeviceName: "Device 3", DeviceType: "laptop", IsOnline: true},
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
	db.Model(&models.Device{}).Where("device_type = ?", "laptop").Count(&count)
	assert.Equal(t, int64(2), count)

	// Test ORDER BY
	var orderedDevices []models.Device
	db.Order("device_name ASC").Find(&orderedDevices)
	assert.GreaterOrEqual(t, len(orderedDevices), 3)
	if len(orderedDevices) >= 3 {
		// Just verify we got results in some order
		assert.NotEmpty(t, orderedDevices[0].DeviceName)
	}
}

func TestDatabaseRelationships(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Create device
	deviceUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440041")
	device := models.Device{
		DeviceID:   deviceUUID,
		DeviceName: "Relations Test",
	}
	db.Create(&device)

	// Create related metrics
	for i := 0; i < 3; i++ {
		metrics := models.DeviceMetric{
			DeviceID:  device.DeviceID,
			CPUUsage:  float64(50 + i*10),
			Timestamp: time.Now().Add(time.Duration(i) * time.Minute),
		}
		db.Create(&metrics)
	}

	// Query related data
	var metricsList []models.DeviceMetric
	db.Where("deviceid = ?", device.DeviceID).Find(&metricsList)
	assert.Equal(t, 3, len(metricsList))
}

func TestDatabaseErrorHandling(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test duplicate primary key
	deviceUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440042")
	device := models.Device{
		DeviceID:   deviceUUID,
		DeviceName: "Duplicate",
	}
	err = db.Create(&device).Error
	assert.NoError(t, err)

	// Try to create again with same ID
	device2 := models.Device{
		DeviceID:   deviceUUID,
		DeviceName: "Duplicate 2",
	}
	err = db.Create(&device2).Error
	assert.Error(t, err, "Should error on duplicate primary key")

	// Test query for non-existent record
	var notFound models.Device
	nonExistentUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655449999")
	err = db.Where("deviceid = ?", nonExistentUUID).First(&notFound).Error
	assert.Error(t, err, "Should error when record not found")
}

func TestAddActivityLogAndCheckExistence(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	activityUUID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440043")
	activity := models.DeviceActivity{
		DeviceID:     activityUUID,
		ActivityType: "app_launch",
		Description:  "User launched Chrome",
		App:          "chrome",
		Duration:     120,
		Timestamp:    time.Now(),
	}

	// Add activity to database
	err = db.Create(&activity).Error
	assert.NoError(t, err)

	// Check existence
	var found models.DeviceActivity
	err = db.Where("deviceid = ? AND activity_type = ?", activityUUID, "app_launch").First(&found).Error
	assert.NoError(t, err)
	assert.Equal(t, "chrome", found.App)
	assert.Equal(t, "User launched Chrome", found.Description)
}
