CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    owner_id TEXT NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (owner_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create index on owner_id for faster queries
CREATE INDEX IF NOT EXISTS idx_projects_owner_id ON projects(owner_id);

-- Create index on deleted_at for soft delete queries
CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at);

-- Create unique constraint on name per owner (excluding soft deleted)
CREATE UNIQUE INDEX IF NOT EXISTS idx_projects_owner_name_unique 
ON projects(owner_id, name) 
WHERE deleted_at IS NULL;

-- Trigger to automatically update updated_at timestamp for projects
CREATE TRIGGER IF NOT EXISTS update_projects_timestamp 
AFTER UPDATE ON projects 
FOR EACH ROW
BEGIN
    UPDATE projects SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;  