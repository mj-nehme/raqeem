package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetPendingCommandsSuccess tests GetPendingCommands with valid database
func TestGetPendingCommandsSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Set up a proper test database
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Save original and restore after test
	originalDB := database.DB
	t.Cleanup(func() {
		database.DB = originalDB
	})
	database.DB = db

	t.Run("Valid query returns empty array for device with no commands", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "valid-device"}}
		c.Request, _ = http.NewRequest("GET", "/devices/valid-device/commands/pending", nil)

		GetPendingCommands(c)

		// Should return 200 OK with empty array when database is working
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "[]")
	})

	t.Run("Valid query returns pending commands", func(t *testing.T) {
		// Create a test device and command
		testDevice := models.Device{
			DeviceID:   sampleUUID,
			DeviceName: "test-pending-device",
		}
		db.Create(&testDevice)

		testCmd := models.DeviceRemoteCommand{
			DeviceID:    sampleUUID,
			CommandText: "get_info",
			Status:      "pending",
		}
		db.Create(&testCmd)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: testDevice.DeviceID.String()}}
		c.Request, _ = http.NewRequest("GET", "/devices/"+testDevice.DeviceID.String()+"/commands/pending", nil)

		GetPendingCommands(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "get_info")
	})
}
