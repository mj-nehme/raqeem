package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mentor-backend/database"
	"mentor-backend/models"
	"mentor-backend/s3"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterDevice registers a new device or updates existing device info
// @Summary Register a device
// @Description Register a new device or update existing device information
// @Tags devices
// @Accept json
// @Produce json
// @Param device body models.Device true "Device information"
// @Success 200 {object} models.Device
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices/register [post]
func RegisterDevice(c *gin.Context) {
	var device models.Device
	if err := c.BindJSON(&device); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	device.LastSeen = time.Now()
	device.IsOnline = true

	// Upsert device
	result := database.DB.Where("id = ?", device.ID).
		Assign(device).
		FirstOrCreate(&device)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, device)
}

// UpdateDeviceMetrics stores new device metrics
func UpdateDeviceMetrics(c *gin.Context) {
	var metrics models.DeviceMetrics
	if err := c.BindJSON(&metrics); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate UUID for the metrics record if not provided
	if metrics.ID == "" {
		metrics.ID = uuid.New().String()
	}

	metrics.Timestamp = time.Now()

	if err := database.DB.Create(&metrics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update device last seen
	database.DB.Model(&models.Device{}).
		Where("id = ?", metrics.DeviceID).
		Updates(map[string]interface{}{
			"last_seen": time.Now(),
			"is_online": true,
		})

	c.JSON(http.StatusOK, metrics)
}

// Activity stores a new activity log entry
func Activity(c *gin.Context) {
	var activity models.DeviceActivities
	if err := c.BindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	activity.Timestamp = time.Now()

	if err := database.DB.Create(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, activity)
}

// UpdateProcessList stores the current process list
func UpdateProcessList(c *gin.Context) {
	var processes []models.DeviceProcesses
	if err := c.BindJSON(&processes); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx := database.DB.Begin()

	// Delete old processes for this device
	if len(processes) > 0 {
		if err := tx.Where("device_id = ?", processes[0].DeviceID).Delete(&models.DeviceProcesses{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Insert new processes
	now := time.Now()
	for i := range processes {
		processes[i].Timestamp = now
		if err := tx.Create(&processes[i]).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, processes)
}

// ListDevices returns all registered devices
// @Summary List all devices
// @Description Get a list of all registered devices with their current status
// @Tags devices
// @Produce json
// @Success 200 {array} models.Device
// @Failure 500 {object} map[string]string
// @Router /devices [get]
func ListDevices(c *gin.Context) {
	devices := make([]models.Device, 0)

	// Mark devices as offline if not seen in last 5 minutes
	database.DB.Model(&models.Device{}).
		Where("last_seen < ?", time.Now().Add(-5*time.Minute)).
		Update("is_online", false)

	if err := database.DB.Find(&devices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, devices)
}

// GetDeviceMetrics returns metrics for a specific device
// @Summary Get device metrics
// @Description Get recent metrics for a specific device
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Param limit query int false "Number of records to return" default(60)
// @Success 200 {array} models.DeviceMetrics
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices/{id}/metrics [get]
func GetDeviceMetrics(c *gin.Context) {
	deviceID := c.Param("id")
	limit := 60 // Last hour by default, one point per minute
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
	}

	metrics := make([]models.DeviceMetrics, 0)
	if err := database.DB.Where("device_id = ?", deviceID).
		Order("timestamp desc").
		Limit(limit).
		Find(&metrics).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, metrics)
}

// GetDeviceProcesseses returns latest known processes for a specific device
func GetDeviceProcesseses(c *gin.Context) {
	deviceID := c.Param("id")
	limit := 100
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
	}

	processes := make([]models.DeviceProcesses, 0)
	// Return most recent snapshot of processes for device (ordered by cpu desc, then timestamp desc)
	if err := database.DB.Where("device_id = ?", deviceID).
		Order("timestamp desc, cpu desc").
		Limit(limit).
		Find(&processes).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, processes)
}

// GetDeviceActivities returns recent activity logs for a device
func GetDeviceActivities(c *gin.Context) {
	deviceID := c.Param("id")
	limit := 100
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
	}

	logs := make([]models.DeviceActivities, 0)
	if err := database.DB.Where("device_id = ?", deviceID).
		Order("timestamp desc").
		Limit(limit).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, logs)
}

// GetDeviceAlertss returns recent alerts for a device
// @Summary Get device alerts
// @Description Get recent alerts for a specific device
// @Tags devices
// @Produce json
// @Param id path string true "Device ID"
// @Param limit query int false "Number of records to return" default(100)
// @Success 200 {array} models.DeviceAlerts
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /devices/{id}/alerts [get]
func GetDeviceAlertss(c *gin.Context) {
	deviceID := c.Param("id")
	limit := 100
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
	}

	alerts := make([]models.DeviceAlerts, 0)
	if err := database.DB.Where("device_id = ?", deviceID).
		Order("timestamp desc").
		Limit(limit).
		Find(&alerts).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alerts)
}

// GetDeviceScreenshotss returns recent screenshots metadata for a device
func GetDeviceScreenshotss(c *gin.Context) {
	deviceID := c.Param("id")
	limit := 50
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
	}

	shots := make([]models.DeviceScreenshots, 0)
	if err := database.DB.Where("device_id = ?", deviceID).
		Order("timestamp desc").
		Limit(limit).
		Find(&shots).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate presigned URLs for screenshots
	resp := make([]gin.H, 0, len(shots))
	for _, s := range shots {
		// Generate presigned URL for the screenshot
		screenshotURL := s3.GeneratePresignedURL(s.Path)

		resp = append(resp, gin.H{
			"id":             s.ID,
			"device_id":      s.DeviceID,
			"timestamp":      s.Timestamp,
			"path":           s.Path,
			"resolution":     s.Resolution,
			"size":           s.Size,
			"url":            screenshotURL,
			"screenshot_url": screenshotURL, // Also provide as screenshot_url for frontend compatibility
		})
	}

	c.JSON(http.StatusOK, resp)
}

// CreateRemoteCommand queues a command for execution on a device
func CreateRemoteCommand(c *gin.Context) {
	var cmd models.DeviceRemoteCommands
	if err := c.BindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd.Status = "pending"
	cmd.CreatedAt = time.Now()

	if err := database.DB.Create(&cmd).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Forward command to devices backend if DEVICES_API_URL is set
	devicesAPIURL := os.Getenv("DEVICES_API_URL")
	if devicesAPIURL != "" {
		go func() {
			payload := map[string]interface{}{
				"command": cmd.Command,
			}
			jsonData, err := json.Marshal(payload)
			if err != nil {
				fmt.Printf("Error marshaling command payload: %v\n", err)
				return
			}
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Post(
				fmt.Sprintf("%s/devices/%s/commands", devicesAPIURL, cmd.DeviceID),
				"application/json",
				bytes.NewBuffer(jsonData),
			)
			if err != nil {
				fmt.Printf("Error forwarding command to devices backend: %v\n", err)
				return
			}
			defer func() {
				if err := resp.Body.Close(); err != nil {
					fmt.Printf("Error closing response body: %v\n", err)
				}
			}()
			if resp.StatusCode >= 400 {
				fmt.Printf("Devices backend returned error status: %d\n", resp.StatusCode)
			}
		}()
	}

	c.JSON(http.StatusOK, cmd)
}

// GetPendingCommands returns pending commands for a device
func GetPendingCommands(c *gin.Context) {
	deviceID := c.Param("id")

	commands := make([]models.DeviceRemoteCommands, 0)
	if err := database.DB.Where("device_id = ? AND status = ?", deviceID, "pending").
		Find(&commands).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commands)
}

// GetDeviceCommands returns command history for a device
func GetDeviceCommands(c *gin.Context) {
	deviceID := c.Param("id")
	limit := 100
	if l := c.Query("limit"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &limit); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid limit parameter"})
			return
		}
	}

	var commands []models.DeviceRemoteCommands
	if err := database.DB.Where("device_id = ?", deviceID).
		Order("created_at desc").
		Limit(limit).
		Find(&commands).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commands)
}

// UpdateCommandStatus updates command execution status
func UpdateCommandStatus(c *gin.Context) {
	var cmd models.DeviceRemoteCommands
	if err := c.BindJSON(&cmd); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if cmd.Status == "completed" || cmd.Status == "failed" {
		cmd.CompletedAt = time.Now()
	}

	if err := database.DB.Model(&models.DeviceRemoteCommands{}).
		Where("id = ?", cmd.ID).
		Updates(map[string]interface{}{
			"status":       cmd.Status,
			"result":       cmd.Result,
			"exit_code":    cmd.ExitCode,
			"completed_at": cmd.CompletedAt,
		}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cmd)
}

// ReportAlert stores a new device alert
func ReportAlert(c *gin.Context) {
	var alert models.DeviceAlerts
	if err := c.BindJSON(&alert); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	alert.Timestamp = time.Now()

	if err := database.DB.Create(&alert).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, alert)
}

// StoreScreenshot stores screenshot metadata forwarded from devices backend
func StoreScreenshot(c *gin.Context) {
	var screenshot models.DeviceScreenshots
	if err := c.BindJSON(&screenshot); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Timestamp is set by database default, no need to set here
	// This ensures consistency with other models and migration defaults

	if err := database.DB.Create(&screenshot).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, screenshot)
}

// ListActivities returns all device activities (global, not per device)
func ListActivities(c *gin.Context) {
	var activities []models.DeviceActivities
	if err := database.DB.Order("timestamp desc").Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, activities)
}
