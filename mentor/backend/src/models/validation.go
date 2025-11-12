package models

import (
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// -------------------- DEVICE --------------------

func (device *Device) ValidateDevice() []string {
	var errors []string

	if strings.TrimSpace(device.DeviceName) == "" {
		errors = append(errors, "device name cannot be empty")
	}

	if len(device.DeviceName) > 255 {
		errors = append(errors, "device name cannot exceed 255 characters")
	}

	validTypes := map[string]bool{
		"laptop":  true,
		"desktop": true,
		"mobile":  true,
		"tablet":  true,
		"server":  true,
		"iot":     true,
	}

	if device.DeviceType != "" && !validTypes[strings.ToLower(device.DeviceType)] {
		errors = append(errors, "invalid device type")
	}

	return errors
}

func (device *Device) IsOnlineRecently(minutes int) bool {
	if minutes <= 0 {
		return device.IsOnline
	}
	threshold := time.Now().Add(-time.Duration(minutes) * time.Minute)
	return device.LastSeen.After(threshold)
}

func (device *Device) GetFormattedType() string {
	if device.DeviceType == "" {
		return "Unknown"
	}
	caser := cases.Title(language.English)
	return caser.String(strings.ToLower(device.DeviceType))
}

// -------------------- ALERTS --------------------

func (alert *DeviceAlert) ValidateAlert() []string {
	var errors []string

	if strings.TrimSpace(alert.AlertID.String()) == "" {
		errors = append(errors, "device ID cannot be empty")
	}

	validLevels := map[string]bool{
		"info":     true,
		"warning":  true,
		"error":    true,
		"critical": true,
	}

	if !validLevels[strings.ToLower(alert.Level)] {
		errors = append(errors, "invalid alert level")
	}

	validTypes := map[string]bool{
		"cpu":      true,
		"memory":   true,
		"disk":     true,
		"network":  true,
		"security": true,
	}

	if !validTypes[strings.ToLower(alert.Type)] {
		errors = append(errors, "invalid alert type")
	}

	if strings.TrimSpace(alert.Message) == "" {
		errors = append(errors, "alert message cannot be empty")
	}

	return errors
}

func (alert *DeviceAlert) IsCritical() bool {
	return strings.ToLower(alert.Level) == "critical"
}

// -------------------- METRICS --------------------

func (metric *DeviceMetric) ValidateDeviceMetric() []string {
	var errors []string

	if strings.TrimSpace(metric.MetricID.String()) == "" {
		errors = append(errors, "device ID cannot be empty")
	}

	if metric.CPUUsage < 0 || metric.CPUUsage > 100 {
		errors = append(errors, "CPU usage must be between 0 and 100")
	}

	if metric.CPUTemp < -50 || metric.CPUTemp > 150 {
		errors = append(errors, "CPU temperature must be between -50 and 150 celsius")
	}

	if metric.MemoryUsed > metric.MemoryTotal && metric.MemoryTotal > 0 {
		errors = append(errors, "memory used cannot exceed memory total")
	}

	if metric.DiskUsed > metric.DiskTotal && metric.DiskTotal > 0 {
		errors = append(errors, "disk used cannot exceed disk total")
	}

	return errors
}

func (metric *DeviceMetric) GetMemoryUsagePercent() float64 {
	if metric.MemoryTotal == 0 {
		return 0
	}
	return float64(metric.MemoryUsed) / float64(metric.MemoryTotal) * 100
}

func (metric *DeviceMetric) GetDiskUsagePercent() float64 {
	if metric.DiskTotal == 0 {
		return 0
	}
	return float64(metric.DiskUsed) / float64(metric.DiskTotal) * 100
}

// -------------------- REMOTE COMMANDS --------------------

func (command *DeviceRemoteCommand) ValidateRemoteCommand() []string {
	var errors []string

	if strings.TrimSpace(command.DeviceID.String()) == "" {
		errors = append(errors, "device ID cannot be empty")
	}

	if strings.TrimSpace(command.CommandText) == "" {
		errors = append(errors, "command cannot be empty")
	}

	validStatuses := map[string]bool{
		"pending":   true,
		"running":   true,
		"completed": true,
		"failed":    true,
	}

	if command.Status != "" && !validStatuses[strings.ToLower(command.Status)] {
		errors = append(errors, "invalid command status")
	}

	return errors
}

func (command *DeviceRemoteCommand) IsCompleted() bool {
	status := strings.ToLower(command.Status)
	return status == "completed" || status == "failed"
}

func (command *DeviceRemoteCommand) IsSuccessful() bool {
	return strings.ToLower(command.Status) == "completed" && command.ExitCode == 0
}
