-- Migration: Create working hours settings table
-- Date: 2025-07-31

CREATE TABLE IF NOT EXISTS working_hours_settings (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    start_hour INTEGER NOT NULL DEFAULT 9,
    end_hour INTEGER NOT NULL DEFAULT 18,
    monday BOOLEAN NOT NULL DEFAULT true,
    tuesday BOOLEAN NOT NULL DEFAULT true,
    wednesday BOOLEAN NOT NULL DEFAULT true,
    thursday BOOLEAN NOT NULL DEFAULT true,
    friday BOOLEAN NOT NULL DEFAULT true,
    saturday BOOLEAN NOT NULL DEFAULT false,
    sunday BOOLEAN NOT NULL DEFAULT false,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create unique index to ensure one settings per project
CREATE UNIQUE INDEX IF NOT EXISTS idx_working_hours_settings_project_id ON working_hours_settings(project_id);

-- Create trigger for updated_at (only if it doesn't exist)
CREATE TRIGGER IF NOT EXISTS update_working_hours_settings_updated_at
    AFTER UPDATE ON working_hours_settings
    FOR EACH ROW
BEGIN
    UPDATE working_hours_settings SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 
