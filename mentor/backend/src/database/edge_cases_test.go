package database

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupTestDBMultipleCalls tests calling SetupTestDB multiple times
func TestSetupTestDBMultipleCalls(t *testing.T) {
	// Ensure we use SQLite
	_ = os.Unsetenv("USE_POSTGRES_FOR_TESTS")

	db1 := SetupTestDB(t)
	require.NotNil(t, db1)

	db2 := SetupTestDB(t)
	require.NotNil(t, db2)

	// Both should work independently
	var count1, count2 int64
	err := db1.Raw("SELECT 1").Scan(&count1).Error
	assert.NoError(t, err)

	err = db2.Raw("SELECT 1").Scan(&count2).Error
	assert.NoError(t, err)

	CleanupTestDB(t, db1)
	CleanupTestDB(t, db2)
}

// TestSetupTestDBAutoMigrationSuccess tests successful auto-migration
func TestSetupTestDBAutoMigrationSuccess(t *testing.T) {
	_ = os.Unsetenv("USE_POSTGRES_FOR_TESTS")

	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Verify all tables were created by checking we can query them
	tables := []string{
		"devices",
		"device_metrics",
		"processes",
		"activities",
		"activity_logs",
		"remote_commands",
		"screenshots",
		"alerts",
	}

	for _, table := range tables {
		var count int64
		// This query will fail if table doesn't exist
		err := db.Table(table).Count(&count).Error
		assert.NoError(t, err, "Table %s should exist", table)
	}
}

// TestSetupTestDBSQLitePragmas tests SQLite pragma settings
func TestSetupTestDBSQLitePragmas(t *testing.T) {
	_ = os.Unsetenv("USE_POSTGRES_FOR_TESTS")

	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Check busy timeout is set
	var busyTimeout int
	err := db.Raw("PRAGMA busy_timeout").Scan(&busyTimeout).Error
	assert.NoError(t, err)
	// Should be greater than 0 (we set it to 5000ms)
	assert.Greater(t, busyTimeout, 0)
}

// TestSetupTestDBCacheSharing tests cache sharing mode
func TestSetupTestDBCacheSharing(t *testing.T) {
	_ = os.Unsetenv("USE_POSTGRES_FOR_TESTS")

	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Verify we can get the underlying SQL DB
	sqlDB, err := db.DB()
	assert.NoError(t, err)
	assert.NotNil(t, sqlDB)

	// Check that the connection is alive
	err = sqlDB.Ping()
	assert.NoError(t, err)
}

// TestSetupTestDBConcurrentAccess tests concurrent access to test database
func TestSetupTestDBConcurrentAccess(t *testing.T) {
	_ = os.Unsetenv("USE_POSTGRES_FOR_TESTS")

	db := SetupTestDB(t)
	require.NotNil(t, db)
	defer CleanupTestDB(t, db)

	// Test concurrent reads and writes
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func(id int) {
			// Try to query
			var result int
			err := db.Raw("SELECT ?", id).Scan(&result).Error
			assert.NoError(t, err)
			assert.Equal(t, id, result)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
}

// TestSetupTestDBWithEmptyEnvironment tests SetupTestDB with no environment variables
func TestSetupTestDBWithEmptyEnvironment(t *testing.T) {
	// Clear all PostgreSQL environment variables
	vars := []string{
		"USE_POSTGRES_FOR_TESTS",
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_DB",
		"POSTGRES_TEST_DB",
	}

	originalVars := make(map[string]string)
	for _, v := range vars {
		originalVars[v] = os.Getenv(v)
		os.Unsetenv(v)
	}

	defer func() {
		for k, v := range originalVars {
			if v != "" {
				_ = os.Setenv(k, v)
			}
		}
	}()

	// Should still work with SQLite
	db := SetupTestDB(t)
	require.NotNil(t, db)

	// Verify it works
	var result int
	err := db.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)

	CleanupTestDB(t, db)
}

// TestCleanupTestDBIdempotent tests calling CleanupTestDB multiple times
func TestCleanupTestDBIdempotent(t *testing.T) {
	os.Unsetenv("USE_POSTGRES_FOR_TESTS")

	db := SetupTestDB(t)
	require.NotNil(t, db)

	// Cleanup multiple times should not cause errors
	CleanupTestDB(t, db)
	CleanupTestDB(t, db)
	CleanupTestDB(t, db)
}

// TestCleanupTestDBWithEmptyTables tests cleanup when tables are already empty
func TestCleanupTestDBWithEmptyTables(t *testing.T) {
	os.Unsetenv("USE_POSTGRES_FOR_TESTS")

	db := SetupTestDB(t)
	require.NotNil(t, db)

	// Tables are already empty after SetupTestDB
	// Cleanup should work fine
	CleanupTestDB(t, db)

	// Verify tables still exist and are empty
	var count int64
	err := db.Table("devices").Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// TestGetEnvOrDefaultWithEmptyString tests getEnvOrDefault with empty string
func TestGetEnvOrDefaultWithEmptyString(t *testing.T) {
	testKey := "TEST_EMPTY_ENV_VAR"

	// Set to empty string
	_ = os.Setenv(testKey, "")
	defer func() {
		_ = os.Unsetenv(testKey)
	}()

	result := getEnvOrDefault(testKey, "default")
	// Empty string should return default
	assert.Equal(t, "default", result)
}

// TestGetEnvOrDefaultWithWhitespace tests getEnvOrDefault with whitespace
func TestGetEnvOrDefaultWithWhitespace(t *testing.T) {
	testKey := "TEST_WHITESPACE_ENV_VAR"

	// Set to whitespace
	_ = os.Setenv(testKey, "   ")
	defer func() {
		_ = os.Unsetenv(testKey)
	}()

	result := getEnvOrDefault(testKey, "default")
	// Whitespace is considered a value
	assert.Equal(t, "   ", result)
}
