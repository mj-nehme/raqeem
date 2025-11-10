package router

import (
	"mentor-backend/controllers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router holds the configuration for HTTP routes
type Router struct {
	engine *gin.Engine
}

// New creates a new router instance
func New() *Router {
	return &Router{
		engine: gin.Default(),
	}
}

// NewWithEngine creates a router with a provided gin engine
func NewWithEngine(engine *gin.Engine) *Router {
	return &Router{
		engine: engine,
	}
}

// SetupCORS configures Cross-Origin Resource Sharing
func (r *Router) SetupCORS() {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	r.engine.Use(cors.New(config))
}

// SetupSwagger configures Swagger documentation routes
func (r *Router) SetupSwagger() {
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.engine.GET("/docs", func(c *gin.Context) {
		c.Redirect(301, "/swagger/index.html")
	})
}

// SetupDeviceRoutes configures all device-related routes
func (r *Router) SetupDeviceRoutes() {
	deviceGroup := r.engine.Group("/devices")
	{
		// Device management
		deviceGroup.POST("/register", controllers.RegisterDevice)
		deviceGroup.GET("", controllers.ListDevices)

		// Device metrics and monitoring
		deviceGroup.POST("/:id/metrics", controllers.UpdateDeviceMetrics)
		deviceGroup.GET("/:id/metrics", controllers.GetDeviceMetrics)

		// Device processes
		deviceGroup.POST("/:id/processes", controllers.UpdateProcessList)
		deviceGroup.GET("/:id/processes", controllers.GetDeviceProcesses)

		// Activity logging
		deviceGroup.POST("/:id/activity", controllers.LogActivity)
		deviceGroup.GET("/:id/activities", controllers.GetDeviceActivities)

		// Alerts
		deviceGroup.POST("/:id/alerts", controllers.ReportAlert)
		deviceGroup.GET("/:id/alerts", controllers.GetDeviceAlerts)

		// Screenshots
		deviceGroup.POST("/:id/screenshots", controllers.StoreScreenshot)
		deviceGroup.GET("/:id/screenshots", controllers.GetDeviceScreenshots)

		// Remote commands
		deviceGroup.POST("/:id/commands", controllers.CreateRemoteCommand)
		deviceGroup.GET("/:id/commands", controllers.GetDeviceCommands)
		deviceGroup.GET("/:id/commands/pending", controllers.GetPendingCommands)
		deviceGroup.PUT("/commands/:commandId/status", controllers.UpdateCommandStatus)
	}
}

// SetupActivityRoutes configures activity-related routes
func (r *Router) SetupActivityRoutes() {
	r.engine.GET("/activities", controllers.ListActivities)
}

// SetupHealthCheck configures health check endpoints
func (r *Router) SetupHealthCheck() {
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "mentor-backend",
		})
	})

	r.engine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}

// SetupAllRoutes configures all application routes
func (r *Router) SetupAllRoutes() {
	r.SetupCORS()
	r.SetupSwagger()
	r.SetupHealthCheck()
	r.SetupDeviceRoutes()
	r.SetupActivityRoutes()
}

// GetEngine returns the underlying gin engine
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// GetRoutes returns information about registered routes
func (r *Router) GetRoutes() gin.RoutesInfo {
	return r.engine.Routes()
}

// Run starts the HTTP server on the specified address
func (r *Router) Run(addr ...string) error {
	return r.engine.Run(addr...)
}
