package controllers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
)

func setupTestDB(t *testing.T) {
	os.Setenv("POSTGRES_USER", os.Getenv("POSTGRES_USER"))
	os.Setenv("POSTGRES_PASSWORD", os.Getenv("POSTGRES_PASSWORD"))
	os.Setenv("POSTGRES_DB", os.Getenv("POSTGRES_DB"))
	os.Setenv("POSTGRES_HOST", os.Getenv("POSTGRES_HOST"))
	os.Setenv("POSTGRES_PORT", os.Getenv("POSTGRES_PORT"))
	database.Connect()
	// Auto-migrate tables
	database.DB.AutoMigrate(&models.Alert{})
}

func TestReportAndGetAlerts(t *testing.T) {
	if os.Getenv("POSTGRES_HOST") == "" {
		t.Skip("POSTGRES_* env vars not set; skipping integration test")
	}
	setupTestDB(t)

	// Ensure clean slate for test device
	deviceID := "test-device-go"
	database.DB.Where("device_id = ?", deviceID).Delete(&models.Alert{})

	// Prepare gin context for ReportAlert
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/devices/"+deviceID+"/alerts", nil)

	alert := models.Alert{
		DeviceID:  deviceID,
		Timestamp: time.Now(),
		Level:     "warning",
		Type:      "cpu_high",
		Message:   "CPU high",
		Value:     95,
		Threshold: 80,
	}
	b, _ := json.Marshal(alert)
	c.Request.Body = io.NopCloser(bytes.NewReader(b))
	c.Request.Header.Set("Content-Type", "application/json")

	ReportAlert(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Prepare gin context for GetDeviceAlerts
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Params = gin.Params{{Key: "id", Value: deviceID}}
	c2.Request, _ = http.NewRequest("GET", "/devices/"+deviceID+"/alerts", nil)

	GetDeviceAlerts(c2)
	if w2.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w2.Code)
	}
}
