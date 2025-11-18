package database

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetEnvInt(t *testing.T) {
	t.Run("returns value from environment", func(t *testing.T) {
		err := os.Setenv("TEST_INT_VAR", "42")
		require.NoError(t, err)
		defer func() { _ = os.Unsetenv("TEST_INT_VAR") }()

		result := getEnvInt("TEST_INT_VAR", 10)
		assert.Equal(t, 42, result)
	})

	t.Run("returns default when env var not set", func(t *testing.T) {
		err := os.Unsetenv("NON_EXISTENT_VAR")
		require.NoError(t, err)

		result := getEnvInt("NON_EXISTENT_VAR", 100)
		assert.Equal(t, 100, result)
	})

	t.Run("returns default when env var is not a valid int", func(t *testing.T) {
		err := os.Setenv("INVALID_INT_VAR", "not_a_number")
		require.NoError(t, err)
		defer func() { _ = os.Unsetenv("INVALID_INT_VAR") }()

		result := getEnvInt("INVALID_INT_VAR", 50)
		assert.Equal(t, 50, result)
	})

	t.Run("returns default when env var is empty", func(t *testing.T) {
		err := os.Setenv("EMPTY_INT_VAR", "")
		require.NoError(t, err)
		defer func() { _ = os.Unsetenv("EMPTY_INT_VAR") }()

		result := getEnvInt("EMPTY_INT_VAR", 75)
		assert.Equal(t, 75, result)
	})

	t.Run("handles negative numbers", func(t *testing.T) {
		err := os.Setenv("NEGATIVE_INT_VAR", "-10")
		require.NoError(t, err)
		defer func() { _ = os.Unsetenv("NEGATIVE_INT_VAR") }()

		result := getEnvInt("NEGATIVE_INT_VAR", 0)
		assert.Equal(t, -10, result)
	})

	t.Run("handles zero value", func(t *testing.T) {
		err := os.Setenv("ZERO_INT_VAR", "0")
		require.NoError(t, err)
		defer func() { _ = os.Unsetenv("ZERO_INT_VAR") }()

		result := getEnvInt("ZERO_INT_VAR", 100)
		assert.Equal(t, 0, result)
	})
}

func TestConnectWithRetry(t *testing.T) {
	t.Run("fails immediately with invalid config", func(t *testing.T) {
		// Set invalid environment variables
		_ = os.Setenv("POSTGRES_USER", "invalid_user")
		_ = os.Setenv("POSTGRES_PASSWORD", "invalid_password")
		_ = os.Setenv("POSTGRES_DB", "invalid_db")
		_ = os.Setenv("POSTGRES_HOST", "invalid_host")
		_ = os.Setenv("POSTGRES_PORT", "99999")
		defer func() {
			_ = os.Unsetenv("POSTGRES_USER")
			_ = os.Unsetenv("POSTGRES_PASSWORD")
			_ = os.Unsetenv("POSTGRES_DB")
			_ = os.Unsetenv("POSTGRES_HOST")
			_ = os.Unsetenv("POSTGRES_PORT")
		}()

		err := connectWithRetry(1, 10*time.Millisecond)
		assert.Error(t, err, "Should fail to connect with invalid configuration")
		assert.Contains(t, err.Error(), "failed to connect after")
	})

	t.Run("respects max retries", func(t *testing.T) {
		// Set invalid environment variables to ensure connection fails
		_ = os.Setenv("POSTGRES_USER", "test_user")
		_ = os.Setenv("POSTGRES_PASSWORD", "test_password")
		_ = os.Setenv("POSTGRES_DB", "test_db")
		_ = os.Setenv("POSTGRES_HOST", "nonexistent_host_12345")
		_ = os.Setenv("POSTGRES_PORT", "5432")
		defer func() {
			_ = os.Unsetenv("POSTGRES_USER")
			_ = os.Unsetenv("POSTGRES_PASSWORD")
			_ = os.Unsetenv("POSTGRES_DB")
			_ = os.Unsetenv("POSTGRES_HOST")
			_ = os.Unsetenv("POSTGRES_PORT")
		}()

		maxRetries := 3
		start := time.Now()
		err := connectWithRetry(maxRetries, 50*time.Millisecond)
		elapsed := time.Since(start)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to connect after 3 attempts")
		// Should have waited at least 50ms + 100ms = 150ms for retries
		assert.GreaterOrEqual(t, elapsed, 150*time.Millisecond)
	})
}

func TestHealthCheck(t *testing.T) {
	t.Run("returns error when DB is nil", func(t *testing.T) {
		// Save current DB and set to nil
		oldDB := DB
		DB = nil
		defer func() { DB = oldDB }()

		err := HealthCheck()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "database connection not initialized")
	})

	t.Run("returns nil for healthy database", func(t *testing.T) {
		db, err := SetupTestDB(t)
		require.NoError(t, err)
		defer CleanupTestDB(t, db)

		// Set global DB for health check
		oldDB := DB
		DB = db
		defer func() { DB = oldDB }()

		err = HealthCheck()
		assert.NoError(t, err, "HealthCheck should succeed for valid database")
	})
}

func TestShutdown(t *testing.T) {
	t.Run("returns nil when DB is nil", func(t *testing.T) {
		oldDB := DB
		DB = nil
		defer func() { DB = oldDB }()

		err := Shutdown()
		assert.NoError(t, err, "Shutdown should handle nil DB gracefully")
	})

	t.Run("closes database connection successfully", func(t *testing.T) {
		db, err := SetupTestDB(t)
		require.NoError(t, err)

		// Set global DB temporarily
		oldDB := DB
		DB = db
		defer func() { DB = oldDB }()

		// Test shutdown
		err = Shutdown()
		assert.NoError(t, err, "Shutdown should close database successfully")
	})
}

func TestMigrate(t *testing.T) {
	t.Run("successfully migrates all models", func(t *testing.T) {
		db, err := SetupTestDB(t)
		require.NoError(t, err)
		defer CleanupTestDB(t, db)

		// migrate function is called internally by SetupTestDB
		// So we just verify tables exist
		err = migrate(db)
		assert.NoError(t, err, "Migration should complete without errors")
	})
}

func TestValidateEnvVars(t *testing.T) {
	t.Run("returns error when required vars are missing", func(t *testing.T) {
		// Clear all required env vars
		requiredVars := []string{
			"POSTGRES_USER",
			"POSTGRES_PASSWORD",
			"POSTGRES_DB",
			"POSTGRES_HOST",
			"POSTGRES_PORT",
		}

		// Save and clear all
		savedVals := make(map[string]string)
		for _, key := range requiredVars {
			savedVals[key] = os.Getenv(key)
			_ = os.Unsetenv(key)
		}

		// Restore after test
		defer func() {
			for key, val := range savedVals {
				if val != "" {
					_ = os.Setenv(key, val)
				}
			}
		}()

		err := validateEnvVars()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required database configuration")
	})

	t.Run("succeeds when all required vars are set", func(t *testing.T) {
		requiredVars := map[string]string{
			"POSTGRES_USER":     "test_user",
			"POSTGRES_PASSWORD": "test_password",
			"POSTGRES_DB":       "test_db",
			"POSTGRES_HOST":     "localhost",
			"POSTGRES_PORT":     "5432",
		}

		// Save current values
		savedVals := make(map[string]string)
		for key := range requiredVars {
			savedVals[key] = os.Getenv(key)
		}

		// Set test values
		for key, val := range requiredVars {
			_ = os.Setenv(key, val)
		}

		// Restore after test
		defer func() {
			for key, val := range savedVals {
				if val != "" {
					_ = os.Setenv(key, val)
				} else {
					_ = os.Unsetenv(key)
				}
			}
		}()

		err := validateEnvVars()
		assert.NoError(t, err)
	})

	t.Run("fails when only some vars are set", func(t *testing.T) {
		// Clear all first
		requiredVars := []string{
			"POSTGRES_USER",
			"POSTGRES_PASSWORD",
			"POSTGRES_DB",
			"POSTGRES_HOST",
			"POSTGRES_PORT",
		}

		savedVals := make(map[string]string)
		for _, key := range requiredVars {
			savedVals[key] = os.Getenv(key)
			_ = os.Unsetenv(key)
		}

		// Set only some
		_ = os.Setenv("POSTGRES_USER", "test_user")
		_ = os.Setenv("POSTGRES_PASSWORD", "test_password")
		// Missing: POSTGRES_DB, POSTGRES_HOST, POSTGRES_PORT

		// Restore after test
		defer func() {
			for key, val := range savedVals {
				if val != "" {
					_ = os.Setenv(key, val)
				}
			}
		}()

		err := validateEnvVars()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required database configuration")
	})
}
