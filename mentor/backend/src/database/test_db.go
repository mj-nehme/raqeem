package database

import (
	"fmt"
	"log"
	"os"
	"testing"

	"mentor-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestDB holds the test database connection
var TestDB *gorm.DB

// SetupTestDB initializes a test database connection
func SetupTestDB(t *testing.T) *gorm.DB {
	// Use test database credentials or defaults
	user := getEnvOrDefault("POSTGRES_USER", "postgres")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "password")
	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port := getEnvOrDefault("POSTGRES_PORT", "5432")
	dbname := getEnvOrDefault("POSTGRES_TEST_DB", "raqeem_test")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		// Skip test if database is not available
		t.Skipf("Test database not available: %v", err)
		return nil
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

	log.Printf("Test database connected successfully: %s", dbname)
	TestDB = db
	return db
}

// CleanupTestDB cleans up test data after each test
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	if db == nil {
		return
	}

	// Clean up all test data in reverse order due to foreign key constraints
	tables := []string{
		"alerts",
		"screenshots",
		"remote_commands",
		"activities",
		"activity_logs",
		"processes",
		"device_metrics",
		"devices",
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)).Error; err != nil {
			log.Printf("Warning: Failed to truncate table %s: %v", table, err)
		}
	}
}

// CreateTestDatabase creates a test database if it doesn't exist
func CreateTestDatabase() error {
	user := getEnvOrDefault("POSTGRES_USER", "postgres")
	password := getEnvOrDefault("POSTGRES_PASSWORD", "password")
	host := getEnvOrDefault("POSTGRES_HOST", "localhost")
	port := getEnvOrDefault("POSTGRES_PORT", "5432")
	dbname := getEnvOrDefault("POSTGRES_TEST_DB", "raqeem_test")

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
