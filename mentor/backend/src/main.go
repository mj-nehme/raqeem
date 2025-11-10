package main

import (
	"log"
	"mentor-backend/controllers"
	"mentor-backend/database"
	"mentor-backend/models"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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

// parseCORSOrigins parses CORS origins from environment variable
func (a *App) parseCORSOrigins() []string {
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	origins := []string{}
	for _, o := range strings.Split(frontendOrigin, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

// setupRouter initializes the Gin router with all routes and middleware
func (a *App) setupRouter() *gin.Engine {
	r := gin.Default()

	// Configure CORS
	origins := a.parseCORSOrigins()
	// Only configure CORS if origins are specified
	if len(origins) > 0 {
		r.Use(cors.New(cors.Config{
			AllowOrigins:     origins,
			AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
			AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			MaxAge:           12 * time.Hour,
		}))
	}

	// Swagger documentation routes
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/docs", func(c *gin.Context) {
		c.Redirect(301, "/swagger/index.html")
	})

	// mentor endpoints
	r.GET("/activities", controllers.ListActivities)

	// health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "mentor-backend"})
	})

	// device ingestion endpoints (devices will POST data here)
	r.POST("/devices/register", controllers.RegisterDevice)
	r.POST("/devices/metrics", controllers.UpdateDeviceMetrics)
	r.POST("/devices/processes", controllers.UpdateProcessList)
	r.POST("/devices/activity", controllers.LogActivity)
	r.POST("/devices/commands", controllers.CreateRemoteCommand)
	r.POST("/devices/screenshots", controllers.StoreScreenshot)
	r.GET("/devices", controllers.ListDevices)
	r.GET("/devices/:id/metrics", controllers.GetDeviceMetrics)
	r.GET("/devices/:id/processes", controllers.GetDeviceProcesses)
	r.GET("/devices/:id/activities", controllers.GetDeviceActivities)
	r.GET("/devices/:id/alerts", controllers.GetDeviceAlerts)
	r.GET("/devices/:id/screenshots", controllers.GetDeviceScreenshots)
	r.GET("/devices/:id/commands/pending", controllers.GetPendingCommands)
	r.GET("/devices/:id/commands", controllers.GetDeviceCommands)
	r.POST("/commands/status", controllers.UpdateCommandStatus)
	r.POST("/devices/:id/alerts", controllers.ReportAlert)

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
