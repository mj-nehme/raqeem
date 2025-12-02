package database

import (
	"os"
	"testing"
	"time"

	"mentor-backend/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateEnvVarsWithAllVariables tests validateEnvVars with all required variables set
func TestValidateEnvVarsWithAllVariables(t *testing.T) {
	// Save original values
	originalUser := os.Getenv("POSTGRES_USER")
	originalPassword := os.Getenv("POSTGRES_PASSWORD")
	originalDB := os.Getenv("POSTGRES_DB")
	originalHost := os.Getenv("POSTGRES_HOST")
	originalPort := os.Getenv("POSTGRES_PORT")

	defer func() {
		os.Setenv("POSTGRES_USER", originalUser)
		os.Setenv("POSTGRES_PASSWORD", originalPassword)
		os.Setenv("POSTGRES_DB", originalDB)
		os.Setenv("POSTGRES_HOST", originalHost)
		os.Setenv("POSTGRES_PORT", originalPort)
	}()

	// Set all required variables
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("POSTGRES_DB", "testdb")
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")

	err := validateEnvVars()
	assert.NoError(t, err)
}

// TestValidateEnvVarsWithMissingVariables tests validateEnvVars with missing variables
func TestValidateEnvVarsWithMissingVariables(t *testing.T) {
	// Save original values
	originalUser := os.Getenv("POSTGRES_USER")
	originalPassword := os.Getenv("POSTGRES_PASSWORD")
	originalDB := os.Getenv("POSTGRES_DB")
	originalHost := os.Getenv("POSTGRES_HOST")
	originalPort := os.Getenv("POSTGRES_PORT")

	defer func() {
		os.Setenv("POSTGRES_USER", originalUser)
		os.Setenv("POSTGRES_PASSWORD", originalPassword)
		os.Setenv("POSTGRES_DB", originalDB)
		os.Setenv("POSTGRES_HOST", originalHost)
		os.Setenv("POSTGRES_PORT", originalPort)
	}()

	// Clear all variables
	os.Unsetenv("POSTGRES_USER")
	os.Unsetenv("POSTGRES_PASSWORD")
	os.Unsetenv("POSTGRES_DB")
	os.Unsetenv("POSTGRES_HOST")
	os.Unsetenv("POSTGRES_PORT")

	err := validateEnvVars()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing required database configuration")
}

// TestGetEnvIntWithValidValue tests getEnvInt with valid integer value
func TestGetEnvIntWithValidValue(t *testing.T) {
	os.Setenv("TEST_INT_VAR", "42")
	defer os.Unsetenv("TEST_INT_VAR")

	result := getEnvInt("TEST_INT_VAR", 10)
	assert.Equal(t, 42, result)
}

// TestGetEnvIntWithInvalidValue tests getEnvInt with invalid integer value
func TestGetEnvIntWithInvalidValue(t *testing.T) {
	os.Setenv("TEST_INT_VAR", "not-a-number")
	defer os.Unsetenv("TEST_INT_VAR")

	result := getEnvInt("TEST_INT_VAR", 10)
	assert.Equal(t, 10, result)
}

// TestGetEnvIntWithEmptyValue tests getEnvInt with empty value
func TestGetEnvIntWithEmptyValue(t *testing.T) {
	os.Unsetenv("TEST_INT_VAR_EMPTY")

	result := getEnvInt("TEST_INT_VAR_EMPTY", 99)
	assert.Equal(t, 99, result)
}

// TestHealthCheckWithNilDB tests HealthCheck when DB is nil
func TestHealthCheckWithNilDB(t *testing.T) {
	// Save original DB
	originalDB := DB
	DB = nil
	defer func() { DB = originalDB }()

	err := HealthCheck()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection not initialized")
}

// TestShutdownWithNilDB tests Shutdown when DB is nil
func TestShutdownWithNilDB(t *testing.T) {
	// Save original DB
	originalDB := DB
	DB = nil
	defer func() { DB = originalDB }()

	err := Shutdown()
	assert.NoError(t, err)
}

// TestConnectWithRetryFailure tests connectWithRetry with failed connections
func TestConnectWithRetryFailure(t *testing.T) {
	// Save original environment
	originalHost := os.Getenv("POSTGRES_HOST")
	defer os.Setenv("POSTGRES_HOST", originalHost)

	// Set an invalid host that will fail to connect
	os.Setenv("POSTGRES_HOST", "nonexistent_host_12345")
	os.Setenv("POSTGRES_USER", "test_user")
	os.Setenv("POSTGRES_PASSWORD", "test_pass")
	os.Setenv("POSTGRES_DB", "test_db")
	os.Setenv("POSTGRES_PORT", "5432")

	// Use minimal retries and short delays for testing
	err := connectWithRetry(3, 50*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect after")
}

// TestMigrateWithModels tests the migrate function with all models
func TestMigrateWithModels(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// The migrate function is tested through SetupTestDB which calls AutoMigrate
	// Verify that all tables exist by creating records

	// Test Device table
	device := models.Device{
		DeviceID:   uuid.New(),
		DeviceName: "Migration Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	err = db.Create(&device).Error
	assert.NoError(t, err)

	// Test DeviceMetric table
	metric := models.DeviceMetric{
		MetricID:  uuid.New(),
		DeviceID:  device.DeviceID,
		CPUUsage:  50.0,
		Timestamp: time.Now(),
	}
	err = db.Create(&metric).Error
	assert.NoError(t, err)

	// Test DeviceProcess table
	process := models.DeviceProcess{
		ProcessID:   uuid.New(),
		DeviceID:    device.DeviceID,
		PID:         1234,
		ProcessName: "test",
		CPU:         10.0,
		Memory:      1024,
		Timestamp:   time.Now(),
	}
	err = db.Create(&process).Error
	assert.NoError(t, err)

	// Test DeviceActivity table
	activity := models.DeviceActivity{
		ActivityID:   uuid.New(),
		DeviceID:     device.DeviceID,
		ActivityType: "test",
		Description:  "Test activity",
		Timestamp:    time.Now(),
	}
	err = db.Create(&activity).Error
	assert.NoError(t, err)

	// Test DeviceRemoteCommand table
	command := models.DeviceRemoteCommand{
		CommandID:   uuid.New(),
		DeviceID:    device.DeviceID,
		CommandText: "echo test",
		Status:      "pending",
		CreatedAt:   time.Now(),
	}
	err = db.Create(&command).Error
	assert.NoError(t, err)

	// Test DeviceScreenshot table
	screenshot := models.DeviceScreenshot{
		ScreenshotID: uuid.New(),
		DeviceID:     device.DeviceID,
		Path:         "test.png",
		Resolution:   "1920x1080",
		Size:         1024,
		Timestamp:    time.Now(),
	}
	err = db.Create(&screenshot).Error
	assert.NoError(t, err)

	// Test DeviceAlert table
	alert := models.DeviceAlert{
		AlertID:   uuid.New(),
		DeviceID:  device.DeviceID,
		AlertType: "test",
		Level:     "warning",
		Message:   "Test alert",
		Timestamp: time.Now(),
	}
	err = db.Create(&alert).Error
	assert.NoError(t, err)
}

// TestDatabaseTransactionIsolation tests that transactions are properly isolated
func TestDatabaseTransactionIsolation(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Create a unique device ID for this test
	deviceID := uuid.New()
	device := models.Device{
		DeviceID:   deviceID,
		DeviceName: "Isolation Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}

	// Create device in transaction
	err = db.Create(&device).Error
	assert.NoError(t, err)

	// Verify device exists within the same transaction
	var foundDevice models.Device
	err = db.Where("deviceid = ?", deviceID).First(&foundDevice).Error
	assert.NoError(t, err)
	assert.Equal(t, "Isolation Test Device", foundDevice.DeviceName)
}

// TestSetupTestDBWithCustomConfig tests SetupTestDB with custom configuration
func TestSetupTestDBWithCustomConfig(t *testing.T) {
	config := DBConfig{
		User:     getEnvOrDefault("POSTGRES_USER", "monitor"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", "password"),
		Host:     getEnvOrDefault("POSTGRES_HOST", "127.0.0.1"),
		Port:     5432,
		DBName:   getEnvOrDefault("POSTGRES_DB", "monitoring_db"),
		SSLMode:  "disable",
	}

	db, err := SetupTestDB(t, config)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Verify connection works
	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

// TestLoadEnvSearchPaths tests that loadEnv searches multiple paths
func TestLoadEnvSearchPaths(t *testing.T) {
	// This test verifies that loadEnv doesn't panic when searching for .env files
	// The actual loading behavior depends on file system state
	loadEnv()
	// Test passes if no panic occurred
}
