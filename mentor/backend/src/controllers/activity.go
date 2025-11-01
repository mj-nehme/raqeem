package controllers

import (
	"mentor-backend/database"
	"mentor-backend/models"
	"mentor-backend/s3"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func ListActivities(c *gin.Context) {
	var activities []models.Activity
	query := database.DB

	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if location := c.Query("location"); location != "" {
		query = query.Where("location = ?", location)
	}
	if start := c.Query("start_time"); start != "" {
		startTime, _ := time.Parse(time.RFC3339, start)
		query = query.Where("timestamp >= ?", startTime)
	}
	if end := c.Query("end_time"); end != "" {
		endTime, _ := time.Parse(time.RFC3339, end)
		query = query.Where("timestamp <= ?", endTime)
	}

	query.Order("timestamp desc").Find(&activities)

	var result []gin.H
	for _, a := range activities {
		result = append(result, gin.H{
			"id":             a.ID,
			"user_id":        a.UserID,
			"location":       a.Location,
			"filename":       a.Filename,
			"timestamp":      a.Timestamp,
			"screenshot_url": s3.GeneratePresignedURL(a.Filename),
		})
	}

	c.JSON(http.StatusOK, result)
}
