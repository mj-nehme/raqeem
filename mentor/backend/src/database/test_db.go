package database

// DBConfig holds database connection variables

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

type DBConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	DBName   string
	SSLMode  string
}

// SetupTestDB initializes a test database connection
func SetupTestDB(t *testing.T, config ...DBConfig) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	if len(config) == 0 {
		config = append(config, DBConfig{
			User:     getEnvOrDefault("POSTGRES_USER", "testusername"),
			Password: getEnvOrDefault("POSTGRES_PASSWORD", "testpassword"),
			Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
			Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
			SSLMode:  "disable",
		})
	}
	dbConfig := config[0]

	// For CI, use the main database; for local testing, use test database
	var dbname string
	if dbConfig.User == "monitor" {
		// CI environment - use the monitoring_db
		dbname = getEnvOrDefault("POSTGRES_DB", "monitoring_db")
	} else {
		// Local environment - use test database
		dbname = getEnvOrDefault("POSTGRES_TEST_DB", "raqeem_test")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		dbConfig.Host, dbConfig.User, dbConfig.Password, dbname, dbConfig.Port)

	fmt.Println("Database Connection: ", dsn)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Errorf("Failed to connect to test database: %v", err)
		return nil, fmt.Errorf("Test database not available: %v", err)
	}
	log.Printf("Test database connected successfully (PostgreSQL): %s", dbname)

	// Auto-migrate all models for testing
	err = db.AutoMigrate(
		&models.Device{},
		&models.DeviceMetric{},
		&models.DeviceProcess{},
		&models.DeviceActivity{},
		&models.DeviceActivity{},
		&models.DeviceRemoteCommand{},
		&models.DeviceScreenshot{},
		&models.DeviceAlert{},
		&models.User{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	TestDB = db
	return db, nil
}

// CleanupTestDB cleans up test data after each test
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	if db == nil {
		return
	}

	// Clean up all test data - using Delete with unscoped to actually remove records
	// Order matters due to foreign key constraints
	db.Unscoped().Where("1 = 1").Delete(&models.DeviceAlert{})
	db.Unscoped().Where("1 = 1").Delete(&models.DeviceScreenshot{})
	db.Unscoped().Where("1 = 1").Delete(&models.DeviceRemoteCommand{})
	db.Unscoped().Where("1 = 1").Delete(&models.DeviceActivity{})
	db.Unscoped().Where("1 = 1").Delete(&models.DeviceActivity{})
	db.Unscoped().Where("1 = 1").Delete(&models.DeviceProcess{})
	db.Unscoped().Where("1 = 1").Delete(&models.DeviceMetric{})
	db.Unscoped().Where("1 = 1").Delete(&models.Device{})
	db.Unscoped().Where("1 = 1").Delete(&models.User{})
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
