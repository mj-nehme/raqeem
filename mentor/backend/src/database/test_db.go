package database

// DBConfig holds database connection variables

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"testing"

	"mentor-backend/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestDB holds the test database connection
var TestDB *gorm.DB

// once ensures AutoMigrate is only run once across all tests
var once sync.Once
var baseConnection *gorm.DB
var migrationError error

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

	// Initialize the base connection and run migrations only once
	once.Do(func() {
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s",
			dbConfig.Host, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.Port, dbConfig.SSLMode)

		log.Printf("Test database connection: host=%s port=%d dbname=%s", dbConfig.Host, dbConfig.Port, dbConfig.DBName)
		baseConnection, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			migrationError = fmt.Errorf("test database not available: %v", err)
			return
		}
		log.Printf("Test database connected successfully (PostgreSQL): %s", dbConfig.DBName)

		// Auto-migrate all models once at connection time (not in transactions)
		// Models must be migrated in order: Device first, then tables with foreign keys to Device
		migrationError = baseConnection.AutoMigrate(&models.Device{})
		if migrationError != nil {
			return
		}
		
		migrationError = baseConnection.AutoMigrate(
			&models.DeviceMetric{},
			&models.DeviceProcess{},
			&models.DeviceActivity{},
			&models.DeviceRemoteCommand{},
			&models.DeviceScreenshot{},
			&models.DeviceAlert{},
			&models.User{},
		)
	})

	if migrationError != nil {
		t.Fatalf("Failed to connect or migrate test database: %v", migrationError)
		return nil, migrationError
	}

	// Begin a transaction for this test
	txDB := baseConnection.Begin()
	if txDB.Error != nil {
		t.Fatalf("Failed to begin transaction: %v", txDB.Error)
	}

	// Register cleanup to rollback transaction
	t.Cleanup(func() {
		txDB.Rollback()
		log.Printf("Test transaction rolled back for %s", t.Name())
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
