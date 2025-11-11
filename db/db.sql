-- Modern schema initialization for Raqeem device monitoring
-- Device-centric architecture using UUID primary keys with proper foreign key relationships

-- Database setup
DROP DATABASE IF EXISTS monitoring_db;
CREATE DATABASE monitoring_db;

-- Tables:
-- devices
-- device_metrics
-- device_processes
-- device_activities
-- device_alerts
-- remote_commands
-- device_screenshots

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Core device table
CREATE TABLE devices (
    device_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    devicename TEXT,
    device_type TEXT,
    os TEXT,
    last_seen TIMESTAMP DEFAULT NOW(),
    is_online BOOLEAN,
    device_location TEXT,
    ip_address TEXT,
    mac_address TEXT,
    current_user_text TEXT                  -- Safe column name for reserved word
);

-- Device metrics with UUID primary key
CREATE TABLE device_metrics (
    metrics_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL,
    metrics_timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    cpu_usage NUMERIC,
    cpu_temp NUMERIC,
    memory_total BIGINT,
    memory_used BIGINT,
    swap_used BIGINT,
    disk_total BIGINT,
    disk_used BIGINT,
    net_bytes_in BIGINT,
    net_bytes_out BIGINT
);

-- Device processes
CREATE TABLE device_processes (
    processes_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL,
    processes_timestamp TIMESTAMP DEFAULT NOW(),
    pid INTEGER NOT NULL,
    pname TEXT NOT NULL,
    cpu NUMERIC,
    memory BIGINT,
    command TEXT
);

-- Device activity logs
CREATE TABLE device_activities (
    activity_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL,
    activity_timestamp TIMESTAMP DEFAULT NOW(),
    activity_type TEXT,
    activity_description TEXT,
    activity_app TEXT,
    duration INTEGER
);

-- Device alerts
CREATE TABLE device_alerts (
    alert_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL,
    alert_timestamp TIMESTAMP DEFAULT NOW(),
    alert_level TEXT,
    alert_type TEXT,
    alert_message TEXT,
    alert_value NUMERIC,
    threshold NUMERIC
);

-- Remote commands for device management
CREATE TABLE device_remote_commands (
    command_id SERIAL PRIMARY KEY,
    device_id UUID NOT NULL,
    command_text TEXT NOT NULL,
    device_status TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    command_result TEXT,
    command_exit_code INTEGER
);

-- Device screenshots
CREATE TABLE device_screenshots (
    screenshot_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL,
    screenshot_timestamp TIMESTAMP DEFAULT NOW(),
    screenshot_path TEXT NOT NULL,
    screenshot_resolution TEXT,
    screenshot_size BIGINT
);

-- Users table (simplified, no UUID foreign keys)
CREATE TABLE users (
    user_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id UUID NOT NULL,
    name TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Performance indexes
CREATE INDEX idx_device_metrics_device_id ON device_metrics(device_id);
CREATE INDEX idx_device_metrics_timestamp ON device_metrics(metrics_timestamp);
CREATE INDEX idx_device_processes_device_id ON device_processes(device_id);
CREATE INDEX idx_device_activities_device_id ON device_activities(device_id);
CREATE INDEX idx_device_alerts_device_id ON device_alerts(device_id);
CREATE INDEX idx_device_screenshots_device_id ON device_screenshots(device_id);
CREATE INDEX idx_remote_commands_device_id ON remote_commands(device_id);
CREATE INDEX idx_screenshots_user_id ON screenshots(user_id);

-- Foreign key constraints for referential integrity
ALTER TABLE device_metrics ADD CONSTRAINT fk_device_metrics_device 
    FOREIGN KEY (device_id) REFERENCES devices(device_id) ON DELETE CASCADE;

ALTER TABLE device_processes ADD CONSTRAINT fk_device_processes_device 
    FOREIGN KEY (device_id) REFERENCES devices(device_id) ON DELETE CASCADE;

ALTER TABLE device_activities ADD CONSTRAINT fk_device_activities_device 
    FOREIGN KEY (device_id) REFERENCES devices(device_id) ON DELETE CASCADE;

ALTER TABLE device_alerts ADD CONSTRAINT fk_device_alerts_device 
    FOREIGN KEY (device_id) REFERENCES devices(device_id) ON DELETE CASCADE;

ALTER TABLE remote_commands ADD CONSTRAINT fk_remote_commands_device 
    FOREIGN KEY (device_id) REFERENCES devices(device_id) ON DELETE CASCADE;

ALTER TABLE device_screenshots ADD CONSTRAINT fk_device_screenshots_device 
    FOREIGN KEY (device_id) REFERENCES devices(device_id) ON DELETE CASCADE;

ALTER TABLE users ADD CONSTRAINT fk_users_device 
    FOREIGN KEY (device_id) REFERENCES devices(device_id) ON DELETE CASCADE;

ALTER TABLE screenshots ADD CONSTRAINT fk_screenshots_user 
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE;