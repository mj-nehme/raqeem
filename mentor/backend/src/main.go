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

// Start initializes and starts the application server with graceful shutdown
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
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logging.Error("Failed to start server", map[string]interface{}{
				"error": err.Error(),
			})
			log.Fatalf("Failed to start server: %v", err)
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

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logging.Error("Server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	logging.Info("Server exited gracefully")
	return nil
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
