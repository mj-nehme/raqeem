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
)

func main() {
	database.Connect()

	// Auto-migrate your models (include device-related models)
	database.DB.AutoMigrate(&models.Activity{})
	database.DB.AutoMigrate(&models.Device{}, &models.DeviceMetrics{}, &models.Process{}, &models.ActivityLog{}, &models.RemoteCommand{}, &models.Screenshot{}, &models.Alert{})

	r := gin.Default()

	// Allow the frontend dev server (vite) to call the API from the browser
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	// You can pass multiple origins as comma-separated in FRONTEND_ORIGIN
	// Support multiple comma-separated origins
	origins := []string{}
	for _, o := range strings.Split(frontendOrigin, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
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
	r.GET("/devices", controllers.ListDevices)
	r.GET("/devices/:id/metrics", controllers.GetDeviceMetrics)
	r.GET("/devices/:id/processes", controllers.GetDeviceProcesses)
	r.GET("/devices/:id/activities", controllers.GetDeviceActivities)
	r.GET("/devices/:id/alerts", controllers.GetDeviceAlerts)
	r.GET("/devices/:id/screenshots", controllers.GetDeviceScreenshots)
	r.GET("/devices/:id/commands/pending", controllers.GetPendingCommands)
	r.POST("/commands/status", controllers.UpdateCommandStatus)
	r.POST("/devices/:id/alerts", controllers.ReportAlert)

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is required (set by Helm chart or .env)")
	}
	log.Fatal(r.Run(":" + port))
}
