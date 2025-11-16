package main

import (
	"context"
	"log"
	"mentor-backend/database"
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
	log.Println("Initializing database connection...")
	database.Connect()
	log.Println("Database setup completed successfully")
	return nil
}

// setupRouter initializes the Gin router with all routes and middleware
func (a *App) setupRouter() *gin.Engine {
	log.Println("Setting up application routes...")
	r := router.New()
	r.SetupAllRoutes()

	a.Router = r.Engine()
	log.Println("Router setup completed successfully")
	return r.Engine()
}

// validatePort ensures the PORT environment variable is set
func (a *App) validatePort() error {
	a.Port = os.Getenv("PORT")
	if a.Port == "" {
		return ErrPortNotSet
	}
	log.Printf("Server will listen on port: %s", a.Port)
	return nil
}

// Start initializes and starts the application server
func (a *App) Start() error {
	log.Println("Starting Raqeem Mentor Backend API...")

	// Setup database
	if err := a.setupDatabase(); err != nil {
		log.Printf("Database setup failed: %v", err)
		return err
	}

	// Setup router
	a.setupRouter()

	// Validate and get port from environment
	if err := a.validatePort(); err != nil {
		log.Printf("Port validation failed: %v", err)
		return err
	}

	// Create HTTP server with graceful shutdown support
	srv := &http.Server{
		Addr:    ":" + a.Port,
		Handler: a.Router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Starting HTTP server on port %s", a.Port)
		log.Println("API documentation available at /swagger/index.html")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Create context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close database connection
	if err := database.Shutdown(); err != nil {
		log.Printf("Error closing database: %v", err)
	}

	log.Println("Server exited")
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
}

func main() {
	app := NewApp()
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
