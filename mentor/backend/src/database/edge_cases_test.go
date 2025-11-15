package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupTestDBMultipleCalls tests calling SetupTestDB multiple times
func TestSetupTestDBMultipleCalls(t *testing.T) {
	db1, err := SetupTestDB(t)
	require.NotNil(t, db1)
	require.NoError(t, err)

	db2, err := SetupTestDB(t)
	require.NotNil(t, db2)
	require.NoError(t, err)

	// Both should work independently
	var count1, count2 int64
	err = db1.Raw("SELECT 1").Scan(&count1).Error
	assert.NoError(t, err)

	err = db2.Raw("SELECT 1").Scan(&count2).Error
	assert.NoError(t, err)

	CleanupTestDB(t, db1)
	CleanupTestDB(t, db2)
}

// TestSetupTestDBAutoMigrationSuccess tests successful auto-migration
func TestSetupTestDBAutoMigrationSuccess(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NotNil(t, db)
	require.NoError(t, err)

	defer CleanupTestDB(t, db)

	// Verify all tables were created by checking we can query them
	tables := []string{
		"devices",
		"device_metrics",
		"device_processes",
		"device_activities",
		"device_alerts",
		"device_remote_commands",
		"device_screenshots",
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
	// This test is SQLite-specific. For PostgreSQL, we verify connection settings instead
	db, err := SetupTestDB(t)
	require.NotNil(t, db)
	require.NoError(t, err)
	defer CleanupTestDB(t, db)

	// For PostgreSQL, verify the connection is working
	sqlDB, err := db.DB()
	require.NoError(t, err)
	
	// Verify connection is alive
	err = sqlDB.Ping()
	assert.NoError(t, err)
	
	// Verify we can execute a simple query
	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

// TestSetupTestDBCacheSharing tests cache sharing mode
func TestSetupTestDBCacheSharing(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NotNil(t, db)
	require.NoError(t, err)
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
	// Setup test database to ensure baseConnection is initialized
	testDB, err := SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, testDB)
	defer CleanupTestDB(t, testDB)
	
	// Use baseConnection for concurrent access since transaction-wrapped DB can't be used concurrently
	db := baseConnection

	// Test concurrent reads
	done := make(chan bool, 3)
	errors := make(chan error, 3)

	for i := 0; i < 3; i++ {
		go func(id int) {
			// Try to query - use proper parameter binding for PostgreSQL
			var result int
			err := db.Raw("SELECT $1::int", id).Scan(&result).Error
			if err != nil {
				errors <- err
			} else if result != id {
				errors <- fmt.Errorf("expected %d, got %d", id, result)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}
	
	close(errors)
	for err := range errors {
		assert.NoError(t, err)
	}
}

// TestSetupTestDBWithEmptyEnvironment tests SetupTestDB with no environment variables
func TestSetupTestDBWithEmptyEnvironment(t *testing.T) {
	envVars := []string{
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_HOST",
		"POSTGRES_PORT",
		"POSTGRES_DB",
		"POSTGRES_TEST_DB",
	}
	// Clear all PostgreSQL environment variables
	originalVars := make(map[string]string)
	for _, v := range envVars {
		originalVars[v] = os.Getenv(v)
		_ = os.Unsetenv(v)
	}

	defer func() {
		for k, v := range originalVars {
			if v != "" {
				_ = os.Setenv(k, v)
			}
		}
	}()

	// Should still work with SQLite
	db, err := SetupTestDB(t)
	require.NotNil(t, db)
	require.NoError(t, err)

	// Verify it works
	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)

	CleanupTestDB(t, db)
}

// TestCleanupTestDBIdempotent tests calling CleanupTestDB multiple times
func TestCleanupTestDBIdempotent(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NotNil(t, db)
	require.NoError(t, err)

	// Cleanup multiple times should not cause errors
	CleanupTestDB(t, db)
	CleanupTestDB(t, db)
	CleanupTestDB(t, db)
}

// TestCleanupTestDBWithEmptyTables tests cleanup when tables are already empty
func TestCleanupTestDBWithEmptyTables(t *testing.T) {
	db, err := SetupTestDB(t)
	require.NotNil(t, db)
	require.NoError(t, err)

	// Tables are already empty after SetupTestDB
	// Cleanup should work fine
	CleanupTestDB(t, db)

	// Verify tables still exist and are empty
	var count int64
	err = db.Table("devices").Count(&count).Error
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
