package main

import (
"context"
"log"
"mentor-backend/database"
"mentor-backend/logging"
"mentor-backend/router"
"net/http"
"os"
"os/signal"
"syscall"
"time"

"github.com/gin-gonic/gin"
"gorm.io/gorm"

_ "mentor-backend/docs" // swagger docs
)

// @title Raqeem Mentor Backend API
// @version 1.0
// @description Device management and monitoring dashboard API for Raqeem IoT platform
// @description
// @description ## Overview
// @description The Mentor Backend provides a centralized dashboard for monitoring and managing IoT devices.
// @description It aggregates telemetry data, provides device management capabilities, and enables remote command execution.
// @description
// @description ## Key Features
// @description - **Device Management**: View and manage all registered devices
// @description - **Metrics Monitoring**: Real-time performance metrics visualization
// @description - **Activity Tracking**: User activity logs and audit trails
// @description - **Alert Management**: Centralized alert aggregation and monitoring
// @description - **Remote Commands**: Execute commands on devices remotely
// @description - **Screenshot Viewing**: Access device screenshots with presigned URLs
// @description
// @description ## Data Flow
// @description The Mentor Backend typically receives data forwarded from the Devices Backend.
// @description It provides query endpoints for frontends and management dashboards.
// @description
// @description ## Authentication
// @description Currently, the API does not require authentication.
// @description Authentication and authorization will be added in future releases.

// @contact.name API Support
// @contact.url https://github.com/mj-nehme/raqeem
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:30081
// @BasePath /

// @schemes http https

// @tag.name devices
// @tag.description Device registration, status, and telemetry endpoints

// @tag.name commands
// @tag.description Remote command execution and status tracking

// @tag.name activities
// @tag.description Activity logging and retrieval across all devices

// App encapsulates the application configuration and dependencies
type App struct {
DB     *gorm.DB
Router *gin.Engine
Port   string
}

// NewApp creates and initializes a new application instance
func NewApp() *App {
return &App{}
}

// setupDatabase initializes the database connection and runs migrations
func (a *App) setupDatabase() error {
<<<<<<< HEAD
database.Connect()
=======
logging.Info("Initializing database connection")
database.Connect()
logging.Info("Database setup completed successfully")
>>>>>>> origin/master
return nil
}

// setupRouter initializes the Gin router with all routes and middleware
func (a *App) setupRouter() *gin.Engine {
<<<<<<< HEAD
=======
logging.Info("Setting up application routes")
>>>>>>> origin/master
r := router.New()
r.SetupAllRoutes()

a.Router = r.Engine()
<<<<<<< HEAD
return r.Engine()
}

// Start initializes and starts the application server with graceful shutdown support
func (a *App) Start() error {
// Initialize logging
if err := logging.InitLogger(); err != nil {
log.Printf("Warning: Failed to initialize structured logging: %v. Using default logging.", err)
} else {
log.Println("Structured logging initialized successfully")
}

// Setup database
if err := a.setupDatabase(); err != nil {
=======
logging.Info("Router setup complete")
return r.Engine()
}

// Start initializes and starts the application server with graceful shutdown
func (a *App) Start() error {
logging.Info("Starting Mentor Backend application")

// Setup database
if err := a.setupDatabase(); err != nil {
logging.Error("Failed to setup database", map[string]interface{}{
"error": err.Error(),
})
>>>>>>> origin/master
return err
}

// Setup router
a.setupRouter()

// Get port from environment
a.Port = os.Getenv("PORT")
if a.Port == "" {
<<<<<<< HEAD
log.Fatal("PORT environment variable is required (set by Helm chart or .env)")
}

// Create HTTP server
srv := &http.Server{
Addr:         ":" + a.Port,
Handler:      a.Router,
ReadTimeout:  15 * time.Second,
WriteTimeout: 15 * time.Second,
IdleTimeout:  60 * time.Second,
}

// Channel to listen for errors coming from the listener
serverErrors := make(chan error, 1)

// Start the server in a goroutine
go func() {
log.Printf("Starting server on port %s", a.Port)
serverErrors <- srv.ListenAndServe()
}()

// Channel to listen for interrupt signal to terminate gracefully
shutdown := make(chan os.Signal, 1)
signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

// Block until we receive a signal or an error
select {
case err := <-serverErrors:
return err

case sig := <-shutdown:
log.Printf("Received shutdown signal: %v. Starting graceful shutdown...", sig)

// Give outstanding requests a deadline for completion
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// Attempt graceful shutdown
if err := srv.Shutdown(ctx); err != nil {
log.Printf("Could not gracefully shutdown the server: %v", err)
// Force close
if closeErr := srv.Close(); closeErr != nil {
return closeErr
}
}

// Close database connection
if err := database.Close(); err != nil {
log.Printf("Error closing database connection: %v", err)
}

log.Println("Server shutdown completed")
}

return nil
=======
return ErrPortNotSet
}

// Create HTTP server with timeouts
srv := &http.Server{
Addr:           ":" + a.Port,
Handler:        a.Router,
ReadTimeout:    15 * time.Second,
WriteTimeout:   15 * time.Second,
IdleTimeout:    60 * time.Second,
MaxHeaderBytes: 1 << 20, // 1 MB
}

// Start server in a goroutine
go func() {
logging.Info("Starting HTTP server", map[string]interface{}{
"port": a.Port,
})
log.Println("API documentation available at /swagger/index.html")
if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
logging.Error("Server error", map[string]interface{}{
"error": err.Error(),
})
log.Fatalf("Server error: %v", err)
}
}()

logging.Info("Server started successfully", map[string]interface{}{
"port": a.Port,
})

// Wait for interrupt signal to gracefully shutdown the server
quit := make(chan os.Signal, 1)
// Accept SIGINT (Ctrl+C) and SIGTERM (docker stop)
signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
<-quit

logging.Info("Shutting down server...")

// Create context with timeout for shutdown (10s for compatibility with master)
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

// Shutdown HTTP server
if err := srv.Shutdown(ctx); err != nil {
logging.Error("Server forced to shutdown", map[string]interface{}{
"error": err.Error(),
})
log.Printf("Server forced to shutdown: %v", err)
}

// Close database connection
if err := database.Shutdown(); err != nil {
logging.Error("Error closing database", map[string]interface{}{
"error": err.Error(),
})
log.Printf("Error closing database: %v", err)
}

logging.Info("Server exited gracefully")
return nil
}

// Custom error types for better error handling
var (
ErrPortNotSet = &AppError{
Message: "PORT environment variable is required (set by Helm chart or .env)",
Code:    "ERR_PORT_NOT_SET",
}
)

// AppError represents application-level errors
type AppError struct {
Message string
Code    string
}

func (e *AppError) Error() string {
return e.Message
>>>>>>> origin/master
}

func main() {
app := NewApp()
if err := app.Start(); err != nil {
<<<<<<< HEAD
=======
logging.Error("Failed to start application", map[string]interface{}{
"error": err.Error(),
})
>>>>>>> origin/master
log.Fatalf("Failed to start application: %v", err)
}
}
