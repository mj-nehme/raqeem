package main

import (
	"log"
	"mentor-backend/database"
	"mentor-backend/models"
	"mentor-backend/router"
	"os"

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
// If DB is already set (for testing), it will use that instead of connecting
func (a *App) setupDatabase() error {
	// Only connect if DB is not already set (allows for dependency injection)
	if a.DB == nil {
		database.Connect()
		a.DB = database.DB
	}

	// Auto-migrate your models (include device-related models)
	if err := a.DB.AutoMigrate(&models.Activity{}); err != nil {
		return err
	}
	if err := a.DB.AutoMigrate(&models.Device{}, &models.DeviceMetrics{}, &models.Process{}, &models.ActivityLog{}, &models.RemoteCommand{}, &models.Screenshot{}, &models.Alert{}); err != nil {
		return err
	}
	return nil
}

	r := router.New()
	r.SetupAllRoutes()

	a.Router = r
	return r
}

// Start initializes and starts the application server
func (a *App) Start() error {
	// Setup database
	if err := a.setupDatabase(); err != nil {
		return err
	}

	// Setup router
	a.setupRouter()

	// Get port from environment
	a.Port = os.Getenv("PORT")
	if a.Port == "" {
		log.Fatal("PORT environment variable is required (set by Helm chart or .env)")
	}

	// Start server
	return a.Router.Run(":" + a.Port)
}

func main() {
	app := NewApp()
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
