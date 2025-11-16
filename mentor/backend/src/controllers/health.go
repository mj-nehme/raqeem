package controllers

import (
	"mentor-backend/database"
	"mentor-backend/s3"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthCheckReady validates all critical dependencies are available.
// This endpoint performs comprehensive health checks on all required services:
// - Database connectivity: Verifies PostgreSQL is accessible and responding
// - MinIO storage: Verifies S3-compatible storage is accessible
//
// Returns 200 OK if all dependencies are healthy, or 503 Service Unavailable
// if any critical dependency is down. Load balancers and orchestrators should
// use this endpoint to determine if the service is ready to accept traffic.
//
// The response includes detailed status for each dependency to aid debugging
// during incidents. When any check fails, the overall status is set to "degraded"
// and a 503 status code is returned to signal the service is not ready.
//
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

	// Check database connectivity
	// This verifies the connection pool is healthy and PostgreSQL is responding
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
	// This verifies S3-compatible storage is accessible for screenshot storage
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

	// Return 503 if any dependency is unhealthy
	// This signals to load balancers and orchestrators that the service is not ready
	if !allHealthy {
		health["status"] = "degraded"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	c.JSON(http.StatusOK, health)
}
