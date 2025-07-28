-- Migration: Create excluded folders table
-- Date: 2025-07-28

CREATE TABLE IF NOT EXISTS excluded_folders (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    folder_path TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_excluded_folders_project_id ON excluded_folders(project_id);
CREATE INDEX IF NOT EXISTS idx_excluded_folders_deleted_at ON excluded_folders(deleted_at);
CREATE INDEX IF NOT EXISTS idx_excluded_folders_created_at ON excluded_folders(created_at);

-- Create unique constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_excluded_folders_unique 
    ON excluded_folders(project_id, folder_path) 
    WHERE deleted_at IS NULL;

-- Create trigger for updated_at
CREATE TRIGGER IF NOT EXISTS update_excluded_folders_updated_at
    AFTER UPDATE ON excluded_folders
    FOR EACH ROW
BEGIN
    UPDATE excluded_folders SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END; 