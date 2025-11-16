package router

import (
	"mentor-backend/controllers"
	"mentor-backend/logging"
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
	engine := gin.Default()
	
	// Add custom middleware for reliability
	engine.Use(CorrelationIDMiddleware())
	engine.Use(RequestLoggerMiddleware())
	engine.Use(ErrorHandlerMiddleware())
	engine.Use(RecoveryMiddleware())
	
	return &Router{
		engine: engine,
	}
}

// SetupAllRoutes configures all routes, middleware, and Swagger documentation
func (r *Router) SetupAllRoutes() {
	// Add request ID middleware first
	r.engine.Use(logging.RequestIDMiddleware())
	// Add request logging middleware
	r.engine.Use(logging.RequestLoggingMiddleware())

	r.setupCORS()
	r.setupSwagger()
	r.setupHealthCheck()
	r.setupActivityRoutes()
	r.setupDeviceRoutes()
}

// setupCORS configures CORS middleware
func (r *Router) setupCORS() {
	// Support both legacy FRONTEND_ORIGIN and new FRONTEND_ORIGINS (comma-separated).
	raw := os.Getenv("FRONTEND_ORIGINS")
	if raw == "" {
		raw = os.Getenv("FRONTEND_ORIGIN")
	}

	origins := []string{}
	for _, o := range strings.Split(raw, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}

	// Development fallback: if nothing specified, use explicit localhost not wildcard.
	if len(origins) == 0 {
		origins = []string{"http://localhost:4000"}
	}

	wildcard := len(origins) == 1 && origins[0] == "*"

	// Credentials only when not wildcard & env requests it.
	allowCredEnv := strings.ToLower(os.Getenv("CORS_ALLOW_CREDENTIALS"))
	allowCredentials := !wildcard && (allowCredEnv == "1" || allowCredEnv == "true" || allowCredEnv == "yes")

	// Allow methods configurable via env, with safe defaults.
	methodsRaw := os.Getenv("CORS_ALLOW_METHODS")
	if methodsRaw == "" {
		methodsRaw = "GET,POST,PUT,PATCH,DELETE,OPTIONS"
	}
	methods := []string{}
	for _, m := range strings.Split(methodsRaw, ",") {
		if trimmed := strings.ToUpper(strings.TrimSpace(m)); trimmed != "" {
			methods = append(methods, trimmed)
		}
	}

	// Allow headers configurable via env.
	headersRaw := os.Getenv("CORS_ALLOW_HEADERS")
	if headersRaw == "" {
		headersRaw = "Origin,Content-Type,Accept,Authorization"
	}
	allowHeaders := []string{}
	for _, h := range strings.Split(headersRaw, ",") {
		if trimmed := strings.TrimSpace(h); trimmed != "" {
			allowHeaders = append(allowHeaders, trimmed)
		}
	}

	exposeRaw := os.Getenv("CORS_EXPOSE_HEADERS")
	if exposeRaw == "" {
		exposeRaw = "Content-Length"
	}
	exposeHeaders := []string{}
	for _, e := range strings.Split(exposeRaw, ",") {
		if trimmed := strings.TrimSpace(e); trimmed != "" {
			exposeHeaders = append(exposeHeaders, trimmed)
		}
	}

	maxAgeRaw := os.Getenv("CORS_MAX_AGE")
	maxAge := 12 * time.Hour
	if maxAgeRaw != "" {
		if parsed, err := time.ParseDuration(maxAgeRaw + "s"); err == nil { // seconds input
			maxAge = parsed
		}
	}

	r.engine.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     methods,
		AllowHeaders:     allowHeaders,
		ExposeHeaders:    exposeHeaders,
		AllowCredentials: allowCredentials,
		MaxAge:           maxAge,
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
<<<<<<< HEAD
	r.engine.GET("/health", controllers.HealthCheck)
=======
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "mentor-backend"})
	})

	// Add a more detailed health check that validates dependencies
	r.engine.GET("/health/ready", controllers.HealthCheckReady)
>>>>>>> origin/master
}

// setupActivityRoutes configures activity-related routes
func (r *Router) setupActivityRoutes() {
	r.engine.GET("/activities", controllers.ListActivities)
}

// setupDeviceRoutes configures all device-related routes
func (r *Router) setupDeviceRoutes() {
	// Device ingestion endpoints (devices will POST data here)
	r.engine.POST("/devices/register", controllers.RegisterDevice)
	r.engine.POST("/devices/metrics", controllers.UpdateDeviceMetric)
	r.engine.POST("/devices/processes", controllers.UpdateProcessList)
	r.engine.POST("/devices/activity", controllers.Activity)
	r.engine.POST("/devices/commands", controllers.CreateRemoteCommand)
	r.engine.POST("/devices/screenshots", controllers.StoreScreenshot)

	// Device query endpoints
	r.engine.GET("/devices", controllers.ListDevices)
	r.engine.GET("/devices/:id/metrics", controllers.GetDeviceMetric)
	r.engine.GET("/devices/:id/processes", controllers.GetDeviceProcesses)
	r.engine.GET("/devices/:id/activities", controllers.GetDeviceActivity)
	r.engine.GET("/devices/:id/alerts", controllers.GetDeviceAlert)
	r.engine.GET("/devices/:id/screenshots", controllers.GetDeviceScreenshot)
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
