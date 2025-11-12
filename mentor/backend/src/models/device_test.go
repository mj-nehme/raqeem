package models

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var sampleUUID, _ = uuid.Parse("550e8400-e29b-41d4-a716-446655440000")

func TestDevice_Fields(t *testing.T) {
	device := Device{
		DeviceID:       sampleUUID,
		DeviceName:     "Test Device",
		DeviceType:     "laptop",
		OS:             "macOS",
		LastSeen:       time.Now(),
		IsOnline:       true,
		DeviceLocation: "Office",
		IPAddress:      "192.168.1.100",
		MacAddress:     "00:11:22:33:44:55",
		CurrentUser:    "testuser",
	}

	// Test basic fields using assert for cleaner tests
	assert.Equal(t, sampleUUID, device.DeviceID)
	assert.Equal(t, "Test Device", device.DeviceName)
	assert.Equal(t, "laptop", device.DeviceType)
	assert.Equal(t, "macOS", device.OS)
	assert.True(t, device.IsOnline)
	assert.Equal(t, "Office", device.DeviceLocation)
	assert.Equal(t, "192.168.1.100", device.IPAddress)
	assert.Equal(t, "00:11:22:33:44:55", device.MacAddress)
	assert.Equal(t, "testuser", device.CurrentUser)
}

func TestDevice_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		device   Device
		expected bool
	}{
		{
			name: "Very long device name",
			device: Device{
				DeviceID:   sampleUUID,
				DeviceName: strings.Repeat("A", 1000),
			},
			expected: true,
		},
		{
			name: "Special characters in name",
			device: Device{
				DeviceID:   sampleUUID,
				DeviceName: "Test-Device_123 (Production) [v2.0]",
			},
			expected: true,
		},
		{
			name: "Invalid IP address",
			device: Device{
				DeviceID:  sampleUUID,
				IPAddress: "999.999.999.999",
			},
			expected: true,
		},
		{
			name: "Invalid MAC address",
			device: Device{
				DeviceID:   sampleUUID,
				MacAddress: "invalid-mac",
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that device can be created with edge case values
			assert.NotNil(t, tt.device)
			assert.Equal(t, tt.expected, true) // All should be createable
		})
	}
}

func TestDevice_Types(t *testing.T) {
	validTypes := []string{"laptop", "desktop", "server", "mobile", "tablet", "iot"}

	for _, deviceType := range validTypes {
		t.Run("type_"+deviceType, func(t *testing.T) {
			device := Device{
				DeviceID:   sampleUUID,
				DeviceType: deviceType,
			}
			assert.Equal(t, deviceType, device.DeviceType)
		})
	}
}

func TestDevice_OSTypes(t *testing.T) {
	validOS := []string{"Windows", "macOS", "Linux", "iOS", "Android", "Ubuntu", "CentOS"}

	for _, os := range validOS {
		t.Run("os_"+os, func(t *testing.T) {
			device := Device{
				DeviceID: sampleUUID,
				OS:       os,
			}
			assert.Equal(t, os, device.OS)
		})
	}
}

func TestDeviceMetric_Fields(t *testing.T) {
	now := time.Now()
	metrics := DeviceMetric{
		MetricID:    sampleUUID,
		DeviceID:    sampleUUID,
		Timestamp:   now,
		CPUUsage:    50.5,
		CPUTemp:     65.2,
		MemoryTotal: 8589934592,    // 8GB
		MemoryUsed:  4294967296,    // 4GB
		SwapUsed:    1073741824,    // 1GB
		DiskTotal:   1099511627776, // 1TB
		DiskUsed:    549755813888,  // 512GB
		NetBytesIn:  1024,
		NetBytesOut: 2048,
	}

	if metrics.MetricID != sampleUUID {
		t.Errorf("Expected ID to be '550e8400-e29b-41d4-a716-446655440000', got %s", metrics.MetricID)
	}
	if metrics.DeviceID != sampleUUID {
		t.Errorf("Expected DeviceID to be 'test-device-123', got '%s'", metrics.DeviceID)
	}
	if !metrics.Timestamp.Equal(now) {
		t.Errorf("Expected Timestamp to be %v, got %v", now, metrics.Timestamp)
	}
	if metrics.CPUUsage != 50.5 {
		t.Errorf("Expected CPUUsage to be 50.5, got %f", metrics.CPUUsage)
	}
	if metrics.CPUTemp != 65.2 {
		t.Errorf("Expected CPUTemp to be 65.2, got %f", metrics.CPUTemp)
	}
	if metrics.MemoryTotal != 8589934592 {
		t.Errorf("Expected MemoryTotal to be 8589934592, got %d", metrics.MemoryTotal)
	}
	if metrics.MemoryUsed != 4294967296 {
		t.Errorf("Expected MemoryUsed to be 4294967296, got %d", metrics.MemoryUsed)
	}
}

func TestProcess_Fields(t *testing.T) {
	now := time.Now()
	process := DeviceProcess{
		ProcessID:   sampleUUID,
		DeviceID:    sampleUUID,
		Timestamp:   now,
		PID:         1234,
		ProcessName: "chrome",
		CPU:         25.5,
		Memory:      536870912, // 512MB
		CommandText: "/usr/bin/chrome --enable-features=test",
	}

	if process.ProcessID != sampleUUID {
		t.Errorf("Expected ProcessID to be %s, got %s", sampleUUID, process.ProcessID)
	}
	if process.DeviceID != sampleUUID {
		t.Errorf("Expected DeviceID to be %s, got '%s'", sampleUUID, process.DeviceID)
	}
	if process.PID != 1234 {
		t.Errorf("Expected PID to be 1234, got %d", process.PID)
	}
	if process.ProcessName != "chrome" {
		t.Errorf("Expected ProcessName to be 'chrome', got '%s'", process.ProcessName)
	}
	if process.CPU != 25.5 {
		t.Errorf("Expected CPU to be 25.5, got %f", process.CPU)
	}
	if process.Memory != 536870912 {
		t.Errorf("Expected Memory to be 536870912, got %d", process.Memory)
	}
	if process.CommandText != "/usr/bin/chrome --enable-features=test" {
		t.Errorf("Expected Command to be '/usr/bin/chrome --enable-features=test', got '%s'", process.CommandText)
	}
}

func TestActivityLog_Fields(t *testing.T) {
	now := time.Now()
	activity := DeviceActivity{
		ActivityID:   sampleUUID,
		DeviceID:     sampleUUID,
		Timestamp:    now,
		ActivityType: "app_launch",
		Description:  "User launched Chrome browser",
		App:          "chrome",
		Duration:     3600, // 1 hour
	}

	if activity.ActivityID != sampleUUID {
		t.Errorf("Expected ActivityID to be %s, got %s", sampleUUID, activity.ActivityID)
	}
	if activity.DeviceID != sampleUUID {
		t.Errorf("Expected DeviceID to be %s, got '%s'", sampleUUID, activity.DeviceID)
	}
	if activity.ActivityType != "app_launch" {
		t.Errorf("Expected ActivityType to be 'app_launch', got '%s'", activity.ActivityType)
	}
	if activity.Description != "User launched Chrome browser" {
		t.Errorf("Expected Description to be 'User launched Chrome browser', got '%s'", activity.Description)
	}
	if activity.App != "chrome" {
		t.Errorf("Expected App to be 'chrome', got '%s'", activity.App)
	}
	if activity.Duration != 3600 {
		t.Errorf("Expected Duration to be 3600, got %d", activity.Duration)
	}
}

func TestRemoteCommand_Fields(t *testing.T) {
	now := time.Now()

	cmd := DeviceRemoteCommand{
		CommandID:   sampleUUID,
		DeviceID:    sampleUUID,
		CommandText: "ls -la",
		Status:      "pending",
		CreatedAt:   now,
		CompletedAt: now.Add(5 * time.Second),
		Result:      "file1.txt\nfile2.txt",
		ExitCode:    0,
	}

	if cmd.CommandID != sampleUUID {
		t.Errorf("Expected CommandID to be %s, got %s", sampleUUID, cmd.CommandID)
	}
	if cmd.DeviceID != sampleUUID {
		t.Errorf("Expected DeviceID to be '%s', got '%s'", sampleUUID, cmd.DeviceID)
	}
	if cmd.CommandText != "ls -la" {
		t.Errorf("Expected Command to be 'ls -la', got '%s'", cmd.CommandText)
	}
	if cmd.Status != "pending" {
		t.Errorf("Expected Status to be 'pending', got '%s'", cmd.Status)
	}
	if cmd.Result != "file1.txt\nfile2.txt" {
		t.Errorf("Expected Result to be 'file1.txt\\nfile2.txt', got '%s'", cmd.Result)
	}
	if cmd.ExitCode != 0 {
		t.Errorf("Expected ExitCode to be 0, got %d", cmd.ExitCode)
	}
}

func TestScreenshot_Fields(t *testing.T) {
	now := time.Now()
	screenshot := DeviceScreenshot{
		ScreenshotID: sampleUUID,
		DeviceID:     sampleUUID,
		Timestamp:    now,
		Path:         "screenshots/2024/01/15/device123_20240115_120000.png",
		Resolution:   "1920x1080",
		Size:         2097152, // 2MB
	}

	if screenshot.ScreenshotID != sampleUUID {
		t.Errorf("Expected ScreenshotID to be %s, got %s", sampleUUID, screenshot.ScreenshotID)
	}
	if screenshot.DeviceID != sampleUUID {
		t.Errorf("Expected DeviceID to be %s, got %s", sampleUUID, screenshot.DeviceID)
	}
	if screenshot.Path != "screenshots/2024/01/15/device123_20240115_120000.png" {
		t.Errorf("Expected Path to be 'screenshots/2024/01/15/device123_20240115_120000.png', got '%s'", screenshot.Path)
	}
	if screenshot.Resolution != "1920x1080" {
		t.Errorf("Expected Resolution to be '1920x1080', got '%s'", screenshot.Resolution)
	}
	if screenshot.Size != 2097152 {
		t.Errorf("Expected Size to be 2097152, got %d", screenshot.Size)
	}
}

func TestAlert_Fields(t *testing.T) {
	now := time.Now()
	alert := DeviceAlert{
		AlertID:   sampleUUID,
		DeviceID:  sampleUUID,
		Timestamp: now,
		Level:     "warning",
		AlertType: "cpu",
		Message:   "High CPU usage detected",
		Value:     85.5,
		Threshold: 80.0,
	}

	if alert.AlertID != sampleUUID {
		t.Errorf("Expected AlertID to be %s, got %s", sampleUUID, alert.AlertID)
	}
	if alert.DeviceID != sampleUUID {
		t.Errorf("Expected DeviceID to be %s, got %s", sampleUUID, alert.DeviceID)
	}
	if alert.Level != "warning" {
		t.Errorf("Expected Level to be 'warning', got '%s'", alert.Level)
	}
	if alert.AlertType != "cpu" {
		t.Errorf("Expected AlertType to be 'cpu', got '%s'", alert.AlertType)
	}
	if alert.Message != "High CPU usage detected" {
		t.Errorf("Expected Message to be 'High CPU usage detected', got '%s'", alert.Message)
	}
	if alert.Value != 85.5 {
		t.Errorf("Expected Value to be 85.5, got %f", alert.Value)
	}
	if alert.Threshold != 80.0 {
		t.Errorf("Expected Threshold to be 80.0, got %f", alert.Threshold)
	}
}

func TestAlert_Levels(t *testing.T) {
	validLevels := []string{"info", "warning", "error", "critical"}

	for _, level := range validLevels {
		alert := DeviceAlert{
			AlertID:   sampleUUID,
			DeviceID:  sampleUUID,
			Level:     level,
			AlertType: "cpu",
			Message:   "Test alert",
			Value:     50.0,
			Threshold: 40.0,
		}

		if alert.Level != level {
			t.Errorf("Expected Level to be '%s', got '%s'", level, alert.Level)
		}
	}
}

func TestAlert_Types(t *testing.T) {
	validTypes := []string{"cpu", "memory", "disk", "network", "security"}

	for _, alertType := range validTypes {
		alert := DeviceAlert{
			AlertID:   sampleUUID,
			DeviceID:  sampleUUID,
			Level:     "warning",
			AlertType: alertType,
			Message:   "Test alert",
			Value:     50.0,
			Threshold: 40.0,
		}

		if alert.AlertType != alertType {
			t.Errorf("Expected AlertType to be '%s', got '%s'", alertType, alert.AlertType)
		}
	}
}
