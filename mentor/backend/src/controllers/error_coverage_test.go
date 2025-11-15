package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestActivityErrorHandling tests error paths in Activity function
func TestActivityErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Test with invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/activity", bytes.NewBufferString("{invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	Activity(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestReportAlertErrorHandling tests error paths in ReportAlert function
func TestReportAlertErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Test with invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/alert", bytes.NewBufferString("{invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	ReportAlert(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestStoreScreenshotErrorHandling tests error paths in StoreScreenshot function
func TestStoreScreenshotErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Test with invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/screenshot", bytes.NewBufferString("{invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	StoreScreenshot(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestUpdateCommandStatusErrorHandling tests error paths
func TestUpdateCommandStatusErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Test with invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "999"}}
	c.Request, _ = http.NewRequest("PUT", "/commands/999/status", bytes.NewBufferString("{invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateCommandStatus(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test with non-existent command - UpdateCommandStatus returns 200 even if command doesn't exist (idempotent)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	nonExistentID := uuid.New()
	c.Params = gin.Params{gin.Param{Key: "id", Value: nonExistentID.String()}}
	statusUpdate := models.DeviceRemoteCommand{
		CommandID: nonExistentID,
		Status:    "completed",
		Result:    "test result",
	}
	body, _ := json.Marshal(statusUpdate)
	c.Request, _ = http.NewRequest("PUT", "/commands/"+nonExistentID.String()+"/status", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateCommandStatus(c)
	assert.Equal(t, http.StatusOK, w.Code) // Returns 200 even if command doesn't exist (idempotent update)
}

// TestGetPendingCommandsErrorHandling tests error paths
func TestGetPendingCommandsErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Test with invalid device ID - returns 200 with empty array (graceful handling)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid-uuid"}}
	c.Request, _ = http.NewRequest("GET", "/devices/invalid-uuid/commands/pending", nil)

	GetPendingCommands(c)
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify it returns an empty array
	var commands []models.DeviceRemoteCommand
	err = json.Unmarshal(w.Body.Bytes(), &commands)
	assert.NoError(t, err)
	assert.Equal(t, 0, len(commands))
}

// TestCreateRemoteCommandErrorHandling tests error paths
func TestCreateRemoteCommandErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Test with invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: uuid.New().String()}}
	c.Request, _ = http.NewRequest("POST", "/commands", bytes.NewBufferString("{invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateRemoteCommand(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Test with valid JSON but empty command text - will create record successfully
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Params = gin.Params{gin.Param{Key: "id", Value: "invalid-uuid"}} // URL param not used
	command := models.DeviceRemoteCommand{
		DeviceID:    uuid.New(),
		CommandText: "test command",
	}
	body, _ := json.Marshal(command)
	c.Request, _ = http.NewRequest("POST", "/commands", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	CreateRemoteCommand(c)
	assert.Equal(t, http.StatusOK, w.Code) // Command creation succeeds with device ID from body
}

// TestUpdateDeviceMetricErrorHandling tests error paths
func TestUpdateDeviceMetricErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Test with invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/metrics", bytes.NewBufferString("{invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	UpdateDeviceMetric(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestRegisterDeviceErrorHandling tests error paths
func TestRegisterDeviceErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Test with invalid JSON
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/devices", bytes.NewBufferString("{invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	RegisterDevice(c)
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
