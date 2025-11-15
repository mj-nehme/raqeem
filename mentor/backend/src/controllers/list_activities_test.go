package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListActivities(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, err := database.SetupTestDB(t)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer database.CleanupTestDB(t, db)
	database.DB = db

	// Create test activities
	activities := []models.DeviceActivity{
		{
			DeviceID:     uuid.New(),
			ActivityType: "login",
			Description:  "User logged in",
			App:          "system",
			Duration:     100,
			Timestamp:    time.Now().Add(-2 * time.Hour),
		},
		{
			DeviceID:     uuid.New(),
			ActivityType: "app_launch",
			Description:  "Launched Firefox",
			App:          "Firefox",
			Duration:     3600,
			Timestamp:    time.Now().Add(-1 * time.Hour),
		},
		{
			DeviceID:     uuid.New(),
			ActivityType: "file_access",
			Description:  "Opened document",
			App:          "Word",
			Duration:     1800,
			Timestamp:    time.Now(),
		},
	}

	for _, activity := range activities {
		db.Create(&activity)
	}

	// Test ListActivities
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/activities", nil)

	ListActivities(c)

	assert.Equal(t, http.StatusOK, w.Code)

	var result []models.DeviceActivity
	err = json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 3)

	// Verify they're ordered by timestamp descending
	if len(result) >= 2 {
		assert.True(t, result[0].Timestamp.After(result[1].Timestamp) || result[0].Timestamp.Equal(result[1].Timestamp))
	}
}
