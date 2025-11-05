-- Migration: Fix screenshots table schema to match SQLAlchemy model
-- 1. Change id from bigint to UUID
-- 2. Add user_id column 
-- 3. Rename path to image_path
-- 4. Add created_at column

-- Check existing data
SELECT COUNT(*) as record_count FROM screenshots;

-- Drop and recreate the table with correct schema
DROP TABLE IF EXISTS screenshots CASCADE;

-- Create with correct schema matching the Screenshot model
CREATE TABLE screenshots (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    image_path TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Create index for performance
CREATE INDEX idx_screenshots_user_id ON screenshots(user_id);
CREATE INDEX idx_screenshots_created_at ON screenshots(created_at);

-- Verify the new schema
\d screenshots;