package models

import (
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// ValidateDevice validates device fields
func (d *Device) ValidateDevice() []string {
	var errors []string

	if strings.TrimSpace(d.ID) == "" {
		errors = append(errors, "device ID cannot be empty")
	}

	if strings.TrimSpace(d.Name) == "" {
		errors = append(errors, "device name cannot be empty")
	}

	if len(d.Name) > 255 {
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

	if d.Type != "" && !validTypes[strings.ToLower(d.Type)] {
		errors = append(errors, "invalid device type")
	}

	return errors
}

// IsOnlineRecently checks if device was online in the last N minutes
func (d *Device) IsOnlineRecently(minutes int) bool {
	if minutes <= 0 {
		return d.IsOnline
	}
	threshold := time.Now().Add(-time.Duration(minutes) * time.Minute)
	return d.LastSeen.After(threshold)
}

// GetFormattedType returns the device type in proper case
func (d *Device) GetFormattedType() string {
	if d.Type == "" {
		return "Unknown"
	}
	caser := cases.Title(language.English)
	return caser.String(strings.ToLower(d.Type))
}

// ValidateAlert validates alert fields
func (a *Alert) ValidateAlert() []string {
	var errors []string

	if strings.TrimSpace(a.DeviceID) == "" {
		errors = append(errors, "device ID cannot be empty")
	}

	validLevels := map[string]bool{
		"info":     true,
		"warning":  true,
		"error":    true,
		"critical": true,
	}

	if !validLevels[strings.ToLower(a.Level)] {
		errors = append(errors, "invalid alert level")
	}

	validTypes := map[string]bool{
		"cpu":      true,
		"memory":   true,
		"disk":     true,
		"network":  true,
		"security": true,
	}

	if !validTypes[strings.ToLower(a.Type)] {
		errors = append(errors, "invalid alert type")
	}

	if strings.TrimSpace(a.Message) == "" {
		errors = append(errors, "alert message cannot be empty")
	}

	return errors
}

// IsCritical checks if alert is critical level
func (a *Alert) IsCritical() bool {
	return strings.ToLower(a.Level) == "critical"
}

// ValidateDeviceMetrics validates device metrics
func (dm *DeviceMetrics) ValidateDeviceMetrics() []string {
	var errors []string

	if strings.TrimSpace(dm.DeviceID) == "" {
		errors = append(errors, "device ID cannot be empty")
	}

	if dm.CPUUsage < 0 || dm.CPUUsage > 100 {
		errors = append(errors, "CPU usage must be between 0 and 100")
	}

	if dm.CPUTemp < -50 || dm.CPUTemp > 150 {
		errors = append(errors, "CPU temperature must be between -50 and 150 celsius")
	}

	if dm.MemoryUsed > dm.MemoryTotal && dm.MemoryTotal > 0 {
		errors = append(errors, "memory used cannot exceed memory total")
	}

	if dm.DiskUsed > dm.DiskTotal && dm.DiskTotal > 0 {
		errors = append(errors, "disk used cannot exceed disk total")
	}

	return errors
}

// GetMemoryUsagePercent returns memory usage as percentage
func (dm *DeviceMetrics) GetMemoryUsagePercent() float64 {
	if dm.MemoryTotal == 0 {
		return 0
	}
	return float64(dm.MemoryUsed) / float64(dm.MemoryTotal) * 100
}

// GetDiskUsagePercent returns disk usage as percentage
func (dm *DeviceMetrics) GetDiskUsagePercent() float64 {
	if dm.DiskTotal == 0 {
		return 0
	}
	return float64(dm.DiskUsed) / float64(dm.DiskTotal) * 100
}

// ValidateRemoteCommand validates remote command fields
func (rc *RemoteCommand) ValidateRemoteCommand() []string {
	var errors []string

	if strings.TrimSpace(rc.DeviceID) == "" {
		errors = append(errors, "device ID cannot be empty")
	}

	if strings.TrimSpace(rc.Command) == "" {
		errors = append(errors, "command cannot be empty")
	}

	validStatuses := map[string]bool{
		"pending":   true,
		"running":   true,
		"completed": true,
		"failed":    true,
	}

	if rc.Status != "" && !validStatuses[strings.ToLower(rc.Status)] {
		errors = append(errors, "invalid command status")
	}

	return errors
}

// IsCompleted checks if command is completed (either successfully or failed)
func (rc *RemoteCommand) IsCompleted() bool {
	status := strings.ToLower(rc.Status)
	return status == "completed" || status == "failed"
}

// IsSuccessful checks if command completed successfully
func (rc *RemoteCommand) IsSuccessful() bool {
	return strings.ToLower(rc.Status) == "completed" && rc.ExitCode == 0
}
