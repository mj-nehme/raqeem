package models

import (
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Validation constants
const (
	// MaxDeviceNameLength is the maximum allowed device name length
	MaxDeviceNameLength = 255
	// MinCPUTemp is the minimum valid CPU temperature in celsius
	MinCPUTemp = -50
	// MaxCPUTemp is the maximum valid CPU temperature in celsius
	MaxCPUTemp = 150
	// MinCPUUsage is the minimum CPU usage percentage
	MinCPUUsage = 0
	// MaxCPUUsage is the maximum CPU usage percentage
	MaxCPUUsage = 100
	// DefaultOnlineThresholdMinutes is the default time in minutes to consider device offline
	DefaultOnlineThresholdMinutes = 5
)

// Valid device types
var validDeviceTypes = map[string]bool{
	"laptop":  true,
	"desktop": true,
	"mobile":  true,
	"tablet":  true,
	"server":  true,
	"iot":     true,
}

// Valid alert levels
var validAlertLevels = map[string]bool{
	"info":     true,
	"warning":  true,
	"error":    true,
	"critical": true,
}

// Valid alert types
var validAlertTypes = map[string]bool{
	"cpu":      true,
	"memory":   true,
	"disk":     true,
	"network":  true,
	"security": true,
}

// Valid command statuses
var validCommandStatuses = map[string]bool{
	"pending":   true,
	"running":   true,
	"completed": true,
	"failed":    true,
}

// AllowedCommands is the whitelist of commands that can be executed on devices
// This must match the whitelist in the devices backend for security
var AllowedCommands = map[string]bool{
	"get_info":        true,
	"status":          true,
	"restart":         true,
	"get_processes":   true,
	"get_logs":        true,
	"restart_service": true,
	"screenshot":      true,
}

// -------------------- DEVICE --------------------

// ValidateDevice validates device fields and returns validation errors
func (device *Device) ValidateDevice() []string {
	var errors []string

	if strings.TrimSpace(device.DeviceName) == "" {
		errors = append(errors, "device name cannot be empty")
	}

	if len(device.DeviceName) > MaxDeviceNameLength {
		errors = append(errors, "device name cannot exceed 255 characters")
	}

	if device.DeviceType != "" && !validDeviceTypes[strings.ToLower(device.DeviceType)] {
		errors = append(errors, "invalid device type")
	}

	return errors
}

// IsOnlineRecently checks if device was seen within specified minutes
func (device *Device) IsOnlineRecently(minutes int) bool {
	if minutes <= 0 {
		return device.IsOnline
	}
	threshold := time.Now().Add(-time.Duration(minutes) * time.Minute)
	return device.LastSeen.After(threshold)
}

// GetFormattedType returns the device type with proper capitalization
func (device *Device) GetFormattedType() string {
	if device.DeviceType == "" {
		return "Unknown"
	}
	caser := cases.Title(language.English)
	return caser.String(strings.ToLower(device.DeviceType))
}

// -------------------- ALERTS --------------------

// ValidateAlert validates alert fields and returns validation errors
func (alert *DeviceAlert) ValidateAlert() []string {
	var errors []string

	if strings.TrimSpace(alert.AlertID.String()) == "" {
		errors = append(errors, "alert ID cannot be empty")
	}

	if !validAlertLevels[strings.ToLower(alert.Level)] {
		errors = append(errors, "invalid alert level (must be: info, warning, error, or critical)")
	}

	if !validAlertTypes[strings.ToLower(alert.AlertType)] {
		errors = append(errors, "invalid alert type (must be: cpu, memory, disk, network, or security)")
	}

	if strings.TrimSpace(alert.Message) == "" {
		errors = append(errors, "alert message cannot be empty")
	}

	return errors
}

// IsCritical returns true if alert level is critical
func (alert *DeviceAlert) IsCritical() bool {
	return strings.ToLower(alert.Level) == "critical"
}

// -------------------- METRICS --------------------

// ValidateDeviceMetric validates metric fields and returns validation errors
func (metric *DeviceMetric) ValidateDeviceMetric() []string {
	var errors []string

	if strings.TrimSpace(metric.MetricID.String()) == "" {
		errors = append(errors, "metric ID cannot be empty")
	}

	if metric.CPUUsage < MinCPUUsage || metric.CPUUsage > MaxCPUUsage {
		errors = append(errors, "CPU usage must be between 0 and 100")
	}

	if metric.CPUTemp < MinCPUTemp || metric.CPUTemp > MaxCPUTemp {
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

// GetMemoryUsagePercent calculates memory usage as a percentage
func (metric *DeviceMetric) GetMemoryUsagePercent() float64 {
	if metric.MemoryTotal == 0 {
		return 0
	}
	return float64(metric.MemoryUsed) / float64(metric.MemoryTotal) * 100
}

// GetDiskUsagePercent calculates disk usage as a percentage
func (metric *DeviceMetric) GetDiskUsagePercent() float64 {
	if metric.DiskTotal == 0 {
		return 0
	}
	return float64(metric.DiskUsed) / float64(metric.DiskTotal) * 100
}

// -------------------- REMOTE COMMANDS --------------------

// ValidateRemoteCommand validates command fields and returns validation errors
func (command *DeviceRemoteCommand) ValidateRemoteCommand() []string {
	var errors []string

	if strings.TrimSpace(command.DeviceID.String()) == "" {
		errors = append(errors, "device ID cannot be empty")
	}

	if strings.TrimSpace(command.CommandText) == "" {
		errors = append(errors, "command cannot be empty")
	}

	// Validate command against whitelist
	commandBase := strings.ToLower(strings.TrimSpace(strings.Split(command.CommandText, " ")[0]))
	if commandBase != "" && !AllowedCommands[commandBase] {
		errors = append(errors, "command not allowed. Allowed commands: get_info, status, restart, get_processes, get_logs, restart_service, screenshot")
	}

	if command.Status != "" && !validCommandStatuses[strings.ToLower(command.Status)] {
		errors = append(errors, "invalid command status (must be: pending, running, completed, or failed)")
	}

	return errors
}

// IsCompleted returns true if command has finished execution (success or failure)
func (command *DeviceRemoteCommand) IsCompleted() bool {
	status := strings.ToLower(command.Status)
	return status == "completed" || status == "failed"
}

// IsSuccessful returns true if command completed successfully
func (command *DeviceRemoteCommand) IsSuccessful() bool {
	return strings.ToLower(command.Status) == "completed" && command.ExitCode == 0
}
