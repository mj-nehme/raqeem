package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

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

// getEnvInt retrieves an integer environment variable with a default value
func getEnvInt(key string, defaultValue int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultValue
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

	// Configure GORM with better defaults for production
	gormConfig := &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		// Use Info level logging in production, Warn for errors
		Logger: logger.Default.LogMode(logger.Info),
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configure connection pool for better performance and reliability
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	// Set connection pool parameters from environment or use sensible defaults
	maxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 25)
	maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 5)
	connMaxLifetime := getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 5)
	connMaxIdleTime := getEnvInt("DB_CONN_MAX_IDLE_MINUTES", 5)

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Minute)
	sqlDB.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Minute)

	log.Printf("Database connection successful (pool: max_open=%d, max_idle=%d)", maxOpenConns, maxIdleConns)
	return nil
}

// connectWithRetry attempts to connect with exponential backoff retry logic
func connectWithRetry(maxRetries int, initialDelay time.Duration) error {
	var err error
	delay := initialDelay

	for attempt := 1; attempt <= maxRetries; attempt++ {
		err = connectWithConfig()
		if err == nil {
			return nil
		}

		if attempt < maxRetries {
			log.Printf("Database connection attempt %d/%d failed: %v. Retrying in %v...", attempt, maxRetries, err, delay)
			time.Sleep(delay)
			// Exponential backoff with cap at 30 seconds
			delay *= 2
			if delay > 30*time.Second {
				delay = 30 * time.Second
			}
		}
	}

	return fmt.Errorf("failed to connect after %d attempts: %v", maxRetries, err)
}

func Connect() {
	// Try connecting with retry logic
	maxRetries := getEnvInt("DB_CONNECT_MAX_RETRIES", 5)
	initialDelay := time.Duration(getEnvInt("DB_CONNECT_INITIAL_DELAY_SECONDS", 2)) * time.Second

	if err := connectWithRetry(maxRetries, initialDelay); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Perform automatic migrations on connect to avoid stale schema
	if err := migrate(DB); err != nil {
		log.Fatalf("Database migration failed: %v", err)
	}
}

// HealthCheck verifies the database connection is alive
func HealthCheck() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	// Ping with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %v", err)
	}

	return nil
}

// Shutdown gracefully closes the database connection
func Shutdown() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %v", err)
	}

	log.Println("Closing database connection...")
	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %v", err)
	}

	log.Println("Database connection closed successfully")
	return nil
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
