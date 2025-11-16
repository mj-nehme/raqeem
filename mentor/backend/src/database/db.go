package database

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"mentor-backend/models"
)

var DB *gorm.DB

// Configuration constants
const (
	// MinPasswordLength is the minimum required password length
	MinPasswordLength = 8
	// MaxRetries for database operations
	MaxRetries = 3
	// EnvSearchDepth is the maximum directory depth to search for .env files
	EnvSearchDepth = 5
)

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
		for i := 0; i < EnvSearchDepth; i++ {
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
		log.Printf("godotenv.Load: %v (not an error if env vars are set externally)", err)
	}
}

// validateEnvVars checks that required environment variables are set
func validateEnvVars() error {
	required := []string{
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"POSTGRES_HOST",
		"POSTGRES_PORT",
	}

	missing := []string{}
	for _, envVar := range required {
		if os.Getenv(envVar) == "" {
			missing = append(missing, envVar)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("required environment variables not set: %v", missing)
	}

	return nil
}

// connectWithConfig attempts to connect to the database and returns an error if it fails.
// This function is separated for testing purposes.
func connectWithConfig() error {
	// Load environment variables from .env file(s)
	loadEnv()

	// Validate required environment variables
	if err := validateEnvVars(); err != nil {
		return err
	}

	user := os.Getenv("POSTGRES_USER")
	password := os.Getenv("POSTGRES_PASSWORD")
	dbname := os.Getenv("POSTGRES_DB")
	host := os.Getenv("POSTGRES_HOST")
	port := os.Getenv("POSTGRES_PORT")

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		host, user, password, dbname, port)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	log.Println("Database connection successful")
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
	log.Println("Running database migrations...")
	for _, m := range steps {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("migration failed for %T: %w", m, err)
		}
	}
	log.Println("Database migrations completed successfully")
	return nil
}
