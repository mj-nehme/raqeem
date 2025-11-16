package main

import (
	"log"
	"mentor-backend/database"
	"mentor-backend/router"
	"os"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	_ "mentor-backend/docs" // swagger docs
)

// @title Raqeem Mentor Backend API
// @version 1.0
// @description Device management and monitoring dashboard API for Raqeem IoT platform
// @description Provides endpoints for device listing, metrics retrieval, alert management, and remote command execution

// @contact.name API Support
// @contact.url https://github.com/mj-nehme/raqeem
// @contact.email support@example.com

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:30081
// @BasePath /

// @schemes http https

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

	// Start server
	log.Printf("Starting HTTP server on port %s", a.Port)
	log.Println("API documentation available at /swagger/index.html")
	return a.Router.Run(":" + a.Port)
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
