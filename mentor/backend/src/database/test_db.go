package database

import (
	"fmt"
	"log"
	"os"
	"testing"

	"mentor-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDB holds the test database connection
var TestDB *gorm.DB

// SetupTestDB initializes a test database connection
func SetupTestDB(t *testing.T) *gorm.DB {
	var db *gorm.DB
	var err error

	// Check if we should use PostgreSQL (only if explicitly configured for CI)
	usePostgres := os.Getenv("USE_POSTGRES_FOR_TESTS") == "true"

	if usePostgres {
		// Use PostgreSQL for tests (CI environment)
		user := getEnvOrDefault("POSTGRES_USER", "postgres")
		password := getEnvOrDefault("POSTGRES_PASSWORD", "password")
		host := getEnvOrDefault("POSTGRES_HOST", "localhost")
		port := getEnvOrDefault("POSTGRES_PORT", "5432")

		// For CI, use the main database; for local testing, use test database
		var dbname string
		if user == "monitor" {
			// CI environment - use the monitoring_db
			dbname = getEnvOrDefault("POSTGRES_DB", "monitoring_db")
		} else {
			// Local environment - use test database
			dbname = getEnvOrDefault("POSTGRES_TEST_DB", "raqeem_test")
		}

		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
			host, user, password, dbname, port)

		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			// Skip test if database is not available
			t.Skipf("Test database not available: %v", err)
			return nil
		}
		log.Printf("Test database connected successfully (PostgreSQL): %s", dbname)
	} else {
		// Use SQLite in-memory database for tests (default, no external dependencies)
		db, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
		if err != nil {
			t.Fatalf("Failed to create SQLite test database: %v", err)
			return nil
		}

		// Enable WAL mode for better concurrency support in SQLite
		sqlDB, err := db.DB()
		if err == nil {
			sqlDB.Exec("PRAGMA journal_mode=WAL;")
			sqlDB.Exec("PRAGMA busy_timeout=5000;") // 5 second timeout for locks
		}

		log.Printf("Test database connected successfully (SQLite in-memory)")
	}

	// Auto-migrate all models for testing
	err = db.AutoMigrate(
		&models.Device{},
		&models.DeviceMetrics{},
		&models.Process{},
		&models.Activity{},
		&models.ActivityLog{},
		&models.RemoteCommand{},
		&models.Screenshot{},
		&models.Alert{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	TestDB = db
	return db
}

// CleanupTestDB cleans up test data after each test
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	if db == nil {
		return
	}

	// Clean up all test data - using Delete with unscoped to actually remove records
	// Order matters due to foreign key constraints
	db.Unscoped().Where("1 = 1").Delete(&models.Alert{})
	db.Unscoped().Where("1 = 1").Delete(&models.Screenshot{})
	db.Unscoped().Where("1 = 1").Delete(&models.RemoteCommand{})
	db.Unscoped().Where("1 = 1").Delete(&models.Activity{})
	db.Unscoped().Where("1 = 1").Delete(&models.ActivityLog{})
	db.Unscoped().Where("1 = 1").Delete(&models.Process{})
	db.Unscoped().Where("1 = 1").Delete(&models.DeviceMetrics{})
	db.Unscoped().Where("1 = 1").Delete(&models.Device{})
}

// CreateTestDatabase creates a test database if it doesn't exist
// This function is only needed for PostgreSQL testing
func CreateTestDatabase() error {
	// Check if we should use PostgreSQL
	usePostgres := os.Getenv("USE_POSTGRES_FOR_TESTS") == "true"

	if !usePostgres {
		// SQLite doesn't need database creation
		return nil
	}

	user := getEnvOrDefault("POSTGRES_USER", "postgres")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "password")
	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port := getEnvOrDefault("POSTGRES_PORT", "5432")

	// For CI, use the main database; for local testing, use test database
	var dbname string
	if user == "monitor" {
		// CI environment - database already exists, no need to create
		return nil
	} else {
		// Local environment - create test database
		dbname = getEnvOrDefault("POSTGRES_TEST_DB", "raqeem_test")
	}

	// Connect to postgres database to create test database
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=postgres port=%s sslmode=disable",
		host, user, password, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %v", err)
	}

	// Create test database if it doesn't exist
	result := db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	if result.Error != nil {
		// Database might already exist, which is fine
		log.Printf("Database creation result: %v", result.Error)
	}

	return nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
