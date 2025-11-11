package router

import (
	"mentor-backend/controllers"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router wraps a gin.Engine and provides route setup methods
type Router struct {
	engine *gin.Engine
}

// New creates a new Router with default gin engine
func New() *Router {
	return &Router{
		engine: gin.Default(),
	}
}

// SetupAllRoutes configures all routes, middleware, and Swagger documentation
func (r *Router) SetupAllRoutes() {
	r.setupCORS()
	r.setupSwagger()
	r.setupHealthCheck()
	r.setupActivityRoutes()
	r.setupDeviceRoutes()
}

// setupCORS configures CORS middleware
func (r *Router) setupCORS() {
	frontendOrigin := os.Getenv("FRONTEND_ORIGIN")
	origins := []string{}
	for _, o := range strings.Split(frontendOrigin, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	// If no origins specified, allow all origins to prevent panic
	if len(origins) == 0 {
		origins = []string{"*"}
	}

	r.engine.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
}

// setupSwagger configures Swagger documentation routes
func (r *Router) setupSwagger() {
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.engine.GET("/docs", func(c *gin.Context) {
		c.Redirect(301, "/swagger/index.html")
	})
}

// setupHealthCheck configures the health check endpoint
func (r *Router) setupHealthCheck() {
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "mentor-backend"})
	})
}

// setupActivityRoutes configures activity-related routes
func (r *Router) setupActivityRoutes() {
	r.engine.GET("/activities", controllers.ListActivities)
}

// setupDeviceRoutes configures all device-related routes
func (r *Router) setupDeviceRoutes() {
	// Device ingestion endpoints (devices will POST data here)
	r.engine.POST("/devices/register", controllers.RegisterDevice)
	r.engine.POST("/devices/metrics", controllers.UpdateDeviceMetrics)
	r.engine.POST("/devices/processes", controllers.UpdateProcessList)
	r.engine.POST("/devices/activity", controllers.Activity)
	r.engine.POST("/devices/commands", controllers.CreateRemoteCommand)
	r.engine.POST("/devices/screenshots", controllers.StoreScreenshot)

	// Device query endpoints
	r.engine.GET("/devices", controllers.ListDevices)
	r.engine.GET("/devices/:id/metrics", controllers.GetDeviceMetrics)
	r.engine.GET("/devices/:id/processes", controllers.GetDeviceProcesseses)
	r.engine.GET("/devices/:id/activities", controllers.GetDeviceActivities)
	r.engine.GET("/devices/:id/alerts", controllers.GetDeviceAlerts)
	r.engine.GET("/devices/:id/screenshots", controllers.GetDeviceScreenshots)
	r.engine.GET("/devices/:id/commands/pending", controllers.GetPendingCommands)
	r.engine.GET("/devices/:id/commands", controllers.GetDeviceCommands)

	// Command and alert endpoints
	r.engine.POST("/commands/status", controllers.UpdateCommandStatus)
	r.engine.POST("/devices/:id/alerts", controllers.ReportAlert)
}

// Run starts the HTTP server on the specified address
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}

// Engine returns the underlying gin.Engine for testing purposes
func (r *Router) Engine() *gin.Engine {
	return r.engine
}
