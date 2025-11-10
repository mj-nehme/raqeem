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
func main() {
	database.Connect()

	// Auto-migrate your models (include device-related models)
	if err := database.DB.AutoMigrate(&models.Activity{}); err != nil {
		log.Fatalf("AutoMigrate Activity failed: %v", err)
	}
	if err := database.DB.AutoMigrate(&models.Device{}, &models.DeviceMetrics{}, &models.Process{}, &models.ActivityLog{}, &models.RemoteCommand{}, &models.Screenshot{}, &models.Alert{}); err != nil {
		log.Fatalf("AutoMigrate device models failed: %v", err)
	}

	r := router.New()
	r.SetupAllRoutes()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is required (set by Helm chart or .env)")
	}
	log.Fatal(r.Run(":" + port))
}
