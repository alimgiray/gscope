-- Migration: Create project update settings table
-- Date: 2025-07-28

CREATE TABLE IF NOT EXISTS project_update_settings (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    is_enabled BOOLEAN DEFAULT FALSE,
    hour INTEGER NOT NULL CHECK (hour >= 0 AND hour <= 23),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_project_update_settings_project_id ON project_update_settings(project_id);
CREATE INDEX IF NOT EXISTS idx_project_update_settings_deleted_at ON project_update_settings(deleted_at);
CREATE INDEX IF NOT EXISTS idx_project_update_settings_created_at ON project_update_settings(created_at);

-- Create unique constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_project_update_settings_unique
    ON project_update_settings(project_id)
    WHERE deleted_at IS NULL;

-- Create trigger for updated_at (only if it doesn't exist)
CREATE TRIGGER IF NOT EXISTS update_project_update_settings_updated_at
    AFTER UPDATE ON project_update_settings
    FOR EACH ROW
BEGIN
    UPDATE project_update_settings SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 