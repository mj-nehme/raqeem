package controllers

import (
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
}
