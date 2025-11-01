CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    device_id TEXT UNIQUE NOT NULL,
    name TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW()
);

CREATE TABLE screenshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    file_path TEXT NOT NULL,
    captured_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE keystrokes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    keylog TEXT NOT NULL,
    logged_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE app_activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    app_name TEXT NOT NULL,
    action TEXT CHECK (action IN ('open', 'close', 'background')),
    activity_time TIMESTAMP DEFAULT NOW()
);

CREATE TABLE device_metrics (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    cpu_usage DOUBLE PRECISION NOT NULL,
    memory_total BIGINT NOT NULL,
    memory_used BIGINT NOT NULL,
    disk_total BIGINT NOT NULL,
    disk_used BIGINT NOT NULL,
    network_rx_bytes BIGINT NOT NULL,
    network_tx_bytes BIGINT NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW()
);

CREATE TABLE device_processes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    pid INTEGER NOT NULL,
    name TEXT NOT NULL,
    cpu_percent DOUBLE PRECISION NOT NULL,
    memory_percent DOUBLE PRECISION NOT NULL,
    status TEXT NOT NULL,
    timestamp TIMESTAMP DEFAULT NOW()
);

CREATE TABLE device_status (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    online BOOLEAN NOT NULL,
    last_heartbeat TIMESTAMP DEFAULT NOW(),
    ip_address TEXT,
    hostname TEXT,
    os_info TEXT,
    last_updated TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_locations_user_id ON locations(user_id);
CREATE INDEX idx_screenshots_user_id ON screenshots(user_id);
CREATE INDEX idx_keystrokes_user_id ON keystrokes(user_id);
CREATE INDEX idx_app_activity_user_id ON app_activity(user_id);
CREATE INDEX idx_device_metrics_user_id ON device_metrics(user_id);
CREATE INDEX idx_device_processes_user_id ON device_processes(user_id);
CREATE INDEX idx_device_status_user_id ON device_status(user_id);
