package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"mentor-backend/models"
)

var DB *gorm.DB

// loadEnv tries multiple locations for a .env file to reduce CWD sensitivity
func loadEnv() {
	// Try common relative locations from various run directories
	candidates := []string{
		".env",
		"../.env",
		"../../.env",
		"../../../.env",
		// project specific
		"mentor/backend/.env",
	}

	// Also walk upwards from current working directory up to root
	if wd, err := os.Getwd(); err == nil {
		dir := wd
		for i := 0; i < 5; i++ { // limit depth to avoid long walks
			candidates = append(candidates, filepath.Join(dir, ".env"))
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}

	// Attempt loading from the first path that exists
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			if err := godotenv.Load(p); err == nil {
				log.Printf("Loaded environment from %s", p)
				return
			}
		}
	}

	// Fallback to default behavior (may be no-op)
	if err := godotenv.Load(); err != nil {
		log.Printf("godotenv.Load: %v", err)
	}
}

// connectWithConfig attempts to connect to the database and returns an error if it fails.
// This function is separated for testing purposes.
func connectWithConfig() error {
	// Load environment variables from .env file(s)
	loadEnv()

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	// Validate required environment variables
	if user == "" || password == "" || dbname == "" || host == "" || port == "" {
		return fmt.Errorf("missing required database configuration: ensure POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB, POSTGRES_HOST, and POSTGRES_PORT are set")
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		// Enable logging for better debugging
		// Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database at %s:%s: %w", host, port, err)
	}

	// Configure connection pool for better performance and reliability
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	// Set connection pool settings
	sqlDB.SetMaxIdleConns(10)           // Maximum number of idle connections
	sqlDB.SetMaxOpenConns(100)          // Maximum number of open connections
	sqlDB.SetConnMaxLifetime(time.Hour) // Maximum lifetime of a connection

	log.Printf("Database connection successful to %s:%s", host, port)
	return nil
}

func Connect() {
	if err := connectWithConfig(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Perform automatic migrations on connect to avoid stale schema
	if err := migrate(DB); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}
}

// migrate runs the schema migrations sequentially
func migrate(db *gorm.DB) error {
	// Migrate in explicit order to avoid FK issues even if constraints are enabled later
	steps := []interface{}{
		&models.Device{},
		&models.DeviceMetric{},
		&models.DeviceProcess{},
		&models.DeviceActivity{},
		&models.DeviceRemoteCommand{},
		&models.DeviceScreenshot{},
		&models.DeviceAlert{},
		&models.User{},
	}
	for _, m := range steps {
		if err := db.AutoMigrate(m); err != nil {
			return err
		}
	}
	return nil
}
