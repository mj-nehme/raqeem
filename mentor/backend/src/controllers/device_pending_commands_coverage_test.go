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
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestGetPendingCommandsDatabaseError tests GetPendingCommands with database error
func TestGetPendingCommandsDatabaseError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	// Save the original DB connection
	originalDB := database.DB
	
	// Set up test to restore DB connection after test completes
	defer func() {
		database.DB = originalDB
	}()
	
	t.Run("Database error returns internal server error", func(t *testing.T) {
		// Create a temporary database without the RemoteCommand table to trigger an error
		tempDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		
		// Explicitly do NOT migrate the RemoteCommand table to cause a query error
		// This simulates a missing table scenario
		
		database.DB = tempDB
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "test-device"}}
		c.Request, _ = http.NewRequest("GET", "/devices/test-device/commands/pending", nil)
		
		GetPendingCommands(c)
		
		// Should return 500 Internal Server Error when database query fails
		assert.Equal(t, http.StatusInternalServerError, w.Code)
		
		// Response should contain an error message
		assert.Contains(t, w.Body.String(), "error")
	})
	
	t.Run("Database error with invalid table triggers error path", func(t *testing.T) {
		// Create a database with the wrong schema to trigger an error
		tempDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
		require.NoError(t, err)
		
		// Create a different table structure that will cause query errors
		type WrongSchema struct {
			ID   uint
			Name string
		}
		err = tempDB.AutoMigrate(&WrongSchema{})
		require.NoError(t, err)
		
		// Try to query remote_commands table which doesn't exist in the schema
		database.DB = tempDB
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "another-device"}}
		c.Request, _ = http.NewRequest("GET", "/devices/another-device/commands/pending", nil)
		
		GetPendingCommands(c)
		
		// Should return 500 Internal Server Error when table doesn't exist
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
	
	t.Run("Valid query after database error recovery", func(t *testing.T) {
		// Set up a proper test database
		db := database.SetupTestDB(t)
		require.NotNil(t, db)
		defer database.CleanupTestDB(t, db)
		database.DB = db
		
		// Ensure the table is migrated
		err := db.AutoMigrate(&models.RemoteCommand{})
		require.NoError(t, err)
		
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "valid-device"}}
		c.Request, _ = http.NewRequest("GET", "/devices/valid-device/commands/pending", nil)
		
		GetPendingCommands(c)
		
		// Should return 200 OK with empty array when database is working
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "[]")
	})
}
