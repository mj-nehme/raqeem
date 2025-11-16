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
	database.Connect()
	return nil
}

// setupRouter initializes the Gin router with all routes and middleware
func (a *App) setupRouter() *gin.Engine {
	r := router.New()
	r.SetupAllRoutes()

	a.Router = r.Engine()
	return r.Engine()
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
