package models

import (
	"time"
)

// Device represents a monitored device
type Device struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name"`
	Type        string    `json:"type"` // laptop, desktop, mobile, etc.
	OS          string    `json:"os"`
	LastSeen    time.Time `json:"last_seen"`
	IsOnline    bool      `json:"is_online"`
	Location    string    `json:"location"`
	IPAddress   string    `json:"ip_address"`
	MacAddress  string    `json:"mac_address"`
	CurrentUser string    `json:"current_user"`
}

// DeviceMetrics stores the current system metrics
type DeviceMetrics struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`

	// CPU metrics
	CPUUsage float64 `json:"cpu_usage"` // percentage
	CPUTemp  float64 `json:"cpu_temp"`  // celsius

	// Memory metrics
	MemoryTotal uint64 `json:"memory_total"` // bytes
	MemoryUsed  uint64 `json:"memory_used"`  // bytes
	SwapUsed    uint64 `json:"swap_used"`    // bytes

	// Disk metrics
	DiskTotal uint64 `json:"disk_total"` // bytes
	DiskUsed  uint64 `json:"disk_used"`  // bytes

	// Network metrics
	NetBytesIn  uint64 `json:"net_bytes_in"`  // bytes/sec
	NetBytesOut uint64 `json:"net_bytes_out"` // bytes/sec
}

// Process represents a running process on the device
type Process struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
	PID       int       `json:"pid"`
	Name      string    `json:"name"`
	CPU       float64   `json:"cpu"`    // percentage
	Memory    uint64    `json:"memory"` // bytes
	Command   string    `json:"command"`
}

// ActivityLog tracks user activity
type ActivityLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	DeviceID    string    `json:"device_id"`
	Timestamp   time.Time `json:"timestamp"`
	Type        string    `json:"type"` // app_launch, file_access, browser, etc.
	Description string    `json:"description"`
	App         string    `json:"app"`
	Duration    int       `json:"duration"` // seconds
}

// RemoteCommand represents a command to be executed on the device
type RemoteCommand struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	DeviceID    string    `json:"device_id"`
	Command     string    `json:"command"`
	Status      string    `json:"status"` // pending, running, completed, failed
	CreatedAt   time.Time `json:"created_at"`
	CompletedAt time.Time `json:"completed_at"`
	Result      string    `json:"result"`
	ExitCode    int       `json:"exit_code"`
}

// Screenshot stores device screen captures
type Screenshot struct {
	ID         uint      `json:"id" gorm:"primaryKey"`
	DeviceID   string    `json:"device_id"`
	Timestamp  time.Time `json:"timestamp"`
	Path       string    `json:"path"` // path in MinIO
	Resolution string    `json:"resolution"`
	Size       int64     `json:"size"` // bytes
}

// Alert represents device monitoring alerts
type Alert struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	DeviceID  string    `json:"device_id"`
	Timestamp time.Time `json:"timestamp"`
	Level     string    `json:"level"` // info, warning, error, critical
	Type      string    `json:"type"`  // cpu, memory, disk, network, security
	Message   string    `json:"message"`
	Value     float64   `json:"value"`     // measured value that triggered alert
	Threshold float64   `json:"threshold"` // threshold that was exceeded
}
