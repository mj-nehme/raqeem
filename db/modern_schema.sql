-- Modern schema initialization for Raqeem device monitoring
-- This replaces the old monitor.sql with schema matching current SQLAlchemy models
-- Device-centric architecture using device_id as text strings

-- Enable UUID generation
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Core device table
CREATE TABLE devices (
    id TEXT PRIMARY KEY,                    -- Allow arbitrary device IDs
    name TEXT,
    type TEXT,
    os TEXT,
    last_seen TIMESTAMP DEFAULT NOW(),
    is_online BOOLEAN,
    location TEXT,
    ip_address TEXT,
    mac_address TEXT,
    current_user_text TEXT                  -- Safe column name for reserved word
);

-- Device metrics with UUID primary key
CREATE TABLE device_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id TEXT NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
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
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW(),
    pid INTEGER NOT NULL,
    name TEXT NOT NULL,
    cpu NUMERIC,
    memory BIGINT,
    command TEXT
);

-- Device activity logs
CREATE TABLE device_activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW(),
    type TEXT,
    description TEXT,
    app TEXT,
    duration INTEGER
);

-- Device alerts
CREATE TABLE device_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW(),
    level TEXT,
    type TEXT,
    message TEXT,
    value NUMERIC,
    threshold NUMERIC
);

-- Remote commands for device management
CREATE TABLE remote_commands (
    id SERIAL PRIMARY KEY,
    device_id TEXT NOT NULL,
    command TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    result TEXT,
    exit_code INTEGER
);

-- Device screenshots
CREATE TABLE device_screenshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW(),
    path TEXT NOT NULL,
    resolution TEXT,
    size BIGINT
);

-- Screenshots table for user-based screenshots (separate from device screenshots)
CREATE TABLE screenshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    image_path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Users table (simplified, no UUID foreign keys)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id TEXT NOT NULL,
    name TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Locations table
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,                 -- Changed from UUID to TEXT
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW()
);

-- Keystrokes table
CREATE TABLE keystrokes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,                 -- Changed from UUID to TEXT
    keylog TEXT NOT NULL,
    logged_at TIMESTAMP DEFAULT NOW()
);

-- App activity table
CREATE TABLE app_activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,                 -- Changed from UUID to TEXT
    app_name TEXT NOT NULL,
    action TEXT CHECK (action IN ('open', 'close', 'background')),
    activity_time TIMESTAMP DEFAULT NOW()
);

-- Performance indexes
CREATE INDEX idx_device_metrics_device_id ON device_metrics(device_id);
CREATE INDEX idx_device_metrics_timestamp ON device_metrics(timestamp);
CREATE INDEX idx_device_processes_device_id ON device_processes(device_id);
CREATE INDEX idx_device_activity_device_id ON device_activity(device_id);
CREATE INDEX idx_device_alerts_device_id ON device_alerts(device_id);
CREATE INDEX idx_device_screenshots_device_id ON device_screenshots(device_id);
CREATE INDEX idx_remote_commands_device_id ON remote_commands(device_id);
CREATE INDEX idx_screenshots_user_id ON screenshots(user_id);
CREATE INDEX idx_locations_user_id ON locations(user_id);
CREATE INDEX idx_keystrokes_user_id ON keystrokes(user_id);
CREATE INDEX idx_app_activity_user_id ON app_activity(user_id);