package main

import (
	"log"
	"mentor-backend/database"
	"mentor-backend/logging"
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
	logging.Info("Connecting to database")
	database.Connect()
	logging.Info("Database connected successfully")
	return nil
}

// setupRouter initializes the Gin router with all routes and middleware
func (a *App) setupRouter() *gin.Engine {
	logging.Info("Setting up router and routes")
	r := router.New()
	r.SetupAllRoutes()

	a.Router = r.Engine()
	logging.Info("Router setup complete")
	return r.Engine()
}

// Start initializes and starts the application server
func (a *App) Start() error {
	logging.Info("Starting Mentor Backend application")

	// Setup database
	if err := a.setupDatabase(); err != nil {
		logging.Error("Failed to setup database", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Setup router
	a.setupRouter()

	// Get port from environment
	a.Port = os.Getenv("PORT")
	if a.Port == "" {
		log.Fatal("PORT environment variable is required (set by Helm chart or .env)")
	}

	logging.Info("Starting HTTP server", map[string]interface{}{
		"port": a.Port,
	})

	// Start server
	return a.Router.Run(":" + a.Port)
}

func main() {
	app := NewApp()
	if err := app.Start(); err != nil {
		logging.Error("Failed to start application", map[string]interface{}{
			"error": err.Error(),
		})
		log.Fatalf("Failed to start application: %v", err)
	}
}
