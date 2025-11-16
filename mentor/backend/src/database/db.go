package database

import (
"context"
"database/sql"
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
"mentor-backend/reliability"
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

var missing []string
for _, key := range required {
if os.Getenv(key) == "" {
missing = append(missing, key)
}
}

if len(missing) > 0 {
return fmt.Errorf("missing required database configuration: %v", missing)
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

// getEnvDuration retrieves a duration environment variable with a default value
func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
if val := os.Getenv(key); val != "" {
if duration, err := time.ParseDuration(val); err == nil {
return duration
}
}
return defaultValue
}

// connectWithConfig attempts to connect to the database and returns an error if it fails.
// This function is separated for testing purposes and includes retry logic for resilience.
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

// Get connection pool configuration from environment variables
maxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 25)
maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 5)
connMaxLifetime := getEnvDuration("DB_CONN_MAX_LIFETIME", 5*time.Minute)
connMaxIdleTime := getEnvDuration("DB_CONN_MAX_IDLE_TIME", 10*time.Minute)

dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
host, user, password, dbname, port)

// Use retry logic for database connection
retryConfig := reliability.DatabaseRetryConfig()
ctx := context.Background()

var err error
err = reliability.RetryWithBackoff(ctx, retryConfig, func() error {
DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
DisableForeignKeyConstraintWhenMigrating: true,
Logger:                                   logger.Default.LogMode(logger.Info),
})
if err != nil {
return fmt.Errorf("failed to connect to database: %v", err)
}

// Get underlying SQL database for connection pool configuration
sqlDB, err := DB.DB()
if err != nil {
return fmt.Errorf("failed to get underlying SQL database: %v", err)
}

// Configure connection pool
sqlDB.SetMaxOpenConns(maxOpenConns)
sqlDB.SetMaxIdleConns(maxIdleConns)
sqlDB.SetConnMaxLifetime(connMaxLifetime)
sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

// Test the connection
if err := sqlDB.Ping(); err != nil {
return fmt.Errorf("failed to ping database: %v", err)
}

return nil
})

if err != nil {
return err
}

log.Printf("Database connection successful (pool: max_open=%d, max_idle=%d, max_lifetime=%v)",
maxOpenConns, maxIdleConns, connMaxLifetime)
return nil
}

// getEnvAsInt retrieves an environment variable as an integer, or returns the default value
func getEnvAsInt(key string, defaultValue int) int {
if value := os.Getenv(key); value != "" {
if intValue, err := strconv.Atoi(value); err == nil {
return intValue
}
}
return defaultValue
}

// getEnvAsDuration retrieves an environment variable as a duration, or returns the default value
func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
if value := os.Getenv(key); value != "" {
if duration, err := time.ParseDuration(value); err == nil {
return duration
}
}
return defaultValue
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

// HealthCheck checks the database connection health
func HealthCheck() error {
if DB == nil {
return fmt.Errorf("database connection is nil")
}

sqlDB, err := DB.DB()
if err != nil {
return fmt.Errorf("failed to get underlying SQL database: %v", err)
}

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := sqlDB.PingContext(ctx); err != nil {
return fmt.Errorf("database ping failed: %v", err)
}

return nil
}

// Close gracefully closes the database connection
func Close() error {
if DB == nil {
return nil
}

sqlDB, err := DB.DB()
if err != nil {
return fmt.Errorf("failed to get underlying SQL database: %v", err)
}

log.Println("Closing database connection...")
if err := sqlDB.Close(); err != nil {
return fmt.Errorf("failed to close database: %v", err)
}

log.Println("Database connection closed successfully")
return nil
}

// GetStats returns database connection pool statistics
func GetStats() sql.DBStats {
if DB == nil {
return sql.DBStats{}
}

sqlDB, err := DB.DB()
if err != nil {
return sql.DBStats{}
}

return sqlDB.Stats()
}
