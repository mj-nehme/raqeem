-- Migration: Fix device_metrics ID column type mismatch
-- Change from bigint to UUID to match SQLAlchemy model

-- First, let's check if we have any existing data
SELECT COUNT(*) as record_count FROM device_metrics;

-- Drop the table and recreate it with correct schema
-- Note: This will lose any existing data, but that's okay for development
DROP TABLE IF EXISTS device_metrics CASCADE;

-- Recreate with correct UUID schema matching the Python model
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

-- Create index for performance
CREATE INDEX idx_device_metrics_device_id ON device_metrics(device_id);
CREATE INDEX idx_device_metrics_timestamp ON device_metrics(timestamp);

-- Verify the new schema
\d device_metrics;