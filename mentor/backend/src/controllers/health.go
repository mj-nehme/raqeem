package controllers

import (
<<<<<<< HEAD
"mentor-backend/database"
"mentor-backend/s3"
"net/http"
"time"

"github.com/gin-gonic/gin"
)

// HealthStatus represents the health status of a service component
type HealthStatus struct {
Status  string                 `json:"status"`
Details map[string]interface{} `json:"details,omitempty"`
}

// HealthCheck performs comprehensive health checks on all dependencies
// @Summary Health check endpoint
// @Description Returns health status of the service and its dependencies
// @Tags health
// @Produce json
// @Success 200 {object} HealthStatus
// @Failure 503 {object} HealthStatus
// @Router /health [get]
func HealthCheck(c *gin.Context) {
startTime := time.Now()
allHealthy := true
details := make(map[string]interface{})

// Check database health
dbHealth := checkDatabaseHealth()
details["database"] = dbHealth
if dbHealth["status"] != "healthy" {
allHealthy = false
}

// Check MinIO/S3 health
s3Health := checkS3Health()
details["s3"] = s3Health
if s3Health["status"] != "healthy" {
allHealthy = false
}

// Add service metadata
details["service"] = "mentor-backend"
details["timestamp"] = time.Now().UTC().Format(time.RFC3339)
details["uptime_ms"] = time.Since(startTime).Milliseconds()

status := "healthy"
httpStatus := http.StatusOK
if !allHealthy {
status = "unhealthy"
httpStatus = http.StatusServiceUnavailable
}

c.JSON(httpStatus, HealthStatus{
Status:  status,
Details: details,
})
}

// checkDatabaseHealth checks if the database connection is healthy
func checkDatabaseHealth() map[string]interface{} {
result := make(map[string]interface{})

if err := database.HealthCheck(); err != nil {
result["status"] = "unhealthy"
result["error"] = err.Error()
return result
}

// Get connection pool stats
stats := database.GetStats()
result["status"] = "healthy"
result["open_connections"] = stats.OpenConnections
result["in_use"] = stats.InUse
result["idle"] = stats.Idle
result["max_open_connections"] = stats.MaxOpenConnections

return result
}

// checkS3Health checks if the S3/MinIO connection is healthy
func checkS3Health() map[string]interface{} {
result := make(map[string]interface{})

if err := s3.HealthCheck(); err != nil {
result["status"] = "unhealthy"
result["error"] = err.Error()
return result
}

result["status"] = "healthy"
return result
=======
	"mentor-backend/database"
	"mentor-backend/s3"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheckReady validates all critical dependencies are available
// @Summary Readiness health check
// @Description Validates database and MinIO connectivity
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 503 {object} map[string]interface{}
// @Router /health/ready [get]
func HealthCheckReady(c *gin.Context) {
	health := gin.H{
		"status":  "ok",
		"service": "mentor-backend",
	}

	checks := gin.H{}
	allHealthy := true

	// Check database
	if err := database.HealthCheck(); err != nil {
		checks["database"] = gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		allHealthy = false
	} else {
		checks["database"] = gin.H{
			"status": "healthy",
		}
	}

	// Check MinIO connectivity
	if err := s3.HealthCheck(); err != nil {
		checks["minio"] = gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
		allHealthy = false
	} else {
		checks["minio"] = gin.H{
			"status": "healthy",
		}
	}

	health["checks"] = checks

	if !allHealthy {
		health["status"] = "degraded"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	c.JSON(http.StatusOK, health)
>>>>>>> origin/master
}
