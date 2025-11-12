package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestDeviceValidation(t *testing.T) {
	tests := []struct {
		name     string
		device   Device
		wantErrs int
	}{
		{
			name: "valid device",
			device: Device{
				DeviceID:   sampleUUID,
				DeviceName: "Test Device",
				DeviceType: "laptop",
			},
			wantErrs: 0,
		},
		{
			name: "empty ID",
			device: Device{
				DeviceID:   uuid.UUID{},
				DeviceName: "Test Device",
				DeviceType: "laptop",
			},
			wantErrs: 1,
		},
		{
			name: "empty name",
			device: Device{
				DeviceID:   sampleUUID,
				DeviceName: "",
				DeviceType: "laptop",
			},
			wantErrs: 1,
		},
		{
			name: "invalid type",
			device: Device{
				DeviceID:   sampleUUID,
				DeviceName: "Test Device",
				DeviceType: "invalid",
			},
			wantErrs: 1,
		},
		{
			name: "name too long",
			device: Device{
				DeviceID:   sampleUUID,
				DeviceName: string(make([]byte, 256)), // 256 characters
				DeviceType: "laptop",
			},
			wantErrs: 1,
		},
		{
			name: "multiple errors",
			device: Device{
				DeviceID:   uuid.UUID{},
				DeviceName: "",
				DeviceType: "invalid",
			},
			wantErrs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.device.ValidateDevice()
			assert.Equal(t, tt.wantErrs, len(errors))
		})
	}
}

func TestDeviceIsOnlineRecently(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		device   Device
		minutes  int
		expected bool
	}{
		{
			name: "online now",
			device: Device{
				IsOnline: true,
				LastSeen: now,
			},
			minutes:  5,
			expected: true,
		},
		{
			name: "offline but seen recently",
			device: Device{
				IsOnline: false,
				LastSeen: now.Add(-2 * time.Minute),
			},
			minutes:  5,
			expected: true,
		},
		{
			name: "offline and not seen recently",
			device: Device{
				IsOnline: false,
				LastSeen: now.Add(-10 * time.Minute),
			},
			minutes:  5,
			expected: false,
		},
		{
			name: "zero minutes should check IsOnline",
			device: Device{
				IsOnline: true,
				LastSeen: now.Add(-10 * time.Minute),
			},
			minutes:  0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.device.IsOnlineRecently(tt.minutes)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeviceGetFormattedType(t *testing.T) {
	tests := []struct {
		name     string
		device   Device
		expected string
	}{
		{"laptop", Device{DeviceType: "laptop"}, "Laptop"},
		{"DESKTOP", Device{DeviceType: "DESKTOP"}, "Desktop"},
		{"MoBiLe", Device{DeviceType: "MoBiLe"}, "Mobile"},
		{"empty", Device{DeviceType: ""}, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.device.GetFormattedType()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAlertValidation(t *testing.T) {
	tests := []struct {
		name     string
		alert    DeviceAlert
		wantErrs int
	}{
		{
			name: "valid alert",
			alert: DeviceAlert{
				DeviceID: sampleUUID,
				Level:    "warning",
				Type:     "cpu",
				Message:  "High CPU usage",
			},
			wantErrs: 0,
		},
		{
			name: "empty device ID",
			alert: DeviceAlert{
				DeviceID: uuid.UUID{},
				Level:    "warning",
				Type:     "cpu",
				Message:  "High CPU usage",
			},
			wantErrs: 1,
		},
		{
			name: "invalid level",
			alert: DeviceAlert{
				DeviceID: sampleUUID,
				Level:    "invalid",
				Type:     "cpu",
				Message:  "High CPU usage",
			},
			wantErrs: 1,
		},
		{
			name: "invalid type",
			alert: DeviceAlert{
				DeviceID: sampleUUID,
				Level:    "warning",
				Type:     "invalid",
				Message:  "High CPU usage",
			},
			wantErrs: 1,
		},
		{
			name: "empty message",
			alert: DeviceAlert{
				DeviceID: sampleUUID,
				Level:    "warning",
				Type:     "cpu",
				Message:  "",
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.alert.ValidateAlert()
			assert.Equal(t, tt.wantErrs, len(errors))
		})
	}
}

func TestAlertIsCritical(t *testing.T) {
	tests := []struct {
		name     string
		alert    DeviceAlert
		expected bool
	}{
		{"critical", DeviceAlert{Level: "critical"}, true},
		{"CRITICAL", DeviceAlert{Level: "CRITICAL"}, true},
		{"warning", DeviceAlert{Level: "warning"}, false},
		{"info", DeviceAlert{Level: "info"}, false},
		{"error", DeviceAlert{Level: "error"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.alert.IsCritical()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeviceMetricValidation(t *testing.T) {
	tests := []struct {
		name     string
		metrics  DeviceMetric
		wantErrs int
	}{
		{
			name: "valid metrics",
			metrics: DeviceMetric{
				DeviceID:    sampleUUID,
				CPUUsage:    50.5,
				CPUTemp:     65.0,
				MemoryTotal: 8000000000,
				MemoryUsed:  4000000000,
				DiskTotal:   1000000000000,
				DiskUsed:    500000000000,
			},
			wantErrs: 0,
		},
		{
			name: "invalid CPU usage",
			metrics: DeviceMetric{
				DeviceID: sampleUUID,
				CPUUsage: 150.0,
			},
			wantErrs: 1,
		},
		{
			name: "invalid CPU temperature",
			metrics: DeviceMetric{
				DeviceID: sampleUUID,
				CPUTemp:  200.0,
			},
			wantErrs: 1,
		},
		{
			name: "memory used exceeds total",
			metrics: DeviceMetric{
				DeviceID:    sampleUUID,
				MemoryTotal: 4000000000,
				MemoryUsed:  8000000000,
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.metrics.ValidateDeviceMetric()
			assert.Equal(t, tt.wantErrs, len(errors))
		})
	}
}

func TestDeviceMetricGetMemoryUsagePercent(t *testing.T) {
	tests := []struct {
		name     string
		metrics  DeviceMetric
		expected float64
	}{
		{
			name: "50% usage",
			metrics: DeviceMetric{
				MemoryTotal: 8000000000,
				MemoryUsed:  4000000000,
			},
			expected: 50.0,
		},
		{
			name: "zero total",
			metrics: DeviceMetric{
				MemoryTotal: 0,
				MemoryUsed:  4000000000,
			},
			expected: 0.0,
		},
		{
			name: "100% usage",
			metrics: DeviceMetric{
				MemoryTotal: 8000000000,
				MemoryUsed:  8000000000,
			},
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metrics.GetMemoryUsagePercent()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeviceMetricGetDiskUsagePercent(t *testing.T) {
	tests := []struct {
		name     string
		metrics  DeviceMetric
		expected float64
	}{
		{
			name: "25% usage",
			metrics: DeviceMetric{
				DiskTotal: 1000000000000,
				DiskUsed:  250000000000,
			},
			expected: 25.0,
		},
		{
			name: "zero total",
			metrics: DeviceMetric{
				DiskTotal: 0,
				DiskUsed:  250000000000,
			},
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.metrics.GetDiskUsagePercent()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoteCommandValidation(t *testing.T) {
	tests := []struct {
		name     string
		command  DeviceRemoteCommand
		wantErrs int
	}{
		{
			name: "valid command",
			command: DeviceRemoteCommand{
				DeviceID:    sampleUUID,
				CommandText: "ls -la",
				Status:      "pending",
			},
			wantErrs: 0,
		},
		{
			name: "empty device ID",
			command: DeviceRemoteCommand{
				DeviceID:    sampleUUID,
				CommandText: "ls -la",
			},
			wantErrs: 1,
		},
		{
			name: "empty command",
			command: DeviceRemoteCommand{
				DeviceID:    sampleUUID,
				CommandText: "",
			},
			wantErrs: 1,
		},
		{
			name: "invalid status",
			command: DeviceRemoteCommand{
				DeviceID:    sampleUUID,
				CommandText: "ls -la",
				Status:      "invalid",
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.command.ValidateRemoteCommand()
			assert.Equal(t, tt.wantErrs, len(errors))
		})
	}
}

func TestRemoteCommandIsCompleted(t *testing.T) {
	tests := []struct {
		name     string
		command  DeviceRemoteCommand
		expected bool
	}{
		{"completed", DeviceRemoteCommand{Status: "completed"}, true},
		{"failed", DeviceRemoteCommand{Status: "failed"}, true},
		{"COMPLETED", DeviceRemoteCommand{Status: "COMPLETED"}, true},
		{"pending", DeviceRemoteCommand{Status: "pending"}, false},
		{"running", DeviceRemoteCommand{Status: "running"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.command.IsCompleted()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRemoteCommandIsSuccessful(t *testing.T) {
	tests := []struct {
		name     string
		command  DeviceRemoteCommand
		expected bool
	}{
		{
			name: "successful",
			command: DeviceRemoteCommand{
				Status:   "completed",
				ExitCode: 0,
			},
			expected: true,
		},
		{
			name: "completed with error",
			command: DeviceRemoteCommand{
				Status:   "completed",
				ExitCode: 1,
			},
			expected: false,
		},
		{
			name: "failed",
			command: DeviceRemoteCommand{
				Status:   "failed",
				ExitCode: 0,
			},
			expected: false,
		},
		{
			name: "still running",
			command: DeviceRemoteCommand{
				Status:   "running",
				ExitCode: 0,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.command.IsSuccessful()
			assert.Equal(t, tt.expected, result)
		})
	}
}
