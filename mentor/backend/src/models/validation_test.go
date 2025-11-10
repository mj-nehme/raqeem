package models

import (
	"testing"
	"time"

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
				ID:   "device1",
				Name: "Test Device",
				Type: "laptop",
			},
			wantErrs: 0,
		},
		{
			name: "empty ID",
			device: Device{
				ID:   "",
				Name: "Test Device",
				Type: "laptop",
			},
			wantErrs: 1,
		},
		{
			name: "empty name",
			device: Device{
				ID:   "device1",
				Name: "",
				Type: "laptop",
			},
			wantErrs: 1,
		},
		{
			name: "invalid type",
			device: Device{
				ID:   "device1",
				Name: "Test Device",
				Type: "invalid",
			},
			wantErrs: 1,
		},
		{
			name: "name too long",
			device: Device{
				ID:   "device1",
				Name: string(make([]byte, 256)), // 256 characters
				Type: "laptop",
			},
			wantErrs: 1,
		},
		{
			name: "multiple errors",
			device: Device{
				ID:   "",
				Name: "",
				Type: "invalid",
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
		{"laptop", Device{Type: "laptop"}, "Laptop"},
		{"DESKTOP", Device{Type: "DESKTOP"}, "Desktop"},
		{"MoBiLe", Device{Type: "MoBiLe"}, "Mobile"},
		{"empty", Device{Type: ""}, "Unknown"},
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
		alert    Alert
		wantErrs int
	}{
		{
			name: "valid alert",
			alert: Alert{
				DeviceID: "device1",
				Level:    "warning",
				Type:     "cpu",
				Message:  "High CPU usage",
			},
			wantErrs: 0,
		},
		{
			name: "empty device ID",
			alert: Alert{
				DeviceID: "",
				Level:    "warning",
				Type:     "cpu",
				Message:  "High CPU usage",
			},
			wantErrs: 1,
		},
		{
			name: "invalid level",
			alert: Alert{
				DeviceID: "device1",
				Level:    "invalid",
				Type:     "cpu",
				Message:  "High CPU usage",
			},
			wantErrs: 1,
		},
		{
			name: "invalid type",
			alert: Alert{
				DeviceID: "device1",
				Level:    "warning",
				Type:     "invalid",
				Message:  "High CPU usage",
			},
			wantErrs: 1,
		},
		{
			name: "empty message",
			alert: Alert{
				DeviceID: "device1",
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
		alert    Alert
		expected bool
	}{
		{"critical", Alert{Level: "critical"}, true},
		{"CRITICAL", Alert{Level: "CRITICAL"}, true},
		{"warning", Alert{Level: "warning"}, false},
		{"info", Alert{Level: "info"}, false},
		{"error", Alert{Level: "error"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.alert.IsCritical()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDeviceMetricsValidation(t *testing.T) {
	tests := []struct {
		name     string
		metrics  DeviceMetrics
		wantErrs int
	}{
		{
			name: "valid metrics",
			metrics: DeviceMetrics{
				DeviceID:    "device1",
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
			metrics: DeviceMetrics{
				DeviceID: "device1",
				CPUUsage: 150.0,
			},
			wantErrs: 1,
		},
		{
			name: "invalid CPU temperature",
			metrics: DeviceMetrics{
				DeviceID: "device1",
				CPUTemp:  200.0,
			},
			wantErrs: 1,
		},
		{
			name: "memory used exceeds total",
			metrics: DeviceMetrics{
				DeviceID:    "device1",
				MemoryTotal: 4000000000,
				MemoryUsed:  8000000000,
			},
			wantErrs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := tt.metrics.ValidateDeviceMetrics()
			assert.Equal(t, tt.wantErrs, len(errors))
		})
	}
}

func TestDeviceMetricsGetMemoryUsagePercent(t *testing.T) {
	tests := []struct {
		name     string
		metrics  DeviceMetrics
		expected float64
	}{
		{
			name: "50% usage",
			metrics: DeviceMetrics{
				MemoryTotal: 8000000000,
				MemoryUsed:  4000000000,
			},
			expected: 50.0,
		},
		{
			name: "zero total",
			metrics: DeviceMetrics{
				MemoryTotal: 0,
				MemoryUsed:  4000000000,
			},
			expected: 0.0,
		},
		{
			name: "100% usage",
			metrics: DeviceMetrics{
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

func TestDeviceMetricsGetDiskUsagePercent(t *testing.T) {
	tests := []struct {
		name     string
		metrics  DeviceMetrics
		expected float64
	}{
		{
			name: "25% usage",
			metrics: DeviceMetrics{
				DiskTotal: 1000000000000,
				DiskUsed:  250000000000,
			},
			expected: 25.0,
		},
		{
			name: "zero total",
			metrics: DeviceMetrics{
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
		command  RemoteCommand
		wantErrs int
	}{
		{
			name: "valid command",
			command: RemoteCommand{
				DeviceID: "device1",
				Command:  "ls -la",
				Status:   "pending",
			},
			wantErrs: 0,
		},
		{
			name: "empty device ID",
			command: RemoteCommand{
				DeviceID: "",
				Command:  "ls -la",
			},
			wantErrs: 1,
		},
		{
			name: "empty command",
			command: RemoteCommand{
				DeviceID: "device1",
				Command:  "",
			},
			wantErrs: 1,
		},
		{
			name: "invalid status",
			command: RemoteCommand{
				DeviceID: "device1",
				Command:  "ls -la",
				Status:   "invalid",
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
		command  RemoteCommand
		expected bool
	}{
		{"completed", RemoteCommand{Status: "completed"}, true},
		{"failed", RemoteCommand{Status: "failed"}, true},
		{"COMPLETED", RemoteCommand{Status: "COMPLETED"}, true},
		{"pending", RemoteCommand{Status: "pending"}, false},
		{"running", RemoteCommand{Status: "running"}, false},
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
		command  RemoteCommand
		expected bool
	}{
		{
			name: "successful",
			command: RemoteCommand{
				Status:   "completed",
				ExitCode: 0,
			},
			expected: true,
		},
		{
			name: "completed with error",
			command: RemoteCommand{
				Status:   "completed",
				ExitCode: 1,
			},
			expected: false,
		},
		{
			name: "failed",
			command: RemoteCommand{
				Status:   "failed",
				ExitCode: 0,
			},
			expected: false,
		},
		{
			name: "still running",
			command: RemoteCommand{
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