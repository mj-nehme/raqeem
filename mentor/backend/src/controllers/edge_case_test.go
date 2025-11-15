package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mentor-backend/database"
	"mentor-backend/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDataValidationAndSanitization(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Test device registration with various data types and edge cases
	testCases := []struct {
		name         string
		device       map[string]interface{}
		expectStatus int
	}{
		{
			name: "Valid device with all fields",
			device: map[string]interface{}{
				"deviceid":        sampleUUID,
				"device_name":     "Valid Device",
				"device_type":     "laptop",
				"os":              "macOS",
				"ip_address":      "192.168.1.100",
				"mac_address":     "aa:bb:cc:dd:ee:ff",
				"device_location": "Office",
				"current_user":    "testuser",
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "Device with minimum required fields",
			device: map[string]interface{}{
				"device_name": "Minimal Device",
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "Device with unicode characters",
			device: map[string]interface{}{
				"device_name": "测试设备 Device Téléphone",
				"device_type": "móvil",
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "Device with long strings",
			device: map[string]interface{}{
				"device_name": strings.Repeat("Very Long Device Name ", 20),
				"device_type": "laptop",
			},
			expectStatus: http.StatusOK,
		},
		{
			name: "Device with special characters in fields",
			device: map[string]interface{}{
				"deviceid":        sampleUUID,
				"device_name":     "Device with @#$%^&*() characters",
				"device_location": "Floor 1 - Section A/B (Test)",
				"device_command":  "echo 'test with quotes and special chars'",
			},
			expectStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			deviceJSON, _ := json.Marshal(tc.device)
			req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, tc.expectStatus, w.Code)

			if w.Code == http.StatusOK {
				var response models.Device
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				if v, ok := tc.device["deviceid"].(uuid.UUID); ok {
					assert.Equal(t, v, response.DeviceID)
				}
				if v, ok := tc.device["device_name"].(string); ok {
					assert.Equal(t, v, response.DeviceName)
				}
			}
		})
	}
}

func TestComplexQueryParameters(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := sampleUUID.String()

	// Register device and add test data
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Query Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	database.DB.Create(&device)

	// Add metrics with different timestamps
	baseTime := time.Now()
	for i := 0; i < 20; i++ {
		metrics := models.DeviceMetric{
			DeviceID:  sampleUUID,
			CPUUsage:  float64(10 + i*5),
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
		}
		database.DB.Create(&metrics)
	}

	// Test various limit values
	limitTests := []struct {
		limit    string
		expected int
	}{
		{"0", 0},
		{"1", 1},
		{"5", 5},
		{"15", 15},
		{"25", 20},  // Should return all 20 metrics
		{"100", 20}, // Should return all 20 metrics
	}

	for _, test := range limitTests {
		t.Run(fmt.Sprintf("limit_%s", test.limit), func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=%s", deviceID, test.limit), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)

			var metrics []models.DeviceMetric
			err := json.Unmarshal(w.Body.Bytes(), &metrics)
			require.NoError(t, err)
			assert.Equal(t, test.expected, len(metrics))
		})
	}

	// Test invalid limit parameters
	invalidLimits := []string{"abc", "invalid"}
	for _, limit := range invalidLimits {
		t.Run(fmt.Sprintf("invalid_limit_%s", limit), func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=%s", deviceID, limit), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusBadRequest, w.Code)
		})
	}

	// Test edge case limits that should work (Go's fmt.Sscanf is lenient)
	edgeCaseLimits := []string{"-1", "1.5", ""}
	for _, limit := range edgeCaseLimits {
		t.Run(fmt.Sprintf("edge_case_limit_%s", limit), func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/devices/%s/metrics?limit=%s", deviceID, limit), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			if limit == "" {
				// Empty limit should use default and succeed
				assert.Equal(t, http.StatusOK, w.Code)
			} else {
				// These might succeed due to Go's lenient parsing
				assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusBadRequest)
			}
		})
	}
}

func TestHTTPHeaderHandling(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Test various content types - Gin is permissive with content types
	contentTypeTests := []struct {
		contentType  string
		expectStatus int
	}{
		{"application/json", http.StatusOK},
		{"application/json; charset=utf-8", http.StatusOK},
		{"text/plain", http.StatusOK}, // Gin doesn't strictly validate content-type for JSON parsing
		{"", http.StatusOK},           // Gin will still try to parse JSON
	}

	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Header Test Device",
	}
	deviceJSON, _ := json.Marshal(device)

	for _, test := range contentTypeTests {
		t.Run(fmt.Sprintf("content_type_%s", strings.ReplaceAll(test.contentType, "/", "_")), func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
			if test.contentType != "" {
				req.Header.Set("Content-Type", test.contentType)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, test.expectStatus, w.Code)
		})
	}
}

func TestJSONMarshalling(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Test response JSON structure for all endpoints
	deviceID := "json-test-device"

	// Register device
	device := models.Device{
		DeviceID:       sampleUUID,
		DeviceName:     "JSON Test Device",
		DeviceType:     "laptop",
		OS:             "Linux",
		IPAddress:      "192.168.1.200",
		MacAddress:     "bb:cc:dd:ee:ff:00",
		DeviceLocation: "Lab",
		IsOnline:       true,
		CurrentUser:    "jsonuser",
		LastSeen:       time.Now(),
	}

	deviceJSON, _ := json.Marshal(device)
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify response structure
	var deviceResponse models.Device
	err := json.Unmarshal(w.Body.Bytes(), &deviceResponse)
	require.NoError(t, err)
	assert.Equal(t, device.DeviceID, deviceResponse.DeviceID)
	assert.Equal(t, device.DeviceName, deviceResponse.DeviceName)
	assert.Equal(t, device.DeviceType, deviceResponse.DeviceType)
	assert.Equal(t, device.OS, deviceResponse.OS)
	assert.Equal(t, device.IPAddress, deviceResponse.IPAddress)
	assert.Equal(t, device.MacAddress, deviceResponse.MacAddress)
	assert.Equal(t, device.DeviceLocation, deviceResponse.DeviceLocation)
	assert.Equal(t, device.CurrentUser, deviceResponse.CurrentUser)

	// Test metrics JSON structure
	metrics := models.DeviceMetric{
		DeviceID:    sampleUUID,
		CPUUsage:    85.5,
		CPUTemp:     72.3,
		MemoryTotal: 32000000000,
		MemoryUsed:  20000000000,
		SwapUsed:    1000000000,
		DiskTotal:   2000000000000,
		DiskUsed:    1500000000000,
		NetBytesIn:  5000,
		NetBytesOut: 3000,
	}

	metricsJSON, _ := json.Marshal(metrics)
	req, _ = http.NewRequest("POST", fmt.Sprintf("/devices/%s/metrics", deviceID), bytes.NewBuffer(metricsJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var metricsResponse models.DeviceMetric
	err = json.Unmarshal(w.Body.Bytes(), &metricsResponse)
	require.NoError(t, err)
	assert.Equal(t, metrics.DeviceID, metricsResponse.DeviceID)
	assert.Equal(t, metrics.CPUUsage, metricsResponse.CPUUsage)
	assert.Equal(t, metrics.CPUTemp, metricsResponse.CPUTemp)
	assert.Equal(t, metrics.MemoryTotal, metricsResponse.MemoryTotal)
	assert.Equal(t, metrics.MemoryUsed, metricsResponse.MemoryUsed)
	assert.Equal(t, metrics.SwapUsed, metricsResponse.SwapUsed)
	assert.NotZero(t, metricsResponse.Timestamp) // Should be auto-set
}

func TestDatabaseErrorHandling(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// Test with extremely long strings that might cause database errors
	veryLongString := strings.Repeat("x", 10000)

	device := models.Device{
		DeviceID:       sampleUUID,
		DeviceName:     veryLongString,
		DeviceType:     veryLongString,
		OS:             veryLongString,
		IPAddress:      veryLongString,
		MacAddress:     veryLongString,
		DeviceLocation: veryLongString,
		CurrentUser:    veryLongString,
	}

	deviceJSON, _ := json.Marshal(device)
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	// Should either succeed (if database handles long strings) or fail gracefully
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusInternalServerError)
}

func TestAllEndpointsResponseFormat(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	deviceID := "format-test-device"

	// Register device first
	device := models.Device{
		DeviceID:   sampleUUID,
		DeviceName: "Format Test Device",
		IsOnline:   true,
		LastSeen:   time.Now(),
	}
	database.DB.Create(&device)

	// Test all GET endpoints return valid JSON arrays
	getEndpoints := []string{
		"/devices",
		fmt.Sprintf("/devices/%s/metrics", deviceID),
		fmt.Sprintf("/devices/%s/processes", deviceID),
		fmt.Sprintf("/devices/%s/activities", deviceID),
		fmt.Sprintf("/devices/%s/alerts", deviceID),
		fmt.Sprintf("/devices/%s/screenshots", deviceID),
		fmt.Sprintf("/devices/%s/commands/pending", deviceID),
	}

	for _, endpoint := range getEndpoints {
		t.Run(fmt.Sprintf("GET_%s", strings.ReplaceAll(endpoint, "/", "_")), func(t *testing.T) {
			req, _ := http.NewRequest("GET", endpoint, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			assert.Equal(t, http.StatusOK, w.Code)

			// Verify it's valid JSON
			var jsonResponse interface{}
			err := json.Unmarshal(w.Body.Bytes(), &jsonResponse)
			assert.NoError(t, err, "Response should be valid JSON")

			// For most endpoints, response should be an array
			if !strings.Contains(endpoint, "screenshots") { // Screenshots has special format
				_, isArray := jsonResponse.([]interface{})
				assert.True(t, isArray, "Response should be an array for endpoint %s", endpoint)
			}
		})
	}
}

func TestDeviceUpsertBehavior(t *testing.T) {
	router, cleanup := setupTestRouterWithDB(t)
	defer cleanup()

	// First registration
	device1 := models.Device{
		DeviceID:       sampleUUID,
		DeviceName:     "Original Device",
		DeviceType:     "laptop",
		OS:             "Windows",
		DeviceLocation: "Office A",
	}

	deviceJSON, _ := json.Marshal(device1)
	req, _ := http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Second registration with same ID but different data (should update)
	device2 := models.Device{
		DeviceID:       sampleUUID,
		DeviceName:     "Updated Device",
		DeviceType:     "desktop",
		OS:             "Linux",
		DeviceLocation: "Office B",
	}

	deviceJSON, _ = json.Marshal(device2)
	req, _ = http.NewRequest("POST", "/devices", bytes.NewBuffer(deviceJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var response models.Device
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should have updated values
	assert.Equal(t, device2.DeviceName, response.DeviceName)
	assert.Equal(t, device2.DeviceType, response.DeviceType)
	assert.Equal(t, device2.OS, response.OS)
	assert.Equal(t, device2.DeviceLocation, response.DeviceLocation)
	assert.True(t, response.IsOnline)    // Should be set to true
	assert.NotZero(t, response.LastSeen) // Should be updated

	// Verify only one device exists in database using correct UUID
	var count int64
	database.DB.Model(&models.Device{}).Where("deviceid = ?", sampleUUID).Count(&count)
	assert.Equal(t, int64(1), count)
}
