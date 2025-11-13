package database

// DBConfig holds database connection variables

import (
	"fmt"
	"log"
	"os"
	"strconv"
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
	Port     int32
	DBName   string
	SSLMode  string
}

// SetupTestDB initializes a test database connection with transaction isolation
// This ensures tests don't persist data - all changes are rolled back automatically
func SetupTestDB(t *testing.T, config ...DBConfig) (*gorm.DB, error) {
	var baseDB *gorm.DB
	var err error

	if len(config) == 0 {
		portStr := getEnvOrDefault("POSTGRES_PORT", "5432")
		var portInt int32 = 5432
		if p, err := strconv.Atoi(portStr); err == nil {
			portInt = int32(p)
		}
		config = append(config, DBConfig{
			User:     getEnvOrDefault("POSTGRES_USER", "monitor"),
			Password: getEnvOrDefault("POSTGRES_PASSWORD", "password"),
			Host:     getEnvOrDefault("POSTGRES_HOST", "127.0.0.1"),
			Port:     portInt,
			DBName:   getEnvOrDefault("POSTGRES_DB", "monitoring_db"),
			SSLMode:  getEnvOrDefault("SSLMODE", "disable"),
		})
	}
	dbConfig := config[0]

	// Always use PostgreSQL with configured or env-provided database name
	if dbConfig.DBName == "" {
		dbConfig.DBName = getEnvOrDefault("POSTGRES_DB", "monitoring_db")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
		dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.Port, dbConfig.SSLMode)

	log.Printf("Test database connection: host=%s port=%d dbname=%s", dbConfig.Host, dbConfig.Port, dbConfig.DBName)
	baseDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Errorf("Failed to connect to test database: %v", err)
		return nil, fmt.Errorf("test database not available: %v", err)
	}
	log.Printf("Test database connected successfully (PostgreSQL): %s", dbConfig.DBName)

	// Auto-migrate all models once at connection time (not in transactions)
	err = baseDB.AutoMigrate(
		&models.Device{},
		&models.DeviceMetric{},
		&models.DeviceProcess{},
		&models.DeviceActivity{},
		&models.DeviceRemoteCommand{},
		&models.DeviceScreenshot{},
		&models.DeviceAlert{},
		&models.User{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Begin a transaction for this test
	txDB := baseDB.Begin()
	if txDB.Error != nil {
		t.Fatalf("Failed to begin transaction: %v", txDB.Error)
	}

	// Register cleanup to rollback transaction and close base connection
	t.Cleanup(func() {
		txDB.Rollback()
		log.Printf("Test transaction rolled back for %s", t.Name())
		
		// Close the base database connection to avoid connection pool exhaustion
		if sqlDB, err := baseDB.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Printf("Error closing DB: %v", err)
			}
		}
	})

	TestDB = txDB
	return txDB, nil
}

// CleanupTestDB is deprecated - transaction rollback handles cleanup automatically
// This function is kept for backward compatibility but does nothing
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	// No-op: cleanup is handled by transaction rollback in SetupTestDB
}

// CreateTestDatabase is deprecated - database should exist before running tests
// This function is kept for backward compatibility but does nothing
func CreateTestDatabase() error {
	// No-op: database must exist before tests run (CI creates it, local requires manual setup)
	return nil
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
