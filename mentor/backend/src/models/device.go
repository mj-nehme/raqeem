package models

import (
	"time"

	"github.com/google/uuid"
)

// Device represents a monitored device.
type Device struct {
	DeviceID       uuid.UUID `json:"deviceid" gorm:"column:deviceid;type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceName     string    `json:"device_name"`
	DeviceType     string    `json:"device_type"`
	OS             string    `json:"os"`
	LastSeen       time.Time `json:"last_seen" gorm:"default:now()"`
	IsOnline       bool      `json:"is_online"`
	DeviceLocation string    `json:"device_location"`
	IPAddress      string    `json:"ip_address"`
	MacAddress     string    `json:"mac_address"`
	CurrentUser    string    `json:"current_user"`

	// Relationships
	Metrics     []DeviceMetric        `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE;"`
	Processes   []DeviceProcess       `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE;"`
	Activities  []DeviceActivity      `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE;"`
	Alerts      []DeviceAlert         `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE;"`
	Commands    []DeviceRemoteCommand `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE;"`
	Screenshots []DeviceScreenshot    `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE;"`
	Users       []User                `gorm:"foreignKey:DeviceID;constraint:OnDelete:CASCADE;"`
}

// DeviceMetric stores system metrics.
type DeviceMetric struct {
	MetricID    uuid.UUID `json:"metricid" gorm:"column:metricid;type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID    uuid.UUID `json:"deviceid" gorm:"column:deviceid"`
	Timestamp   time.Time `json:"timestamp" gorm:"default:now()"`
	CPUUsage    float64   `json:"cpu_usage"`
	CPUTemp     float64   `json:"cpu_temp"`
	MemoryTotal uint64    `json:"memory_total"`
	MemoryUsed  uint64    `json:"memory_used"`
	SwapUsed    uint64    `json:"swap_used"`
	DiskTotal   uint64    `json:"disk_total"`
	DiskUsed    uint64    `json:"disk_used"`
	NetBytesIn  uint64    `json:"net_bytes_in"`
	NetBytesOut uint64    `json:"net_bytes_out"`
}

// DeviceProcess represents a running process.
type DeviceProcess struct {
	ProcessID   uuid.UUID `json:"processid" gorm:"column:processid;type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID    uuid.UUID `json:"deviceid" gorm:"column:deviceid"`
	Timestamp   time.Time `json:"timestamp" gorm:"default:now()"`
	PID         int       `json:"pid" gorm:"column:pid"`
	ProcessName string    `json:"process_name"`
	CPU         float64   `json:"cpu"`
	Memory      uint64    `json:"memory"`
	CommandText string    `json:"command_text"`
}

// DeviceActivity tracks user activity on the device.
type DeviceActivity struct {
	ActivityID   uuid.UUID `json:"activityid" gorm:"column:activityid;type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID     uuid.UUID `json:"deviceid" gorm:"column:deviceid"`
	Timestamp    time.Time `json:"timestamp" gorm:"default:now()"`
	ActivityType string    `json:"activity_type"`
	Description  string    `json:"description"`
	App          string    `json:"app" gorm:"column:app;type:text"`
	Duration     int       `json:"duration"`
}

// DeviceAlert represents alerts raised by monitoring.
type DeviceAlert struct {
	AlertID   uuid.UUID `json:"alertid" gorm:"column:alertid;type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID  uuid.UUID `json:"deviceid" gorm:"column:deviceid"`
	Timestamp time.Time `json:"timestamp" gorm:"default:now()"`
	Level     string    `json:"level"`
	AlertType string    `json:"alert_type"`
	Message   string    `json:"message"`
	Value     float64   `json:"value"`
	Threshold float64   `json:"threshold"`
}

// DeviceRemoteCommand represents a command sent remotely.
type DeviceRemoteCommand struct {
	CommandID   uuid.UUID `json:"commandid" gorm:"column:commandid;type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID    uuid.UUID `json:"deviceid" gorm:"column:deviceid"`
	CommandText string    `json:"command_text"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at" gorm:"default:now()"`
	CompletedAt time.Time `json:"completed_at"`
	Result      string    `json:"result"`
	ExitCode    int       `json:"exit_code"`
}

// DeviceScreenshot stores screen captures.
type DeviceScreenshot struct {
	ScreenshotID uuid.UUID `json:"screenshotid" gorm:"column:screenshot_id;type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID     uuid.UUID `json:"deviceid" gorm:"column:device_id"`
	Timestamp    time.Time `json:"timestamp" gorm:"column:screenshot_timestamp;default:now()"`
	Path         string    `json:"path" gorm:"column:screenshot_path"`
	Resolution   string    `json:"resolution" gorm:"column:screenshot_resolution"`
	Size         int64     `json:"size" gorm:"column:screenshot_size"`
}

// User represents a user linked to a device.
type User struct {
	UserID    uuid.UUID `json:"userid" gorm:"column:userid;type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID  uuid.UUID `json:"deviceid" gorm:"column:deviceid"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at" gorm:"default:now()"`
}
