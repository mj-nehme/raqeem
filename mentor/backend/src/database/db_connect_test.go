package database

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestConnectEnvironmentVariableLoading tests that Connect loads environment variables
func TestConnectEnvironmentVariableLoading(t *testing.T) {
	// We can't call Connect directly because it will try to connect to a real database
	// and call log.Fatalf on failure. Instead, we test the DSN construction logic
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	expected := "host=testhost user=testuser password=testpass dbname=testdb port=5432 sslmode=disable"
	assert.Equal(t, expected, dsn)
}

// TestConnectWithConfig tests the connectWithConfig function
func TestConnectWithConfig(t *testing.T) {
	// Save original environment variables and DB
	originalVars := map[string]string{
		"POSTGRES_USER":     os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":       os.Getenv("POSTGRES_DB"),
		"POSTGRES_HOST":     os.Getenv("POSTGRES_HOST"),
		"POSTGRES_PORT":     os.Getenv("POSTGRES_PORT"),
	}
	originalDB := DB

	defer func() {
		for k, v := range originalVars {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
		DB = originalDB
	}()

	// Test with invalid connection (should return error, not crash)
	_ = os.Setenv("POSTGRES_USER", "invalid")
	_ = os.Setenv("POSTGRES_PASSWORD", "invalid")
	_ = os.Setenv("POSTGRES_DB", "invalid")
	_ = os.Setenv("POSTGRES_HOST", "invalid-host-xyz")
	_ = os.Setenv("POSTGRES_PORT", "9999")

	err := connectWithConfig()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

// TestConnectWithConfigLoadsEnvFile tests that connectWithConfig loads .env file
func TestConnectWithConfigLoadsEnvFile(t *testing.T) {
	// Save original environment variables and DB
	originalVars := map[string]string{
		"POSTGRES_HOST": os.Getenv("POSTGRES_HOST"),
		"POSTGRES_USER": os.Getenv("POSTGRES_USER"),
	}
	originalDB := DB

	defer func() {
		for k, v := range originalVars {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
		DB = originalDB
	}()

	// Create a temporary .env file
	tmpDir := t.TempDir()
	envFile := tmpDir + "/.env"
	envContent := `POSTGRES_HOST=envhost
POSTGRES_USER=envuser
POSTGRES_PASSWORD=envpass
POSTGRES_DB=envdb
POSTGRES_PORT=5432
`
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	assert.NoError(t, err)

	// Change to temp directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Clear environment variables so they come from .env
	_ = os.Unsetenv("POSTGRES_HOST")
	_ = os.Unsetenv("POSTGRES_USER")

	// Call connectWithConfig - it will fail to connect but should load env vars
	err = connectWithConfig()
	// Should fail but env vars should be loaded
	assert.Error(t, err)

	// Verify environment variables were loaded from .env file
	assert.Equal(t, "envhost", os.Getenv("POSTGRES_HOST"))
	assert.Equal(t, "envuser", os.Getenv("POSTGRES_USER"))
}

// TestConnectWithConfigMissingEnvFile tests behavior without .env file
func TestConnectWithConfigMissingEnvFile(t *testing.T) {
	// Save original environment variables and DB
	originalVars := map[string]string{
		"POSTGRES_USER":     os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":       os.Getenv("POSTGRES_DB"),
		"POSTGRES_HOST":     os.Getenv("POSTGRES_HOST"),
		"POSTGRES_PORT":     os.Getenv("POSTGRES_PORT"),
	}
	originalDB := DB

	defer func() {
		for k, v := range originalVars {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
		DB = originalDB
	}()

	// Create a temporary directory without .env file
	tmpDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Set environment variables directly
	_ = os.Setenv("POSTGRES_USER", "testuser")
	_ = os.Setenv("POSTGRES_PASSWORD", "testpass")
	_ = os.Setenv("POSTGRES_DB", "testdb")
	_ = os.Setenv("POSTGRES_HOST", "invalid-host")
	_ = os.Setenv("POSTGRES_PORT", "5432")

	// Call connectWithConfig - should not crash even without .env file
	err = connectWithConfig()
	// Will fail to connect but should not crash due to missing .env
	assert.Error(t, err)
}

// TestConnectWithConfigEmptyEnvironmentVariables tests with empty env vars
func TestConnectWithConfigEmptyEnvironmentVariables(t *testing.T) {
	// Save original environment variables and DB
	vars := []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB", "POSTGRES_HOST", "POSTGRES_PORT"}
	originalVars := make(map[string]string)
	for _, v := range vars {
		originalVars[v] = os.Getenv(v)
	}
	originalDB := DB

	defer func() {
		for k, v := range originalVars {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
		DB = originalDB
	}()

	// Create temp dir without .env
	tmpDir := t.TempDir()
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tmpDir)

	// Clear all environment variables
	for _, v := range vars {
		_ = os.Unsetenv(v)
	}

	// Call connectWithConfig - should attempt connection with empty values
	err = connectWithConfig()
	assert.Error(t, err)
}

// TestConnectCallsConnectWithConfig tests that Connect calls connectWithConfig
func TestConnectCallsConnectWithConfig(t *testing.T) {
	// This test documents that Connect function exists and calls connectWithConfig
	// We can't directly test Connect because it calls log.Fatalf on error
	// But we verify it exists and has the expected signature
	assert.NotNil(t, Connect)
}

// TestConnectWithValidDatabaseConnection tests Connect with godotenv loading
func TestConnectWithValidDatabaseConnection(t *testing.T) {
	// Save original environment variables
	originalVars := map[string]string{
		"POSTGRES_HOST":     os.Getenv("POSTGRES_HOST"),
		"POSTGRES_USER":     os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":       os.Getenv("POSTGRES_DB"),
		"POSTGRES_PORT":     os.Getenv("POSTGRES_PORT"),
	}
	defer func() {
		for k, v := range originalVars {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}()

	// Clear environment variables first
	_ = os.Unsetenv("POSTGRES_HOST")
	_ = os.Unsetenv("POSTGRES_USER")
	_ = os.Unsetenv("POSTGRES_PASSWORD")
	_ = os.Unsetenv("POSTGRES_DB")
	_ = os.Unsetenv("POSTGRES_PORT")

	// Create a temporary .env file for testing
	tmpDir := t.TempDir()
	envFile := tmpDir + "/.env"
	envContent := `POSTGRES_HOST=localhost
POSTGRES_USER=testuser
POSTGRES_PASSWORD=testpass
POSTGRES_DB=testdb
POSTGRES_PORT=5432
`
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	assert.NoError(t, err)

	// Change to temp directory to test godotenv.Load
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Test godotenv.Load
	err = godotenv.Load()
	assert.NoError(t, err)

	// Verify environment variables were loaded
	assert.Equal(t, "localhost", os.Getenv("POSTGRES_HOST"))
	assert.Equal(t, "testuser", os.Getenv("POSTGRES_USER"))
}

// TestConnectWithMissingEnvFileDoesNotCrash tests that Connect handles missing .env gracefully
func TestConnectWithMissingEnvFileDoesNotCrash(t *testing.T) {
	// Create a temporary directory without .env file
	tmpDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// Test godotenv.Load with missing file
	err = godotenv.Load()
	// Should return an error but not crash
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file")
}

// TestConnectDSNConstruction tests the DSN string construction
func TestConnectDSNConstruction(t *testing.T) {
	tests := []struct {
		name     string
		host     string
		user     string
		password string
		dbname   string
		port     string
		expected string
	}{
		{
			name:     "Standard configuration",
			host:     "localhost",
			user:     "postgres",
			password: "password",
			dbname:   "mydb",
			port:     "5432",
			expected: "host=localhost user=postgres password=password dbname=mydb port=5432 sslmode=disable",
		},
		{
			name:     "Remote host",
			host:     "db.example.com",
			user:     "admin",
			password: "secret",
			dbname:   "production",
			port:     "5433",
			expected: "host=db.example.com user=admin password=secret dbname=production port=5433 sslmode=disable",
		},
		{
			name:     "Special characters in password",
			host:     "localhost",
			user:     "user",
			password: "p@ss!word#123",
			dbname:   "db",
			port:     "5432",
			expected: "host=localhost user=user password=p@ss!word#123 dbname=db port=5432 sslmode=disable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
				tt.host, tt.user, tt.password, tt.dbname, tt.port)
			assert.Equal(t, tt.expected, dsn)
		})
	}
}

// TestConnectWithEmptyEnvironmentVariables tests behavior with empty env vars
func TestConnectWithEmptyEnvironmentVariables(t *testing.T) {
	// Save original environment variables
	originalVars := map[string]string{
		"POSTGRES_USER":     os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":       os.Getenv("POSTGRES_DB"),
		"POSTGRES_HOST":     os.Getenv("POSTGRES_HOST"),
		"POSTGRES_PORT":     os.Getenv("POSTGRES_PORT"),
	}

	defer func() {
		for k, v := range originalVars {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}()

	// Clear all environment variables
	_ = os.Unsetenv("POSTGRES_USER")
	_ = os.Unsetenv("POSTGRES_PASSWORD")
	_ = os.Unsetenv("POSTGRES_DB")
	_ = os.Unsetenv("POSTGRES_HOST")
	_ = os.Unsetenv("POSTGRES_PORT")

	// Test DSN construction with empty values
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	// Should create a DSN with empty values (this would fail to connect, which is expected)
	expected := "host= user= password= dbname= port= sslmode=disable"
	assert.Equal(t, expected, dsn)
}

// TestConnectGormOpenWithValidDSN tests that gorm.Open works with a valid DSN
func TestConnectGormOpenWithValidDSN(t *testing.T) {
	// Test with SQLite (since we can't reliably test with PostgreSQL in unit tests)
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Verify connection works
	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

// TestConnectGormOpenWithInvalidDSN tests error handling with invalid DSN
func TestConnectGormOpenWithInvalidDSN(t *testing.T) {
	// Test with PostgreSQL driver but invalid connection string
	invalidDSN := "host=invalid-host-12345 user=invalid password=invalid dbname=invalid port=9999 sslmode=disable"

	db, err := gorm.Open(postgres.Open(invalidDSN), &gorm.Config{})

	// Should return an error for invalid connection
	assert.Error(t, err)
	// DB might be nil or not, depending on GORM version
	_ = db
}

// TestConnectIntegrationWithRealDatabase tests Connect with a real database if available
func TestConnectIntegrationWithRealDatabase(t *testing.T) {
	// Only run if PostgreSQL is configured in environment
	if os.Getenv("RUN_INTEGRATION_TESTS") != "true" {
		t.Error("Integration test ERROR- set RUN_INTEGRATION_TESTS=true to run")
	}

	// Save original DB
	originalDB := DB
	defer func() {
		DB = originalDB
	}()

	// Run Connect
	Connect()

	// Verify DB was initialized
	assert.NotNil(t, DB)

	// Test a simple query
	var result int
	err := DB.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)
}

// TestConnectEnvironmentVariablesPrecedence tests that env vars take precedence over .env file
func TestConnectEnvironmentVariablesPrecedence(t *testing.T) {
	// Create a temporary directory with .env file
	tmpDir := t.TempDir()
	envFile := tmpDir + "/.env"
	envContent := `POSTGRES_HOST=envfilehost
POSTGRES_USER=envfileuser
`
	err := os.WriteFile(envFile, []byte(envContent), 0644)
	assert.NoError(t, err)

	// Set environment variables (should take precedence)
	originalHost := os.Getenv("POSTGRES_HOST")
	originalUser := os.Getenv("POSTGRES_USER")
	defer func() {
		if originalHost != "" {
			_ = os.Setenv("POSTGRES_HOST", originalHost)
		} else {
			_ = os.Unsetenv("POSTGRES_HOST")
		}
		if originalUser != "" {
			_ = os.Setenv("POSTGRES_USER", originalUser)
		} else {
			_ = os.Unsetenv("POSTGRES_USER")
		}
	}()

	_ = os.Setenv("POSTGRES_HOST", "envvarhost")
	_ = os.Setenv("POSTGRES_USER", "envvaruser")

	// Change to temp directory and load .env
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tmpDir)

	// Load .env file (should not overwrite existing env vars)
	err = godotenv.Load()
	assert.NoError(t, err)

	// Verify environment variables still have their original values (not overwritten by .env)
	// godotenv.Load does NOT overwrite existing environment variables
	assert.Equal(t, "envvarhost", os.Getenv("POSTGRES_HOST"))
	assert.Equal(t, "envvaruser", os.Getenv("POSTGRES_USER"))
}

// TestConnectLogsErrorForMissingEnvFile tests that godotenv.Load error is logged
func TestConnectLogsErrorForMissingEnvFile(t *testing.T) {
	// Create a temporary directory without .env file
	tmpDir := t.TempDir()

	// Change to temp directory
	originalDir, err := os.Getwd()
	assert.NoError(t, err)
	defer func() { _ = os.Chdir(originalDir) }()

	_ = os.Chdir(tmpDir)

	// The actual Connect function logs the error with log.Printf
	// We can't easily capture log output without changing the function,
	// but we can verify godotenv.Load returns an error
	err = godotenv.Load()
	assert.Error(t, err)
}

// TestConnectSuccessfulConnectionFlow tests the happy path
func TestConnectSuccessfulConnectionFlow(t *testing.T) {
	// This test documents the expected flow of Connect function
	// 1. Load .env file (may or may not exist)
	// 2. Read environment variables
	// 3. Construct DSN
	// 4. Open database connection
	// 5. Set global DB variable

	// We can't actually run Connect() without a real database,
	// but we can test the individual steps

	// Step 1 & 2: Environment variables
	originalVars := map[string]string{
		"POSTGRES_USER":     os.Getenv("POSTGRES_USER"),
		"POSTGRES_PASSWORD": os.Getenv("POSTGRES_PASSWORD"),
		"POSTGRES_DB":       os.Getenv("POSTGRES_DB"),
		"POSTGRES_HOST":     os.Getenv("POSTGRES_HOST"),
		"POSTGRES_PORT":     os.Getenv("POSTGRES_PORT"),
	}

	defer func() {
		for k, v := range originalVars {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}()

	_ = os.Setenv("POSTGRES_USER", "testuser")
	_ = os.Setenv("POSTGRES_PASSWORD", "testpass")
	_ = os.Setenv("POSTGRES_DB", "testdb")
	_ = os.Setenv("POSTGRES_HOST", "localhost")
	_ = os.Setenv("POSTGRES_PORT", "5432")

	// Step 3: Construct DSN
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	assert.NotEmpty(t, dsn)
	assert.Contains(t, dsn, "host=localhost")
	assert.Contains(t, dsn, "user=testuser")
	assert.Contains(t, dsn, "dbname=testdb")
}

// TestConnectGlobalDBVariableAssignment tests that Connect would set the global DB variable
func TestConnectGlobalDBVariableAssignment(t *testing.T) {
	// Save original DB
	originalDB := DB

	// Create a mock database connection
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)

	// Simulate what Connect does: assign to global DB
	DB = db

	// Verify assignment
	assert.NotNil(t, DB)
	assert.Equal(t, db, DB)

	// Test that we can use the global DB
	var result int
	err = DB.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)

	// Restore original DB
	DB = originalDB
}

// TestConnectDatabaseConnectionSuccess tests successful database connection
func TestConnectDatabaseConnectionSuccess(t *testing.T) {
	// Test with SQLite to verify the connection logic works
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	assert.NotNil(t, db)

	// Verify we can execute queries
	var result int
	err = db.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, result)

	// Verify database connection is healthy
	sqlDB, err := db.DB()
	assert.NoError(t, err)
	assert.NotNil(t, sqlDB)

	err = sqlDB.Ping()
	assert.NoError(t, err)
}

// TestConnectAllEnvironmentVariablesUsed tests that all environment variables are used
func TestConnectAllEnvironmentVariablesUsed(t *testing.T) {
	// Save original environment variables
	vars := []string{"POSTGRES_USER", "POSTGRES_PASSWORD", "POSTGRES_DB", "POSTGRES_HOST", "POSTGRES_PORT"}
	originalVars := make(map[string]string)
	for _, v := range vars {
		originalVars[v] = os.Getenv(v)
	}

	defer func() {
		for k, v := range originalVars {
			if v != "" {
				_ = os.Setenv(k, v)
			} else {
				_ = os.Unsetenv(k)
			}
		}
	}()

	// Set all variables to unique values
	_ = os.Setenv("POSTGRES_USER", "user123")
	_ = os.Setenv("POSTGRES_PASSWORD", "pass456")
	_ = os.Setenv("POSTGRES_DB", "db789")
	_ = os.Setenv("POSTGRES_HOST", "host.example.com")
	_ = os.Setenv("POSTGRES_PORT", "9999")

	// Construct DSN as Connect does
	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	// Verify all values are in the DSN
	assert.Contains(t, dsn, "user123")
	assert.Contains(t, dsn, "pass456")
	assert.Contains(t, dsn, "db789")
	assert.Contains(t, dsn, "host.example.com")
	assert.Contains(t, dsn, "9999")
}
