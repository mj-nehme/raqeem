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

// Start initializes and starts the application server with graceful shutdown support
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
}

func main() {
	app := NewApp()
	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start application: %v", err)
	}
}
